package ulid

import (
	"testing"

	"github.com/aatuh/validate/v3/types"
)

func TestULID_ValidULIDs(t *testing.T) {
	// Test valid ULID formats (Crockford base32)
	validULIDs := []string{
		"01ARZ3NDEKTSV4RRFFQ69G5FAV",
		"7ZZZZZZZZZZZZZZZZZZZZZZZZZ",
		"00000000000000000000000000",
		"0123456789ABCDEFGHJKMNPQRS",
	}

	for _, ulid := range validULIDs {
		if fe := validateULIDString(&types.Compiler{}, ulid); fe.Code != "" {
			t.Errorf("Expected valid ULID %q to pass, got error: %s", ulid, fe.Code)
		}
	}
}

func TestULID_InvalidULIDs(t *testing.T) {
	// Test invalid ULID formats
	invalidULIDs := []string{
		"not-a-ulid",
		"01ARZ3NDEKTSV4RRFFQ69G5FA",   // too short (25 chars)
		"01ARZ3NDEKTSV4RRFFQ69G5FAVA", // too long (27 chars)
		"01ARZ3NDEKTSV4RRFFQ69G5FAI",  // contains 'I'
		"01ARZ3NDEKTSV4RRFFQ69G5FAL",  // contains 'L'
		"01ARZ3NDEKTSV4RRFFQ69G5FAO",  // contains 'O'
		"01ARZ3NDEKTSV4RRFFQ69G5FAU",  // contains 'U'
		"01ARZ3NDEKTSV4RRFFQ69G5FAi",  // contains lowercase 'i'
		"01ARZ3NDEKTSV4RRFFQ69G5FAl",  // contains lowercase 'l'
		"01ARZ3NDEKTSV4RRFFQ69G5FAo",  // contains lowercase 'o'
		"01ARZ3NDEKTSV4RRFFQ69G5FAu",  // contains lowercase 'u'
		"",
		"123",
		"!@#$%^&*()",
	}

	for _, ulid := range invalidULIDs {
		if fe := validateULIDString(&types.Compiler{}, ulid); fe.Code == "" {
			t.Errorf("Expected invalid ULID %q to fail, but it passed", ulid)
		}
	}
}

func TestULID_LengthValidation(t *testing.T) {
	// Test exact length requirement (26 characters)
	testCases := []struct {
		ulid  string
		valid bool
		desc  string
	}{
		{"01ARZ3NDEKTSV4RRFFQ69G5FAV", true, "exactly 26 chars"},
		{"01ARZ3NDEKTSV4RRFFQ69G5FA", false, "25 chars (too short)"},
		{"01ARZ3NDEKTSV4RRFFQ69G5FAVA", false, "27 chars (too long)"},
		{"", false, "empty string"},
	}

	for _, tc := range testCases {
		fe := validateULIDString(&types.Compiler{}, tc.ulid)
		if tc.valid && fe.Code != "" {
			t.Errorf("Expected %s to be valid, got error: %s", tc.desc, fe.Code)
		}
		if !tc.valid && fe.Code == "" {
			t.Errorf("Expected %s to be invalid, but it passed", tc.desc)
		}
	}
}
