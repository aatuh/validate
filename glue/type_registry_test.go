package glue

import (
	"errors"
	"strings"
	"testing"

	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/translator"
	"github.com/aatuh/validate/v3/types"
)

func TestValidate_WithTypeValidatorSupportsTagsAndBuilder(t *testing.T) {
	typeName := uniqueGlueTypeName(t)
	base := New()
	v := base.WithTypeValidator(typeName, glueStringTypeFactory{want: "ok", code: "type.local"})

	if _, err := base.FromTag(typeName); err == nil {
		t.Fatalf("base instance compiled per-instance custom type")
	}

	tagFn, err := v.FromTag(typeName)
	if err != nil {
		t.Fatalf("FromTag returned error: %v", err)
	}
	if err := tagFn("ok"); err != nil {
		t.Fatalf("tag validator rejected valid value: %v", err)
	}
	requireGlueErrorCode(t, tagFn("bad"), "type.local")

	builderFn := v.CustomType(typeName).Build()
	if err := builderFn("ok"); err != nil {
		t.Fatalf("builder validator rejected valid value: %v", err)
	}
	requireGlueErrorCode(t, builderFn("bad"), "type.local")
}

func TestValidate_WithTypeValidatorSupportsNestedCollectionTags(t *testing.T) {
	typeName := uniqueGlueTypeName(t)
	base := New()
	v := base.WithTypeValidator(typeName, glueStringTypeFactory{want: "ok", code: "type.local"})

	for _, tag := range []string{
		"slice;foreach=(" + typeName + ")",
		"map;keys=(" + typeName + ")",
		"map;values=(" + typeName + ")",
	} {
		t.Run(tag, func(t *testing.T) {
			if _, err := base.FromTag(tag); err == nil {
				t.Fatalf("base instance compiled per-instance custom type in %q", tag)
			}
			fn, err := v.FromTag(tag)
			if err != nil {
				t.Fatalf("FromTag(%q): %v", tag, err)
			}

			var valid, invalid any
			switch {
			case strings.HasPrefix(tag, "slice"):
				valid = []string{"ok"}
				invalid = []string{"bad"}
			case strings.Contains(tag, "keys="):
				valid = map[string]string{"ok": "ignored"}
				invalid = map[string]string{"bad": "ignored"}
			default:
				valid = map[string]string{"id": "ok"}
				invalid = map[string]string{"id": "bad"}
			}

			if err := fn(valid); err != nil {
				t.Fatalf("tag validator rejected valid value: %v", err)
			}
			requireGlueErrorCode(t, fn(invalid), "type.local")
		})
	}

	manualFn := v.CompileRules([]types.Rule{
		types.NewRule(types.KSlice, nil),
		types.NewRule(types.KForEach, map[string]any{
			"rules": []types.Rule{types.NewRule(types.Kind(typeName), nil)},
		}),
	})
	requireGlueErrorCode(t, manualFn([]string{"bad"}), "type.local")

	builderFn := v.Map().
		ValuesRules(types.NewRule(types.Kind(typeName), nil)).
		Build()
	requireGlueErrorCode(t, builderFn(map[string]string{"id": "bad"}), "type.local")
}

type glueStringTypeFactory struct {
	want string
	code string
}

func (f glueStringTypeFactory) CreateValidator(_ translator.Translator) types.TypeValidator {
	return glueStringTypeValidator{want: f.want, code: f.code}
}

type glueStringTypeValidator struct {
	want string
	code string
}

func (v glueStringTypeValidator) Validate(value any) error {
	if value == v.want {
		return nil
	}
	return verrs.Errors{verrs.FieldError{Code: v.code}}
}

func requireGlueErrorCode(t *testing.T, err error, code string) {
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

func uniqueGlueTypeName(t *testing.T) string {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	return "audit_" + name
}
