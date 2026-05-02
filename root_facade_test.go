package validate

import (
	"context"
	"errors"
	"testing"
	"time"

	verrs "github.com/aatuh/validate/v3/errors"
)

func TestRootFacade_ExpandedExportsAndPluginTranslations(t *testing.T) {
	v := New()

	if err := v.Float().Finite().Positive().Build()(1.5); err != nil {
		t.Fatalf("root float builder failed: %v", err)
	}
	if err := v.Map().MinKeys(1).Build()(map[string]int{"a": 1}); err != nil {
		t.Fatalf("root map builder failed: %v", err)
	}
	if err := v.Time().After(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)).Build()(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("root time builder failed: %v", err)
	}

	err := v.CheckTag("string;email", "not-an-email")
	if err == nil {
		t.Fatalf("invalid email passed")
	}
	var es Errors
	if !errors.As(err, &es) || len(es) == 0 || es[0].Code != "string.email.invalid" {
		t.Fatalf("expected stable plugin code, got %v", err)
	}
	if es[0].Msg != "invalid email address" {
		t.Fatalf("expected plugin default translation, got %#v", es[0])
	}

	_ = verrs.CodeRequired
}

func TestRootFacade_OptionsAndContextHelpers(t *testing.T) {
	v := New()

	err := CheckTagWithOpts(v, "string;min=5;max=2", "abc", CompileOpts{CollectAll: true})
	var es Errors
	if !errors.As(err, &es) || len(es) != 2 {
		t.Fatalf("CheckTagWithOpts errors = %#v, want two structured errors", err)
	}

	ctxFn, err := FromTagContextWithOpts(v, "string;min=2", CompileOpts{})
	if err != nil {
		t.Fatalf("FromTagContextWithOpts returned error: %v", err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := ctxFn(canceled, "abc"); !errors.Is(err, context.Canceled) {
		t.Fatalf("context validator error = %v, want context.Canceled", err)
	}

	type Input struct {
		Name string `validate:"string;min=5;max=2"`
	}
	err = ValidateStructContextWithOpts(context.Background(), v, Input{Name: "abc"}, ValidateOpts{CollectAllRules: true})
	if !errors.As(err, &es) || len(es) != 2 {
		t.Fatalf("ValidateStructContextWithOpts errors = %#v, want two structured errors", err)
	}
}
