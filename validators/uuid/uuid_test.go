package uuid

import (
	"testing"

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
