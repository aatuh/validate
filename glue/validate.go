package glue

import (
	"context"

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

// WithRuleCompiler returns a copy with a per-instance custom rule compiler.
func (v *Validate) WithRuleCompiler(
	kind types.Kind, rc types.RuleCompiler,
) *Validate {
	return &Validate{
		engine: v.engine.WithRuleCompiler(kind, rc),
	}
}

// WithContextRuleCompiler returns a copy with a per-instance context-aware
// custom rule compiler.
func (v *Validate) WithContextRuleCompiler(
	kind types.Kind, rc types.ContextRuleCompiler,
) *Validate {
	return &Validate{
		engine: v.engine.WithContextRuleCompiler(kind, rc),
	}
}

// WithStructRuleCompiler returns a copy with a per-instance struct rule compiler.
func (v *Validate) WithStructRuleCompiler(
	kind types.Kind, compiler core.StructRuleCompiler,
) *Validate {
	return &Validate{
		engine: v.engine.WithStructRuleCompiler(kind, compiler),
	}
}

// WithTypeValidator returns a copy with a per-instance custom type validator.
func (v *Validate) WithTypeValidator(
	name string, factory types.TypeValidatorFactory,
) *Validate {
	return &Validate{
		engine: v.engine.WithTypeValidator(name, factory),
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

// FromRulesWithOpts creates a validator function from rule tokens with options.
func (v *Validate) FromRulesWithOpts(
	rules []string,
	opts types.CompileOpts,
) (func(any) error, error) {
	return v.engine.FromRulesWithOpts(rules, opts)
}

// FromRulesContext creates a context-aware validator from rule tokens.
func (v *Validate) FromRulesContext(
	rules []string,
) (types.ContextValidatorFunc, error) {
	return v.engine.FromRulesContext(rules)
}

// FromRulesContextWithOpts creates a context-aware validator from rule tokens
// with options.
func (v *Validate) FromRulesContextWithOpts(
	rules []string,
	opts types.CompileOpts,
) (types.ContextValidatorFunc, error) {
	return v.engine.FromRulesContextWithOpts(rules, opts)
}

// FromTag compiles a single tag string into a validator function.
func (v *Validate) FromTag(tag string) (func(any) error, error) {
	if tag == "" {
		return func(any) error { return nil }, nil
	}
	return v.engine.FromRules([]string{tag})
}

// FromTagWithOpts compiles a single tag string into a validator function with options.
func (v *Validate) FromTagWithOpts(tag string, opts types.CompileOpts) (func(any) error, error) {
	if tag == "" {
		return func(any) error { return nil }, nil
	}
	return v.engine.FromRulesWithOpts([]string{tag}, opts)
}

// FromTagContext compiles a single tag string into a context-aware validator.
func (v *Validate) FromTagContext(tag string) (types.ContextValidatorFunc, error) {
	if tag == "" {
		return func(context.Context, any) error { return nil }, nil
	}
	return v.engine.FromRulesContext([]string{tag})
}

// FromTagContextWithOpts compiles a single tag string into a context-aware
// validator with options.
func (v *Validate) FromTagContextWithOpts(tag string, opts types.CompileOpts) (types.ContextValidatorFunc, error) {
	if tag == "" {
		return func(context.Context, any) error { return nil }, nil
	}
	return v.engine.FromRulesContextWithOpts([]string{tag}, opts)
}

// CompileRules compiles AST rules into a validator function.
func (v *Validate) CompileRules(rules []types.Rule) func(any) error {
	return v.engine.CompileRules(rules)
}

// CompileRulesE compiles AST rules into a validator function and returns
// compile-time custom-rule errors.
func (v *Validate) CompileRulesE(rules []types.Rule) (func(any) error, error) {
	return v.engine.CompileRulesE(rules)
}

// CompileRulesWithOpts compiles AST rules into a validator function with options.
func (v *Validate) CompileRulesWithOpts(rules []types.Rule, opts types.CompileOpts) func(any) error {
	return v.engine.CompileRulesWithOpts(rules, opts)
}

// CompileRulesWithOptsE compiles AST rules with options and returns compile errors.
func (v *Validate) CompileRulesWithOptsE(rules []types.Rule, opts types.CompileOpts) (func(any) error, error) {
	return v.engine.CompileRulesWithOptsE(rules, opts)
}

// CompileRulesContext compiles AST rules into a context-aware validator.
func (v *Validate) CompileRulesContext(rules []types.Rule) types.ContextValidatorFunc {
	return v.engine.CompileRulesContext(rules)
}

// CompileRulesContextE compiles AST rules into a context-aware validator and
// returns compile errors.
func (v *Validate) CompileRulesContextE(rules []types.Rule) (types.ContextValidatorFunc, error) {
	return v.engine.CompileRulesContextE(rules)
}

// CompileRulesContextWithOpts compiles AST rules into a context-aware validator
// with options.
func (v *Validate) CompileRulesContextWithOpts(rules []types.Rule, opts types.CompileOpts) types.ContextValidatorFunc {
	return v.engine.CompileRulesContextWithOpts(rules, opts)
}

// CompileRulesContextWithOptsE compiles AST rules into a context-aware validator
// with options and returns compile errors.
func (v *Validate) CompileRulesContextWithOptsE(rules []types.Rule, opts types.CompileOpts) (types.ContextValidatorFunc, error) {
	return v.engine.CompileRulesContextWithOptsE(rules, opts)
}

// CheckTag compiles a tag and validates a single value.
func (v *Validate) CheckTag(tag string, value any) error {
	fn, err := v.FromTag(tag)
	if err != nil {
		return err
	}
	return fn(value)
}

// CheckTagWithOpts compiles a tag with options and validates a single value.
func (v *Validate) CheckTagWithOpts(tag string, value any, opts types.CompileOpts) error {
	fn, err := v.FromTagWithOpts(tag, opts)
	if err != nil {
		return err
	}
	return fn(value)
}

// CheckTagContext compiles a tag and validates a single value with context.
func (v *Validate) CheckTagContext(ctx context.Context, tag string, value any) error {
	fn, err := v.FromTagContext(tag)
	if err != nil {
		return err
	}
	return fn(ctx, value)
}

// CheckTagContextWithOpts compiles a tag with options and validates a single
// value with context.
func (v *Validate) CheckTagContextWithOpts(ctx context.Context, tag string, value any, opts types.CompileOpts) error {
	fn, err := v.FromTagContextWithOpts(tag, opts)
	if err != nil {
		return err
	}
	return fn(ctx, value)
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

// ValidateStructContext validates a struct using `validate` tags with context.
func (v *Validate) ValidateStructContext(ctx context.Context, s any) error {
	return v.Struct().ValidateStructContext(ctx, s)
}

// ValidateStructWithOpts validates a struct with advanced options.
func (v *Validate) ValidateStructWithOpts(
	s any, opts core.ValidateOpts,
) error {
	return v.Struct().ValidateStructWithOpts(s, opts)
}

// ValidateStructContextWithOpts validates a struct with context and advanced options.
func (v *Validate) ValidateStructContextWithOpts(
	ctx context.Context, s any, opts core.ValidateOpts,
) error {
	return v.Struct().ValidateStructContextWithOpts(ctx, s, opts)
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

// Float returns a floating-point validator builder.
func (v *Validate) Float() *FloatBuilder {
	return NewFloatBuilder(v.engine)
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

// Array returns an array validator builder.
func (v *Validate) Array() *ArrayBuilder {
	return NewArrayBuilder(v.engine)
}

// Map returns a map validator builder.
func (v *Validate) Map() *MapBuilder {
	return NewMapBuilder(v.engine)
}

// Time returns a time.Time validator builder.
func (v *Validate) Time() *TimeBuilder {
	return NewTimeBuilder(v.engine)
}

// CustomType returns a custom type validator builder for the given type name.
// The type must be registered with WithTypeValidator or types.RegisterGlobalType before use.
func (v *Validate) CustomType(typeName string) *CustomTypeBuilder {
	return &CustomTypeBuilder{
		engine:   v.engine,
		typeName: typeName,
		rules:    []types.Rule{types.NewRule(types.Kind(typeName), nil)},
	}
}
