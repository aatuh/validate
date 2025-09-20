package validators

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"

	"github.com/aatuh/validate/translator"
)

// StringValidator defines a function that validates a string.
type StringValidator func(string) error

// StringValidators provides string validation methods.
type StringValidators struct {
	translator translator.Translator
}

// NewStringValidators creates a new StringValidators instance.
func NewStringValidators(
	t translator.Translator,
) *StringValidators {
	return &StringValidators{translator: t}
}

// WithString applies a series of string validators to a value.
func (sv *StringValidators) WithString(
	validators ...StringValidator,
) func(any) error {
	return func(value any) error {
		s, err := sv.toString(value)
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

// Length returns a validator that checks for exact length.
func (sv *StringValidators) Length(n int) StringValidator {
	return func(s string) error {
		if len(s) != n {
			return errors.New(sv.translate("string.length", n))
		}
		return nil
	}
}

func (sv *StringValidators) MinLength(n int) StringValidator {
	return func(s string) error {
		if len(s) < n {
			return errors.New(sv.translate("string.minLength", n))
		}
		return nil
	}
}

func (sv *StringValidators) MaxLength(n int) StringValidator {
	return func(s string) error {
		if len(s) > n {
			return errors.New(sv.translate("string.maxLength", n))
		}
		return nil
	}
}

func (sv *StringValidators) OneOf(
	values ...string,
) StringValidator {
	return func(s string) error {
		for _, val := range values {
			if strings.EqualFold(s, val) {
				return nil
			}
		}
		return fmt.Errorf("must be one of %s",
			strings.Join(values, ", "))
	}
}

func (sv *StringValidators) Email() StringValidator {
	return func(s string) error {
		if len(s) > 254 {
			return errors.New(sv.translate("string.maxLength", 254))
		}
		if _, err := mail.ParseAddress(s); err != nil {
			return errors.New(sv.translate("string.email.invalid"))
		}
		return nil
	}
}

// Regex returns a validator that ensures the string matches the pattern.
func (sv *StringValidators) Regex(pattern string) StringValidator {
	re, err := regexp.Compile(pattern)
	if err != nil {
		// Always fail if the pattern is invalid.
		return func(s string) error {
			return errors.New(
				sv.translate("string.regex.invalidPattern", pattern),
			)
		}
	}
	return func(s string) error {
		if !re.MatchString(s) {
			return errors.New(
				sv.translate("string.regex.noMatch", pattern),
			)
		}
		return nil
	}
}

// BuildStringValidator builds a composite string validator from tokens.
// Expected tag: "string;min=3;max=10;regex=^a.*z$".
func BuildStringValidator(
	sv *StringValidators, rules []string,
) (func(any) error, error) {
	var fns []StringValidator
	for _, rule := range rules[1:] {
		parts := strings.SplitN(rule, "=", 2)
		key := strings.ToLower(parts[0])
		var param string
		if len(parts) == 2 {
			param = parts[1]
		}
		switch key {
		case "len":
			n, err := strconv.Atoi(param)
			if err != nil {
				return nil, fmt.Errorf(
					"invalid parameter for len: %w", err)
			}
			fns = append(fns, sv.Length(n))
		case "min":
			n, err := strconv.Atoi(param)
			if err != nil {
				return nil, fmt.Errorf(
					"invalid parameter for min: %w", err)
			}
			fns = append(fns, sv.MinLength(n))
		case "max":
			n, err := strconv.Atoi(param)
			if err != nil {
				return nil, fmt.Errorf(
					"invalid parameter for max: %w", err)
			}
			fns = append(fns, sv.MaxLength(n))
		case "oneof":
			opts := strings.Split(param, " ")
			fns = append(fns, sv.OneOf(opts...))
		case "email":
			fns = append(fns, sv.Email())
		case "regex":
			fns = append(fns, sv.Regex(param))
		default:
			return nil, fmt.Errorf(
				"unknown string validator: %s", key)
		}
	}
	return sv.WithString(fns...), nil
}

func (sv *StringValidators) toString(
	value any,
) (string, error) {
	if s, ok := value.(string); ok {
		return s, nil
	}
	if stringer, ok := value.(fmt.Stringer); ok {
		return stringer.String(), nil
	}
	return "", errors.New("cannot convert value to string")
}

func (sv *StringValidators) translate(
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
