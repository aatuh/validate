package uuid

import (
	"unicode"

	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/types"
)

// UUID-specific error codes
const (
	CodeUUIDInvalid = "string.uuid.invalid"
)

// DefaultUUIDTranslations returns default English translations for UUID validation errors.
func DefaultUUIDTranslations() map[string]string {
	return map[string]string{
		"string.uuid.invalid": "invalid UUID format",
	}
}

// KUUID is the rule kind for UUID validation.
const KUUID types.Kind = "uuid"

func init() {
	types.RegisterRule(KUUID, compileUUID)
}

func compileUUID(c *types.Compiler, _ types.Rule) (func(any) error, error) {
	return func(v any) error {
		s, ok := v.(string)
		if !ok {
			msg := c.T("string.type", "expected string", nil)
			return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
		}
		if fe := validateUUIDString(c, s); fe.Code != "" {
			return verrs.Errors{fe}
		}
		return nil
	}, nil
}

// validateUUIDString checks canonical UUID format and uses translator.
func validateUUIDString(c *types.Compiler, s string) verrs.FieldError {
	const L = 36
	if len(s) != L || s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return verrs.FieldError{
			Code: CodeUUIDInvalid,
			Msg:  c.T(CodeUUIDInvalid, "invalid UUID format", nil),
		}
	}
	for i, r := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			continue
		}
		if !isHex(r) {
			return verrs.FieldError{
				Code: CodeUUIDInvalid,
				Msg:  c.T(CodeUUIDInvalid, "invalid UUID format", nil),
			}
		}
	}
	return verrs.FieldError{}
}

func isHex(r rune) bool {
	return unicode.IsDigit(r) || ('a' <= r && r <= 'f') || ('A' <= r && r <= 'F')
}
