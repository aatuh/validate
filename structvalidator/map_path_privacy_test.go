package structvalidator

import (
	"errors"
	"strings"
	"testing"

	"github.com/aatuh/validate/v3/core"
	verrs "github.com/aatuh/validate/v3/errors"
)

func TestStruct_MapKeyPathPreservesShortOrdinaryKeys(t *testing.T) {
	v := core.New().WithTranslator(dummyTr{})
	sv := NewStructValidator(v)

	err := sv.ValidateStructWithOpts(mapPrivacyInput{
		Items: map[string]mapPrivacyItem{"id": {Code: ""}},
	}, core.ValidateOpts{FieldNameFunc: JSONFieldName})

	es := requireStructMapPrivacyErrors(t, err)
	if len(es) != 1 {
		t.Fatalf("errors = %#v, want one error", es)
	}
	if es[0].Path != "items[id].code" {
		t.Fatalf("path = %q, want short key path", es[0].Path)
	}
	if es[0].Code != verrs.CodeStringMin {
		t.Fatalf("code = %q, want %q", es[0].Code, verrs.CodeStringMin)
	}
}

func TestStruct_MapKeyPathRedactsLongAndSensitiveKeys(t *testing.T) {
	v := core.New().WithTranslator(dummyTr{})
	sv := NewStructValidator(v)

	longKey := strings.Repeat("a", 96)
	sensitiveKey := "session_token=secret-value"

	for _, raw := range []string{longKey, sensitiveKey} {
		t.Run(raw[:8], func(t *testing.T) {
			err := sv.ValidateStructWithOpts(mapPrivacyInput{
				Items: map[string]mapPrivacyItem{raw: {Code: ""}},
			}, core.ValidateOpts{FieldNameFunc: JSONFieldName})

			es := requireStructMapPrivacyErrors(t, err)
			if len(es) != 1 {
				t.Fatalf("errors = %#v, want one error", es)
			}
			if es[0].Path != "items[<redacted>].code" {
				t.Fatalf("path = %q, want redacted map key path", es[0].Path)
			}
			if strings.Contains(es[0].Path, raw) || strings.Contains(es.Error(), raw) {
				t.Fatalf("error leaked raw map key %q: %#v", raw, es)
			}
			if es[0].Code != verrs.CodeStringMin {
				t.Fatalf("code = %q, want %q", es[0].Code, verrs.CodeStringMin)
			}
		})
	}
}

type mapPrivacyInput struct {
	Items map[string]mapPrivacyItem `json:"items"`
}

type mapPrivacyItem struct {
	Code string `json:"code" validate:"string;min=2"`
}

func requireStructMapPrivacyErrors(t *testing.T, err error) verrs.Errors {
	t.Helper()
	if err == nil {
		t.Fatalf("got nil error, want structured errors")
	}
	var es verrs.Errors
	if !errors.As(err, &es) {
		t.Fatalf("got %T %v, want structured errors", err, err)
	}
	return es
}
