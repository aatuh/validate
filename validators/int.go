package validators

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aatuh/validate/translator"
)

// IntValidator defines a function that validates an int64.
type IntValidator func(int64) error

// IntValidators provides integer validation methods.
type IntValidators struct {
	translator translator.Translator
}

func NewIntValidators(
	t translator.Translator,
) *IntValidators {
	return &IntValidators{translator: t}
}

// WithInt applies IntValidators converting the value to int64 first.
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

func (iv *IntValidators) MinInt(min int64) IntValidator {
	return func(i int64) error {
		if i < min {
			return errors.New(iv.translate("int.min", min))
		}
		return nil
	}
}

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
