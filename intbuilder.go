package validate

// IntValidatorBuilder provides a fluent interface for building integer validators.
type IntValidatorBuilder struct {
	v          *Validate
	validators []IntValidator
}

// IntBuilder returns a new IntValidatorBuilder using the given Validate instance.
func (v *Validate) IntBuilder() *IntValidatorBuilder {
	return &IntValidatorBuilder{
		v:          v,
		validators: []IntValidator{},
	}
}

// WithMin adds a minimum value requirement.
func (b *IntValidatorBuilder) WithMin(min int64) *IntValidatorBuilder {
	b.validators = append(b.validators, b.v.MinInt(min))
	return b
}

// WithMax adds a maximum value requirement.
func (b *IntValidatorBuilder) WithMax(max int64) *IntValidatorBuilder {
	b.validators = append(b.validators, b.v.MaxInt(max))
	return b
}

// Build composes all added integer validators into one.
func (b *IntValidatorBuilder) Build() func(any) error {
	return b.v.WithInt(b.validators...)
}
