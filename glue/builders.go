package glue

import (
	"github.com/aatuh/validate/v3/core"
	"github.com/aatuh/validate/v3/types"
)

// StringBuilder accumulates string validation rules.
type StringBuilder struct {
	rules  []types.Rule
	engine *core.Engine
}

func (b *StringBuilder) Length(n int) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KLength, map[string]any{"n": int64(n)}))
	return b
}

func (b *StringBuilder) MinLength(n int) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMinLength, map[string]any{"n": int64(n)}))
	return b
}

func (b *StringBuilder) MaxLength(n int) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMaxLength, map[string]any{"n": int64(n)}))
	return b
}

func (b *StringBuilder) OneOf(vals ...string) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KOneOf, map[string]any{"values": vals}))
	return b
}

func (b *StringBuilder) MinRunes(n int) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMinRunes, map[string]any{"n": int64(n)}))
	return b
}

func (b *StringBuilder) MaxRunes(n int) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMaxRunes, map[string]any{"n": int64(n)}))
	return b
}

func (b *StringBuilder) Regex(pat string) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KRegex, map[string]any{"pattern": pat}))
	return b
}

func (b *StringBuilder) OmitEmpty() *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KOmitempty, nil))
	return b
}

func (b *StringBuilder) Build() func(any) error {
	return b.engine.CompileRules(b.rules)
}

// IntBuilder accumulates integer validation rules.
type IntBuilder struct {
	rules  []types.Rule
	exact  bool
	engine *core.Engine
}

// NewIntBuilder creates a new IntBuilder with the base type rule.
func NewIntBuilder(exact bool, engine *core.Engine) *IntBuilder {
	builder := &IntBuilder{
		rules:  []types.Rule{},
		exact:  exact,
		engine: engine,
	}

	// Set the base type rule
	if exact {
		builder.rules = append(builder.rules, types.NewRule(types.KInt64, map[string]any{}))
	} else {
		builder.rules = append(builder.rules, types.NewRule(types.KInt, map[string]any{}))
	}

	return builder
}

func (b *IntBuilder) MinInt(n int64) *IntBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMinInt, map[string]any{"n": n}))
	return b
}

func (b *IntBuilder) MaxInt(n int64) *IntBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMaxInt, map[string]any{"n": n}))
	return b
}

func (b *IntBuilder) OmitEmpty() *IntBuilder {
	b.rules = append(b.rules, types.NewRule(types.KOmitempty, nil))
	return b
}

func (b *IntBuilder) Build() func(any) error {
	return b.engine.CompileRules(b.rules)
}

// BoolBuilder accumulates boolean validation rules.
type BoolBuilder struct {
	rules  []types.Rule
	engine *core.Engine
}

// NewBoolBuilder creates a new BoolBuilder with the base type rule.
func NewBoolBuilder(engine *core.Engine) *BoolBuilder {
	return &BoolBuilder{
		rules:  []types.Rule{types.NewRule(types.KBool, nil)},
		engine: engine,
	}
}

func (b *BoolBuilder) Build() func(any) error {
	return b.engine.CompileRules(b.rules)
}

func (b *BoolBuilder) OmitEmpty() *BoolBuilder {
	b.rules = append(b.rules, types.NewRule(types.KOmitempty, nil))
	return b
}

// SliceBuilder accumulates slice validation rules.
type SliceBuilder struct {
	engine *core.Engine
	rules  []types.Rule
}

func (b *SliceBuilder) Length(n int) *SliceBuilder {
	b.rules = append(b.rules, types.NewRule(types.KSliceLength, map[string]any{"n": int64(n)}))
	return b
}

func (b *SliceBuilder) MinLength(n int) *SliceBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMinSliceLength, map[string]any{"n": int64(n)}))
	return b
}

func (b *SliceBuilder) MaxLength(n int) *SliceBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMaxSliceLength, map[string]any{"n": int64(n)}))
	return b
}

func (b *SliceBuilder) ForEach(elemValidator func(any) error) *SliceBuilder {
	b.rules = append(b.rules, types.NewRule(types.KForEach, map[string]any{"validator": elemValidator}))
	return b
}

// ForEachRules applies inner rules to each slice element.
// This form is cache-friendly (no function args).
func (b *SliceBuilder) ForEachRules(inner ...types.Rule) *SliceBuilder {
	if len(inner) == 0 {
		return b
	}
	// Convert to []types.Rule slice for the compiler
	innerRules := make([]types.Rule, len(inner))
	copy(innerRules, inner)
	r := types.NewRule(types.KForEach, map[string]any{"rules": innerRules})
	b.rules = append(b.rules, r)
	return b
}

// ForEachStringBuilder copies rules from a StringBuilder as element rules.
func (b *SliceBuilder) ForEachStringBuilder(sb *StringBuilder) *SliceBuilder {
	if sb == nil {
		return b
	}
	cp := append([]types.Rule(nil), sb.rules...)
	return b.ForEachRules(cp...)
}

func (b *SliceBuilder) OmitEmpty() *SliceBuilder {
	b.rules = append(b.rules, types.NewRule(types.KOmitempty, nil))
	return b
}

func (b *SliceBuilder) Build() func(any) error {
	return b.engine.CompileRules(b.rules)
}

// CustomTypeBuilder accumulates custom type validation rules.
type CustomTypeBuilder struct {
	engine   *core.Engine
	typeName string
	rules    []types.Rule
}

func (b *CustomTypeBuilder) Build() func(any) error {
	return b.engine.CompileRules(b.rules)
}

func (b *CustomTypeBuilder) OmitEmpty() *CustomTypeBuilder {
	b.rules = append(b.rules, types.NewRule(types.KOmitempty, nil))
	return b
}
