package structvalidator

import (
	"context"
	"errors"
	"testing"

	"github.com/aatuh/validate/v3/core"
	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/types"
)

func TestStruct_RequiredIfAndRequiredUnless(t *testing.T) {
	v := core.New().WithTranslator(dummyTr{})
	sv := NewStructValidator(v)

	type Input struct {
		Kind      string `json:"kind"`
		Company   string `json:"company" validate:"string;requiredIf=Kind,business"`
		FirstName string `json:"first_name" validate:"string;requiredUnless=Kind,business"`
	}

	err := sv.ValidateStructWithOpts(Input{Kind: "business"}, core.ValidateOpts{FieldNameFunc: JSONFieldName})
	requireStructFieldError(t, err, "company", verrs.CodeRequiredIf, nil)

	err = sv.ValidateStructWithOpts(Input{Kind: "personal"}, core.ValidateOpts{FieldNameFunc: JSONFieldName})
	requireStructFieldError(t, err, "first_name", verrs.CodeRequiredUnless, nil)

	if err := sv.ValidateStructWithOpts(Input{Kind: "business", Company: "Acme"}, core.ValidateOpts{FieldNameFunc: JSONFieldName}); err != nil {
		t.Fatalf("valid business input failed: %v", err)
	}
	if err := sv.ValidateStructWithOpts(Input{Kind: "personal", FirstName: "Ada"}, core.ValidateOpts{FieldNameFunc: JSONFieldName}); err != nil {
		t.Fatalf("valid personal input failed: %v", err)
	}
}

func TestStruct_RequiredIfPointerReferencesAndMalformedTags(t *testing.T) {
	v := core.New().WithTranslator(dummyTr{})
	sv := NewStructValidator(v)

	active := "active"
	type Input struct {
		Status *string `validate:"string"`
		Token  *string `validate:"string;requiredIf=Status,active"`
	}
	err := sv.ValidateStruct(Input{Status: &active})
	requireStructFieldError(t, err, "Token", verrs.CodeRequiredIf, nil)

	type MissingReference struct {
		Value string `validate:"string;requiredIf=Missing,yes"`
	}
	requireStructFieldError(t, sv.ValidateStruct(MissingReference{}), "Value", verrs.CodeFieldReference, "Missing")

	type Malformed struct {
		Value string `validate:"string;requiredIf=Kind"`
	}
	requireStructFieldError(t, sv.ValidateStruct(Malformed{}), "Value", verrs.CodeUnknown, nil)
}

func TestStruct_RequiredIfDereferencesInterfacesAndShortCircuits(t *testing.T) {
	v := core.New().WithTranslator(dummyTr{})
	sv := NewStructValidator(v)

	status := "active"
	type Input struct {
		Status any    `validate:""`
		Token  string `validate:"string;requiredIf=Status,active;min=10"`
	}

	err := sv.ValidateStructWithOpts(Input{Status: &status}, core.ValidateOpts{CollectAllRules: true})
	assertStructCodes(t, err, []string{verrs.CodeRequiredIf})
}

func TestStruct_CollectAllRulesAndContext(t *testing.T) {
	v := core.New().WithTranslator(dummyTr{})
	sv := NewStructValidator(v)

	type Input struct {
		Name string `validate:"string;min=5;max=2"`
	}
	err := sv.ValidateStructWithOpts(Input{Name: "abc"}, core.ValidateOpts{CollectAllRules: true})
	assertStructCodes(t, err, []string{verrs.CodeStringMin, verrs.CodeStringMax})

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sv.ValidateStructContext(canceled, Input{Name: "abc"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("ValidateStructContext error = %v, want context.Canceled", err)
	}
}

func TestStruct_CustomRuleReceivesContext(t *testing.T) {
	type ctxKey string
	const key ctxKey = "allow"

	v := core.New().WithStructRuleCompiler("ctxAllowed", func(rule types.Rule) (core.StructRuleFunc, error) {
		return func(ctx core.StructRuleContext) error {
			if ctx.Context.Value(key) == true {
				return nil
			}
			return verrs.Errors{verrs.FieldError{Code: "field.context"}}
		}, nil
	})
	sv := NewStructValidator(v)

	type Input struct {
		Name string `validate:"string;struct:ctxAllowed"`
	}
	if err := sv.ValidateStructContext(context.WithValue(context.Background(), key, true), Input{Name: "Ada"}); err != nil {
		t.Fatalf("context-aware struct rule rejected allowed context: %v", err)
	}
	requireStructFieldError(t, sv.ValidateStructContext(context.Background(), Input{Name: "Ada"}), "Name", "field.context", nil)
}

func assertStructCodes(t *testing.T, err error, want []string) {
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
