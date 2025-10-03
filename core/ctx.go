package core

import "context"

// CheckFunc validates a single value and returns an error if invalid.
type CheckFunc func(v any) error

// CheckFuncCtx is the context-aware variant of CheckFunc.
type CheckFuncCtx func(ctx context.Context, v any) error

// WithContext adapts a CheckFunc to a CheckFuncCtx that ignores ctx.
func WithContext(f CheckFunc) CheckFuncCtx {
	if f == nil {
		return nil
	}
	return func(ctx context.Context, v any) error { return f(v) }
}

// WithoutContext adapts a CheckFuncCtx to a CheckFunc.
func WithoutContext(f CheckFuncCtx) CheckFunc {
	if f == nil {
		return nil
	}
	return func(v any) error { return f(context.Background(), v) }
}
