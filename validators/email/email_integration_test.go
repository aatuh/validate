package email_test

import (
	"strings"
	"testing"

	"github.com/aatuh/validate/v3/core"
	"github.com/aatuh/validate/v3/structvalidator"
	"github.com/aatuh/validate/v3/translator"
)

func TestEmail_Integration_EndToEnd(t *testing.T) {
	// End-to-end test via the main validation library
	v := core.New()
	sv := structvalidator.NewStructValidator(v)

	type User struct {
		Email string `validate:"string;email"`
	}

	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"valid email", "user@example.com", true},
		{"valid with subdomain", "user@sub.example.com", true},
		{"valid with plus", "user+tag@example.org", true},
		{"invalid format", "not-an-email", false},
		{"missing @", "userexample.com", false},
		{"empty", "", false},
		{"too long", "a" + strings.Repeat("b", 250) + "@example.com", false},
		{"display name", "John Doe <user@example.com>", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{Email: tt.email}
			err := sv.ValidateStruct(user)

			if tt.valid && err != nil {
				t.Errorf("Expected valid email %q to pass, got error: %v", tt.email, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("Expected invalid email %q to fail, but it passed", tt.email)
			}
		})
	}
}

func TestEmail_Integration_FromRules(t *testing.T) {
	// Using FromRules through the main library
	v := core.New()

	validator, err := v.FromRules([]string{"string;email"})
	if err != nil {
		t.Fatalf("Failed to create validator from rules: %v", err)
	}

	if err := validator("user@example.com"); err != nil {
		t.Errorf("Expected valid email to pass, got error: %v", err)
	}
	if err := validator("invalid-email"); err == nil {
		t.Error("Expected invalid email to fail, but it passed")
	}
}

func TestEmail_Integration_WithTranslator(t *testing.T) {
	msgs := map[string]string{
		"string.email.invalid": "adresse email invalide",
		"string.email.tooLong": "adresse email trop longue",
	}
	tr := translator.NewSimpleTranslator(msgs)

	v := core.New().WithTranslator(tr)
	sv := structvalidator.NewStructValidator(v)

	type User struct {
		Email string `validate:"string;email"`
	}

	user := User{Email: "invalid-email"}
	err := sv.ValidateStruct(user)
	if err == nil {
		t.Error("Expected invalid email to fail")
	}

	if err != nil && !strings.Contains(err.Error(), "adresse email invalide") {
		t.Errorf("Expected custom translation, got: %v", err)
	}
}

func TestEmail_Integration_PluginSystem(t *testing.T) {
	v := core.New()

	validator, err := v.FromRules([]string{"string;email"})
	if err != nil {
		t.Fatalf("Failed to create email validator: %v", err)
	}

	if err := validator("test@example.com"); err != nil {
		t.Errorf("Expected valid email to pass, got error: %v", err)
	}

	err = validator("invalid-email")
	if err == nil {
		t.Error("Expected invalid email to fail, but it passed")
	}

	if err != nil && !strings.Contains(err.Error(), "string.email.invalid") {
		t.Errorf("Expected email error code, got: %v", err)
	}
}
