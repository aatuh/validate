package uuid

import (
	"errors"
	"testing"

	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/types"
)

func TestUUID_ValidUUIDs(t *testing.T) {
	// Test valid UUID formats
	validUUIDs := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		"6ba7b811-9dad-11d1-80b4-00c04fd430c8",
		"00000000-0000-0000-0000-000000000000",
		"ffffffff-ffff-ffff-ffff-ffffffffffff",
	}

	for _, uuid := range validUUIDs {
		if fe := validateUUIDString(&types.Compiler{}, uuid); fe.Code != "" {
			t.Errorf("Expected valid UUID %q to pass, got error: %s", uuid, fe.Code)
		}
	}
}

func TestUUID_InvalidUUIDs(t *testing.T) {
	// Test invalid UUID formats
	invalidUUIDs := []string{
		"not-a-uuid",
		"550e8400-e29b-41d4-a716-44665544000",   // too short
		"550e8400-e29b-41d4-a716-4466554400000", // too long
		"550e8400-e29b-41d4-a716-44665544000g",  // invalid character
		"550e8400-e29b-41d4-a716446655440000",   // missing hyphen
		"550e8400e29b-41d4-a716-446655440000",   // missing hyphen
		"550e8400-e29b41d4-a716-446655440000",   // missing hyphen
		"550e8400-e29b-41d4a716-446655440000",   // missing hyphen
		"550e8400-e29b-41d4-a716446655440000",   // missing hyphen
		"",
		"123",
	}

	for _, uuid := range invalidUUIDs {
		if fe := validateUUIDString(&types.Compiler{}, uuid); fe.Code == "" {
			t.Errorf("Expected invalid UUID %q to fail, but it passed", uuid)
		}
	}
}

func TestIsHex(t *testing.T) {
	// Test hex character detection
	hexChars := "0123456789abcdefABCDEF"
	for _, r := range hexChars {
		if !isHex(r) {
			t.Errorf("Expected %c to be recognized as hex", r)
		}
	}

	nonHexChars := "ghijklmnopqrstuvwxyzGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()"
	for _, r := range nonHexChars {
		if isHex(r) {
			t.Errorf("Expected %c to NOT be recognized as hex", r)
		}
	}
}

func TestUUIDVersionRules(t *testing.T) {
	tests := []struct {
		name    string
		kind    types.Kind
		valid   string
		invalid string
	}{
		{"v1", KUUIDv1, "6ba7b810-9dad-11d1-80b4-00c04fd430c8", "550e8400-e29b-41d4-a716-446655440000"},
		{"v3", KUUIDv3, "6fa459ea-ee8a-3ca4-894e-db77e160355e", "550e8400-e29b-41d4-a716-446655440000"},
		{"v4", KUUIDv4, "550e8400-e29b-41d4-a716-446655440000", "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
		{"v5", KUUIDv5, "2ed6657d-e927-568b-95e1-2665a8aea6a2", "550e8400-e29b-41d4-a716-446655440000"},
		{"v6", KUUIDv6, "1ef21d2f-1207-6660-8c4f-419efbd44d48", "550e8400-e29b-41d4-a716-446655440000"},
		{"v7", KUUIDv7, "01890f13-a93c-7cc2-98e5-9f8c7e2b8a6f", "550e8400-e29b-41d4-a716-446655440000"},
		{"v8", KUUIDv8, "01890f13-a93c-8cc2-98e5-9f8c7e2b8a6f", "550e8400-e29b-41d4-a716-446655440000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := types.NewCompiler(nil).Compile([]types.Rule{types.NewRule(tt.kind, nil)})
			if err := fn(tt.valid); err != nil {
				t.Fatalf("valid UUID version failed: %v", err)
			}
			requireUUIDCode(t, fn(tt.invalid), CodeUUIDVersion)
			requireUUIDCode(t, fn(""), CodeUUIDInvalid)
			requireUUIDCode(t, fn(123), verrs.CodeStringType)
		})
	}
}

func TestUUIDVersionRules_RejectNonRFC4122Variant(t *testing.T) {
	fn := types.NewCompiler(nil).Compile([]types.Rule{types.NewRule(KUUIDv4, nil)})
	requireUUIDCode(t, fn("550e8400-e29b-41d4-c716-446655440000"), CodeUUIDVersion)
}

func requireUUIDCode(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("got nil error, want %q", want)
	}
	var es verrs.Errors
	if !errors.As(err, &es) || len(es) == 0 {
		t.Fatalf("got %T %v, want structured errors", err, err)
	}
	if es[0].Code != want {
		t.Fatalf("code = %q, want %q; errors=%#v", es[0].Code, want, es)
	}
}
