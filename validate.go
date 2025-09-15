package validate

import (
	"fmt"
)

// Validate is the main struct that holds custom validation rules and an
// optional translator.
type Validate struct {
	customRules map[string]func(any) error
	translator  Translator
}

// NewValidate creates a new Validate instance.
func NewValidate(customRules map[string]func(any) error) *Validate {
	return &Validate{
		customRules: customRules,
	}
}

// WithTranslator sets a Translator for localized error messages.
func (v *Validate) WithTranslator(t Translator) {
	v.translator = t
}

// translate returns a translated message if a translator is set.
func (v *Validate) translate(key string, params ...any) string {
	if v.translator != nil {
		return v.translator.T(key, params...)
	}
	// Default behavior: use fmt.Sprintf on the key.
	return fmt.Sprintf(key, params...)
}

// FromRules creates a validator function from a slice of rules.
// The first element indicates the type (e.g., "string", "int", "slice", etc.).
func (v *Validate) FromRules(rules []string) (func(any) error, error) {
	if rule, ok := v.customRules[rules[0]]; ok {
		return rule, nil
	}
	switch rules[0] {
	case "string":
		return BuildStringValidator(v, rules)
	case "int", "int64":
		return buildIntValidator(v, rules, rules[0])
	case "bool":
		return v.WithBool(), nil
	case "slice":
		return BuildSliceValidator(v, rules)
	default:
		return nil, fmt.Errorf("unknown validator type: %s", rules[0])
	}
}
