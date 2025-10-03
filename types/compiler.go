package types

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode/utf8"

	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/translator"
)

// RuleCompiler compiles a single Rule into a validate function.
// Implementations may precompute heavy state (e.g., compiled regex).
type RuleCompiler func(c *Compiler, rule Rule) (func(any) error, error)

// globalRegistry holds globally registered custom rule compilers.
// NewCompiler copies these into the per-compiler registry.
var globalRegistry = map[Kind]RuleCompiler{}

// RegisterRule registers a global custom Rule compiler. Call this at init.
func RegisterRule(kind Kind, rc RuleCompiler) {
	globalRegistry[kind] = rc
}

// Compiler compiles rules into validator functions.
type Compiler struct {
	translator translator.Translator
	custom     map[Kind]RuleCompiler
}

// NewCompiler creates a new compiler with the given translator.
func NewCompiler(t translator.Translator) *Compiler {
	// Copy global registry so compilers can be customized per instance
	copied := make(map[Kind]RuleCompiler, len(globalRegistry))
	for k, v := range globalRegistry {
		copied[k] = v
	}
	return &Compiler{translator: t, custom: copied}
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

// Compile compiles a slice of rules into a validator function.
func (c *Compiler) Compile(rules []Rule) ValidatorFunc {
	if len(rules) == 0 {
		return func(any) error { return nil }
	}

	// Pre-compile regexes and other expensive operations
	compiledRules := make([]compiledRule, len(rules))
	for i, rule := range rules {
		compiledRules[i] = c.compileRule(rule)
	}

	return func(v any) error {
		for _, rule := range compiledRules {
			if err := rule.validate(v); err != nil {
				return err
			}
		}
		return nil
	}
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
}

func (c *Compiler) compileRule(rule Rule) compiledRule {
	// Allow custom compilers to handle the rule first
	if rc, ok := c.custom[rule.Kind]; ok {
		if fn, err := rc(c, rule); err == nil && fn != nil {
			return compiledRule{validate: fn}
		}
	}
	switch rule.Kind {
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
	case KRegex:
		pattern := c.getStringArg(rule, "pattern", "")
		re, err := c.compileRegexSafe(pattern) // returns (*regexp.Regexp, error)
		if err != nil {
			// Compile must still succeed; create a closure that reports the error
			return compiledRule{validate: func(v any) error {
				msg := c.translateMessage(
					verrs.CodeStringRegexInvalidPattern,
					"invalid regex pattern: %s",
					[]any{pattern},
				)
				return verrs.Errors{verrs.FieldError{
					Path: "", Code: verrs.CodeStringRegexInvalidPattern, Msg: msg,
				}}
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
				elemValidator := c.Compile(innerRules)
				return compiledRule{validate: func(v any) error {
					return c.validateForEach(v, elemValidator)
				}}
			}
		}
		// Fallback to Elem for backward compatibility
		if rule.Elem != nil {
			elemValidator := c.Compile([]Rule{*rule.Elem})
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
	case KBool:
		return compiledRule{validate: c.validateBool}
	default:
		return compiledRule{validate: func(any) error {
			return fmt.Errorf("unknown rule kind: %s", rule.Kind)
		}}
	}
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

// Validation methods
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
		msg := c.translateMessage("string.regex.invalidPattern", "invalid regex pattern: %s", []any{pattern})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringRegexInvalidPattern, Msg: msg}}
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

func (c *Compiler) validateSlice(v any) error {
	if v == nil {
		msg := c.translateMessage("slice.type", "expected slice", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceType, Msg: msg}}
	}
	if reflect.TypeOf(v).Kind() != reflect.Slice {
		msg := c.translateMessage("slice.type", "expected slice", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceType, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateSliceLength(v any, n int) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		msg := c.translateMessage("slice.type", "expected slice", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceType, Msg: msg}}
	}
	if rv.Len() != n {
		msg := c.translateMessage("slice.length", fmt.Sprintf("length must be %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceLength, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateMinSliceLength(v any, n int) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		msg := c.translateMessage("slice.type", "expected slice", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceType, Msg: msg}}
	}
	if rv.Len() < n {
		msg := c.translateMessage("slice.min", fmt.Sprintf("minimum length is %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceMin, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateMaxSliceLength(v any, n int) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		msg := c.translateMessage("slice.type", "expected slice", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceType, Msg: msg}}
	}
	if rv.Len() > n {
		msg := c.translateMessage("slice.max", fmt.Sprintf("maximum length is %d", n), []any{n})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceMax, Msg: msg}}
	}
	return nil
}

func (c *Compiler) validateForEach(v any, elemValidator ValidatorFunc) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		msg := c.translateMessage("slice.type", "expected slice", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeSliceType, Msg: msg}}
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

func (c *Compiler) validateBool(v any) error {
	if _, ok := v.(bool); !ok {
		msg := c.translateMessage("bool.type", "expected boolean", []any{})
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeBoolType, Msg: msg}}
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
