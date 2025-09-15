package validate

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// IntValidator defines a function that validates an int64.
type IntValidator func(int64) error

// WithInt applies a series of IntValidators to a value (converting it via toInt64).
func (v *Validate) WithInt(validators ...IntValidator) func(any) error {
	return func(value any) error {
		i, err := v.toInt64(value)
		if err != nil {
			return err
		}
		for _, validator := range validators {
			if err := validator(i); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithInt64 applies a series of IntValidators assuming the value is exactly an int64.
func (v *Validate) WithInt64(validators ...IntValidator) func(any) error {
	return func(value any) error {
		i, err := v.toExplicitInt64(value)
		if err != nil {
			return err
		}
		for _, validator := range validators {
			if err := validator(i); err != nil {
				return err
			}
		}
		return nil
	}
}

// MinInt returns a validator that checks if an integer is at least min.
func (v *Validate) MinInt(min int64) IntValidator {
	return func(i int64) error {
		if i < min {
			return errors.New(v.translate("int.min", min))
		}
		return nil
	}
}

// MaxInt returns a validator that checks if an integer is at most max.
func (v *Validate) MaxInt(max int64) IntValidator {
	return func(i int64) error {
		if i > max {
			return errors.New(v.translate("int.max", max))
		}
		return nil
	}
}

// buildIntValidator builds an integer validator from a list of rules.
// rules[0] is expected to be "int" or "int64".
func buildIntValidator(
	v *Validate, rules []string, intType string,
) (func(any) error, error) {
	var validators []IntValidator
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
				return nil, fmt.Errorf(v.translate("int.invalidMinParameter")+": %w", err)
			}
			validators = append(validators, v.MinInt(n))
		case "max":
			n, err := strconv.ParseInt(param, 10, 64)
			if err != nil {
				return nil, fmt.Errorf(v.translate("int.invalidMaxParameter")+": %w", err)
			}
			validators = append(validators, v.MaxInt(n))
		default:
			if intType == "int64" {
				return nil, fmt.Errorf(v.translate("int.unknownInt64Validator"), key)
			}
			return nil, fmt.Errorf(v.translate("int.unknownIntValidator"), key)
		}
	}
	if intType == "int64" {
		return v.WithInt64(validators...), nil
	}
	return v.WithInt(validators...), nil
}

func (v *Validate) toInt64(value any) (int64, error) {
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
		return 0, errors.New(v.translate("int.notInteger"))
	}
}

func (v *Validate) toExplicitInt64(value any) (int64, error) {
	if val, ok := value.(int64); ok {
		return val, nil
	}
	return 0, errors.New(v.translate("int.notInt64"))
}
