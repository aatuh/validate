package email

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode"

	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/types"
)

// Email-specific error codes
const (
	CodeEmailInvalid = "string.email.invalid"
	CodeEmailTooLong = "string.email.tooLong"
)

// DefaultEmailTranslations returns default English translations for email validation errors.
func DefaultEmailTranslations() map[string]string {
	return map[string]string{
		"string.email.invalid":           "invalid email address",
		"string.email.tooLong":           "email is too long",
		"string.email.empty":             "email cannot be empty",
		"string.email.format":            "invalid email format",
		"string.email.bareOnly":          "email must not include a display name",
		"string.email.localLength":       "local part length is invalid",
		"string.email.domainLength":      "domain length is invalid",
		"string.email.localDots":         "local part cannot start or end with '.'",
		"string.email.domainLabels":      "domain must have at least two labels",
		"string.email.domainLabelLength": "domain label length is invalid",
		"string.email.domainChars":       "domain contains invalid characters",
		"string.email.domainHyphen":      "domain label cannot start or end with '-'",
		"string.email.tld":               "top-level domain is too short",
	}
}

// KEmail is the rule kind for email validation.
const KEmail types.Kind = "email"

func init() {
	types.RegisterRule(KEmail, compileEmail)
}

func compileEmail(c *types.Compiler, _ types.Rule) (func(any) error, error) {
	return func(v any) error {
		s, ok := v.(string)
		if !ok {
			msg := c.T("string.type", "expected string", nil)
			return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
		}
		if err := validate(s); err != nil {
			msg := c.T(CodeEmailInvalid, "invalid email format", nil)
			return verrs.Errors{verrs.FieldError{Path: "", Code: CodeEmailInvalid, Msg: msg}}
		}
		return nil
	}, nil
}

// validate enforces a bare address with reasonable ASCII domain rules.
func validate(s string) error {
	const maxLen = 255

	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("string.email.empty")
	}
	if len(s) > maxLen {
		return fmt.Errorf("string.email.tooLong")
	}
	if strings.Count(s, "@") != 1 {
		return fmt.Errorf("string.email.format")
	}
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return fmt.Errorf("string.email.format")
	}
	if addr.Address != s {
		return fmt.Errorf("string.email.bareOnly")
	}
	local, domain, _ := strings.Cut(addr.Address, "@")
	if len(local) == 0 || len(local) > 64 {
		return fmt.Errorf("string.email.localLength")
	}
	if len(domain) == 0 || len(domain) > 253 {
		return fmt.Errorf("string.email.domainLength")
	}
	if strings.HasPrefix(local, ".") || strings.HasSuffix(local, ".") {
		return fmt.Errorf("string.email.localDots")
	}
	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return fmt.Errorf("string.email.domainLabels")
	}
	for _, lab := range labels {
		if l := len(lab); l == 0 || l > 63 {
			return fmt.Errorf("string.email.domainLabelLength")
		}
		for i, r := range lab {
			if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-') {
				return fmt.Errorf("string.email.domainChars")
			}
			if (i == 0 || i == len(lab)-1) && r == '-' {
				return fmt.Errorf("string.email.domainHyphen")
			}
		}
	}
	tld := labels[len(labels)-1]
	if len(tld) < 2 {
		return fmt.Errorf("string.email.tld")
	}
	return nil
}
