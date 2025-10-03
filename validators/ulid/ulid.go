package ulid

import (
	"strings"

	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/types"
)

// ULID-specific error codes
const (
	CodeULIDInvalid = "string.ulid.invalid"
)

// DefaultULIDTranslations returns default English translations for ULID validation errors.
func DefaultULIDTranslations() map[string]string {
	return map[string]string{
		"string.ulid.invalid": "invalid ULID format",
	}
}

// KULID is the rule kind for ULID validation.
const KULID types.Kind = "ulid"

func init() {
	types.RegisterRule(KULID, compileULID)
}

func compileULID(c *types.Compiler, _ types.Rule) (func(any) error, error) {
	return func(v any) error {
		s, ok := v.(string)
		if !ok {
			msg := c.T("string.type", "expected string", nil)
			return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
		}
		if fe := validateULIDString(c, s); fe.Code != "" {
			return verrs.Errors{fe}
		}
		return nil
	}, nil
}

// validateULIDString checks Crockford base32 ULID format.
func validateULIDString(c *types.Compiler, s string) verrs.FieldError {
	if len(s) != 26 {
		return verrs.FieldError{
			Code: CodeULIDInvalid,
			Msg:  c.T(CodeULIDInvalid, "invalid ULID format", nil),
		}
	}
	const invalid = "ILOU"
	for _, r := range s {
		switch {
		case '0' <= r && r <= '9':
			// ok
		case 'A' <= r && r <= 'Z':
			if strings.ContainsRune(invalid, r) {
				return verrs.FieldError{
					Code: CodeULIDInvalid,
					Msg:  c.T(CodeULIDInvalid, "invalid ULID format", nil),
				}
			}
		default:
			return verrs.FieldError{
				Code: CodeULIDInvalid,
				Msg:  c.T(CodeULIDInvalid, "invalid ULID format", nil),
			}
		}
	}
	return verrs.FieldError{}
}
