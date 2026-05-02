// Package domain registers universal zero-dependency string format validators.
package domain

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net"
	"regexp"
	"strings"
	"time"

	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/translator"
	"github.com/aatuh/validate/v3/types"
)

const (
	KSlug      types.Kind = "slug"
	KSemVer    types.Kind = "semver"
	KJSON      types.Kind = "json"
	KJWT       types.Kind = "jwt"
	KBase64    types.Kind = "base64"
	KBase64URL types.Kind = "base64url"
	KHex       types.Kind = "hex"
	KMAC       types.Kind = "mac"
	KE164      types.Kind = "e164"
	KFQDN      types.Kind = "fqdn"
	KDate      types.Kind = "date"
	KRFC3339   types.Kind = "rfc3339"
	KLuhn      types.Kind = "luhn"
)

const (
	CodeSlugInvalid      = verrs.CodeStringSlugInvalid
	CodeSemVerInvalid    = verrs.CodeStringSemVerInvalid
	CodeJSONInvalid      = verrs.CodeStringJSONInvalid
	CodeJWTInvalid       = verrs.CodeStringJWTInvalid
	CodeBase64Invalid    = verrs.CodeStringBase64Invalid
	CodeBase64URLInvalid = verrs.CodeStringBase64URLInvalid
	CodeHexInvalid       = verrs.CodeStringHexInvalid
	CodeMACInvalid       = verrs.CodeStringMACInvalid
	CodeE164Invalid      = verrs.CodeStringE164Invalid
	CodeFQDNInvalid      = verrs.CodeStringFQDNInvalid
	CodeDateInvalid      = verrs.CodeStringDateInvalid
	CodeRFC3339Invalid   = verrs.CodeStringRFC3339Invalid
	CodeLuhnInvalid      = verrs.CodeStringLuhnInvalid
)

type stringFormatRule struct {
	kind       types.Kind
	code       string
	defaultMsg string
	validate   func(string) bool
}

var semverPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-(?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

func init() {
	for _, rule := range []stringFormatRule{
		{KSlug, CodeSlugInvalid, "must be a valid slug", isSlug},
		{KSemVer, CodeSemVerInvalid, "must be a valid semantic version", isSemVer},
		{KJSON, CodeJSONInvalid, "must be valid JSON", isJSON},
		{KJWT, CodeJWTInvalid, "must be a structurally valid JWT", isJWT},
		{KBase64, CodeBase64Invalid, "must be valid base64", isBase64},
		{KBase64URL, CodeBase64URLInvalid, "must be valid base64url", isBase64URL},
		{KHex, CodeHexInvalid, "must be valid hexadecimal", isHexString},
		{KMAC, CodeMACInvalid, "must be a valid MAC address", isMAC},
		{KE164, CodeE164Invalid, "must be a valid E.164 phone number", isE164},
		{KFQDN, CodeFQDNInvalid, "must be a valid fully qualified domain name", isFQDN},
		{KDate, CodeDateInvalid, "must be a valid date", isDate},
		{KRFC3339, CodeRFC3339Invalid, "must be a valid RFC3339 timestamp", isRFC3339},
		{KLuhn, CodeLuhnInvalid, "must pass the Luhn checksum", isLuhn},
	} {
		types.RegisterRule(rule.kind, compileStringFormat(rule))
	}
	translator.RegisterDefaultEnglishTranslations(DefaultDomainTranslations())
}

func DefaultDomainTranslations() map[string]string {
	return map[string]string{
		CodeSlugInvalid:      "must be a valid slug",
		CodeSemVerInvalid:    "must be a valid semantic version",
		CodeJSONInvalid:      "must be valid JSON",
		CodeJWTInvalid:       "must be a structurally valid JWT",
		CodeBase64Invalid:    "must be valid base64",
		CodeBase64URLInvalid: "must be valid base64url",
		CodeHexInvalid:       "must be valid hexadecimal",
		CodeMACInvalid:       "must be a valid MAC address",
		CodeE164Invalid:      "must be a valid E.164 phone number",
		CodeFQDNInvalid:      "must be a valid fully qualified domain name",
		CodeDateInvalid:      "must be a valid date",
		CodeRFC3339Invalid:   "must be a valid RFC3339 timestamp",
		CodeLuhnInvalid:      "must pass the Luhn checksum",
	}
}

func compileStringFormat(rule stringFormatRule) types.RuleCompiler {
	return func(c *types.Compiler, _ types.Rule) (func(any) error, error) {
		return func(v any) error {
			s, ok := v.(string)
			if !ok {
				msg := c.T(verrs.CodeStringType, "expected string", nil)
				return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
			}
			if !rule.validate(s) {
				msg := c.T(rule.code, rule.defaultMsg, nil)
				return verrs.Errors{verrs.FieldError{Path: "", Code: rule.code, Msg: msg}}
			}
			return nil
		}, nil
	}
}

func isSlug(s string) bool {
	if s == "" || s[0] == '-' || s[len(s)-1] == '-' {
		return false
	}
	previousHyphen := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			previousHyphen = false
		case r >= '0' && r <= '9':
			previousHyphen = false
		case r == '-':
			if previousHyphen {
				return false
			}
			previousHyphen = true
		default:
			return false
		}
	}
	return true
}

func isSemVer(s string) bool {
	return semverPattern.MatchString(s)
}

func isJSON(s string) bool {
	return json.Valid([]byte(s))
}

func isJWT(s string) bool {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
	}
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil || !json.Valid(header) {
		return false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !json.Valid(payload) {
		return false
	}
	_, err = base64.RawURLEncoding.DecodeString(parts[2])
	return err == nil
}

func isBase64(s string) bool {
	if s == "" {
		return false
	}
	if _, err := base64.StdEncoding.DecodeString(s); err == nil {
		return true
	}
	_, err := base64.RawStdEncoding.DecodeString(s)
	return err == nil
}

func isBase64URL(s string) bool {
	if s == "" || strings.ContainsAny(s, "+/") {
		return false
	}
	if _, err := base64.URLEncoding.DecodeString(s); err == nil {
		return true
	}
	_, err := base64.RawURLEncoding.DecodeString(s)
	return err == nil
}

func isHexString(s string) bool {
	if s == "" {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

func isMAC(s string) bool {
	if s == "" {
		return false
	}
	_, err := net.ParseMAC(s)
	return err == nil
}

func isE164(s string) bool {
	if len(s) < 3 || len(s) > 16 || s[0] != '+' || s[1] == '0' {
		return false
	}
	for i := 1; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func isFQDN(s string) bool {
	if s == "" || len(s) > 253 {
		return false
	}
	name := strings.TrimSuffix(s, ".")
	if name == "" || !strings.Contains(name, ".") {
		return false
	}
	labels := strings.Split(name, ".")
	for _, label := range labels {
		if !isDomainLabel(label) {
			return false
		}
	}
	return len(labels[len(labels)-1]) >= 2
}

func isDomainLabel(label string) bool {
	if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
		return false
	}
	for i := 0; i < len(label); i++ {
		c := label[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' {
			continue
		}
		return false
	}
	return true
}

func isDate(s string) bool {
	t, err := time.Parse("2006-01-02", s)
	return err == nil && t.Format("2006-01-02") == s
}

func isRFC3339(s string) bool {
	_, err := time.Parse(time.RFC3339Nano, s)
	return err == nil
}

func isLuhn(s string) bool {
	if len(s) < 2 {
		return false
	}
	sum := 0
	double := false
	for i := len(s) - 1; i >= 0; i-- {
		c := s[i]
		if c < '0' || c > '9' {
			return false
		}
		n := int(c - '0')
		if double {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		double = !double
	}
	return sum%10 == 0
}
