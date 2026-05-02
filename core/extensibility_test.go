package core

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/translator"
	"github.com/aatuh/validate/v3/types"
)

func TestWithRuleCompiler_IsPerInstance(t *testing.T) {
	base := New()
	custom := base.WithRuleCompiler("even", func(c *types.Compiler, rule types.Rule) (func(any) error, error) {
		return func(v any) error {
			n, ok := v.(int)
			if !ok || n%2 != 0 {
				return verrs.Errors{verrs.FieldError{Code: "number.even", Msg: "must be even"}}
			}
			return nil
		}, nil
	})

	if _, err := base.FromRules([]string{"string;even"}); err == nil {
		t.Fatalf("base instance should not compile unregistered custom rule")
	}
	fn := base.CompileRules([]types.Rule{types.NewRule("even", nil)})
	if err := fn(2); err == nil {
		t.Fatalf("base instance should not know per-instance compiler")
	}

	fn = custom.CompileRules([]types.Rule{types.NewRule("even", nil)})
	if err := fn(2); err != nil {
		t.Fatalf("custom compiler rejected even value: %v", err)
	}
	if err := fn(3); err == nil {
		t.Fatalf("custom compiler accepted odd value")
	}
}

func TestWithTypeValidator_IsPerInstanceAndPrefersLocalType(t *testing.T) {
	globalName := uniqueCoreTypeName(t, "global")
	types.RegisterGlobalType(globalName, coreStringTypeFactory{want: "global", code: "type.global"})

	base := New()
	local := base.WithTypeValidator(globalName, coreStringTypeFactory{want: "local", code: "type.local"})

	globalFn, err := base.FromRules([]string{globalName})
	if err != nil {
		t.Fatalf("global FromRules: %v", err)
	}
	requireCoreErrorCode(t, globalFn("local"), "type.global")
	if err := globalFn("global"); err != nil {
		t.Fatalf("global validator rejected valid value: %v", err)
	}

	localFn, err := local.FromRules([]string{globalName})
	if err != nil {
		t.Fatalf("local FromRules: %v", err)
	}
	requireCoreErrorCode(t, localFn("global"), "type.local")
	if err := localFn("local"); err != nil {
		t.Fatalf("local validator rejected valid value: %v", err)
	}

	localOnlyName := uniqueCoreTypeName(t, "local_only")
	isolated := base.WithTypeValidator(localOnlyName, coreStringTypeFactory{want: "ok", code: "type.localOnly"})
	if _, err := base.FromRules([]string{localOnlyName}); err == nil {
		t.Fatalf("base instance compiled per-instance custom type")
	}
	localOnlyFn, err := isolated.FromRules([]string{localOnlyName})
	if err != nil {
		t.Fatalf("isolated FromRules: %v", err)
	}
	if err := localOnlyFn("ok"); err != nil {
		t.Fatalf("isolated tag validator rejected valid value: %v", err)
	}

	manualFn := isolated.CompileRules([]types.Rule{types.NewRule(types.Kind(localOnlyName), nil)})
	requireCoreErrorCode(t, manualFn("bad"), "type.localOnly")
}

func TestCompileRulesE_ReturnsCustomCompilerErrors(t *testing.T) {
	compileErr := errors.New("compile failed")
	v := New().WithRuleCompiler("broken", func(c *types.Compiler, rule types.Rule) (func(any) error, error) {
		return nil, compileErr
	})

	if _, err := v.CompileRulesE([]types.Rule{types.NewRule("broken", nil)}); !errors.Is(err, compileErr) {
		t.Fatalf("CompileRulesE error = %v, want wrapped compile error", err)
	}

	fn := v.CompileRules([]types.Rule{types.NewRule("broken", nil)})
	if err := fn(123); !errors.Is(err, compileErr) {
		t.Fatalf("CompileRules validation error = %v, want wrapped compile error", err)
	}

	if _, err := v.FromRules([]string{"int;broken"}); !errors.Is(err, compileErr) {
		t.Fatalf("FromRules error = %v, want wrapped compile error", err)
	}

	if _, err := v.FromRules([]string{"slice;foreach=(int;broken)"}); !errors.Is(err, compileErr) {
		t.Fatalf("nested FromRules error = %v, want wrapped compile error", err)
	}
}

func TestGlobalRegistriesConcurrentAccess(t *testing.T) {
	const n = 50
	var wg sync.WaitGroup
	errs := make(chan error, n*2)
	for i := 0; i < n; i++ {
		i := i
		wg.Add(2)
		go func() {
			defer wg.Done()
			kind := types.Kind(fmt.Sprintf("concurrentRule%d", i))
			types.RegisterRule(kind, func(c *types.Compiler, rule types.Rule) (func(any) error, error) {
				return func(any) error { return nil }, nil
			})
			fn := types.NewCompiler(nil).Compile([]types.Rule{types.NewRule(kind, nil)})
			errs <- fn(nil)
		}()
		go func() {
			defer wg.Done()
			name := fmt.Sprintf("concurrentType%d", i)
			types.RegisterGlobalType(name, testTypeFactory{})
			if _, ok := types.GetGlobalTypeValidator(name, nil); !ok {
				errs <- errors.New("registered type missing")
				return
			}
			errs <- nil
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent registry access failed: %v", err)
		}
	}
}

type testTypeFactory struct{}

func (testTypeFactory) CreateValidator(_ translator.Translator) types.TypeValidator {
	return testTypeValidator{}
}

type testTypeValidator struct{}

func (testTypeValidator) Validate(any) error { return nil }

type coreStringTypeFactory struct {
	want string
	code string
}

func (f coreStringTypeFactory) CreateValidator(_ translator.Translator) types.TypeValidator {
	return coreStringTypeValidator{want: f.want, code: f.code}
}

type coreStringTypeValidator struct {
	want string
	code string
}

func (v coreStringTypeValidator) Validate(value any) error {
	if value == v.want {
		return nil
	}
	return verrs.Errors{verrs.FieldError{Code: v.code}}
}

func requireCoreErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil {
		t.Fatalf("got nil error, want %q", code)
	}
	var es verrs.Errors
	if !errors.As(err, &es) || len(es) == 0 {
		t.Fatalf("got %T %v, want structured error", err, err)
	}
	if es[0].Code != code {
		t.Fatalf("code = %q, want %q; errors=%#v", es[0].Code, code, es)
	}
}

func uniqueCoreTypeName(t *testing.T, suffix string) string {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	return "audit_" + name + "_" + suffix
}
