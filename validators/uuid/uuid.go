package uuid

import (
	"unicode"

	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/translator"
	"github.com/aatuh/validate/v3/types"
)

// UUID-specific error codes
const (
	CodeUUIDInvalid = "string.uuid.invalid"
	CodeUUIDVersion = verrs.CodeStringUUIDVersion
)

// DefaultUUIDTranslations returns default English translations for UUID validation errors.
func DefaultUUIDTranslations() map[string]string {
	return map[string]string{
		"string.uuid.invalid": "invalid UUID format",
		"string.uuid.version": "invalid UUID version",
	}
}

// KUUID is the rule kind for UUID validation.
const (
	KUUID   types.Kind = "uuid"
	KUUIDv1 types.Kind = "uuidv1"
	KUUIDv3 types.Kind = "uuidv3"
	KUUIDv4 types.Kind = "uuidv4"
	KUUIDv5 types.Kind = "uuidv5"
	KUUIDv6 types.Kind = "uuidv6"
	KUUIDv7 types.Kind = "uuidv7"
	KUUIDv8 types.Kind = "uuidv8"
)

func init() {
	types.RegisterRule(KUUID, compileUUID)
	for _, rule := range []struct {
		kind    types.Kind
		version byte
	}{
		{KUUIDv1, '1'},
		{KUUIDv3, '3'},
		{KUUIDv4, '4'},
		{KUUIDv5, '5'},
		{KUUIDv6, '6'},
		{KUUIDv7, '7'},
		{KUUIDv8, '8'},
	} {
		types.RegisterRule(rule.kind, compileUUIDVersion(rule.version))
	}
	// Register UUID as a custom type
	types.RegisterGlobalType("uuid", &UUIDTypeValidatorFactory{})
	translator.RegisterDefaultEnglishTranslations(DefaultUUIDTranslations())
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

func compileUUIDVersion(version byte) types.RuleCompiler {
	return func(c *types.Compiler, _ types.Rule) (func(any) error, error) {
		return func(v any) error {
			s, ok := v.(string)
			if !ok {
				msg := c.T("string.type", "expected string", nil)
				return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
			}
			if fe := validateUUIDString(c, s); fe.Code != "" {
				return verrs.Errors{fe}
			}
			if s[14] != version || !isRFC4122Variant(s[19]) {
				return verrs.Errors{verrs.FieldError{
					Code: CodeUUIDVersion,
					Msg:  c.T(CodeUUIDVersion, "invalid UUID version", nil),
				}}
			}
			return nil
		}, nil
	}
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

func isRFC4122Variant(b byte) bool {
	switch b {
	case '8', '9', 'a', 'A', 'b', 'B':
		return true
	default:
		return false
	}
}
