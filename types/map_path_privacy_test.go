package types

import (
	"errors"
	"strings"
	"testing"

	verrs "github.com/aatuh/validate/v3/errors"
)

func TestCompiler_MapKeyPathPreservesShortOrdinaryKeys(t *testing.T) {
	fn := compileMapPrivacyTag(t, "map;values=(string;min=2)")

	es := requireMapPrivacyErrors(t, fn(map[string]string{"id": ""}))
	if len(es) != 1 {
		t.Fatalf("errors = %#v, want one error", es)
	}
	if es[0].Path != "[id]" {
		t.Fatalf("path = %q, want short key path", es[0].Path)
	}
	if es[0].Code != verrs.CodeStringMin {
		t.Fatalf("code = %q, want %q", es[0].Code, verrs.CodeStringMin)
	}
}

func TestCompiler_MapKeyPathRedactsLongAndSensitiveKeys(t *testing.T) {
	longKey := strings.Repeat("a", 96)
	sensitiveKey := "access_token=secret-value"

	tests := []struct {
		name  string
		tag   string
		value any
		raw   string
		code  string
	}{
		{
			name:  "values long key",
			tag:   "map;values=(string;min=2)",
			value: map[string]string{longKey: ""},
			raw:   longKey,
			code:  verrs.CodeStringMin,
		},
		{
			name:  "keys sensitive key",
			tag:   "map;keys=(string;max=3)",
			value: map[string]string{sensitiveKey: "ignored"},
			raw:   sensitiveKey,
			code:  verrs.CodeStringMax,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := compileMapPrivacyTag(t, tt.tag)
			es := requireMapPrivacyErrors(t, fn(tt.value))
			if len(es) != 1 {
				t.Fatalf("errors = %#v, want one error", es)
			}
			if es[0].Path != "[<redacted>]" {
				t.Fatalf("path = %q, want redacted map key path", es[0].Path)
			}
			if strings.Contains(es[0].Path, tt.raw) || strings.Contains(es.Error(), tt.raw) {
				t.Fatalf("error leaked raw map key %q: %#v", tt.raw, es)
			}
			if es[0].Code != tt.code {
				t.Fatalf("code = %q, want %q", es[0].Code, tt.code)
			}
		})
	}
}

func compileMapPrivacyTag(t *testing.T, tag string) ValidatorFunc {
	t.Helper()
	rules, err := ParseTag(tag)
	if err != nil {
		t.Fatalf("ParseTag(%q): %v", tag, err)
	}
	fn, err := NewCompiler(nil).CompileE(rules)
	if err != nil {
		t.Fatalf("CompileE(%q): %v", tag, err)
	}
	return fn
}

func requireMapPrivacyErrors(t *testing.T, err error) verrs.Errors {
	t.Helper()
	if err == nil {
		t.Fatalf("got nil error, want structured errors")
	}
	var es verrs.Errors
	if !errors.As(err, &es) {
		t.Fatalf("got %T %v, want structured errors", err, err)
	}
	return es
}
