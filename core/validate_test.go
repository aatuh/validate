package core

import (
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3/translator"
)

type keyEchoTr struct{}

// T returns "<key> <params...>" for deterministic assertions.
func (keyEchoTr) T(key string, params ...any) string {
	if len(params) == 0 {
		return key
	}
	return key + " " + fmt.Sprint(params...)
}

func TestFromRules_KnownTypes(t *testing.T) {
	tr := translator.NewSimpleTranslator(
		translator.DefaultEnglishTranslations(),
	)

	v := New().WithTranslator(tr)

	// string rules
	sf, err := v.FromRules([]string{"string", "min=2", "max=4"})
	if err != nil {
		t.Fatalf("string build: %v", err)
	}
	if err := sf("a"); err == nil {
		t.Fatalf("want min error")
	}
	if err := sf("abcd"); err != nil {
		t.Fatalf("want ok, got %v", err)
	}
	if err := sf("abcde"); err == nil {
		t.Fatalf("want max error")
	}

	// int rules
	ifn, err := v.FromRules([]string{"int", "min=1", "max=3"})
	if err != nil {
		t.Fatalf("int build: %v", err)
	}
	if err := ifn(int(0)); err == nil {
		t.Fatalf("int min fail expected")
	}
	if err := ifn(int64(2)); err != nil {
		t.Fatalf("int pass expected, got %v", err)
	}

	// bool rules
	bf, err := v.FromRules([]string{"bool"})
	if err != nil {
		t.Fatalf("bool build: %v", err)
	}
	if err := bf(true); err != nil {
		t.Fatalf("bool true should pass: %v", err)
	}
	if err := bf("nope"); err == nil {
		t.Fatalf("non-bool should fail")
	}

	// slice rules
	sf2, err := v.FromRules([]string{"slice", "min=1"})
	if err != nil {
		t.Fatalf("slice build: %v", err)
	}
	if err := sf2([]any{1}); err != nil {
		t.Fatalf("[]any should pass, got %v", err)
	}
	if err := sf2(123); err == nil {
		t.Fatalf("non-slice input should fail")
	}
}

func TestFromRules_ErrorsAndCustom(t *testing.T) {
	v := NewWithCustomRules(map[string]func(any) error{
		"custom": func(a any) error { return nil },
	})

	// Custom hit.
	fn, err := v.FromRules([]string{"custom"})
	if err != nil {
		t.Fatalf("custom build: %v", err)
	}
	if err := fn(nil); err != nil {
		t.Fatalf("custom exec: %v", err)
	}

	// Unknown validator type.
	if _, err := v.FromRules([]string{"nope"}); err == nil {
		t.Fatalf("want unknown type error")
	}

	// Builder errors bubble up (bad int param).
	if _, err := v.FromRules([]string{"int", "min=abc"}); err == nil {
		t.Fatalf("want builder parse error")
	}

	// Regex invalid pattern returns function that errors on use.
	tr := translator.NewSimpleTranslator(
		translator.DefaultEnglishTranslations(),
	)
	v = v.WithTranslator(tr)
	fn2, err := v.FromRules([]string{"string", "regex=("})
	if err != nil {
		t.Fatalf("regex invalid pattern should not error at build: %v", err)
	}
	if err := fn2("anything"); err == nil {
		t.Fatalf("invalid pattern must error on use")
	}
}

func TestPathSeparator_Set_And_IgnoreEmpty(t *testing.T) {
	v := New()
	if v.pathSep != "." {
		t.Fatalf("default pathSep should be '.', got %q", v.pathSep)
	}
	// Set to a custom separator.
	v = v.PathSeparator(":")
	if v.pathSep != ":" {
		t.Fatalf("PathSeparator did not take effect, got %q", v.pathSep)
	}
	// Empty string should be ignored (remain the same).
	v = v.PathSeparator("")
	if v.pathSep != ":" {
		t.Fatalf("empty PathSeparator should be ignored, got %q", v.pathSep)
	}
}
