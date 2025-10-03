package glue

import (
	"strings"
	"testing"
	"time"

	"github.com/aatuh/validate/v3/translator"
)

func TestRegex_SafetyMeasures(t *testing.T) {
	tr := translator.NewSimpleTranslator(
		translator.DefaultEnglishTranslations(),
	)
	v := New().WithTranslator(tr)

	tests := []struct {
		name        string
		pattern     string
		input       string
		shouldPass  bool
		description string
	}{
		{name: "normal regex", pattern: "a.*z", input: "abcz", shouldPass: true,
			description: "normal regex should work"},
		{name: "anchored regex", pattern: "^a.*z$", input: "abcz", shouldPass: true,
			description: "anchored regex should work"},
		{name: "input too long", pattern: ".*",
			input: strings.Repeat("a", 10001), shouldPass: false,
			description: "input over 10k characters should fail"},
		{name: "catastrophic backtracking prevention", pattern: "a+",
			input: strings.Repeat("a", 1000), shouldPass: true,
			description: "long input under limit should work"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := v.String().Regex(tt.pattern).Build()

			done := make(chan bool, 1)
			var err error

			go func() {
				err = validator(tt.input)
				done <- true
			}()

			select {
			case <-done:
			case <-time.After(1 * time.Second):
				t.Fatalf("regex validation timed out - possible catastrophic backtracking")
			}

			if tt.shouldPass && err != nil {
				t.Errorf("expected validation to pass, got error: %v", err)
			}
			if !tt.shouldPass && err == nil {
				t.Errorf("expected validation to fail, got no error")
			}
		})
	}
}

func TestRegex_AnchorSafety(t *testing.T) {
	tr := translator.NewSimpleTranslator(
		translator.DefaultEnglishTranslations(),
	)
	v := New().WithTranslator(tr)

	validator := v.String().Regex("a.*z").Build()

	if err := validator("xabcz"); err == nil {
		t.Errorf("expected anchored regex to fail on input with prefix")
	}
	if err := validator("abcz"); err != nil {
		t.Errorf("expected anchored regex to pass on exact match: %v", err)
	}
}

func TestRegex_InputLengthLimit(t *testing.T) {
	tr := translator.NewSimpleTranslator(
		translator.DefaultEnglishTranslations(),
	)
	v := New().WithTranslator(tr)

	validator := v.String().Regex(".*").Build()

	exactLimit := strings.Repeat("a", 10000)
	if err := validator(exactLimit); err != nil {
		t.Errorf("input at limit should pass: %v", err)
	}

	overLimit := strings.Repeat("a", 10001)
	if err := validator(overLimit); err == nil {
		t.Errorf("input over limit should fail")
	}
}
