package email

import (
	"strings"
	"testing"
)

func TestEmail_ValidAddresses(t *testing.T) {
	// Test valid email addresses
	validEmails := []string{
		"user@example.com",
		"test.email@domain.co.uk",
		"user+tag@example.org",
		"user123@subdomain.example.com",
	}

	for _, email := range validEmails {
		if err := validate(email); err != nil {
			t.Errorf("Expected valid email %q to pass, got error: %v", email, err)
		}
	}
}

func TestEmail_InvalidAddresses(t *testing.T) {
	// Test invalid email addresses
	invalidEmails := []string{
		"not-an-email",
		"@example.com",
		"user@",
		"user@.com",
		"user..name@example.com",
		"user@example..com",
		"",
		"user@example",
		"user@example.c",
	}

	for _, email := range invalidEmails {
		if err := validate(email); err == nil {
			t.Errorf("Expected invalid email %q to fail, but it passed", email)
		}
	}
}

func TestEmail_TooLong(t *testing.T) {
	// Create an email that's too long (over 255 characters)
	longEmail := "a" + strings.Repeat("b", 250) + "@example.com"
	if err := validate(longEmail); err == nil {
		t.Errorf("Expected long email to fail, but it passed")
	}
}

func TestEmail_DisplayNames(t *testing.T) {
	// Test that display names are rejected (bare addresses only)
	displayNameEmails := []string{
		"John Doe <user@example.com>",
		"\"John Doe\" <user@example.com>",
	}

	for _, email := range displayNameEmails {
		if err := validate(email); err == nil {
			t.Errorf("Expected display name email %q to fail, but it passed", email)
		}
	}
}
