package validators

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/aatuh/validate/v3/translator"
)

// StringValidator defines a function that validates a string.
//
// This type represents a validation function that takes a string value and
// returns an error if validation fails.
type StringValidator func(string) error

// StringValidators provides string validation methods.
//
// Fields:
//   - translator: Optional translator for localized error messages.
type StringValidators struct {
	translator translator.Translator
}

// NewStringValidators creates a new StringValidators instance.
//
// Parameters:
//   - t: Optional translator for localized error messages.
//
// Returns:
//   - *StringValidators: A new StringValidators instance.
func NewStringValidators(
	t translator.Translator,
) *StringValidators {
	return &StringValidators{translator: t}
}

// Translator returns the translator instance.
//
// Returns:
//   - translator.Translator: The translator instance.
func (sv *StringValidators) Translator() translator.Translator {
	return sv.translator
}

// WithString applies a series of string validators to a value.
//
// Parameters:
//   - validators: Variable number of string validators to apply.
//
// Returns:
//   - func(any) error: A validator function that validates any value.
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
//
// Parameters:
//   - n: The exact length the string must have.
//
// Returns:
//   - StringValidator: A validator function that checks exact length.
func (sv *StringValidators) Length(n int) StringValidator {
	return func(s string) error {
		if len(s) != n {
			return errors.New(sv.translate("string.length", n))
		}
		return nil
	}
}

// MinLength returns a validator that checks for minimum length.
//
// Parameters:
//   - n: The minimum length the string must have.
//
// Returns:
//   - StringValidator: A validator function that checks minimum length.
func (sv *StringValidators) MinLength(n int) StringValidator {
	return func(s string) error {
		if len(s) < n {
			return errors.New(sv.translate("string.minLength", n))
		}
		return nil
	}
}

// MaxLength returns a validator that checks for maximum length.
//
// Parameters:
//   - n: The maximum length the string can have.
//
// Returns:
//   - StringValidator: A validator function that checks maximum length.
func (sv *StringValidators) MaxLength(n int) StringValidator {
	return func(s string) error {
		if len(s) > n {
			return errors.New(sv.translate("string.maxLength", n))
		}
		return nil
	}
}

// MinRunes returns a validator that checks for minimum number of Unicode runes.
//
// Parameters:
//   - n: The minimum number of runes the string must have.
//
// Returns:
//   - StringValidator: A validator function that checks minimum rune count.
func (sv *StringValidators) MinRunes(n int) StringValidator {
	return func(s string) error {
		if utf8.RuneCountInString(s) < n {
			return errors.New(sv.translate("string.minRunes", n))
		}
		return nil
	}
}

// MaxRunes returns a validator that checks for maximum number of Unicode runes.
//
// Parameters:
//   - n: The maximum number of runes the string can have.
//
// Returns:
//   - StringValidator: A validator function that checks maximum rune count.
func (sv *StringValidators) MaxRunes(n int) StringValidator {
	return func(s string) error {
		if utf8.RuneCountInString(s) > n {
			return errors.New(sv.translate("string.maxRunes", n))
		}
		return nil
	}
}

// OneOf returns a validator that checks if the string is one of the specified
// values.
//
// Parameters:
//   - values: Variable number of allowed string values.
//
// Returns:
//   - StringValidator: A validator function that checks if string is in the
//     allowed values.
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

// Regex returns a validator that ensures the string matches the pattern.
// It includes safety measures against catastrophic backtracking and enforces
// reasonable input length limits.
func (sv *StringValidators) Regex(pattern string) StringValidator {
	// Add safety anchors to prevent catastrophic backtracking
	safePattern := pattern
	if !strings.HasPrefix(pattern, "^") {
		safePattern = "^" + safePattern
	}
	if !strings.HasSuffix(pattern, "$") {
		safePattern = safePattern + "$"
	}

	re, err := regexp.Compile(safePattern)
	if err != nil {
		// Always fail if the pattern is invalid.
		return func(s string) error {
			return errors.New(
				sv.translate("string.regex.invalidPattern", pattern),
			)
		}
	}

	return func(s string) error {
		// Enforce maximum input length to prevent DoS attacks
		const maxInputLength = 10000
		if len(s) > maxInputLength {
			return errors.New(
				sv.translate("string.regex.inputTooLong", maxInputLength),
			)
		}

		// Use the pre-compiled regex for performance
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
