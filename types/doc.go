// Package types provides the core validation engine types and rule system.
//
// The types package contains the fundamental building blocks for the validation
// system including canonical rule representations, compilation of rules into
// executable validator functions, parsing of struct tags into rules, and rule
// type identifiers. It provides the unified AST (Abstract Syntax Tree) for
// validation rules, ensuring consistency between tag-based and builder-based
// validation. The package includes comprehensive regex safety measures and
// Unicode-aware string validation, and is designed to be extensible, allowing
// custom rule types to be registered and compiled through the plugin system.
package types
