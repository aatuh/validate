package types

import (
	"sort"
	"sync"

	"github.com/aatuh/validate/v3/translator"
)

// TypeValidator defines the interface for custom type validators.
// Custom types must implement this interface to be registered.
type TypeValidator interface {
	// Validate validates a value of this custom type.
	Validate(value any) error
}

// TypeValidatorFactory creates type validators for a specific custom type.
type TypeValidatorFactory interface {
	// CreateValidator creates a new validator instance for this type.
	CreateValidator(translator translator.Translator) TypeValidator
}

// TypeRegistry manages custom type validators and provides a way to register
// and retrieve type validators without the core engine knowing about specific types.
type TypeRegistry struct {
	mu    sync.RWMutex
	types map[string]TypeValidatorFactory
}

// NewTypeRegistry creates a new type registry.
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		types: make(map[string]TypeValidatorFactory),
	}
}

// Clone returns an independent copy of the registry.
func (r *TypeRegistry) Clone() *TypeRegistry {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	cp := NewTypeRegistry()
	for name, factory := range r.types {
		cp.types[name] = factory
	}
	return cp
}

// RegisterType registers a type validator factory for a given type name.
func (r *TypeRegistry) RegisterType(name string, factory TypeValidatorFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.types[name] = factory
}

// GetTypeValidator creates a new type validator instance for the given type.
func (r *TypeRegistry) GetTypeValidator(name string, translator translator.Translator) (TypeValidator, bool) {
	r.mu.RLock()
	factory, exists := r.types[name]
	r.mu.RUnlock()
	if !exists {
		return nil, false
	}
	return factory.CreateValidator(translator), true
}

// GetSupportedTypes returns a list of all registered custom types.
func (r *TypeRegistry) GetSupportedTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := make([]string, 0, len(r.types))
	for name := range r.types {
		types = append(types, name)
	}
	sort.Strings(types)
	return types
}

// IsTypeRegistered checks if a type is registered.
func (r *TypeRegistry) IsTypeRegistered(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.types[name]
	return exists
}

// globalTypeRegistry holds the global type registry for custom types.
var globalTypeRegistry = NewTypeRegistry()

// RegisterGlobalType registers a process-wide type in the global registry.
// Duplicate names overwrite earlier factories.
func RegisterGlobalType(name string, factory TypeValidatorFactory) {
	globalTypeRegistry.RegisterType(name, factory)
}

// GetGlobalTypeValidator gets a type validator from the global registry.
func GetGlobalTypeValidator(name string, translator translator.Translator) (TypeValidator, bool) {
	return globalTypeRegistry.GetTypeValidator(name, translator)
}

// GetGlobalSupportedTypes returns all globally registered types.
func GetGlobalSupportedTypes() []string {
	return globalTypeRegistry.GetSupportedTypes()
}

// IsGlobalTypeRegistered checks if a type is registered globally.
func IsGlobalTypeRegistered(name string) bool {
	return globalTypeRegistry.IsTypeRegistered(name)
}
