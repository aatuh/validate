package types

import (
	"context"
	"errors"
	"testing"

	verrs "github.com/aatuh/validate/v3/errors"
)

func TestCompiler_CollectAllOptInPreservesFailFastDefault(t *testing.T) {
	rules := []Rule{
		NewRule(KString, nil),
		NewRule(KMinLength, map[string]any{"n": 5}),
		NewRule(KMaxLength, map[string]any{"n": 2}),
	}

	failFast := NewCompiler(nil).Compile(rules)
	assertCodes(t, failFast("abc"), []string{verrs.CodeStringMin})

	collectAll, err := NewCompiler(nil).CompileWithOptsE(rules, CompileOpts{CollectAll: true})
	if err != nil {
		t.Fatalf("CompileWithOptsE returned error: %v", err)
	}
	assertCodes(t, collectAll("abc"), []string{verrs.CodeStringMin, verrs.CodeStringMax})
}

func TestCompiler_CollectAllRequiredAndOmitEmptyShortCircuit(t *testing.T) {
	collectAll, err := NewCompiler(nil).CompileWithOptsE([]Rule{
		NewRule(KRequired, nil),
		NewRule(KString, nil),
		NewRule(KMinLength, map[string]any{"n": 5}),
	}, CompileOpts{CollectAll: true})
	if err != nil {
		t.Fatalf("CompileWithOptsE returned error: %v", err)
	}
	assertCodes(t, collectAll(""), []string{verrs.CodeRequired})

	omitEmpty, err := NewCompiler(nil).CompileWithOptsE([]Rule{
		NewRule(KOmitempty, nil),
		NewRule(KString, nil),
		NewRule(KMinLength, map[string]any{"n": 5}),
	}, CompileOpts{CollectAll: true})
	if err != nil {
		t.Fatalf("CompileWithOptsE returned error: %v", err)
	}
	if err := omitEmpty(""); err != nil {
		t.Fatalf("omitempty should skip zero value, got %v", err)
	}
}

func TestCompiler_ContextCompilation(t *testing.T) {
	type ctxKey string
	const key ctxKey = "tenant"

	c := NewCompiler(nil)
	c.RegisterContextRule("ctxTenant", func(c *Compiler, rule Rule) (ContextValidatorFunc, error) {
		return func(ctx context.Context, value any) error {
			if ctx.Value(key) == "allowed" {
				return nil
			}
			return verrs.Errors{verrs.FieldError{Code: "context.tenant"}}
		}, nil
	})

	fn, err := c.CompileContextE([]Rule{NewRule("ctxTenant", nil)})
	if err != nil {
		t.Fatalf("CompileContextE returned error: %v", err)
	}
	if err := fn(context.WithValue(context.Background(), key, "allowed"), "value"); err != nil {
		t.Fatalf("context validator rejected allowed context: %v", err)
	}
	assertCodes(t, fn(context.Background(), "value"), []string{"context.tenant"})

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	builtIn, err := NewCompiler(nil).CompileContextE([]Rule{NewRule(KString, nil)})
	if err != nil {
		t.Fatalf("CompileContextE built-in: %v", err)
	}
	if err := builtIn(canceled, "value"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled context error = %v, want context.Canceled", err)
	}
}

func assertCodes(t *testing.T, err error, want []string) {
	t.Helper()
	if len(want) == 0 {
		if err != nil {
			t.Fatalf("got error %v, want nil", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("got nil error, want codes %v", want)
	}
	var es verrs.Errors
	if !errors.As(err, &es) {
		t.Fatalf("got %T %v, want structured errors", err, err)
	}
	if len(es) != len(want) {
		t.Fatalf("codes = %#v, want %v", es, want)
	}
	for i, code := range want {
		if es[i].Code != code {
			t.Fatalf("code[%d] = %q, want %q; errors=%#v", i, es[i].Code, code, es)
		}
	}
}
