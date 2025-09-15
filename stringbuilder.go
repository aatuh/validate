package validate

// StringValidatorBuilder provides a fluent interface for building string validators.
type StringValidatorBuilder struct {
	v          *Validate
	validators []StringValidator
}

// StringBuilder returns a new StringValidatorBuilder using the given Validate instance.
func (v *Validate) StringBuilder() *StringValidatorBuilder {
	return &StringValidatorBuilder{
		v:          v,
		validators: []StringValidator{},
	}
}

// WithMin adds a minimum length requirement.
func (b *StringValidatorBuilder) WithMin(n int) *StringValidatorBuilder {
	b.validators = append(b.validators, b.v.MinLength(n))
	return b
}

// WithMax adds a maximum length requirement.
func (b *StringValidatorBuilder) WithMax(n int) *StringValidatorBuilder {
	b.validators = append(b.validators, b.v.MaxLength(n))
	return b
}

// WithLength adds an exact length requirement.
func (b *StringValidatorBuilder) WithLength(n int) *StringValidatorBuilder {
	b.validators = append(b.validators, b.v.Length(n))
	return b
}

// WithRegex adds a regex pattern requirement.
func (b *StringValidatorBuilder) WithRegex(pattern string) *StringValidatorBuilder {
	b.validators = append(b.validators, b.v.Regex(pattern))
	return b
}

// WithEmail adds an email format requirement.
func (b *StringValidatorBuilder) WithEmail() *StringValidatorBuilder {
	b.validators = append(b.validators, b.v.Email())
	return b
}

// WithOneOf adds an allowed values requirement.
func (b *StringValidatorBuilder) WithOneOf(values ...string) *StringValidatorBuilder {
	b.validators = append(b.validators, b.v.OneOf(values...))
	return b
}

// Build composes all added validators into one.
func (b *StringValidatorBuilder) Build() func(any) error {
	return b.v.WithString(b.validators...)
}
