package types

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"net/netip"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/internal/pathutil"
	"github.com/aatuh/validate/v3/translator"
)

// RuleCompiler compiles a single Rule into a validate function.
// Implementations may precompute heavy state (e.g., compiled regex).
type RuleCompiler func(c *Compiler, rule Rule) (func(any) error, error)

// globalRegistry holds globally registered custom rule compilers.
// NewCompiler copies these into the per-compiler registry.
var (
	globalRegistry   = map[Kind]RuleCompiler{}
	globalRegistryMu sync.RWMutex
)

// RegisterRule registers a global custom Rule compiler. Call this at init.
func RegisterRule(kind Kind, rc RuleCompiler) {
	globalRegistryMu.Lock()
	defer globalRegistryMu.Unlock()
	globalRegistry[kind] = rc
}

// Compiler compiles rules into validator functions.
type Compiler struct {
	translator    translator.Translator
	custom        map[Kind]RuleCompiler
	contextCustom map[Kind]ContextRuleCompiler
	types         *TypeRegistry
}

// NewCompiler creates a new compiler with the given translator.
func NewCompiler(t translator.Translator) *Compiler {
	// Copy global registry so compilers can be customized per instance
	globalRegistryMu.RLock()
	defer globalRegistryMu.RUnlock()
	copied := make(map[Kind]RuleCompiler, len(globalRegistry))
	for k, v := range globalRegistry {
		copied[k] = v
	}
	return &Compiler{translator: t, custom: copied, contextCustom: map[Kind]ContextRuleCompiler{}}
}

// translateMessage returns a translated message if translator is available, otherwise returns the default message.
func (c *Compiler) translateMessage(code string, defaultMsg string, params []any) string {
	if c.translator != nil {
		if translated := c.translator.T(code, params...); translated != "" {
			return translated
		}
	}
	// If no translator or translation failed, return the default message
	return defaultMsg
}

// T returns a translated message for the given code, or defaultMsg if
// no translator or translation is available. This is a public proxy to
// enable external validator plugins to use the compiler's translator.
func (c *Compiler) T(code string, defaultMsg string, params []any) string {
	return c.translateMessage(code, defaultMsg, params)
}

// RegisterRule registers a custom rule compiler for this compiler instance.
func (c *Compiler) RegisterRule(kind Kind, rc RuleCompiler) {
	if c.custom == nil {
		c.custom = map[Kind]RuleCompiler{}
	}
	c.custom[kind] = rc
}

// RegisterContextRule registers a context-aware custom rule compiler for this
// compiler instance. It is used only by context-aware compile APIs.
func (c *Compiler) RegisterContextRule(kind Kind, rc ContextRuleCompiler) {
	if c.contextCustom == nil {
		c.contextCustom = map[Kind]ContextRuleCompiler{}
	}
	c.contextCustom[kind] = rc
}

// SetTypeRegistry sets per-compiler custom type validators.
func (c *Compiler) SetTypeRegistry(registry *TypeRegistry) {
	c.types = registry.Clone()
}

// RegisterType registers a custom type validator for this compiler instance.
func (c *Compiler) RegisterType(name string, factory TypeValidatorFactory) {
	if c.types == nil {
		c.types = NewTypeRegistry()
	}
	c.types.RegisterType(name, factory)
}

// Compile compiles a slice of rules into a validator function.
func (c *Compiler) Compile(rules []Rule) ValidatorFunc {
	fn, err := c.CompileE(rules)
	if err != nil {
		return func(any) error { return err }
	}
	return fn
}

// CompileE compiles a slice of rules into a validator function and returns
// compile-time custom-rule errors to callers that need to fail early.
func (c *Compiler) CompileE(rules []Rule) (ValidatorFunc, error) {
	return c.CompileWithOptsE(rules, CompileOpts{})
}

// CompileWithOpts compiles rules with options and converts compile errors into
// validation errors for compatibility with Compile.
func (c *Compiler) CompileWithOpts(rules []Rule, opts CompileOpts) ValidatorFunc {
	fn, err := c.CompileWithOptsE(rules, opts)
	if err != nil {
		return func(any) error { return err }
	}
	return fn
}

// CompileWithOptsE compiles a slice of rules with options.
func (c *Compiler) CompileWithOptsE(rules []Rule, opts CompileOpts) (ValidatorFunc, error) {
	if len(rules) == 0 {
		return func(any) error { return nil }, nil
	}

	// Pre-compile regexes and other expensive operations
	compiledRules := make([]compiledRule, 0, len(rules))
	hasOmitEmpty := false
	hasRequired := false
	for _, rule := range rules {
		if rule.Kind == KOmitempty {
			hasOmitEmpty = true
			continue
		}
		if rule.Kind == KRequired {
			hasRequired = true
			continue
		}
		compiled := c.compileRule(rule)
		if compiled.err != nil {
			return nil, compiled.err
		}
		compiledRules = append(compiledRules, compiled)
	}

	return func(v any) error {
		if hasOmitEmpty && isZeroValue(v) {
			return nil
		}
		if hasRequired && isZeroValue(v) {
			return c.validateRequired(v)
		}
		if opts.CollectAll {
			var acc verrs.Errors
			for _, rule := range compiledRules {
				if err := rule.validate(v); err != nil {
					appendCollectedErrors(&acc, err)
				}
			}
			if len(acc) > 0 {
				return acc
			}
			return nil
		}
		for _, rule := range compiledRules {
			if err := rule.validate(v); err != nil {
				return err
			}
		}
		return nil
	}, nil
}

// CompileContext compiles rules into a context-aware validator.
func (c *Compiler) CompileContext(rules []Rule) ContextValidatorFunc {
	fn, err := c.CompileContextE(rules)
	if err != nil {
		return func(context.Context, any) error { return err }
	}
	return fn
}

// CompileContextE compiles rules into a context-aware validator.
func (c *Compiler) CompileContextE(rules []Rule) (ContextValidatorFunc, error) {
	return c.CompileContextWithOptsE(rules, CompileOpts{})
}

// CompileContextWithOpts compiles rules into a context-aware validator with
// options and converts compile errors into validation errors.
func (c *Compiler) CompileContextWithOpts(rules []Rule, opts CompileOpts) ContextValidatorFunc {
	fn, err := c.CompileContextWithOptsE(rules, opts)
	if err != nil {
		return func(context.Context, any) error { return err }
	}
	return fn
}

// CompileContextWithOptsE compiles rules into a context-aware validator with
// options.
func (c *Compiler) CompileContextWithOptsE(rules []Rule, opts CompileOpts) (ContextValidatorFunc, error) {
	if len(rules) == 0 {
		return func(context.Context, any) error { return nil }, nil
	}

	compiledRules := make([]compiledContextRule, 0, len(rules))
	hasOmitEmpty := false
	hasRequired := false
	for _, rule := range rules {
		if rule.Kind == KOmitempty {
			hasOmitEmpty = true
			continue
		}
		if rule.Kind == KRequired {
			hasRequired = true
			continue
		}
		compiled := c.compileContextRule(rule)
		if compiled.err != nil {
			return nil, compiled.err
		}
		compiledRules = append(compiledRules, compiled)
	}

	return func(ctx context.Context, v any) error {
		if ctx == nil {
			ctx = context.Background()
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if hasOmitEmpty && isZeroValue(v) {
			return nil
		}
		if hasRequired && isZeroValue(v) {
			return c.validateRequired(v)
		}
		if opts.CollectAll {
			var acc verrs.Errors
			for _, rule := range compiledRules {
				if err := ctx.Err(); err != nil {
					return err
				}
				if err := rule.validate(ctx, v); err != nil {
					appendCollectedErrors(&acc, err)
				}
			}
			if len(acc) > 0 {
				return acc
			}
			return nil
		}
		for _, rule := range compiledRules {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := rule.validate(ctx, v); err != nil {
				return err
			}
		}
		return nil
	}, nil
}

func appendCollectedErrors(acc *verrs.Errors, err error) {
	var es verrs.Errors
	if errors.As(err, &es) {
		*acc = append(*acc, es...)
		return
	}
	*acc = append(*acc, verrs.FieldError{Code: verrs.CodeUnknown, Msg: err.Error()})
}

// isZeroValue reports whether v is the zero value for its dynamic type.
func isZeroValue(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	// Treat nil interface/pointer/map/slice as empty
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	case reflect.Slice:
		if rv.IsNil() {
			return true
		}
		return rv.Len() == 0
	case reflect.Map:
		if rv.IsNil() {
			return true
		}
		return rv.Len() == 0
	case reflect.String:
		return rv.Len() == 0
	}
	// For other kinds, compare to zero
	z := reflect.Zero(rv.Type())
	return reflect.DeepEqual(rv.Interface(), z.Interface())
}

// CompileField compiles rules for struct field validation.
func (c *Compiler) CompileField(rules []Rule) FieldValidator {
	validator := c.Compile(rules)
	return func(field any) error {
		return validator(field)
	}
}

type compiledRule struct {
	validate func(any) error
	err      error
}

type compiledContextRule struct {
	validate ContextValidatorFunc
	err      error
}

func (c *Compiler) compileContextRule(rule Rule) compiledContextRule {
	if rc, ok := c.contextCustom[rule.Kind]; ok {
		fn, err := rc(c, rule)
		if err != nil {
			return compiledContextRule{err: fmt.Errorf("compile rule %s: %w", safeRuleKindForError(rule.Kind), err)}
		}
		if fn != nil {
			return compiledContextRule{validate: fn}
		}
	}
	compiled := c.compileRule(rule)
	if compiled.err != nil {
		return compiledContextRule{err: compiled.err}
	}
	return compiledContextRule{validate: func(ctx context.Context, v any) error {
		if ctx == nil {
			ctx = context.Background()
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		return compiled.validate(v)
	}}
}

func (c *Compiler) compileRule(rule Rule) compiledRule {
	// Allow custom compilers to handle the rule first
	if rc, ok := c.custom[rule.Kind]; ok {
		fn, err := rc(c, rule)
		if err != nil {
			return compiledRule{err: fmt.Errorf("compile rule %s: %w", safeRuleKindForError(rule.Kind), err)}
		}
		if fn != nil {
			return compiledRule{validate: fn}
		}
	}
	switch rule.Kind {
	case KRequired:
		return compiledRule{validate: c.validateRequired}
	case KString:
		return compiledRule{validate: c.validateString}
	case KLength:
		n := c.getIntArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error {
			return c.validateLength(v, n)
		}}
	case KMinLength:
		n := c.getIntArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error {
			return c.validateMinLength(v, n)
		}}
	case KMaxLength:
		n := c.getIntArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error {
			return c.validateMaxLength(v, n)
		}}
	case KMinRunes:
		n := c.getIntArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error {
			return c.validateMinRunes(v, n)
		}}
	case KMaxRunes:
		n := c.getIntArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error {
			return c.validateMaxRunes(v, n)
		}}
	case KNonEmpty:
		return compiledRule{validate: c.validateNonEmpty}
	case KContains:
		value := c.getStringArg(rule, "value", "")
		return compiledRule{validate: func(v any) error {
			return c.validateStringContains(v, value, true)
		}}
	case KNotContains:
		value := c.getStringArg(rule, "value", "")
		return compiledRule{validate: func(v any) error {
			return c.validateStringContains(v, value, false)
		}}
	case KPrefix:
		value := c.getStringArg(rule, "value", "")
		return compiledRule{validate: func(v any) error {
			return c.validateStringPrefix(v, value)
		}}
	case KSuffix:
		value := c.getStringArg(rule, "value", "")
		return compiledRule{validate: func(v any) error {
			return c.validateStringSuffix(v, value)
		}}
	case KURL:
		return compiledRule{validate: c.validateURL}
	case KHostname:
		return compiledRule{validate: c.validateHostname}
	case KIP:
		return compiledRule{validate: func(v any) error { return c.validateIP(v, "") }}
	case KIPv4:
		return compiledRule{validate: func(v any) error { return c.validateIP(v, "4") }}
	case KIPv6:
		return compiledRule{validate: func(v any) error { return c.validateIP(v, "6") }}
	case KCIDR:
		return compiledRule{validate: c.validateCIDR}
	case KASCII:
		return compiledRule{validate: c.validateASCII}
	case KAlpha:
		return compiledRule{validate: c.validateAlpha}
	case KAlnum:
		return compiledRule{validate: c.validateAlnum}
	case KRegex:
		pattern := c.getStringArg(rule, "pattern", "")
		re, err := c.compileRegexSafe(pattern) // returns (*regexp.Regexp, error)
		if err != nil {
			// Compile must still succeed; create a closure that reports the error
			return compiledRule{validate: func(v any) error {
				return c.invalidRegexPatternError(pattern)
			}}
		}
		return compiledRule{validate: func(v any) error {
			// Pass pattern for nil-regex cases in validateRegex
			return c.validateRegexWithPattern(v, re, pattern)
		}}
	case KOneOf:
		values := c.getStringSliceArg(rule, "values", nil)
		return compiledRule{validate: func(v any) error {
			return c.validateOneOf(v, values)
		}}
	case KInt:
		return compiledRule{validate: c.validateInt}
	case KInt64:
		return compiledRule{validate: c.validateInt64}
	case KMinInt:
		n := c.getInt64Arg(rule, "n", 0)
		return compiledRule{validate: func(v any) error {
			return c.validateMinInt(v, n)
		}}
	case KMaxInt:
		n := c.getInt64Arg(rule, "n", 0)
		return compiledRule{validate: func(v any) error {
			return c.validateMaxInt(v, n)
		}}
	case KFloat:
		return compiledRule{validate: c.validateFloat}
	case KMinNumber:
		n := c.getFloatArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error { return c.validateNumberMin(v, n) }}
	case KMaxNumber:
		n := c.getFloatArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error { return c.validateNumberMax(v, n) }}
	case KGreaterThan:
		n := c.getFloatArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error { return c.validateNumberCompare(v, n, "gt") }}
	case KGreaterThanEqual:
		n := c.getFloatArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error { return c.validateNumberCompare(v, n, "gte") }}
	case KLessThan:
		n := c.getFloatArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error { return c.validateNumberCompare(v, n, "lt") }}
	case KLessThanEqual:
		n := c.getFloatArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error { return c.validateNumberCompare(v, n, "lte") }}
	case KBetween:
		min := c.getFloatArg(rule, "min", 0)
		max := c.getFloatArg(rule, "max", 0)
		return compiledRule{validate: func(v any) error { return c.validateNumberBetween(v, min, max) }}
	case KPositive:
		return compiledRule{validate: c.validateNumberPositive}
	case KNonNegative:
		return compiledRule{validate: c.validateNumberNonNegative}
	case KFinite:
		return compiledRule{validate: c.validateNumberFinite}
	case KSlice:
		return compiledRule{validate: c.validateSlice}
	case KSliceLength:
		n := c.getIntArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error {
			return c.validateSliceLength(v, n)
		}}
	case KMinSliceLength:
		n := c.getIntArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error {
			return c.validateMinSliceLength(v, n)
		}}
	case KMaxSliceLength:
		n := c.getIntArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error {
			return c.validateMaxSliceLength(v, n)
		}}
	case KForEach:
		// Check if there are inner rules from tag parsing
		if rules, ok := rule.Args["rules"]; ok {
			if innerRules, ok := rules.([]Rule); ok {
				elemValidator, err := c.CompileE(innerRules)
				if err != nil {
					return compiledRule{err: err}
				}
				return compiledRule{validate: func(v any) error {
					return c.validateForEach(v, elemValidator)
				}}
			}
		}
		// Fallback to Elem for backward compatibility
		if rule.Elem != nil {
			elemValidator, err := c.CompileE([]Rule{*rule.Elem})
			if err != nil {
				return compiledRule{err: err}
			}
			return compiledRule{validate: func(v any) error {
				return c.validateForEach(v, elemValidator)
			}}
		}
		// Check if there's a validator function in the args
		if validator, ok := rule.Args["validator"]; ok {
			if elemValidator, ok := validator.(func(any) error); ok {
				return compiledRule{validate: func(v any) error {
					return c.validateForEach(v, elemValidator)
				}}
			}
		}
		return compiledRule{validate: func(any) error { return nil }}
	case KSliceUnique:
		return compiledRule{validate: c.validateSliceUnique}
	case KSliceContains:
		value := rule.Args["value"]
		return compiledRule{validate: func(v any) error { return c.validateSliceContains(v, value) }}
	case KArray:
		return compiledRule{validate: c.validateArray}
	case KArrayLength:
		n := c.getIntArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error {
			return c.validateArrayLength(v, n)
		}}
	case KMinArrayLength:
		n := c.getIntArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error {
			return c.validateMinArrayLength(v, n)
		}}
	case KMaxArrayLength:
		n := c.getIntArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error {
			return c.validateMaxArrayLength(v, n)
		}}
	case KArrayForEach:
		if rules, ok := rule.Args["rules"]; ok {
			if innerRules, ok := rules.([]Rule); ok {
				elemValidator, err := c.CompileE(innerRules)
				if err != nil {
					return compiledRule{err: err}
				}
				return compiledRule{validate: func(v any) error {
					return c.validateArrayForEach(v, elemValidator)
				}}
			}
		}
		if rule.Elem != nil {
			elemValidator, err := c.CompileE([]Rule{*rule.Elem})
			if err != nil {
				return compiledRule{err: err}
			}
			return compiledRule{validate: func(v any) error {
				return c.validateArrayForEach(v, elemValidator)
			}}
		}
		if validator, ok := rule.Args["validator"]; ok {
			if elemValidator, ok := validator.(func(any) error); ok {
				return compiledRule{validate: func(v any) error {
					return c.validateArrayForEach(v, elemValidator)
				}}
			}
		}
		return compiledRule{validate: func(any) error { return nil }}
	case KArrayUnique:
		return compiledRule{validate: c.validateArrayUnique}
	case KArrayContains:
		value := rule.Args["value"]
		return compiledRule{validate: func(v any) error { return c.validateArrayContains(v, value) }}
	case KMap:
		return compiledRule{validate: c.validateMap}
	case KMapLength:
		n := c.getIntArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error { return c.validateMapLength(v, n) }}
	case KMinMapKeys:
		n := c.getIntArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error { return c.validateMinMapKeys(v, n) }}
	case KMaxMapKeys:
		n := c.getIntArg(rule, "n", 0)
		return compiledRule{validate: func(v any) error { return c.validateMaxMapKeys(v, n) }}
	case KMapKeys:
		rules, _ := rule.Args["rules"].([]Rule)
		keyValidator, err := c.CompileE(rules)
		if err != nil {
			return compiledRule{err: err}
		}
		return compiledRule{validate: func(v any) error { return c.validateMapKeys(v, keyValidator) }}
	case KMapValues:
		rules, _ := rule.Args["rules"].([]Rule)
		valueValidator, err := c.CompileE(rules)
		if err != nil {
			return compiledRule{err: err}
		}
		return compiledRule{validate: func(v any) error { return c.validateMapValues(v, valueValidator) }}
	case KBool:
		return compiledRule{validate: c.validateBool}
	case KBoolTrue:
		return compiledRule{validate: func(v any) error { return c.validateBoolValue(v, true) }}
	case KBoolFalse:
		return compiledRule{validate: func(v any) error { return c.validateBoolValue(v, false) }}
	case KTime:
		return compiledRule{validate: c.validateTime}
	case KTimeNotZero:
		return compiledRule{validate: c.validateTimeNotZero}
	case KTimeBefore:
		target := c.getTimeArg(rule, "time")
		return compiledRule{validate: func(v any) error { return c.validateTimeBefore(v, target) }}
	case KTimeAfter:
		target := c.getTimeArg(rule, "time")
		return compiledRule{validate: func(v any) error { return c.validateTimeAfter(v, target) }}
	case KTimeBetween:
		start := c.getTimeArg(rule, "start")
		end := c.getTimeArg(rule, "end")
		return compiledRule{validate: func(v any) error { return c.validateTimeBetween(v, start, end) }}
	default:
		// Check if it's a custom type
		if c.isTypeRegistered(string(rule.Kind)) {
			return compiledRule{validate: c.validateCustomType(rule.Kind)}
		}
		return compiledRule{err: unknownRuleKindError(rule.Kind)}
	}
}

func unknownRuleKindError(kind Kind) error {
	msg := fmt.Sprintf("unknown rule kind: %s", safeRuleKindForError(kind))
	return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeUnknown, Msg: msg}}
}

func safeRuleKindForError(kind Kind) string {
	const maxLen = 80
	value := string(kind)
	if len(value) <= maxLen {
		return value
	}
	return value[:maxLen] + "..."
}

// Helper methods for argument extraction
func (c *Compiler) getIntArg(rule Rule, key string, defaultVal int) int {
	if val, ok := rule.Args[key]; ok {
		if n, ok := val.(int); ok {
			return n
		}
		if n, ok := val.(int64); ok {
			return int(n)
		}
	}
	return defaultVal
}

func (c *Compiler) getInt64Arg(rule Rule, key string, defaultVal int64) int64 {
	if val, ok := rule.Args[key]; ok {
		if n, ok := val.(int64); ok {
			return n
		}
	}
	return defaultVal
}

func (c *Compiler) getStringArg(
	rule Rule,
	key string,
	defaultVal string,
) string {
	if val, ok := rule.Args[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return defaultVal
}

func (c *Compiler) getStringSliceArg(
	rule Rule,
	key string,
	defaultVal []string,
) []string {
	if val, ok := rule.Args[key]; ok {
		if slice, ok := val.([]string); ok {
			return slice
		}
	}
	return defaultVal
}

func (c *Compiler) getFloatArg(rule Rule, key string, defaultVal float64) float64 {
	if val, ok := rule.Args[key]; ok {
		switch n := val.(type) {
		case float64:
			return n
		case float32:
			return float64(n)
		case int:
			return float64(n)
		case int64:
			return float64(n)
		}
	}
	return defaultVal
}

func (c *Compiler) getTimeArg(rule Rule, key string) time.Time {
	if val, ok := rule.Args[key]; ok {
		if t, ok := val.(time.Time); ok {
			return t
		}
	}
	return time.Time{}
}

// Validation methods
func (c *Compiler) validateRequired(v any) error {
	if isZeroValue(v) {
		msg := c.translateMessage(verrs.CodeRequired, "value is required", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeRequired, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateString(v any) error {
	if _, ok := v.(string); !ok {
		msg := c.translateMessage("string.type", "expected string", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateLength(v any, n int) error {
	s, ok := v.(string)
	if !ok {
		msg := c.translateMessage("string.type", "expected string", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}
	if len(s) != n {
		msg := c.translateMessage("string.length", fmt.Sprintf("length must be %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringLength, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateMinLength(v any, n int) error {
	s, ok := v.(string)
	if !ok {
		msg := c.translateMessage("string.type", "expected string", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}
	if len(s) < n {
		msg := c.translateMessage("string.min", fmt.Sprintf("minimum length is %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringMin, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateMaxLength(v any, n int) error {
	s, ok := v.(string)
	if !ok {
		msg := c.translateMessage("string.type", "expected string", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}
	if len(s) > n {
		msg := c.translateMessage("string.max", fmt.Sprintf("maximum length is %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringMax, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateRegexWithPattern(v any, regex *regexp.Regexp, pattern string) error {
	s, ok := v.(string)
	if !ok {
		msg := c.translateMessage("string.type", "expected string", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}

	// Check if regex is nil (compilation failed)
	if regex == nil {
		return c.invalidRegexPatternError(pattern)
	}

	// Enforce maximum input length to prevent DoS attacks
	const maxInputLength = 10000
	if len(s) > maxInputLength {
		msg := c.translateMessage("string.regex.inputTooLong", fmt.Sprintf("input too long (max %d characters)", maxInputLength), []any{maxInputLength})
		return verrs.Errors{verrs.FieldError{
			Path: "",
			Code: verrs.CodeStringRegexInputTooLong,
			Msg:  msg,
		}}
	}

	if !regex.MatchString(s) {
		msg := c.translateMessage("string.regex.noMatch", "does not match required pattern", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringRegexNoMatch, Msg: msg}}
	}
	return nil
}

// Backward-compat wrapper (without pattern context)
func (c *Compiler) validateRegex(v any, regex *regexp.Regexp) error {
	return c.validateRegexWithPattern(v, regex, "")
}

func (c *Compiler) validateOneOf(v any, values []string) error {
	s, ok := v.(string)
	if !ok {
		msg := c.translateMessage("string.type", "expected string", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}
	for _, val := range values {
		if s == val {
			return nil
		}
	}
	msg := c.translateMessage("string.oneof", fmt.Sprintf("must be one of: %s", strings.Join(values, ", ")), []any{strings.Join(values, ", ")})
	return verrs.Errors{verrs.FieldError{
		Path: "",
		Code: verrs.CodeStringOneOf,
		Msg:  msg,
	}}
}

func (c *Compiler) validateNonEmpty(v any) error {
	s, ok := v.(string)
	if !ok {
		msg := c.translateMessage("string.type", "expected string", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}
	if s == "" {
		msg := c.translateMessage("string.nonempty", "must not be empty", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringNonEmpty, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateStringContains(v any, value string, shouldContain bool) error {
	s, ok := v.(string)
	if !ok {
		msg := c.translateMessage("string.type", "expected string", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}
	contains := strings.Contains(s, value)
	if shouldContain && !contains {
		msg := c.translateMessage("string.contains", "must contain required text", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringContains, Msg: msg}}
	}
	if !shouldContain && contains {
		msg := c.translateMessage("string.notContains", "must not contain prohibited text", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringNotContains, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateStringPrefix(v any, value string) error {
	s, ok := v.(string)
	if !ok {
		msg := c.translateMessage("string.type", "expected string", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}
	if !strings.HasPrefix(s, value) {
		msg := c.translateMessage("string.prefix", "must have required prefix", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringPrefix, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateStringSuffix(v any, value string) error {
	s, ok := v.(string)
	if !ok {
		msg := c.translateMessage("string.type", "expected string", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}
	if !strings.HasSuffix(s, value) {
		msg := c.translateMessage("string.suffix", "must have required suffix", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringSuffix, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateURL(v any) error {
	s, ok := v.(string)
	if !ok {
		msg := c.translateMessage("string.type", "expected string", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" || !isValidHostPort(u.Host) {
		msg := c.translateMessage("string.url", "must be a valid absolute URL", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringURL, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateHostname(v any) error {
	s, ok := v.(string)
	if !ok {
		msg := c.translateMessage("string.type", "expected string", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}
	if !isValidHostname(s) {
		msg := c.translateMessage("string.hostname", "must be a valid hostname", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringHost, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateIP(v any, version string) error {
	s, ok := v.(string)
	if !ok {
		msg := c.translateMessage("string.type", "expected string", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}
	addr, err := netip.ParseAddr(s)
	if err != nil || (version == "4" && !addr.Is4()) || (version == "6" && !addr.Is6()) {
		msg := c.translateMessage("string.ip", "must be a valid IP address", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringIP, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateCIDR(v any) error {
	s, ok := v.(string)
	if !ok {
		msg := c.translateMessage("string.type", "expected string", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}
	if _, err := netip.ParsePrefix(s); err != nil {
		msg := c.translateMessage("string.cidr", "must be a valid CIDR prefix", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringCIDR, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateASCII(v any) error {
	return c.validateStringRunes(v, verrs.CodeStringASCII, "string.ascii", func(r rune) bool { return r <= 127 })
}

func (c *Compiler) validateAlpha(v any) error {
	return c.validateStringRunes(v, verrs.CodeStringAlpha, "string.alpha", unicode.IsLetter)
}

func (c *Compiler) validateAlnum(v any) error {
	return c.validateStringRunes(v, verrs.CodeStringAlnum, "string.alnum", func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r)
	})
}

func (c *Compiler) validateStringRunes(v any, code, key string, okFn func(rune) bool) error {
	s, ok := v.(string)
	if !ok {
		msg := c.translateMessage("string.type", "expected string", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}
	for _, r := range s {
		if !okFn(r) {
			msg := c.translateMessage(key, key, nil)
			return verrs.Errors{verrs.FieldError{Path: "", Code: code, Msg: msg}}
		}
	}
	return nil
}

func (c *Compiler) validateInt(v any) error {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return nil
	default:
		msg := c.translateMessage("int.type", "expected integer", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeIntType, Msg: msg}}
	}
}

func (c *Compiler) validateInt64(v any) error {
	switch v.(type) {
	case int64:
		return nil
	default:
		msg := c.translateMessage("int64.type", "expected int64", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeInt64Type, Msg: msg}}
	}
}

func (c *Compiler) validateMinInt(v any, n int64) error {
	val, err := c.toInt64(v)
	if err != nil {
		msg := c.translateMessage("int.type", "expected integer", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeIntType, Msg: msg}}
	}
	if val < n {
		msg := c.translateMessage("int.min", fmt.Sprintf("minimum value is %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeIntMin, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateMaxInt(v any, n int64) error {
	val, err := c.toInt64(v)
	if err != nil {
		msg := c.translateMessage("int.type", "expected integer", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeIntType, Msg: msg}}
	}
	if val > n {
		msg := c.translateMessage("int.max", fmt.Sprintf("maximum value is %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeIntMax, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateFloat(v any) error {
	switch v.(type) {
	case float32, float64:
		return nil
	default:
		msg := c.translateMessage("float.type", "expected floating-point number", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeFloatType, Msg: msg}}
	}
}

func (c *Compiler) validateNumberMin(v any, n float64) error {
	val, ok := toNumberFloat64(v)
	if !ok {
		return c.numberTypeError()
	}
	if val < n {
		msg := c.translateMessage("number.min", fmt.Sprintf("minimum value is %g", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeNumberMin, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateNumberMax(v any, n float64) error {
	val, ok := toNumberFloat64(v)
	if !ok {
		return c.numberTypeError()
	}
	if val > n {
		msg := c.translateMessage("number.max", fmt.Sprintf("maximum value is %g", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeNumberMax, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateNumberCompare(v any, n float64, op string) error {
	val, ok := toNumberFloat64(v)
	if !ok {
		return c.numberTypeError()
	}
	var pass bool
	var code, key string
	switch op {
	case "gt":
		pass, code, key = val > n, verrs.CodeNumberGreaterThan, "number.gt"
	case "gte":
		pass, code, key = val >= n, verrs.CodeNumberGreaterThanEqual, "number.gte"
	case "lt":
		pass, code, key = val < n, verrs.CodeNumberLessThan, "number.lt"
	case "lte":
		pass, code, key = val <= n, verrs.CodeNumberLessThanEqual, "number.lte"
	}
	if !pass {
		msg := c.translateMessage(key, key, []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: code, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateNumberBetween(v any, min, max float64) error {
	val, ok := toNumberFloat64(v)
	if !ok {
		return c.numberTypeError()
	}
	if val < min || val > max {
		msg := c.translateMessage("number.between", fmt.Sprintf("must be between %g and %g", min, max), []any{min, max})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeNumberBetween, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateNumberPositive(v any) error {
	val, ok := toNumberFloat64(v)
	if !ok {
		return c.numberTypeError()
	}
	if val <= 0 {
		msg := c.translateMessage("number.positive", "must be positive", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeNumberPositive, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateNumberNonNegative(v any) error {
	val, ok := toNumberFloat64(v)
	if !ok {
		return c.numberTypeError()
	}
	if val < 0 {
		msg := c.translateMessage("number.nonnegative", "must be nonnegative", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeNumberNonNeg, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateNumberFinite(v any) error {
	val, ok := toNumberFloat64(v)
	if !ok {
		return c.numberTypeError()
	}
	if math.IsInf(val, 0) || math.IsNaN(val) {
		msg := c.translateMessage("number.finite", "must be finite", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeNumberFinite, Msg: msg}}
	}
	return nil
}

func (c *Compiler) numberTypeError() error {
	msg := c.translateMessage("number.type", "expected number", nil)
	return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeNumberType, Msg: msg}}
}

func (c *Compiler) validateSlice(v any) error {
	_, err := c.sliceValue(v)
	return err
}

func (c *Compiler) validateSliceLength(v any, n int) error {
	rv, err := c.sliceValue(v)
	if err != nil {
		return err
	}
	if rv.Len() != n {
		msg := c.translateMessage("slice.length", fmt.Sprintf("length must be %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceLength, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateMinSliceLength(v any, n int) error {
	rv, err := c.sliceValue(v)
	if err != nil {
		return err
	}
	if rv.Len() < n {
		msg := c.translateMessage("slice.min", fmt.Sprintf("minimum length is %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceMin, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateMaxSliceLength(v any, n int) error {
	rv, err := c.sliceValue(v)
	if err != nil {
		return err
	}
	if rv.Len() > n {
		msg := c.translateMessage("slice.max", fmt.Sprintf("maximum length is %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceMax, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateForEach(v any, elemValidator ValidatorFunc) error {
	rv, err := c.sliceValue(v)
	if err != nil {
		return err
	}

	var acc verrs.Errors
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		if err := elemValidator(elem); err != nil {
			var es verrs.Errors
			if errors.As(err, &es) {
				// Prefix each child path with [i]
				for _, fe := range es {
					fe.Path = fmt.Sprintf("[%d]%s", i, fe.Path)
					acc = append(acc, fe)
				}
				continue
			}
			// Fallback for non-structured errors
			acc = append(acc, verrs.FieldError{
				Path: fmt.Sprintf("[%d]", i),
				Code: verrs.CodeUnknown,
				Msg:  err.Error(),
			})
		}
	}

	if len(acc) > 0 {
		return acc
	}
	return nil
}

func (c *Compiler) sliceValue(v any) (reflect.Value, error) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Slice {
		return reflect.Value{}, c.sliceTypeError()
	}
	return rv, nil
}

func (c *Compiler) sliceTypeError() error {
	msg := c.translateMessage("slice.type", "expected slice", []any{})
	return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceType, Msg: msg}}
}

func (c *Compiler) validateSliceUnique(v any) error {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Slice {
		msg := c.translateMessage("slice.type", "expected slice", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceType, Msg: msg}}
	}
	seenComparable := map[any]struct{}{}
	seenFallback := map[string]struct{}{}
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		var key any = elem
		if elem != nil && !reflect.TypeOf(elem).Comparable() {
			fallback := fmt.Sprintf("%#v", elem)
			if _, ok := seenFallback[fallback]; ok {
				return c.sliceUniqueError()
			}
			seenFallback[fallback] = struct{}{}
			continue
		}
		if _, ok := seenComparable[key]; ok {
			return c.sliceUniqueError()
		}
		seenComparable[key] = struct{}{}
	}
	return nil
}

func (c *Compiler) sliceUniqueError() error {
	msg := c.translateMessage("slice.unique", "must contain unique elements", nil)
	return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceUnique, Msg: msg}}
}

func (c *Compiler) validateSliceContains(v any, want any) error {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Slice {
		msg := c.translateMessage("slice.type", "expected slice", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceType, Msg: msg}}
	}
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		if reflect.DeepEqual(elem, want) || fmt.Sprint(elem) == fmt.Sprint(want) {
			return nil
		}
	}
	msg := c.translateMessage("slice.contains", "must contain required element", nil)
	return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceContains, Msg: msg}}
}

func (c *Compiler) validateArray(v any) error {
	_, err := c.arrayValue(v)
	return err
}

func (c *Compiler) validateArrayLength(v any, n int) error {
	rv, err := c.arrayValue(v)
	if err != nil {
		return err
	}
	if rv.Len() != n {
		msg := c.translateMessage("array.length", fmt.Sprintf("length must be %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeArrayLength, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateMinArrayLength(v any, n int) error {
	rv, err := c.arrayValue(v)
	if err != nil {
		return err
	}
	if rv.Len() < n {
		msg := c.translateMessage("array.min", fmt.Sprintf("minimum length is %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeArrayMin, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateMaxArrayLength(v any, n int) error {
	rv, err := c.arrayValue(v)
	if err != nil {
		return err
	}
	if rv.Len() > n {
		msg := c.translateMessage("array.max", fmt.Sprintf("maximum length is %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeArrayMax, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateArrayForEach(v any, elemValidator ValidatorFunc) error {
	rv, err := c.arrayValue(v)
	if err != nil {
		return err
	}

	var acc verrs.Errors
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		if err := elemValidator(elem); err != nil {
			var es verrs.Errors
			if errors.As(err, &es) {
				for _, fe := range es {
					fe.Path = fmt.Sprintf("[%d]%s", i, fe.Path)
					acc = append(acc, fe)
				}
				continue
			}
			acc = append(acc, verrs.FieldError{
				Path: fmt.Sprintf("[%d]", i),
				Code: verrs.CodeUnknown,
				Msg:  err.Error(),
			})
		}
	}
	if len(acc) > 0 {
		return acc
	}
	return nil
}

func (c *Compiler) arrayValue(v any) (reflect.Value, error) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Array {
		return reflect.Value{}, c.arrayTypeError()
	}
	return rv, nil
}

func (c *Compiler) arrayTypeError() error {
	msg := c.translateMessage("array.type", "expected array", []any{})
	return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeArrayType, Msg: msg}}
}

func (c *Compiler) validateArrayUnique(v any) error {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Array {
		return c.arrayTypeError()
	}
	seenComparable := map[any]struct{}{}
	seenFallback := map[string]struct{}{}
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		var key any = elem
		if elem != nil && !reflect.TypeOf(elem).Comparable() {
			fallback := fmt.Sprintf("%#v", elem)
			if _, ok := seenFallback[fallback]; ok {
				return c.arrayUniqueError()
			}
			seenFallback[fallback] = struct{}{}
			continue
		}
		if _, ok := seenComparable[key]; ok {
			return c.arrayUniqueError()
		}
		seenComparable[key] = struct{}{}
	}
	return nil
}

func (c *Compiler) arrayUniqueError() error {
	msg := c.translateMessage("array.unique", "must contain unique elements", nil)
	return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeArrayUnique, Msg: msg}}
}

func (c *Compiler) validateArrayContains(v any, want any) error {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Array {
		return c.arrayTypeError()
	}
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		if reflect.DeepEqual(elem, want) || fmt.Sprint(elem) == fmt.Sprint(want) {
			return nil
		}
	}
	msg := c.translateMessage("array.contains", "must contain required element", nil)
	return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeArrayContains, Msg: msg}}
}

func (c *Compiler) validateMap(v any) error {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Map {
		msg := c.translateMessage("map.type", "expected map", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeMapType, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateMapLength(v any, n int) error {
	rv, err := c.mapValue(v)
	if err != nil {
		return err
	}
	if rv.Len() != n {
		msg := c.translateMessage("map.length", fmt.Sprintf("length must be %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeMapLength, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateMinMapKeys(v any, n int) error {
	rv, err := c.mapValue(v)
	if err != nil {
		return err
	}
	if rv.Len() < n {
		msg := c.translateMessage("map.minkeys", fmt.Sprintf("minimum key count is %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeMapMinKeys, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateMaxMapKeys(v any, n int) error {
	rv, err := c.mapValue(v)
	if err != nil {
		return err
	}
	if rv.Len() > n {
		msg := c.translateMessage("map.maxkeys", fmt.Sprintf("maximum key count is %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeMapMaxKeys, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateMapKeys(v any, keyValidator ValidatorFunc) error {
	rv, err := c.mapValue(v)
	if err != nil {
		return err
	}
	return c.validateMapItems(rv, keyValidator, true)
}

func (c *Compiler) validateMapValues(v any, valueValidator ValidatorFunc) error {
	rv, err := c.mapValue(v)
	if err != nil {
		return err
	}
	return c.validateMapItems(rv, valueValidator, false)
}

func (c *Compiler) validateMapItems(rv reflect.Value, validator ValidatorFunc, keys bool) error {
	var acc verrs.Errors
	for _, key := range sortedMapKeys(rv) {
		var target any
		if keys {
			target = key.Interface()
		} else {
			target = rv.MapIndex(key).Interface()
		}
		if err := validator(target); err != nil {
			pathPrefix := pathutil.MapKeySegment(key.Interface())
			var es verrs.Errors
			if errors.As(err, &es) {
				for _, fe := range es {
					fe.Path = pathPrefix + fe.Path
					acc = append(acc, fe)
				}
				continue
			}
			code := verrs.CodeMapValues
			if keys {
				code = verrs.CodeMapKeys
			}
			acc = append(acc, verrs.FieldError{Path: pathPrefix, Code: code, Msg: err.Error()})
		}
	}
	if len(acc) > 0 {
		return acc
	}
	return nil
}

func (c *Compiler) mapValue(v any) (reflect.Value, error) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() || rv.Kind() != reflect.Map {
		msg := c.translateMessage("map.type", "expected map", nil)
		return reflect.Value{}, verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeMapType, Msg: msg}}
	}
	return rv, nil
}

func (c *Compiler) validateBool(v any) error {
	if _, ok := v.(bool); !ok {
		msg := c.translateMessage("bool.type", "expected boolean", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeBoolType, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateBoolValue(v any, want bool) error {
	b, ok := v.(bool)
	if !ok {
		msg := c.translateMessage("bool.type", "expected boolean", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeBoolType, Msg: msg}}
	}
	if b != want {
		code := verrs.CodeBoolFalse
		key := "bool.false"
		if want {
			code = verrs.CodeBoolTrue
			key = "bool.true"
		}
		msg := c.translateMessage(key, key, nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: code, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateTime(v any) error {
	if _, ok := v.(time.Time); !ok {
		msg := c.translateMessage("time.type", "expected time.Time", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeTimeType, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateTimeNotZero(v any) error {
	t, ok := v.(time.Time)
	if !ok {
		return c.validateTime(v)
	}
	if t.IsZero() {
		msg := c.translateMessage("time.notzero", "must not be zero", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeTimeNotZero, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateTimeBefore(v any, target time.Time) error {
	t, ok := v.(time.Time)
	if !ok {
		return c.validateTime(v)
	}
	if !t.Before(target) {
		msg := c.translateMessage("time.before", fmt.Sprintf("must be before %s", target.Format(time.RFC3339Nano)), []any{target.Format(time.RFC3339Nano)})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeTimeBefore, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateTimeAfter(v any, target time.Time) error {
	t, ok := v.(time.Time)
	if !ok {
		return c.validateTime(v)
	}
	if !t.After(target) {
		msg := c.translateMessage("time.after", fmt.Sprintf("must be after %s", target.Format(time.RFC3339Nano)), []any{target.Format(time.RFC3339Nano)})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeTimeAfter, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateTimeBetween(v any, start, end time.Time) error {
	t, ok := v.(time.Time)
	if !ok {
		return c.validateTime(v)
	}
	if t.Before(start) || t.After(end) {
		msg := c.translateMessage("time.between", fmt.Sprintf("must be between %s and %s", start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano)), []any{start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano)})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeTimeBetween, Msg: msg}}
	}
	return nil
}

// Helper methods

func (c *Compiler) toInt64(v any) (int64, error) {
	if val, ok := toInt64(v); ok {
		return val, nil
	}
	return 0, fmt.Errorf("cannot convert %T to int64", v)
}

func (c *Compiler) validateMinRunes(v any, n int) error {
	s, ok := v.(string)
	if !ok {
		msg := c.translateMessage("string.type", "expected string", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}
	if utf8.RuneCountInString(s) < n {
		msg := c.translateMessage("string.minRunes", fmt.Sprintf("minimum rune count is %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringMinRunes, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateMaxRunes(v any, n int) error {
	s, ok := v.(string)
	if !ok {
		msg := c.translateMessage("string.type", "expected string", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}
	if utf8.RuneCountInString(s) > n {
		msg := c.translateMessage("string.maxRunes", fmt.Sprintf("maximum rune count is %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringMaxRunes, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateCustomType(kind Kind) func(any) error {
	return func(v any) error {
		if c.types != nil {
			if validator, exists := c.types.GetTypeValidator(string(kind), c.translator); exists {
				return validator.Validate(v)
			}
		}
		validator, exists := GetGlobalTypeValidator(string(kind), c.translator)
		if !exists {
			return fmt.Errorf("custom type %s not found", kind)
		}
		return validator.Validate(v)
	}
}

func (c *Compiler) isTypeRegistered(name string) bool {
	if c.types != nil && c.types.IsTypeRegistered(name) {
		return true
	}
	return IsGlobalTypeRegistered(name)
}

func toNumberFloat64(v any) (float64, bool) {
	switch x := v.(type) {
	case int:
		return float64(x), true
	case int8:
		return float64(x), true
	case int16:
		return float64(x), true
	case int32:
		return float64(x), true
	case int64:
		return float64(x), true
	case uint:
		return float64(x), true
	case uint8:
		return float64(x), true
	case uint16:
		return float64(x), true
	case uint32:
		return float64(x), true
	case uint64:
		return float64(x), true
	case float32:
		return float64(x), true
	case float64:
		return x, true
	default:
		return 0, false
	}
}

func sortedMapKeys(rv reflect.Value) []reflect.Value {
	keys := rv.MapKeys()
	sort.Slice(keys, func(i, j int) bool {
		left := fmt.Sprint(keys[i].Interface())
		right := fmt.Sprint(keys[j].Interface())
		if left == right {
			return keys[i].Type().String() < keys[j].Type().String()
		}
		return left < right
	})
	return keys
}

func isValidHostPort(hostport string) bool {
	host := hostport
	if h, _, err := net.SplitHostPort(hostport); err == nil {
		host = h
	}
	if addr, err := netip.ParseAddr(strings.Trim(host, "[]")); err == nil {
		return addr.IsValid()
	}
	return isValidHostname(host)
}

func isValidHostname(host string) bool {
	host = strings.TrimSuffix(host, ".")
	if host == "" || len(host) > 253 {
		return false
	}
	labels := strings.Split(host, ".")
	for _, label := range labels {
		if label == "" || len(label) > 63 {
			return false
		}
		for i, r := range label {
			ok := unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-'
			if !ok || (r == '-' && (i == 0 || i == len(label)-1)) {
				return false
			}
		}
	}
	return true
}
