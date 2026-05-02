package structvalidator

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/aatuh/validate/v3/core"
	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/internal/pathutil"
	"github.com/aatuh/validate/v3/types"
)

// StructValidator provides struct validation functionality.
//
// Fields:
//   - validator: The underlying Validate instance for validation rules.
type StructValidator struct{ validator *core.Validate }

// NewStructValidator creates a new StructValidator instance.
//
// Parameters:
//   - v: The Validate instance to use for validation rules.
//
// Returns:
//   - *StructValidator: A new StructValidator instance.
func NewStructValidator(v *core.Validate) *StructValidator {
	return &StructValidator{validator: v}
}

// ValidateStruct keeps backward compatibility and uses default options.
//
// Parameters:
//   - s: The struct to validate.
//
// Returns:
//   - error: Validation errors if any, nil if valid.
func (sv *StructValidator) ValidateStruct(s any) error {
	return sv.ValidateStructWithOpts(s, core.ValidateOpts{})
}

// ValidateStructContext validates a struct using `validate` tags with context.
func (sv *StructValidator) ValidateStructContext(ctx context.Context, s any) error {
	return sv.ValidateStructContextWithOpts(ctx, s, core.ValidateOpts{})
}

// ValidateStructWithOpts validates s, honoring StopOnFirst and PathSep.
// Expected tag example: `validate:"string;min=3;max=10"`.
//
// Parameters:
//   - s: The struct to validate.
//   - opts: Validation options including StopOnFirst and PathSep.
//
// Returns:
//   - error: Validation errors if any, nil if valid.
func (sv *StructValidator) ValidateStructWithOpts(
	s any, opts core.ValidateOpts,
) error {
	return sv.ValidateStructContextWithOpts(context.Background(), s, opts)
}

// ValidateStructContextWithOpts validates s with context and options.
func (sv *StructValidator) ValidateStructContextWithOpts(
	ctx context.Context,
	s any,
	opts core.ValidateOpts,
) error {
	if ctx == nil {
		ctx = context.Background()
	}
	opts = core.ApplyOpts(sv.validator, opts)

	val := reflect.ValueOf(s)
	typ := reflect.TypeOf(s)
	if !val.IsValid() {
		return fmt.Errorf("ValidateStruct: expected struct, got %T", s)
	}

	// Dereference pointer if necessary.
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return fmt.Errorf("ValidateStruct: expected struct, got %T", s)
		}
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("ValidateStruct: expected struct, got %T", s)
	}

	var errs verrs.Errors
	var terminalErr error

	// walkStruct returns true to continue, false to stop early.
	var walkStruct func(v reflect.Value, t reflect.Type, path string) bool
	walkStruct = func(v reflect.Value, t reflect.Type, path string) bool {
		for i := 0; i < v.NumField(); i++ {
			if err := ctx.Err(); err != nil {
				terminalErr = err
				return false
			}
			ft := t.Field(i)
			fv := v.Field(i)

			// Skip unexported fields.
			if ft.PkgPath != "" {
				continue
			}

			displayName := fieldDisplayName(ft, opts)
			fieldPath := fieldPathJoin(path, displayName, opts.PathSep)

			// Recurse into structs/slices/maps when no tag is present.
			tag := ft.Tag.Get("validate")
			if tag == "" {
				// Dereference pointer before checking kind
				derefFv := derefPointer(fv)
				switch derefFv.Kind() {
				case reflect.Struct:
					if !walkStruct(derefFv, derefFv.Type(), fieldPath) &&
						opts.StopOnFirst {
						return false
					}
					continue
				case reflect.Slice, reflect.Array:
					for j := 0; j < derefFv.Len(); j++ {
						ep := fieldPath + "[" + strconv.Itoa(j) + "]"
						ev := derefFv.Index(j)
						// Dereference pointer in slice elements
						derefEv := derefPointer(ev)
						if derefEv.Kind() == reflect.Struct {
							if !walkStruct(derefEv, derefEv.Type(), ep) &&
								opts.StopOnFirst {
								return false
							}
						}
					}
					continue
				case reflect.Map:
					for _, mk := range sortedMapKeys(derefFv) {
						ev := derefFv.MapIndex(mk)
						ep := fieldPath + pathutil.MapKeySegment(mk.Interface())
						// Dereference pointer in map values
						derefEv := derefPointer(ev)
						if derefEv.Kind() == reflect.Struct {
							if !walkStruct(derefEv, derefEv.Type(), ep) &&
								opts.StopOnFirst {
								return false
							}
						}
					}
					continue
				default:
					continue
				}
			}

			// Validate with rules from tag.
			tokens := types.SplitTag(tag)
			rules, structRules, err := splitStructRules(tokens)
			if err != nil {
				errs = append(errs, verrs.FieldError{Path: fieldPath, Code: verrs.CodeUnknown, Msg: err.Error()})
				if opts.StopOnFirst {
					return false
				}
				continue
			}
			ctxFn := func(context.Context, any) error { return nil }
			if len(rules) > 0 {
				ctxFn, err = sv.validator.FromRulesContextWithOpts(rules, types.CompileOpts{CollectAll: opts.CollectAllRules})
				if err != nil {
					errs = append(errs, verrs.FieldError{
						Path: fieldPath, Code: verrs.CodeUnknown,
						Msg: err.Error(),
					})
					if opts.StopOnFirst {
						return false
					}
					continue
				}
			}
			fieldValue := valueForValidation(fv)
			if err := validateStructRules(ctx, fieldValue, v, ft, structRules, fieldPath, opts, sv.validator); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					terminalErr = err
					return false
				}
				var fieldErrors verrs.Errors
				if errors.As(err, &fieldErrors) {
					errs = append(errs, fieldErrors...)
				} else {
					errs = append(errs, verrs.FieldError{Path: fieldPath, Code: verrs.CodeUnknown, Msg: err.Error()})
				}
				if opts.StopOnFirst {
					return false
				}
				if !opts.CollectAllRules || hasRequiredFailure(err) {
					continue
				}
			}
			if err := ctxFn(ctx, fieldValue); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					terminalErr = err
					return false
				}
				appendValidationErrors(&errs, err, fieldPath, opts)
				if opts.StopOnFirst {
					return false
				}
			}
		}
		return true
	}

	// Start the walk from the root.
	walkStruct(val, typ, "")

	if terminalErr != nil {
		return terminalErr
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

// derefPointer dereferences a pointer value recursively until it reaches a non-pointer type.
func derefPointer(v reflect.Value) reflect.Value {
	for v.IsValid() && v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v
}

// fieldPathJoin joins path parts with a custom separator.
// Handles bracket-prefixed paths (e.g., "[0]", "[key]") by concatenating without separator.
func fieldPathJoin(base, name, sep string) string {
	if base == "" {
		return name
	}
	if name == "" {
		return base
	}
	if sep == "" {
		sep = "."
	}
	// If the child path starts with a bracket, concatenate directly without separator
	if len(name) > 0 && name[0] == '[' {
		return base + name
	}
	return base + sep + name
}

// JSONFieldName returns a field's JSON tag name, falling back to the Go name.
func JSONFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" || name == "-" {
		return field.Name
	}
	return name
}

func fieldDisplayName(field reflect.StructField, opts core.ValidateOpts) string {
	if opts.FieldNameFunc == nil {
		return field.Name
	}
	if name := opts.FieldNameFunc(field); name != "" {
		return name
	}
	return field.Name
}

func valueForValidation(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	v = derefPointer(v)
	if !v.IsValid() {
		return nil
	}
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}
	if !v.CanInterface() {
		return nil
	}
	return v.Interface()
}

const (
	structRuleEqual          types.Kind = "eqField"
	structRuleNotEqual       types.Kind = "neField"
	structRuleRequiredWith   types.Kind = "requiredWith"
	structRuleRequiredIf     types.Kind = "requiredIf"
	structRuleRequiredUnless types.Kind = "requiredUnless"
)

func splitStructRules(tokens []string) ([]string, []types.Rule, error) {
	if len(tokens) == 0 {
		return tokens, nil, nil
	}
	out := make([]string, 0, len(tokens))
	structRules := make([]types.Rule, 0, 2)
	for _, token := range tokens {
		switch {
		case strings.HasPrefix(token, "eqField="):
			structRules = append(structRules, types.NewRule(structRuleEqual, map[string]any{"field": strings.TrimPrefix(token, "eqField=")}))
		case strings.HasPrefix(token, "neField="):
			structRules = append(structRules, types.NewRule(structRuleNotEqual, map[string]any{"field": strings.TrimPrefix(token, "neField=")}))
		case strings.HasPrefix(token, "requiredWith="):
			structRules = append(structRules, types.NewRule(structRuleRequiredWith, map[string]any{"field": strings.TrimPrefix(token, "requiredWith=")}))
		case strings.HasPrefix(token, "requiredIf="):
			rule, err := parseConditionalRequiredRule(structRuleRequiredIf, token, "requiredIf=")
			if err != nil {
				return nil, nil, err
			}
			structRules = append(structRules, rule)
		case strings.HasPrefix(token, "requiredUnless="):
			rule, err := parseConditionalRequiredRule(structRuleRequiredUnless, token, "requiredUnless=")
			if err != nil {
				return nil, nil, err
			}
			structRules = append(structRules, rule)
		case strings.HasPrefix(token, "struct:"):
			rule, err := parseStructCustomRule(token)
			if err != nil {
				return nil, nil, err
			}
			structRules = append(structRules, rule)
		default:
			out = append(out, token)
		}
	}
	return out, structRules, nil
}

func parseConditionalRequiredRule(kind types.Kind, token, prefix string) (types.Rule, error) {
	raw := strings.TrimPrefix(token, prefix)
	field, value, ok := strings.Cut(raw, ",")
	if !ok || field == "" {
		return types.Rule{}, fmt.Errorf("%s requires Field,value", strings.TrimSuffix(prefix, "="))
	}
	return types.NewRule(kind, map[string]any{"field": field, "value": value}), nil
}

func parseStructCustomRule(token string) (types.Rule, error) {
	raw := strings.TrimPrefix(token, "struct:")
	name, value, hasValue := strings.Cut(raw, "=")
	if name == "" {
		return types.Rule{}, fmt.Errorf("struct rule name cannot be empty")
	}
	var args map[string]any
	if hasValue {
		args = map[string]any{"value": value}
	}
	return types.NewRule(types.Kind(name), args), nil
}

func validateStructRules(
	runtimeCtx context.Context,
	value any,
	owner reflect.Value,
	field reflect.StructField,
	rules []types.Rule,
	path string,
	opts core.ValidateOpts,
	v *core.Validate,
) error {
	if len(rules) == 0 {
		return nil
	}
	var errs verrs.Errors
	for _, rule := range rules {
		if err := runtimeCtx.Err(); err != nil {
			return err
		}
		fn, err := compileStructRule(rule, v)
		if err != nil {
			errs = append(errs, verrs.FieldError{Path: path, Code: verrs.CodeUnknown, Msg: err.Error()})
			if !opts.CollectAllRules {
				return errs
			}
			continue
		}
		ctx := core.StructRuleContext{
			Path:       path,
			Field:      field,
			Value:      value,
			Owner:      owner,
			Rule:       rule,
			Context:    runtimeCtx,
			Translator: v.Translator(),
		}
		if err := fn(ctx); err != nil {
			appendValidationErrors(&errs, err, path, opts)
			if !opts.CollectAllRules || hasRequiredFailure(err) {
				return errs
			}
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func compileStructRule(rule types.Rule, v *core.Validate) (core.StructRuleFunc, error) {
	if compiler, ok := v.StructRuleCompiler(rule.Kind); ok {
		fn, err := compiler(rule)
		if err != nil {
			return nil, err
		}
		if fn == nil {
			return nil, fmt.Errorf("struct rule %s returned nil validator", rule.Kind)
		}
		return fn, nil
	}
	switch rule.Kind {
	case structRuleEqual:
		field, err := structRuleFieldArg(rule)
		if err != nil {
			return nil, err
		}
		return func(ctx core.StructRuleContext) error {
			other, ok := ctx.FieldValue(field)
			if !ok {
				return fieldReferenceError(ctx, field)
			}
			if !reflect.DeepEqual(ctx.Value, other) {
				return verrs.Errors{verrs.FieldError{Code: verrs.CodeFieldEqual, Msg: translate(ctx.Translator, verrs.CodeFieldEqual, "must match the referenced field")}}
			}
			return nil
		}, nil
	case structRuleNotEqual:
		field, err := structRuleFieldArg(rule)
		if err != nil {
			return nil, err
		}
		return func(ctx core.StructRuleContext) error {
			other, ok := ctx.FieldValue(field)
			if !ok {
				return fieldReferenceError(ctx, field)
			}
			if reflect.DeepEqual(ctx.Value, other) {
				return verrs.Errors{verrs.FieldError{Code: verrs.CodeFieldNotEqual, Msg: translate(ctx.Translator, verrs.CodeFieldNotEqual, "must differ from the referenced field")}}
			}
			return nil
		}, nil
	case structRuleRequiredWith:
		field, err := structRuleFieldArg(rule)
		if err != nil {
			return nil, err
		}
		return func(ctx core.StructRuleContext) error {
			other, ok := ctx.FieldValue(field)
			if !ok {
				return fieldReferenceError(ctx, field)
			}
			if !isZeroValue(other) && isZeroValue(ctx.Value) {
				return verrs.Errors{verrs.FieldError{Code: verrs.CodeRequiredWith, Msg: translate(ctx.Translator, verrs.CodeRequiredWith, "value is required")}}
			}
			return nil
		}, nil
	case structRuleRequiredIf:
		field, want, err := structRuleConditionArgs(rule)
		if err != nil {
			return nil, err
		}
		return func(ctx core.StructRuleContext) error {
			other, ok := ctx.FieldValue(field)
			if !ok {
				return fieldReferenceError(ctx, field)
			}
			if fmt.Sprint(other) == want && isZeroValue(ctx.Value) {
				return verrs.Errors{verrs.FieldError{Code: verrs.CodeRequiredIf, Msg: translate(ctx.Translator, verrs.CodeRequiredIf, "value is required")}}
			}
			return nil
		}, nil
	case structRuleRequiredUnless:
		field, want, err := structRuleConditionArgs(rule)
		if err != nil {
			return nil, err
		}
		return func(ctx core.StructRuleContext) error {
			other, ok := ctx.FieldValue(field)
			if !ok {
				return fieldReferenceError(ctx, field)
			}
			if fmt.Sprint(other) != want && isZeroValue(ctx.Value) {
				return verrs.Errors{verrs.FieldError{Code: verrs.CodeRequiredUnless, Msg: translate(ctx.Translator, verrs.CodeRequiredUnless, "value is required")}}
			}
			return nil
		}, nil
	default:
		return nil, fmt.Errorf("unknown struct rule kind: %s", rule.Kind)
	}
}

func structRuleFieldArg(rule types.Rule) (string, error) {
	field, _ := rule.Args["field"].(string)
	if field == "" {
		return "", fmt.Errorf("struct rule %s requires a field argument", rule.Kind)
	}
	return field, nil
}

func structRuleConditionArgs(rule types.Rule) (string, string, error) {
	field, err := structRuleFieldArg(rule)
	if err != nil {
		return "", "", err
	}
	value, _ := rule.Args["value"].(string)
	return field, value, nil
}

func fieldReferenceError(ctx core.StructRuleContext, field string) error {
	return verrs.Errors{verrs.FieldError{
		Code:  verrs.CodeFieldReference,
		Param: field,
		Msg:   translate(ctx.Translator, verrs.CodeFieldReference, "invalid referenced field"),
	}}
}

func appendValidationErrors(errs *verrs.Errors, err error, fieldPath string, opts core.ValidateOpts) {
	var fieldErrors verrs.Errors
	if errors.As(err, &fieldErrors) {
		for _, fe := range fieldErrors {
			fe.Path = fieldPathJoin(fieldPath, fe.Path, opts.PathSep)
			*errs = append(*errs, fe)
		}
		return
	}
	*errs = append(*errs, verrs.FieldError{
		Path: fieldPath, Code: verrs.CodeUnknown,
		Msg: err.Error(),
	})
}

func hasRequiredFailure(err error) bool {
	var fieldErrors verrs.Errors
	if !errors.As(err, &fieldErrors) {
		return false
	}
	for _, fe := range fieldErrors {
		switch fe.Code {
		case verrs.CodeRequired, verrs.CodeRequiredWith, verrs.CodeRequiredIf, verrs.CodeRequiredUnless:
			return true
		}
	}
	return false
}

func translate(tr interface {
	T(string, ...any) string
}, key, fallback string) string {
	if tr == nil {
		return fallback
	}
	if msg := tr.T(key); msg != "" {
		return msg
	}
	return fallback
}

func isZeroValue(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice, reflect.Map:
		return rv.IsNil() || rv.Len() == 0
	case reflect.String:
		return rv.Len() == 0
	}
	return reflect.DeepEqual(v, reflect.Zero(rv.Type()).Interface())
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
