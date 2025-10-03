package validators

import (
	"github.com/aatuh/validate/v3/translator"
	"github.com/aatuh/validate/v3/types"
)

// IntValidatorFactory creates integer validators.
type IntValidatorFactory struct{}

// CreateValidator creates a new integer validator instance.
func (f *IntValidatorFactory) CreateValidator(t translator.Translator) Validator {
	return &IntValidatorImpl{
		iv:    NewIntValidators(t),
		rules: []types.Rule{},
	}
}

// IntValidatorImpl implements the Validator interface for integer validation.
type IntValidatorImpl struct {
	iv    *IntValidators
	rules []types.Rule
	exact bool // true for int64, false for int
}

// Build creates a validation function from the accumulated rules.
func (v *IntValidatorImpl) Build() func(any) error {
	if len(v.rules) == 0 {
		// If no specific rules, just validate that it's an integer
		if v.exact {
			return v.iv.WithInt64()
		}
		return v.iv.WithInt()
	}

	// Convert rules to validators and combine them
	var validators []IntValidator
	for _, rule := range v.rules {
		validator := v.ruleToValidator(rule)
		if validator != nil {
			validators = append(validators, validator)
		}
	}

	if v.exact {
		return v.iv.WithInt64(validators...)
	}
	return v.iv.WithInt(validators...)
}

// GetRules returns the current rules.
func (v *IntValidatorImpl) GetRules() []types.Rule {
	return v.rules
}

// SetRules sets the rules for this validator.
func (v *IntValidatorImpl) SetRules(rules []types.Rule) {
	v.rules = rules
}

// AddRule adds a single rule to the validator.
func (v *IntValidatorImpl) AddRule(rule types.Rule) {
	v.rules = append(v.rules, rule)
}

// SetExact sets whether this validator should be exact (int64) or not (int).
func (v *IntValidatorImpl) SetExact(exact bool) {
	v.exact = exact
}

// ruleToValidator converts a Rule to an IntValidator function.
func (v *IntValidatorImpl) ruleToValidator(rule types.Rule) IntValidator {
	switch rule.Kind {
	case types.KMinInt:
		if n, ok := rule.Args["n"].(int64); ok {
			return v.iv.MinInt(n)
		}
	case types.KMaxInt:
		if n, ok := rule.Args["n"].(int64); ok {
			return v.iv.MaxInt(n)
		}
	}
	return nil
}
