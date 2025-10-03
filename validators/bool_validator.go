package validators

import (
	"github.com/aatuh/validate/v3/translator"
	"github.com/aatuh/validate/v3/types"
)

// BoolValidatorFactory creates boolean validators.
type BoolValidatorFactory struct{}

// CreateValidator creates a new boolean validator instance.
func (f *BoolValidatorFactory) CreateValidator(t translator.Translator) Validator {
	return &BoolValidatorImpl{
		bv:    NewBoolValidators(t),
		rules: []types.Rule{},
	}
}

// BoolValidatorImpl implements the Validator interface for boolean validation.
type BoolValidatorImpl struct {
	bv    *BoolValidators
	rules []types.Rule
}

// Build creates a validation function from the accumulated rules.
func (v *BoolValidatorImpl) Build() func(any) error {
	// For boolean validation, we typically just check the type
	// since there are no common boolean-specific rules
	return v.bv.WithBool()
}

// GetRules returns the current rules.
func (v *BoolValidatorImpl) GetRules() []types.Rule {
	return v.rules
}

// SetRules sets the rules for this validator.
func (v *BoolValidatorImpl) SetRules(rules []types.Rule) {
	v.rules = rules
}

// AddRule adds a single rule to the validator.
func (v *BoolValidatorImpl) AddRule(rule types.Rule) {
	v.rules = append(v.rules, rule)
}
