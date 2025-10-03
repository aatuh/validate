package validators

import (
	"github.com/aatuh/validate/v3/translator"
	"github.com/aatuh/validate/v3/types"
)

// Registry manages validator implementations and provides a way to register
// and retrieve validators without the core engine knowing about specific types.
type Registry struct {
	validators map[string]ValidatorFactory
}

// ValidatorFactory creates validator instances for a specific type.
type ValidatorFactory interface {
	CreateValidator(translator translator.Translator) Validator
}

// Validator defines the interface that all validators must implement.
type Validator interface {
	// Build creates a validation function from the accumulated rules.
	Build() func(any) error

	// GetRules returns the current rules.
	GetRules() []types.Rule

	// SetRules sets the rules for this validator.
	SetRules(rules []types.Rule)
}

// NewRegistry creates a new validator registry.
func NewRegistry() *Registry {
	return &Registry{
		validators: make(map[string]ValidatorFactory),
	}
}

// RegisterValidator registers a validator factory for a given type name.
func (r *Registry) RegisterValidator(name string, factory ValidatorFactory) {
	r.validators[name] = factory
}

// GetValidator creates a new validator instance for the given type.
func (r *Registry) GetValidator(name string, translator translator.Translator) (Validator, bool) {
	factory, exists := r.validators[name]
	if !exists {
		return nil, false
	}
	return factory.CreateValidator(translator), true
}

// GetSupportedTypes returns a list of all registered validator types.
func (r *Registry) GetSupportedTypes() []string {
	types := make([]string, 0, len(r.validators))
	for name := range r.validators {
		types = append(types, name)
	}
	return types
}
