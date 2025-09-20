package validate

import (
	"context"
	"testing"
)

func TestWithContext_And_WithoutContext(t *testing.T) {
	// CheckFunc adapted to ctx.
	calls := 0
	base := func(v any) error {
		calls++
		return nil
	}
	cf := WithContext(base)
	if err := cf(context.Background(), 123); err != nil {
		t.Fatalf("WithContext exec err: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}

	// CheckFuncCtx adapted to non-ctx.
	calls = 0
	baseCtx := func(ctx context.Context, v any) error {
		calls++
		return nil
	}
	cf2 := WithoutContext(baseCtx)
	if err := cf2("x"); err != nil {
		t.Fatalf("WithoutContext exec err: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestApplyOpts_And_WithDefaults(t *testing.T) {
	// With nil validator, default PathSep should be ".".
	o := ApplyOpts(nil, ValidateOpts{})
	if o.PathSep != "." {
		t.Fatalf("default PathSep '.' expected, got %q", o.PathSep)
	}

	// With validator that has a custom separator.
	v := New().PathSeparator(":")
	o2 := ApplyOpts(v, ValidateOpts{})
	if o2.PathSep != ":" {
		t.Fatalf("want PathSep ':', got %q", o2.PathSep)
	}

	// Explicit PathSep in opts should not be overridden.
	o3 := ApplyOpts(v, ValidateOpts{PathSep: "|"})
	if o3.PathSep != "|" {
		t.Fatalf("explicit PathSep should be kept, got %q", o3.PathSep)
	}

	// WithDefaults currently a no-op; call to cover.
	_ = ValidateOpts{}.WithDefaults()
}
