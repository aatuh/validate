package validators

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/aatuh/validate/translator"
)

// SliceValidator validates a slice presented as []any.
type SliceValidator func([]any) error

type SliceValidators struct {
	translator translator.Translator
}

func NewSliceValidators(
	t translator.Translator,
) *SliceValidators {
	return &SliceValidators{translator: t}
}

// WithSlice applies validators to any slice type via reflection.
func (sv *SliceValidators) WithSlice(
	validators ...SliceValidator,
) func(any) error {
	return func(value any) error {
		s, err := sv.toSlice(value)
		if err != nil {
			return err
		}
		for _, v := range validators {
			if err := v(s); err != nil {
				return err
			}
		}
		return nil
	}
}

func (sv *SliceValidators) SliceLength(n int) SliceValidator {
	return func(s []any) error {
		if len(s) != n {
			return errors.New(sv.translate("slice.length", n))
		}
		return nil
	}
}

func (sv *SliceValidators) MinSliceLength(n int) SliceValidator {
	return func(s []any) error {
		if len(s) < n {
			return errors.New(sv.translate("slice.min", n))
		}
		return nil
	}
}

func (sv *SliceValidators) MaxSliceLength(n int) SliceValidator {
	return func(s []any) error {
		if len(s) > n {
			return errors.New(sv.translate("slice.max", n))
		}
		return nil
	}
}

// ForEach applies an element validator to every element in the slice.
func (sv *SliceValidators) ForEach(
	elementValidator func(any) error,
) SliceValidator {
	return func(s []any) error {
		for idx, elem := range s {
			if err := elementValidator(elem); err != nil {
				return errors.New(sv.translate(
					"slice.element", idx, err.Error(),
				))
			}
		}
		return nil
	}
}

// BuildSliceValidator builds a composite slice validator from tag tokens.
// Example: "slice;min=1;max=5".
func BuildSliceValidator(
	sv *SliceValidators, rules []string,
) (func(any) error, error) {
	var fns []SliceValidator
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
				return nil, fmt.Errorf(
					"%s: %w",
					sv.translate("slice.invalidLenParameter"), err,
				)
			}
			fns = append(fns, sv.SliceLength(int(n)))
		case "min":
			n, err := strconv.ParseInt(param, 10, 64)
			if err != nil {
				return nil, fmt.Errorf(
					"%s: %w",
					sv.translate("slice.invalidMinParameter"), err,
				)
			}
			fns = append(fns, sv.MinSliceLength(int(n)))
		case "max":
			n, err := strconv.ParseInt(param, 10, 64)
			if err != nil {
				return nil, fmt.Errorf(
					"%s: %w",
					sv.translate("slice.invalidMaxParameter"), err,
				)
			}
			fns = append(fns, sv.MaxSliceLength(int(n)))
		default:
			return nil, errors.New(
				sv.translate("slice.unknownValidator", key),
			)
		}
	}
	return sv.WithSlice(fns...), nil
}

// toSlice converts any slice/array to []any for uniform validation.
func (sv *SliceValidators) toSlice(value any) ([]any, error) {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		n := rv.Len()
		out := make([]any, n)
		for i := 0; i < n; i++ {
			out[i] = rv.Index(i).Interface()
		}
		return out, nil
	default:
		return nil, errors.New(sv.translate("slice.notSlice"))
	}
}

func (sv *SliceValidators) translate(
	key string, params ...any,
) string {
	if sv.translator != nil {
		return sv.translator.T(key, params...)
	}
	if len(params) == 0 {
		return key
	}
	return key + ": " + fmt.Sprint(params...)
}
