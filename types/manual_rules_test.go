package types

import (
	"errors"
	"testing"

	verrs "github.com/aatuh/validate/v3/errors"
)

func TestCompiler_ManualSliceRulesReturnTypeErrorsForMalformedInputs(t *testing.T) {
	tests := []struct {
		name  string
		rules []Rule
		input any
	}{
		{
			name:  "slice length nil",
			rules: []Rule{NewRule(KSliceLength, map[string]any{"n": 1})},
			input: nil,
		},
		{
			name:  "min slice length nil",
			rules: []Rule{NewRule(KMinSliceLength, map[string]any{"n": 1})},
			input: nil,
		},
		{
			name:  "max slice length nil",
			rules: []Rule{NewRule(KMaxSliceLength, map[string]any{"n": 1})},
			input: nil,
		},
		{
			name: "foreach nil",
			rules: []Rule{NewRule(KForEach, map[string]any{
				"rules": []Rule{NewRule(KString, nil)},
			})},
			input: nil,
		},
		{
			name: "foreach elem nil",
			rules: []Rule{NewRuleWithElem(KForEach, nil, &Rule{
				Kind: KString,
			})},
			input: nil,
		},
		{
			name:  "slice length wrong type",
			rules: []Rule{NewRule(KSliceLength, map[string]any{"n": 1})},
			input: 123,
		},
		{
			name: "foreach wrong type",
			rules: []Rule{NewRule(KForEach, map[string]any{
				"rules": []Rule{NewRule(KString, nil)},
			})},
			input: "not a slice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, err := NewCompiler(nil).CompileE(tt.rules)
			if err != nil {
				t.Fatalf("CompileE returned error: %v", err)
			}

			var got error
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("validator panicked: %v", r)
					}
				}()
				got = fn(tt.input)
			}()

			assertErrorCode(t, got, verrs.CodeSliceType)
		})
	}
}

func assertErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatalf("got nil error, want code %q", code)
	}
	var es verrs.Errors
	if !errors.As(err, &es) || len(es) == 0 {
		t.Fatalf("got %T %v, want structured error code %q", err, err, code)
	}
	if es[0].Code != code {
		t.Fatalf("code = %q, want %q; errors=%#v", es[0].Code, code, es)
	}
}
