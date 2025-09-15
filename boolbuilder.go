package validate

import "fmt"

// BoolValidatorBuilder provides a fluent interface for building bool validators.
type BoolValidatorBuilder struct {
	v          *Validate
	validators []BoolValidator
}

// BoolBuilder returns a new BoolValidatorBuilder using the given Validate instance.
func (v *Validate) BoolBuilder() *BoolValidatorBuilder {
	return &BoolValidatorBuilder{
		v:          v,
		validators: []BoolValidator{},
	}
}

// MustBeTrue adds a validator that ensures the bool is true.
func (b *BoolValidatorBuilder) MustBeTrue() *BoolValidatorBuilder {
	b.validators = append(b.validators, func(val bool) error {
		if !val {
			return fmt.Errorf("must be true")
		}
		return nil
	})
	return b
}

// MustBeFalse adds a validator that ensures the bool is false.
func (b *BoolValidatorBuilder) MustBeFalse() *BoolValidatorBuilder {
	b.validators = append(b.validators, func(val bool) error {
		if val {
			return fmt.Errorf("must be false")
		}
		return nil
	})
	return b
}

// Build composes all added bool validators into one.
func (b *BoolValidatorBuilder) Build() func(any) error {
	return b.v.WithBool(b.validators...)
}
