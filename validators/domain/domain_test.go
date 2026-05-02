package domain

import "testing"

func TestStringFormatValidators(t *testing.T) {
	tests := []struct {
		name    string
		valid   string
		invalid string
		check   func(string) bool
	}{
		{"slug", "alpha-123", "Alpha_123", isSlug},
		{"semver", "1.2.3-alpha.1+build.5", "01.2.3", isSemVer},
		{"json", `{"ok":true}`, `{bad`, isJSON},
		{"jwt", "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiIxMjMifQ.c2lnbmF0dXJl", "a.b.c", isJWT},
		{"base64", "dmFsaWQ=", "not-base64", isBase64},
		{"base64url", "dmFsaWQ", "not/base64url", isBase64URL},
		{"hex", "deadBEEF", "abc", isHexString},
		{"mac", "01:23:45:67:89:ab", "01:23:45", isMAC},
		{"e164", "+358401234567", "+012345", isE164},
		{"fqdn", "api.example.com", "localhost", isFQDN},
		{"date", "2026-05-08", "2026-02-29", isDate},
		{"rfc3339", "2026-05-08T10:30:00Z", "2026-05-08", isRFC3339},
		{"luhn", "79927398713", "79927398714", isLuhn},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check(tt.valid) {
				t.Fatalf("valid value rejected")
			}
			if tt.check(tt.invalid) {
				t.Fatalf("invalid value accepted")
			}
			if tt.check("") {
				t.Fatalf("empty value accepted")
			}
		})
	}
}
