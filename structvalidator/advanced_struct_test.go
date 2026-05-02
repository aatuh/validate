package structvalidator

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/aatuh/validate/v3/core"
	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/types"
)

func TestStruct_TaggedPointersRequiredAndOmitEmpty(t *testing.T) {
	v := core.New().WithTranslator(dummyTr{})
	sv := NewStructValidator(v)

	type Input struct {
		Name *string `validate:"string;required;min=2"`
		Nick *string `validate:"string;omitempty;min=2"`
	}

	err := sv.ValidateStruct(Input{})
	if err == nil {
		t.Fatalf("nil required pointer should fail")
	}
	var es verrs.Errors
	if !errors.As(err, &es) || len(es) == 0 || es[0].Code != verrs.CodeRequired {
		t.Fatalf("expected required code, got %v", err)
	}

	name := "Al"
	if err := sv.ValidateStruct(Input{Name: &name}); err != nil {
		t.Fatalf("valid pointer input failed: %v", err)
	}
}

func TestStruct_DeterministicMapTraversalAndJSONFieldNames(t *testing.T) {
	v := core.New().WithTranslator(dummyTr{})
	sv := NewStructValidator(v)

	type Item struct {
		Code string `json:"code" validate:"string;min=2"`
	}
	type Input struct {
		Items map[string]Item `json:"items"`
	}

	in := Input{Items: map[string]Item{
		"b": {Code: ""},
		"a": {Code: ""},
	}}
	err := sv.ValidateStructWithOpts(in, core.ValidateOpts{FieldNameFunc: JSONFieldName})
	if err == nil {
		t.Fatalf("want validation errors")
	}
	var es verrs.Errors
	if !errors.As(err, &es) {
		t.Fatalf("expected structured errors, got %T", err)
	}
	got := []string{es[0].Path, es[1].Path}
	want := []string{"items[a].code", "items[b].code"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %#v, want %#v", got, want)
	}
}

func TestStruct_CrossFieldRules(t *testing.T) {
	v := core.New().WithTranslator(dummyTr{})
	sv := NewStructValidator(v)

	type Input struct {
		Password string `validate:"string;required"`
		Confirm  string `validate:"string;eqField=Password"`
		Token    string `validate:"string;requiredWith=Password"`
		Other    string `validate:"string;neField=Password"`
	}

	err := sv.ValidateStruct(Input{Password: "secret", Confirm: "mismatch", Other: "secret"})
	if err == nil {
		t.Fatalf("want cross-field errors")
	}
	got := err.Error()
	for _, code := range []string{verrs.CodeFieldEqual, verrs.CodeRequiredWith, verrs.CodeFieldNotEqual} {
		if !strings.Contains(got, code) {
			t.Fatalf("expected %s in %q", code, got)
		}
	}

	if err := sv.ValidateStruct(Input{Password: "secret", Confirm: "secret", Token: "token", Other: "different"}); err != nil {
		t.Fatalf("valid cross-field input failed: %v", err)
	}
}

func TestStruct_InvalidCrossFieldReferences(t *testing.T) {
	v := core.New().WithTranslator(dummyTr{})
	sv := NewStructValidator(v)

	t.Run("eqField missing with JSON field name and pointer owner", func(t *testing.T) {
		type Input struct {
			Password string `json:"password"`
			Confirm  string `json:"confirm_password" validate:"string;eqField=Missing"`
		}

		err := sv.ValidateStructWithOpts(&Input{Password: "secret", Confirm: "secret"}, core.ValidateOpts{
			FieldNameFunc: JSONFieldName,
		})
		requireStructFieldError(t, err, "confirm_password", verrs.CodeFieldReference, "Missing")
	})

	t.Run("neField unexported reference", func(t *testing.T) {
		type Input struct {
			secret string
			Public string `validate:"string;neField=secret"`
		}

		err := sv.ValidateStruct(Input{secret: "hidden", Public: "shown"})
		requireStructFieldError(t, err, "Public", verrs.CodeFieldReference, "secret")
	})

	t.Run("requiredWith missing reference", func(t *testing.T) {
		type Input struct {
			Value string `validate:"string;requiredWith=Missing"`
		}

		err := sv.ValidateStruct(Input{})
		requireStructFieldError(t, err, "Value", verrs.CodeFieldReference, "Missing")
	})
}

func TestStruct_CustomStructRuleFieldValueMissingOrInaccessible(t *testing.T) {
	var sawMissing, sawUnexported bool

	v := core.New().WithTranslator(dummyTr{}).WithStructRuleCompiler("requiresField", func(rule types.Rule) (core.StructRuleFunc, error) {
		fieldName, _ := rule.Args["value"].(string)
		return func(ctx core.StructRuleContext) error {
			if _, ok := ctx.FieldValue(fieldName); ok {
				return nil
			}
			switch fieldName {
			case "Missing":
				sawMissing = true
			case "secret":
				sawUnexported = true
			}
			return verrs.Errors{verrs.FieldError{Code: "field.customReference", Param: fieldName}}
		}, nil
	})
	sv := NewStructValidator(v)

	type MissingInput struct {
		Name string `validate:"string;struct:requiresField=Missing"`
	}
	type UnexportedInput struct {
		secret string
		Name   string `validate:"string;struct:requiresField=secret"`
	}

	requireStructFieldError(t, sv.ValidateStruct(MissingInput{Name: "Ada"}), "Name", "field.customReference", "Missing")
	requireStructFieldError(t, sv.ValidateStruct(UnexportedInput{secret: "hidden", Name: "Ada"}), "Name", "field.customReference", "secret")
	if !sawMissing || !sawUnexported {
		t.Fatalf("custom rule did not observe missing=%v unexported=%v", sawMissing, sawUnexported)
	}
}

func TestStruct_CustomStructRuleCompiler(t *testing.T) {
	var sawPath, sawTranslator, sawOtherField bool

	v := core.New().WithTranslator(dummyTr{}).WithStructRuleCompiler("matchesField", func(rule types.Rule) (core.StructRuleFunc, error) {
		fieldName, _ := rule.Args["value"].(string)
		return func(ctx core.StructRuleContext) error {
			sawPath = ctx.Path == "Confirm"
			sawTranslator = ctx.Translator != nil
			other, ok := ctx.FieldValue(fieldName)
			sawOtherField = ok && other == "alpha12345"
			if ctx.Value != other {
				return verrs.Errors{verrs.FieldError{Code: "field.matches", Msg: "must match"}}
			}
			return nil
		}, nil
	})
	sv := NewStructValidator(v)

	type Input struct {
		Password string `validate:"string;required"`
		Confirm  string `validate:"string;struct:matchesField=Password"`
	}

	err := sv.ValidateStruct(Input{Password: "alpha12345", Confirm: "mismatch"})
	if err == nil {
		t.Fatalf("want custom struct rule error")
	}
	var es verrs.Errors
	if !errors.As(err, &es) || len(es) != 1 {
		t.Fatalf("expected one structured error, got %v", err)
	}
	if es[0].Path != "Confirm" || es[0].Code != "field.matches" {
		t.Fatalf("error = %#v, want path Confirm and code field.matches", es[0])
	}
	if !sawPath || !sawTranslator || !sawOtherField {
		t.Fatalf("custom struct context incomplete: path=%v translator=%v other=%v", sawPath, sawTranslator, sawOtherField)
	}

	if err := sv.ValidateStruct(Input{Password: "alpha12345", Confirm: "alpha12345"}); err != nil {
		t.Fatalf("valid custom struct rule input failed: %v", err)
	}
}

func requireStructFieldError(t *testing.T, err error, path, code string, param any) {
	t.Helper()
	if err == nil {
		t.Fatalf("got nil error, want path %q code %q", path, code)
	}
	var es verrs.Errors
	if !errors.As(err, &es) || len(es) == 0 {
		t.Fatalf("got %T %v, want structured errors", err, err)
	}
	for _, fe := range es {
		if fe.Path == path && fe.Code == code && reflect.DeepEqual(fe.Param, param) {
			return
		}
	}
	t.Fatalf("errors = %#v, want path %q code %q param %#v", es, path, code, param)
}
