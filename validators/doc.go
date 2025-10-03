// Package validators provides built-in validation rules and plugin architecture.
//
// The validators package contains built-in validation rules for common types
// (string, int, slice, bool) and a plugin architecture for extensible validation
// domains. It provides both built-in validation rules and a plugin architecture
// that allows domain-specific validators (email, UUID, ULID) to be registered
// and used seamlessly with the main validation system. The package is designed
// to be both comprehensive for common use cases and extensible for domain-specific
// validation requirements.
package validators
