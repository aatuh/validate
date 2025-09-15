package validate

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
)

// StringValidator defines a function that validates a string.
type StringValidator func(string) error

// WithString applies a series of string validators to a value.
func (v *Validate) WithString(validators ...StringValidator) func(any) error {
	return func(value any) error {
		s, err := v.toString(value)
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

// Length returns a validator that checks if a string is exactly n characters
// long.
func (v *Validate) Length(n int) StringValidator {
	return func(s string) error {
		if len(s) != n {
			return errors.New(v.translate("string.length", n))
		}
		return nil
	}
}

// MinLength returns a validator that ensures a string has at least n
// characters.
func (v *Validate) MinLength(n int) StringValidator {
	return func(s string) error {
		if len(s) < n {
			return errors.New(v.translate("string.minLength", n))
		}
		return nil
	}
}

// MaxLength returns a validator that ensures a string has at most n characters.
func (v *Validate) MaxLength(n int) StringValidator {
	return func(s string) error {
		if len(s) > n {
			return fmt.Errorf(v.translate("string.maxLength", n))
		}
		return nil
	}
}

// OneOf returns a validator that checks if a string is one of the allowed
// values.
func (v *Validate) OneOf(values ...string) StringValidator {
	return func(s string) error {
		for _, val := range values {
			if strings.EqualFold(s, val) {
				return nil
			}
		}
		return fmt.Errorf("must be one of %s", strings.Join(values, ", "))
	}
}

// Email returns a validator that checks if a string is a valid email address.
func (v *Validate) Email() StringValidator {
	return func(s string) error {
		if len(s) > 254 {
			return errors.New(v.translate("string.maxLength", 254))
		}
		if _, err := mail.ParseAddress(s); err != nil {
			return errors.New(v.translate("string.email.invalid"))
		}
		return nil
	}
}

// Regex returns a validator that ensures the string matches the given regex
// pattern.
func (v *Validate) Regex(pattern string) StringValidator {
	re, err := regexp.Compile(pattern)
	if err != nil {
		// Always fail if the pattern is invalid.
		return func(s string) error {
			return errors.New(v.translate("string.regex.invalidPattern", pattern))
		}
	}
	return func(s string) error {
		if !re.MatchString(s) {
			return errors.New(v.translate("string.regex.noMatch", pattern))
		}
		return nil
	}
}

// BuildStringValidator builds a composite string validator from a list of
// rules. Expected tag format: "string;min=3;max=10;regex=^a.*z$"
func BuildStringValidator(v *Validate, rules []string) (func(any) error, error) {
	var validators []StringValidator
	// rules[0] is "string"
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
				return nil, fmt.Errorf("invalid parameter for len: %w", err)
			}
			validators = append(validators, v.Length(n))
		case "min":
			n, err := strconv.Atoi(param)
			if err != nil {
				return nil, fmt.Errorf("invalid parameter for min: %w", err)
			}
			validators = append(validators, v.MinLength(n))
		case "max":
			n, err := strconv.Atoi(param)
			if err != nil {
				return nil, fmt.Errorf("invalid parameter for max: %w", err)
			}
			validators = append(validators, v.MaxLength(n))
		case "oneof":
			opts := strings.Split(param, " ")
			validators = append(validators, v.OneOf(opts...))
		case "email":
			validators = append(validators, v.Email())
		case "regex":
			validators = append(validators, v.Regex(param))
		default:
			return nil, fmt.Errorf("unknown string validator: %s", key)
		}
	}
	return v.WithString(validators...), nil
}

func (v *Validate) toString(value any) (string, error) {
	if s, ok := value.(string); ok {
		return s, nil
	}
	if stringer, ok := value.(fmt.Stringer); ok {
		return stringer.String(), nil
	}
	return "", errors.New("cannot convert value to string")
}
