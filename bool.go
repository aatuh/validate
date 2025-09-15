package validate

import (
	"errors"
)

// BoolValidator is a function that validates a bool.
type BoolValidator func(b bool) error

// WithBool is a function that validates a bool.
func (v *Validate) WithBool(validators ...BoolValidator) func(value any) error {
	return func(value any) error {
		b, err := v.toBool(value)
		if err != nil {
			return err
		}
		for _, validator := range validators {
			if err := validator(b); err != nil {
				return err
			}
		}
		return nil
	}
}

func (v *Validate) toBool(value any) (bool, error) {
	b, ok := value.(bool)
	if !ok {
		return false, errors.New(v.translate("bool.notBool"))
	}
	return b, nil
}
