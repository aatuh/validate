package validate

import (
	"testing"

	"github.com/aatuh/validate/translator"
)

type echoTr struct{}

func (echoTr) T(key string, params ...any) string { return key }

func TestBuilders_String_Int_Int64_Slice_Bool(t *testing.T) {
	tr := translator.NewSimpleTranslator(
		translator.DefaultEnglishTranslations(),
	)
	v := New().WithTranslator(tr)

	// String builder: MinLength + MaxLength chain.
	sv := v.String().MinLength(2).MaxLength(4).Build()
	if err := sv("a"); err == nil {
		t.Fatalf("string min should fail")
	}
	if err := sv("abcd"); err != nil {
		t.Fatalf("string within bounds should pass: %v", err)
	}
	if err := sv("abcde"); err == nil {
		t.Fatalf("string max should fail")
	}

	// String builder: Regex only.
	svRegex := v.String().Regex("^a.*z$").Build()
	if err := svRegex("abcz"); err != nil {
		t.Fatalf("regex match should pass: %v", err)
	}
	if err := svRegex("hello"); err == nil {
		t.Fatalf("regex non-match should fail")
	}

	// String builder: Email only (terminal rule).
	emailFn := v.String().Email()
	if err := emailFn("user@example.com"); err != nil {
		t.Fatalf("valid email should pass: %v", err)
	}
	if err := emailFn("not-an-email"); err == nil {
		t.Fatalf("invalid email should fail")
	}

	// Int builder: generic WithInt accepts all int kinds.
	iv := v.Int().MinInt(1).MaxInt(3).Build()
	if err := iv(int(0)); err == nil {
		t.Fatalf("int min should fail")
	}
	if err := iv(int64(2)); err != nil {
		t.Fatalf("int in-range should pass: %v", err)
	}
	if err := iv(int32(4)); err == nil {
		t.Fatalf("int max should fail")
	}

	// Int64 builder: requires exactly int64.
	iv64 := v.Int64().MinInt(0).Build()
	if err := iv64(int(1)); err == nil {
		t.Fatalf("int should not be accepted by Int64 builder")
	}
	if err := iv64(int64(1)); err != nil {
		t.Fatalf("int64 should be accepted: %v", err)
	}

	// Slice builder: ForEach string MinLength(1) over []string values.
	elem := v.String().MinLength(1).Build()
	sl := v.Slice().ForEach(elem).Build()
	if err := sl([]string{"ok", ""}); err == nil {
		t.Fatalf("second element empty should fail")
	}
	if err := sl([]string{"a", "b", "c"}); err != nil {
		t.Fatalf("all non-empty should pass: %v", err)
	}

	// Bool builder.
	bf := v.Bool()
	if err := bf(true); err != nil {
		t.Fatalf("bool true should pass: %v", err)
	}
	if err := bf("nope"); err == nil {
		t.Fatalf("non-bool should fail")
	}
}

func TestStringBuilder_Length_And_OneOf(t *testing.T) {
	tr := translator.NewSimpleTranslator(
		translator.DefaultEnglishTranslations(),
	)
	v := New().WithTranslator(tr)

	// Exact length rule.
	fn := v.String().Length(3).Build()
	if err := fn("ab"); err == nil {
		t.Fatalf("length=3 should fail for 'ab'")
	}
	if err := fn("abcd"); err == nil {
		t.Fatalf("length=3 should fail for 'abcd'")
	}
	if err := fn("abc"); err != nil {
		t.Fatalf("length=3 should pass for 'abc': %v", err)
	}

	// OneOf rule.
	fn2 := v.String().OneOf("red", "green").Build()
	if err := fn2("green"); err != nil {
		t.Fatalf("oneof should pass: %v", err)
	}
	if err := fn2("blue"); err == nil {
		t.Fatalf("oneof should fail for blue")
	}
}

func TestSliceBuilder_Length_Min_Max_And_ForEach(t *testing.T) {
	tr := translator.NewSimpleTranslator(
		translator.DefaultEnglishTranslations(),
	)
	v := New().WithTranslator(tr)

	// Exact slice length.
	lfn := v.Slice().Length(2).Build()
	if err := lfn([]int{1}); err == nil {
		t.Fatalf("slice len=2 should fail for len=1")
	}
	if err := lfn([]int{1, 2}); err != nil {
		t.Fatalf("slice len=2 should pass: %v", err)
	}
	if err := lfn([]int{1, 2, 3}); err == nil {
		t.Fatalf("slice len=2 should fail for len=3")
	}

	// Min/Max slice length.
	rfn := v.Slice().MinSliceLength(1).MaxSliceLength(3).Build()
	if err := rfn([]string{}); err == nil {
		t.Fatalf("min slice length should fail")
	}
	if err := rfn([]string{"a", "b"}); err != nil {
		t.Fatalf("min/max slice length should pass: %v", err)
	}
	if err := rfn([]string{"a", "b", "c", "d"}); err == nil {
		t.Fatalf("max slice length should fail")
	}

	// ForEach builder path with a string elem rule.
	elem := v.String().MinLength(1).Build()
	ffn := v.Slice().ForEach(elem).Build()
	if err := ffn([]string{"ok", ""}); err == nil {
		t.Fatalf("foreach should fail on empty second element")
	}
	if err := ffn([]string{"a", "b"}); err != nil {
		t.Fatalf("foreach should pass: %v", err)
	}
}
