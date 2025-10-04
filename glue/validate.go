package glue

import (
	"github.com/aatuh/validate/v3/core"
	"github.com/aatuh/validate/v3/structvalidator"
	"github.com/aatuh/validate/v3/translator"
	"github.com/aatuh/validate/v3/types"
)

// Validate provides the main validation API that glues together the core
// engine with specific validator implementations.
type Validate struct {
	engine *core.Engine
}

// New creates a new Validate instance with sensible defaults.
func New() *Validate {
	engine := core.NewEngine()
	return &Validate{engine: engine}
}

// NewWithTranslator returns a Validate configured with the provided
// translator while keeping other defaults.
func NewWithTranslator(tr translator.Translator) *Validate {
	engine := core.NewEngine().WithTranslator(tr)
	return &Validate{engine: engine}
}

// NewBare returns a Validate without installing a default translator.
// Useful for advanced setups that manage translations differently.
func NewBare() *Validate {
	return &Validate{engine: core.NewEngine()}
}

// WithCustomRule returns a copy with an additional custom rule.
func (v *Validate) WithCustomRule(
	name string, rule func(any) error,
) *Validate {
	return &Validate{
		engine: v.engine.WithCustomRule(name, rule),
	}
}

// WithTranslator sets a Translator and returns a new Validate.
func (v *Validate) WithTranslator(t translator.Translator) *Validate {
	return &Validate{
		engine: v.engine.WithTranslator(t),
	}
}

// PathSeparator customizes the nested field path separator.
func (v *Validate) PathSeparator(sep string) *Validate {
	return &Validate{
		engine: v.engine.PathSeparator(sep),
	}
}

// FromRules creates a validator function from rule tokens.
func (v *Validate) FromRules(
	rules []string,
) (func(any) error, error) {
	return v.engine.FromRules(rules)
}

// FromTag compiles a single tag string into a validator function.
func (v *Validate) FromTag(tag string) (func(any) error, error) {
	if tag == "" {
		return func(any) error { return nil }, nil
	}
	return v.engine.FromRules([]string{tag})
}

// CompileRules compiles AST rules into a validator function.
func (v *Validate) CompileRules(rules []types.Rule) func(any) error {
	return v.engine.CompileRules(rules)
}

// CheckTag compiles a tag and validates a single value.
func (v *Validate) CheckTag(tag string, value any) error {
	fn, err := v.FromTag(tag)
	if err != nil {
		return err
	}
	return fn(value)
}

// CheckRules compiles AST rules and validates a single value.
func (v *Validate) CheckRules(rules []types.Rule, value any) error {
	return v.engine.CompileRules(rules)(value)
}

// Struct returns a struct validator bound to this Validate's engine.
func (v *Validate) Struct() *structvalidator.StructValidator {
	return structvalidator.NewStructValidator((*core.Validate)(v.engine))
}

// ValidateStruct validates a struct using `validate` tags with defaults.
func (v *Validate) ValidateStruct(s any) error {
	return v.Struct().ValidateStruct(s)
}

// ValidateStructWithOpts validates a struct with advanced options.
func (v *Validate) ValidateStructWithOpts(
	s any, opts core.ValidateOpts,
) error {
	return v.Struct().ValidateStructWithOpts(s, opts)
}

// String returns a string validator builder.
func (v *Validate) String() *StringBuilder {
	return &StringBuilder{
		rules:  []types.Rule{types.NewRule(types.KString, nil)},
		engine: v.engine,
	}
}

// Int returns an integer validator builder.
func (v *Validate) Int() *IntBuilder {
	return NewIntBuilder(false, v.engine)
}

// Int64 returns an int64 validator builder.
func (v *Validate) Int64() *IntBuilder {
	return NewIntBuilder(true, v.engine)
}

// Bool returns a boolean validator builder.
func (v *Validate) Bool() *BoolBuilder {
	return NewBoolBuilder(v.engine)
}

// Slice returns a slice validator builder.
func (v *Validate) Slice() *SliceBuilder {
	// For now, we'll create a basic slice builder
	// In a full implementation, you'd have a SliceValidators type
	return &SliceBuilder{
		engine: v.engine,
		rules:  []types.Rule{types.NewRule(types.KSlice, nil)},
	}
}

// CustomType returns a custom type validator builder for the given type name.
// The type must be registered using types.RegisterGlobalType before use.
func (v *Validate) CustomType(typeName string) *CustomTypeBuilder {
	return &CustomTypeBuilder{
		engine:   v.engine,
		typeName: typeName,
		rules:    []types.Rule{types.NewRule(types.Kind(typeName), nil)},
	}
}
