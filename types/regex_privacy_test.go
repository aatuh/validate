package types

import (
	"errors"
	"strings"
	"testing"

	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/translator"
)

func TestCompiler_InvalidRegexPatternMessageIsCappedAndRedacted(t *testing.T) {
	tr := translator.NewSimpleTranslator(translator.DefaultEnglishTranslations())
	c := NewCompiler(tr)

	tests := []struct {
		name      string
		pattern   string
		forbidden []string
	}{
		{
			name:    "long pattern is capped",
			pattern: "(" + strings.Repeat("a", 200),
			forbidden: []string{
				strings.Repeat("a", 120),
			},
		},
		{
			name:    "sensitive-looking pattern is redacted",
			pattern: "token=sk_live_" + strings.Repeat("x", 48) + "(",
			forbidden: []string{
				"token=",
				"sk_live",
				strings.Repeat("x", 24),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, err := c.CompileE([]Rule{NewRule(KRegex, map[string]any{"pattern": tt.pattern})})
			if err != nil {
				t.Fatalf("CompileE returned error: %v", err)
			}
			err = fn("anything")
			es := requireErrorsWithCode(t, err, verrs.CodeStringRegexInvalidPattern)
			if len(es[0].Msg) > len("invalid regex pattern: ")+maxRegexPatternMessageRunes+len("...") {
				t.Fatalf("message too long: len=%d msg=%q", len(es[0].Msg), es[0].Msg)
			}
			for _, forbidden := range tt.forbidden {
				if strings.Contains(es[0].Msg, forbidden) || strings.Contains(err.Error(), forbidden) {
					t.Fatalf("invalid regex error exposed %q in %q", forbidden, err.Error())
				}
			}
		})
	}
}

func requireErrorsWithCode(t *testing.T, err error, code string) verrs.Errors {
	t.Helper()
	if err == nil {
		t.Fatalf("got nil error, want code %q", code)
	}
	var es verrs.Errors
	if !errors.As(err, &es) || len(es) == 0 {
		t.Fatalf("got %T %v, want structured errors", err, err)
	}
	if es[0].Code != code {
		t.Fatalf("code = %q, want %q; errors=%#v", es[0].Code, code, es)
	}
	return es
}
