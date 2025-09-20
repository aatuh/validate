package validate

import "github.com/aatuh/validate/validators"

// StringBuilder accumulates string validation rules and can be finished
// with Build() or by calling a terminal rule (like Email()) which returns
// the final func(any) error.
type StringBuilder struct {
	sv    *validators.StringValidators
	rules []validators.StringValidator
}

func (b *StringBuilder) Length(n int) *StringBuilder {
	b.rules = append(b.rules, b.sv.Length(n))
	return b
}

func (b *StringBuilder) MinLength(n int) *StringBuilder {
	b.rules = append(b.rules, b.sv.MinLength(n))
	return b
}

func (b *StringBuilder) MaxLength(n int) *StringBuilder {
	b.rules = append(b.rules, b.sv.MaxLength(n))
	return b
}

func (b *StringBuilder) OneOf(vals ...string) *StringBuilder {
	b.rules = append(b.rules, b.sv.OneOf(vals...))
	return b
}

// Email is terminal for convenience in examples. It appends the rule
// and returns the finished validator func.
func (b *StringBuilder) Email() func(any) error {
	b.rules = append(b.rules, b.sv.Email())
	return b.sv.WithString(b.rules...)
}

// Regex is not terminal, allowing more chaining. Call Build() to finish.
func (b *StringBuilder) Regex(pat string) *StringBuilder {
	b.rules = append(b.rules, b.sv.Regex(pat))
	return b
}

// Build returns the composite validator func.
func (b *StringBuilder) Build() func(any) error {
	return b.sv.WithString(b.rules...)
}

// IntBuilder accumulates integer rules; Build() finishes it.
type IntBuilder struct {
	iv    *validators.IntValidators
	rules []validators.IntValidator
	exact bool // true => require exactly int64 at call time
}

func (b *IntBuilder) MinInt(n int64) *IntBuilder {
	b.rules = append(b.rules, b.iv.MinInt(n))
	return b
}

func (b *IntBuilder) MaxInt(n int64) *IntBuilder {
	b.rules = append(b.rules, b.iv.MaxInt(n))
	return b
}

// Build returns a func(any) error that validates the provided value.
func (b *IntBuilder) Build() func(any) error {
	if b.exact {
		return b.iv.WithInt64(b.rules...)
	}
	return b.iv.WithInt(b.rules...)
}

// SliceBuilder accumulates slice rules; Build() finishes it.
type SliceBuilder struct {
	sv    *validators.SliceValidators
	rules []validators.SliceValidator
}

func (b *SliceBuilder) Length(n int) *SliceBuilder {
	b.rules = append(b.rules, b.sv.SliceLength(n))
	return b
}

func (b *SliceBuilder) MinSliceLength(n int) *SliceBuilder {
	b.rules = append(b.rules, b.sv.MinSliceLength(n))
	return b
}

func (b *SliceBuilder) MaxSliceLength(n int) *SliceBuilder {
	b.rules = append(b.rules, b.sv.MaxSliceLength(n))
	return b
}

// ForEach applies an element validator to each item in the slice.
func (b *SliceBuilder) ForEach(
	elem func(any) error,
) *SliceBuilder {
	b.rules = append(b.rules, b.sv.ForEach(elem))
	return b
}

func (b *SliceBuilder) Build() func(any) error {
	return b.sv.WithSlice(b.rules...)
}
