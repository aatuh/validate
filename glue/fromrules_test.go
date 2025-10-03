package glue

import "testing"

type dummyTr struct{}

func (dummyTr) T(key string, params ...any) string { return key }

func TestValidate_FromRules_String(t *testing.T) {
	v := New().WithTranslator(dummyTr{})
	fn, err := v.FromRules([]string{"string", "min=2", "max=4"})
	if err != nil {
		t.Fatalf("build err %v", err)
	}
	if err := fn("a"); err == nil {
		t.Fatalf("want min failure")
	}
	if err := fn("abcd"); err != nil {
		t.Fatalf("want pass, got %v", err)
	}
	if err := fn("abcde"); err == nil {
		t.Fatalf("want max failure")
	}
}

func TestValidate_Builders_Fluent(t *testing.T) {
	v := New().WithTranslator(dummyTr{})
	sfn := v.String().MinLength(2).MaxLength(3).Build()
	if err := sfn("a"); err == nil {
		t.Fatalf("want min failure")
	}
	if err := sfn("abc"); err != nil {
		t.Fatalf("want pass, got %v", err)
	}
	if err := sfn("abcd"); err == nil {
		t.Fatalf("want max failure")
	}

	ifn := v.Int().MinInt(1).MaxInt(3).Build()
	if err := ifn(int(0)); err == nil {
		t.Fatalf("want min fail")
	}
	if err := ifn(int64(2)); err != nil {
		t.Fatalf("want pass, got %v", err)
	}
	if err := ifn(int64(4)); err == nil {
		t.Fatalf("want max fail")
	}
}
