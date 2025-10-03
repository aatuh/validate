package validators

import (
	"github.com/aatuh/validate/v3/translator"
)

// DefaultRegistry returns a registry with all built-in validators registered.
func DefaultRegistry() *Registry {
	registry := NewRegistry()

	// Register built-in validators
	registry.RegisterValidator("string", &StringValidatorFactory{})
	registry.RegisterValidator("int", &IntValidatorFactory{})
	registry.RegisterValidator("int64", &IntValidatorFactory{})
	registry.RegisterValidator("bool", &BoolValidatorFactory{})

	return registry
}

// CreateValidator creates a validator for the given type using the default registry.
func CreateValidator(typeName string, translator translator.Translator) (Validator, bool) {
	registry := DefaultRegistry()
	return registry.GetValidator(typeName, translator)
}

// GetSupportedTypes returns all supported validator types from the default registry.
func GetSupportedTypes() []string {
	registry := DefaultRegistry()
	return registry.GetSupportedTypes()
}
