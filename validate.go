package validate

import (
	"fmt"

	"github.com/aatuh/validate/translator"
	"github.com/aatuh/validate/validators"
)

// Validate is the main struct that holds custom validation rules and an
// optional translator.
type Validate struct {
	customRules map[string]func(any) error
	translator  translator.Translator
	pathSep     string
}

// New creates a new Validate with sane defaults.
func New() *Validate {
	return &Validate{
		customRules: make(map[string]func(any) error),
		pathSep:     ".",
	}
}

// NewWithCustomRules creates a new Validate with pre-registered rules.
func NewWithCustomRules(
	custom map[string]func(any) error,
) *Validate {
	v := New()
	for k, fn := range custom {
		v.customRules[k] = fn
	}
	return v
}

// WithTranslator sets a Translator for localized error messages and
// returns the receiver for chaining.
func (v *Validate) WithTranslator(
	t translator.Translator,
) *Validate {
	v.translator = t
	return v
}

// PathSeparator customizes the separator used in nested field paths.
// Example: "User.Addresses[2].Zip".
func (v *Validate) PathSeparator(sep string) *Validate {
	if sep != "" {
		v.pathSep = sep
	}
	return v
}

// FromRules creates a validator func from rule tokens. The first token
// is the type: "string", "int", "int64", "slice", "bool".
func (v *Validate) FromRules(
	rules []string,
) (func(any) error, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("empty rules")
	}
	if rule, ok := v.customRules[rules[0]]; ok {
		return rule, nil
	}

	sv := validators.NewStringValidators(v.translator)
	iv := validators.NewIntValidators(v.translator)
	sl := validators.NewSliceValidators(v.translator)
	bv := validators.NewBoolValidators(v.translator)

	switch rules[0] {
	case "string":
		return validators.BuildStringValidator(sv, rules)
	case "int", "int64":
		return validators.BuildIntValidator(iv, rules, rules[0])
	case "bool":
		return bv.WithBool(), nil
	case "slice":
		return validators.BuildSliceValidator(sl, rules)
	default:
		return nil, fmt.Errorf("unknown validator type: %s", rules[0])
	}
}

// String returns a fluent string validator builder.
func (v *Validate) String() *StringBuilder {
	return &StringBuilder{
		sv:    validators.NewStringValidators(v.translator),
		rules: nil,
	}
}

// Int returns a fluent integer validator builder that accepts any
// Go int type at call time.
func (v *Validate) Int() *IntBuilder {
	return &IntBuilder{
		iv:    validators.NewIntValidators(v.translator),
		exact: false,
	}
}

// Int64 returns a fluent builder that requires exactly int64.
func (v *Validate) Int64() *IntBuilder {
	return &IntBuilder{
		iv:    validators.NewIntValidators(v.translator),
		exact: true,
	}
}

// Slice returns a fluent slice validator builder. It accepts any slice
// element type at call time.
func (v *Validate) Slice() *SliceBuilder {
	return &SliceBuilder{
		sv:    validators.NewSliceValidators(v.translator),
		rules: nil,
	}
}

// Bool returns a bool validator func (no rules yet, kept for symmetry).
func (v *Validate) Bool() func(any) error {
	return validators.NewBoolValidators(v.translator).WithBool()
}
