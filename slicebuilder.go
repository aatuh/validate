package validate

// SliceValidatorBuilder provides a fluent interface for building slice validators.
type SliceValidatorBuilder struct {
	v          *Validate
	validators []SliceValidator
}

// SliceBuilder returns a new SliceValidatorBuilder using the given Validate instance.
func (v *Validate) SliceBuilder() *SliceValidatorBuilder {
	return &SliceValidatorBuilder{
		v:          v,
		validators: []SliceValidator{},
	}
}

// WithLength adds a validator that ensures the slice has exactly n elements.
func (b *SliceValidatorBuilder) WithLength(n int) *SliceValidatorBuilder {
	b.validators = append(b.validators, b.v.SliceLength(n))
	return b
}

// WithMinLength adds a validator that ensures the slice has at least n elements.
func (b *SliceValidatorBuilder) WithMinLength(n int) *SliceValidatorBuilder {
	b.validators = append(b.validators, b.v.MinSliceLength(n))
	return b
}

// WithMaxLength adds a validator that ensures the slice has at most n elements.
func (b *SliceValidatorBuilder) WithMaxLength(n int) *SliceValidatorBuilder {
	b.validators = append(b.validators, b.v.MaxSliceLength(n))
	return b
}

// ForEach adds a validator that applies the given elementValidator to every element in the slice.
func (b *SliceValidatorBuilder) ForEach(elementValidator func(any) error) *SliceValidatorBuilder {
	b.validators = append(b.validators, b.v.ForEach(elementValidator))
	return b
}

// Build composes all added slice validators into one.
func (b *SliceValidatorBuilder) Build() func(any) error {
	return b.v.WithSlice(b.validators...)
}
