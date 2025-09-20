package validators

import (
	"fmt"
	"strings"
	"testing"
)

type dummyTr struct{}

// Include params in the returned string so tests can assert on indexes, etc.
func (dummyTr) T(key string, params ...any) string {
	if len(params) == 0 {
		return key
	}
	return key + " " + fmt.Sprint(params...)
}

func TestString_Length_Min_Max(t *testing.T) {
	sv := NewStringValidators(dummyTr{})
	fn := sv.WithString(
		sv.Length(3),
	)
	if err := fn("ab"); err == nil {
		t.Fatalf("want length=3 error")
	}
	if err := fn("abcd"); err == nil {
		t.Fatalf("want length=3 error")
	}
	if err := fn("abc"); err != nil {
		t.Fatalf("got %v", err)
	}

	fn2 := sv.WithString(sv.MinLength(2), sv.MaxLength(4))
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"", false},
		{"a", false},
		{"ab", true},
		{"abcd", true},
		{"abcde", false},
	} {
		err := fn2(tc.in)
		if tc.want && err != nil {
			t.Fatalf("input=%q unexpected err %v", tc.in, err)
		}
		if !tc.want && err == nil {
			t.Fatalf("input=%q want error", tc.in)
		}
	}
}

func TestString_OneOf_Email_Regex(t *testing.T) {
	sv := NewStringValidators(dummyTr{})

	one := sv.WithString(sv.OneOf("red", "green", "blue"))
	if err := one("Green"); err != nil {
		t.Fatalf("oneof should pass case-insensitively")
	}
	if err := one("yellow"); err == nil {
		t.Fatalf("oneof should fail for yellow")
	}

	email := sv.WithString(sv.Email())
	if err := email("user@example.com"); err != nil {
		t.Fatalf("valid email got %v", err)
	}
	// Use length 255 to exceed 254-limit and ensure failure.
	long := strings.Repeat("a", 249) + "@x.com" // 249 + 6 = 255
	if err := email(long); err == nil {
		t.Fatalf("email length should fail")
	}

	// Regex: invalid pattern is handled and always errors on use.
	bad := sv.WithString(sv.Regex("("))
	if err := bad("anything"); err == nil {
		t.Fatalf("invalid regex should error at use")
	}

	// Regex: valid noMatch path.
	re := sv.WithString(sv.Regex("^a.*z$"))
	if err := re("hello"); err == nil {
		t.Fatalf("regex should not match")
	}
	if err := re("abcz"); err != nil {
		t.Fatalf("regex should match, got %v", err)
	}
}

func TestBuildStringValidator_FromTokens(t *testing.T) {
	sv := NewStringValidators(dummyTr{})
	fn, err := BuildStringValidator(sv, []string{
		"string", "min=2", "max=3",
	})
	if err != nil {
		t.Fatalf("build err %v", err)
	}
	for _, tc := range []struct {
		in   string
		want bool
	}{
		{"a", false},
		{"ab", true},
		{"abc", true},
		{"abcd", false},
	} {
		err := fn(tc.in)
		if tc.want && err != nil {
			t.Fatalf("input=%q unexpected err %v", tc.in, err)
		}
		if !tc.want && err == nil {
			t.Fatalf("input=%q want error", tc.in)
		}
	}
}

type hasString string

func (h hasString) String() string { return "S" }

func TestString_toString_StringerAndError(t *testing.T) {
	sv := NewStringValidators(nil)

	// With no inner rules, WithString should still call toString and pass.
	fn := sv.WithString()
	if err := fn(hasString("ok")); err != nil {
		t.Fatalf("Stringer should be accepted: %v", err)
	}

	// Non-string and non-Stringer should fail.
	if err := fn(123.45); err == nil {
		t.Fatalf("non-string should error")
	}
}

func TestString_Translator_Nil_Branch(t *testing.T) {
	// Nil translator triggers fallback "key: params" path.
	sv := NewStringValidators(nil)
	fn := sv.WithString(sv.MinLength(2))
	if err := fn("a"); err == nil {
		t.Fatalf("min length should fail")
	}
}
