package validators

import (
	"errors"
	"fmt"

	"github.com/aatuh/validate/translator"
)

// BoolValidator is a function that validates a bool.
type BoolValidator func(b bool) error

// BoolValidators provides boolean validation methods.
type BoolValidators struct {
	translator translator.Translator
}

// NewBoolValidators creates a new BoolValidators instance.
func NewBoolValidators(t translator.Translator) *BoolValidators {
	return &BoolValidators{translator: t}
}

// WithBool is a function that validates a bool.
func (bv *BoolValidators) WithBool(validators ...BoolValidator) func(value any) error {
	return func(value any) error {
		b, err := bv.toBool(value)
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

func (bv *BoolValidators) toBool(value any) (bool, error) {
	b, ok := value.(bool)
	if !ok {
		return false, errors.New(bv.translate("bool.notBool"))
	}
	return b, nil
}

func (bv *BoolValidators) translate(key string, params ...any) string {
	if bv.translator != nil {
		return bv.translator.T(key, params...)
	}
	// Default behavior without a translator: avoid printf-style formatting
	// so vet does not treat this as a printf wrapper. Concatenate instead.
	if len(params) == 0 {
		return key
	}
	return key + ": " + fmt.Sprint(params...)
}
