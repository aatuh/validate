package validators

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aatuh/validate/v3/translator"
)

// IntValidator defines a function that validates an int64.
//
// This type represents a validation function that takes an int64 value and
// returns an error if validation fails.
type IntValidator func(int64) error

// IntValidators provides integer validation methods.
//
// Fields:
//   - translator: Optional translator for localized error messages.
type IntValidators struct {
	translator translator.Translator
}

// NewIntValidators creates a new IntValidators instance.
//
// Parameters:
//   - t: Optional translator for localized error messages.
//
// Returns:
//   - *IntValidators: A new IntValidators instance.
func NewIntValidators(
	t translator.Translator,
) *IntValidators {
	return &IntValidators{translator: t}
}

// Translator returns the translator instance.
//
// Returns:
//   - translator.Translator: The translator instance.
func (iv *IntValidators) Translator() translator.Translator {
	return iv.translator
}

// WithInt applies IntValidators converting the value to int64 first.
//
// Parameters:
//   - validators: Variable number of integer validators to apply.
//
// Returns:
//   - func(any) error: A validator function that validates any value.
func (iv *IntValidators) WithInt(
	validators ...IntValidator,
) func(any) error {
	return func(value any) error {
		i, err := iv.toInt64(value)
		if err != nil {
			return err
		}
		for _, v := range validators {
			if err := v(i); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithInt64 applies IntValidators requiring exactly int64.
//
// Parameters:
//   - validators: Variable number of integer validators to apply.
//
// Returns:
//   - func(any) error: A validator function that validates int64 values.
func (iv *IntValidators) WithInt64(
	validators ...IntValidator,
) func(any) error {
	return func(value any) error {
		i, err := iv.toExplicitInt64(value)
		if err != nil {
			return err
		}
		for _, v := range validators {
			if err := v(i); err != nil {
				return err
			}
		}
		return nil
	}
}

// MinInt returns a validator that checks for minimum integer value.
//
// Parameters:
//   - min: The minimum integer value allowed.
//
// Returns:
//   - IntValidator: A validator function that checks minimum value.
func (iv *IntValidators) MinInt(min int64) IntValidator {
	return func(i int64) error {
		if i < min {
			return errors.New(iv.translate("int.min", min))
		}
		return nil
	}
}

// MaxInt returns a validator that checks for maximum integer value.
//
// Parameters:
//   - max: The maximum integer value allowed.
//
// Returns:
//   - IntValidator: A validator function that checks maximum value.
func (iv *IntValidators) MaxInt(max int64) IntValidator {
	return func(i int64) error {
		if i > max {
			return errors.New(iv.translate("int.max", max))
		}
		return nil
	}
}

// BuildIntValidator builds an integer validator from tokens.
// rules[0] == "int" or "int64".
//
// Parameters:
//   - iv: IntValidators instance for creating validators.
//   - rules: Slice of rule tokens.
//   - intType: The integer type ("int" or "int64").
//
// Returns:
//   - func(any) error: The compiled validator function.
//   - error: An error if the rules are invalid.
func BuildIntValidator(
	iv *IntValidators, rules []string, intType string,
) (func(any) error, error) {
	var fns []IntValidator
	for _, rule := range rules[1:] {
		parts := strings.SplitN(rule, "=", 2)
		key := strings.ToLower(parts[0])
		var param string
		if len(parts) == 2 {
			param = parts[1]
		}
		switch key {
		case "min":
			n, err := strconv.ParseInt(param, 10, 64)
			if err != nil {
				return nil, fmt.Errorf(
					"%s: %w",
					iv.translate("int.invalidMinParameter"), err,
				)
			}
			fns = append(fns, iv.MinInt(n))
		case "max":
			n, err := strconv.ParseInt(param, 10, 64)
			if err != nil {
				return nil, fmt.Errorf(
					"%s: %w",
					iv.translate("int.invalidMaxParameter"), err,
				)
			}
			fns = append(fns, iv.MaxInt(n))
		default:
			if intType == "int64" {
				return nil, errors.New(
					iv.translate("int.unknownInt64Validator", key),
				)
			}
			return nil, errors.New(
				iv.translate("int.unknownIntValidator", key),
			)
		}
	}
	if intType == "int64" {
		return iv.WithInt64(fns...), nil
	}
	return iv.WithInt(fns...), nil
}

func (iv *IntValidators) toInt64(value any) (int64, error) {
	switch val := value.(type) {
	case int:
		return int64(val), nil
	case int8:
		return int64(val), nil
	case int16:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case int64:
		return val, nil
	default:
		return 0, errors.New(iv.translate("int.notInteger"))
	}
}

func (iv *IntValidators) toExplicitInt64(value any) (int64, error) {
	if val, ok := value.(int64); ok {
		return val, nil
	}
	return 0, errors.New(iv.translate("int.notInt64"))
}

func (iv *IntValidators) translate(
	key string, params ...any,
) string {
	if iv.translator != nil {
		return iv.translator.T(key, params...)
	}
	if len(params) == 0 {
		return key
	}
	return key + ": " + fmt.Sprint(params...)
}
