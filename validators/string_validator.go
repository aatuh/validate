package validators

import (
	"github.com/aatuh/validate/v3/translator"
	"github.com/aatuh/validate/v3/types"
)

// StringValidatorFactory creates string validators.
type StringValidatorFactory struct{}

// CreateValidator creates a new string validator instance.
func (f *StringValidatorFactory) CreateValidator(t translator.Translator) Validator {
	return &StringValidatorImpl{
		sv:    NewStringValidators(t),
		rules: []types.Rule{},
	}
}

// StringValidatorImpl implements the Validator interface for string validation.
type StringValidatorImpl struct {
	sv    *StringValidators
	rules []types.Rule
}

// Build creates a validation function from the accumulated rules.
func (v *StringValidatorImpl) Build() func(any) error {
	if len(v.rules) == 0 {
		// If no specific rules, just validate that it's a string
		return v.sv.WithString()
	}

	// Convert rules to validators and combine them
	var validators []StringValidator
	for _, rule := range v.rules {
		validator := v.ruleToValidator(rule)
		if validator != nil {
			validators = append(validators, validator)
		}
	}

	return v.sv.WithString(validators...)
}

// GetRules returns the current rules.
func (v *StringValidatorImpl) GetRules() []types.Rule {
	return v.rules
}

// SetRules sets the rules for this validator.
func (v *StringValidatorImpl) SetRules(rules []types.Rule) {
	v.rules = rules
}

// AddRule adds a single rule to the validator.
func (v *StringValidatorImpl) AddRule(rule types.Rule) {
	v.rules = append(v.rules, rule)
}

// ruleToValidator converts a Rule to a StringValidator function.
func (v *StringValidatorImpl) ruleToValidator(rule types.Rule) StringValidator {
	switch rule.Kind {
	case types.KLength:
		if n, ok := rule.Args["n"].(int64); ok {
			return v.sv.Length(int(n))
		}
	case types.KMinLength:
		if n, ok := rule.Args["n"].(int64); ok {
			return v.sv.MinLength(int(n))
		}
	case types.KMaxLength:
		if n, ok := rule.Args["n"].(int64); ok {
			return v.sv.MaxLength(int(n))
		}
	case types.KMinRunes:
		if n, ok := rule.Args["n"].(int64); ok {
			return v.sv.MinRunes(int(n))
		}
	case types.KMaxRunes:
		if n, ok := rule.Args["n"].(int64); ok {
			return v.sv.MaxRunes(int(n))
		}
	case types.KRegex:
		if pattern, ok := rule.Args["pattern"].(string); ok {
			return v.sv.Regex(pattern)
		}
	case types.KOneOf:
		if values, ok := rule.Args["values"].([]string); ok {
			return v.sv.OneOf(values...)
		}
	}
	return nil
}
