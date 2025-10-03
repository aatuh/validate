package core

import (
	"github.com/aatuh/validate/v3/types"
)

// ValidatorBuilder defines the interface for building validators.
// This allows the core engine to work with any validator implementation
// without knowing the specific types.
type ValidatorBuilder interface {
	Build() func(any) error
}

// RuleBuilder defines the interface for building rules.
// This allows validators to construct rules that the core engine can compile.
type RuleBuilder interface {
	GetRules() []types.Rule
	SetRules(rules []types.Rule)
}

// ValidatorFactory defines the interface for creating validator instances.
// This allows the core engine to create validators without knowing their
// specific implementation details.
type ValidatorFactory interface {
	CreateValidator(validatorType string) ValidatorBuilder
}

// ValidatorRegistry defines the interface for registering custom validators.
// This allows the core engine to work with plugin-style validators.
type ValidatorRegistry interface {
	RegisterValidator(name string, factory func() ValidatorBuilder)
	GetValidator(name string) (ValidatorBuilder, bool)
}
