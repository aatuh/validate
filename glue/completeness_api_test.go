package glue

import (
	"context"
	"errors"
	"testing"

	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/types"
)

func TestValidate_CollectAllAPIsAndBuilderMethods(t *testing.T) {
	v := New()

	err := v.CheckTagWithOpts("string;min=5;max=2", "abc", types.CompileOpts{CollectAll: true})
	requireGlueCodes(t, err, []string{verrs.CodeStringMin, verrs.CodeStringMax})

	fn := v.String().MinLength(5).MaxLength(2).BuildAll()
	requireGlueCodes(t, fn("abc"), []string{verrs.CodeStringMin, verrs.CodeStringMax})

	failFast := v.String().MinLength(5).MaxLength(2).Build()
	requireGlueCodes(t, failFast("abc"), []string{verrs.CodeStringMin})
}

func TestValidate_ContextAPIsAndBuilders(t *testing.T) {
	type ctxKey string
	const key ctxKey = "ok"

	v := New().WithContextRuleCompiler("ctxOK", func(c *types.Compiler, rule types.Rule) (types.ContextValidatorFunc, error) {
		return func(ctx context.Context, value any) error {
			if ctx.Value(key) == true {
				return nil
			}
			return verrs.Errors{verrs.FieldError{Code: "context.ok"}}
		}, nil
	})

	if err := v.CheckTagContext(context.WithValue(context.Background(), key, true), "string;ctxOK", "value"); err != nil {
		t.Fatalf("CheckTagContext rejected allowed context: %v", err)
	}
	requireGlueCodes(t, v.CheckTagContext(context.Background(), "string;ctxOK", "value"), []string{"context.ok"})

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	fn := v.String().MinLength(2).BuildContext()
	if err := fn(canceled, "abc"); !errors.Is(err, context.Canceled) {
		t.Fatalf("BuildContext canceled error = %v, want context.Canceled", err)
	}

	collectAll := v.String().MinLength(5).MaxLength(2).BuildContextWithOpts(types.CompileOpts{CollectAll: true})
	requireGlueCodes(t, collectAll(context.Background(), "abc"), []string{verrs.CodeStringMin, verrs.CodeStringMax})
}

func requireGlueCodes(t *testing.T, err error, want []string) {
	t.Helper()
	if err == nil {
		t.Fatalf("got nil error, want codes %v", want)
	}
	var es verrs.Errors
	if !errors.As(err, &es) {
		t.Fatalf("got %T %v, want structured errors", err, err)
	}
	if len(es) != len(want) {
		t.Fatalf("errors = %#v, want codes %v", es, want)
	}
	for i, code := range want {
		if es[i].Code != code {
			t.Fatalf("code[%d] = %q, want %q; errors=%#v", i, es[i].Code, code, es)
		}
	}
}
