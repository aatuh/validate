package validate

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// SliceValidator defines a function that validates a slice.
type SliceValidator func([]any) error

// WithSlice applies a series of SliceValidators to a value.
func (v *Validate) WithSlice(validators ...SliceValidator) func(any) error {
	return func(value any) error {
		s, err := v.toSlice(value)
		if err != nil {
			return err
		}
		for _, validator := range validators {
			if err := validator(s); err != nil {
				return err
			}
		}
		return nil
	}
}

// SliceLength returns a validator that ensures the slice has exactly n elements.
func (v *Validate) SliceLength(n int) SliceValidator {
	return func(s []any) error {
		if len(s) != n {
			return errors.New(v.translate("slice.length", n))
		}
		return nil
	}
}

// MinSliceLength returns a validator that ensures the slice has at least n elements.
func (v *Validate) MinSliceLength(n int) SliceValidator {
	return func(s []any) error {
		if len(s) < n {
			return errors.New(v.translate("slice.min", n))
		}
		return nil
	}
}

// MaxSliceLength returns a validator that ensures the slice has at most n elements.
func (v *Validate) MaxSliceLength(n int) SliceValidator {
	return func(s []any) error {
		if len(s) > n {
			return errors.New(v.translate("slice.max", n))
		}
		return nil
	}
}

// ForEach applies an element validator to every element in the slice.
func (v *Validate) ForEach(elementValidator func(any) error) SliceValidator {
	return func(s []any) error {
		for idx, elem := range s {
			if err := elementValidator(elem); err != nil {
				return errors.New(v.translate("slice.element", idx, err.Error()))
			}
		}
		return nil
	}
}

// BuildSliceValidator builds a composite slice validator from a list of rules.
// Expected tag format: "slice;min=1;max=5"
func BuildSliceValidator(v *Validate, rules []string) (func(any) error, error) {
	var validators []SliceValidator
	// rules[0] is "slice"
	for _, rule := range rules[1:] {
		parts := strings.SplitN(rule, "=", 2)
		key := strings.ToLower(parts[0])
		var param string
		if len(parts) == 2 {
			param = parts[1]
		}
		switch key {
		case "len":
			n, err := strconv.ParseInt(param, 10, 64)
			if err != nil {
				return nil, fmt.Errorf(v.translate("slice.invalidLenParameter")+": %w", err)
			}
			validators = append(validators, v.SliceLength(int(n)))
		case "min":
			n, err := strconv.ParseInt(param, 10, 64)
			if err != nil {
				return nil, fmt.Errorf(v.translate("slice.invalidMinParameter")+": %w", err)
			}
			validators = append(validators, v.MinSliceLength(int(n)))
		case "max":
			n, err := strconv.ParseInt(param, 10, 64)
			if err != nil {
				return nil, fmt.Errorf(v.translate("slice.invalidMaxParameter")+": %w", err)
			}
			validators = append(validators, v.MaxSliceLength(int(n)))
		default:
			return nil, fmt.Errorf(v.translate("slice.unknownValidator"), key)
		}
	}
	return v.WithSlice(validators...), nil
}

func (v *Validate) toSlice(value any) ([]any, error) {
	if s, ok := value.([]any); ok {
		return s, nil
	}
	return nil, errors.New(v.translate("slice.notSlice"))
}
