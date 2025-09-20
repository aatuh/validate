package structvalidator

import (
	"errors"
	"strings"
	"testing"

	"github.com/aatuh/validate"
)

type dummyTr struct{}

func (dummyTr) T(key string, params ...any) string { return key }

type Profile struct {
	Email string `validate:"string;email"`
}

type User struct {
	Name    string   `validate:"string;min=2"`
	Age     int      `validate:"int;min=1"`
	Tags    []string `validate:"slice;min=1"`
	Profile Profile
}

func TestStruct_Basic_Aggregate(t *testing.T) {
	v := validate.New().WithTranslator(dummyTr{})
	sv := NewStructValidator(v)

	u := User{
		Name:    "A",
		Age:     0,
		Tags:    []string{},
		Profile: Profile{Email: "no-at-symbol"},
	}
	err := sv.ValidateStruct(u)
	if err == nil {
		t.Fatalf("want aggregated errors")
	}
	got := err.Error()
	wantSubs := []string{
		// builder messages are translator keys here.
		"string.minLength",
		"int.min",
		"slice.min",
		"string.email",
	}
	for _, w := range wantSubs {
		if !strings.Contains(got, w) {
			t.Fatalf("want substring %q in %q", w, got)
		}
	}
}

func TestStruct_StopOnFirst_And_PathSep(t *testing.T) {
	v := validate.New().WithTranslator(dummyTr{}).PathSeparator(":")
	sv := NewStructValidator(v)

	u := struct {
		A string `validate:"string;min=2"`
		B string `validate:"string;min=3"`
	}{A: "", B: ""}

	// Stop on first should report only A.
	err := sv.ValidateStructWithOpts(u, validate.ValidateOpts{
		StopOnFirst: true,
	})
	if err == nil {
		t.Fatalf("want error")
	}
	if strings.Count(err.Error(), ";") > 0 {
		t.Fatalf("want single error, got %q", err.Error())
	}
	// Ensure PathSep applied with nested struct.
	type N struct {
		X string `validate:"string;min=2"`
	}
	type R struct{ N N }
	r := R{N: N{X: ""}}
	err = sv.ValidateStruct(r)
	if err == nil {
		t.Fatalf("want error")
	}
	if !strings.Contains(err.Error(), "N:X") {
		t.Fatalf("want custom sep in path, got %q", err.Error())
	}
}

func TestStruct_NonStruct(t *testing.T) {
	v := validate.New().WithTranslator(dummyTr{})
	sv := NewStructValidator(v)
	err := sv.ValidateStruct(42)
	if err == nil {
		t.Fatalf("want error for non-struct")
	}
	if !strings.Contains(err.Error(), "expected struct") {
		t.Fatalf("unexpected message: %v", err)
	}
}

func TestStruct_SliceOfStructs_Recurse(t *testing.T) {
	v := validate.New().WithTranslator(dummyTr{})
	sv := NewStructValidator(v)

	type Item struct {
		Code string `validate:"string;min=2"`
	}
	type Basket struct {
		Items []Item
	}
	b := Basket{Items: []Item{{Code: ""}, {Code: "ok"}}}
	err := sv.ValidateStruct(b)
	if err == nil {
		t.Fatalf("want error in Items[0].Code")
	}
	if !strings.Contains(err.Error(), "Items[0]") {
		t.Fatalf("want index in path, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "string.minLength") {
		t.Fatalf("want string.minLength key")
	}
}

func TestStruct_MapOfStructs_Recurse(t *testing.T) {
	v := validate.New().WithTranslator(dummyTr{})
	sv := NewStructValidator(v)

	type Item struct {
		Code string `validate:"string;min=2"`
	}
	type Bag struct {
		M map[string]Item
	}
	b := Bag{M: map[string]Item{"k1": {Code: ""}}}
	err := sv.ValidateStruct(b)
	if err == nil {
		t.Fatalf("want error")
	}
	if !strings.Contains(err.Error(), "M[k1]") {
		t.Fatalf("want map key in path, got %q", err.Error())
	}
}

func TestStruct_OK(t *testing.T) {
	v := validate.New().WithTranslator(dummyTr{})
	sv := NewStructValidator(v)

	ok := User{
		Name:    "Ok",
		Age:     2,
		Tags:    []string{"x"},
		Profile: Profile{Email: "u@x.com"},
	}
	if err := sv.ValidateStruct(ok); err != nil {
		t.Fatalf("unexpected err %v", err)
	}
}

// guard unused import errors for "errors" on some Go versions.
var _ = errors.New
