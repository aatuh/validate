package validators

import "testing"

func TestInt_Min_Max_WithInt(t *testing.T) {
	iv := NewIntValidators(dummyTr{})
	fn, err := BuildIntValidator(iv, []string{
		"int", "min=1", "max=3",
	}, "int")
	if err != nil {
		t.Fatalf("build err %v", err)
	}
	type in struct {
		v  any
		ok bool
	}
	cases := []in{
		{int(0), false},
		{int(1), true},
		{int32(2), true},
		{int64(3), true},
		{int64(4), false},
	}
	for _, c := range cases {
		err := fn(c.v)
		if c.ok && err != nil {
			t.Fatalf("value=%v unexpected err %v", c.v, err)
		}
		if !c.ok && err == nil {
			t.Fatalf("value=%v want error", c.v)
		}
	}
}

func TestInt_WithInt64_RequiresExact(t *testing.T) {
	iv := NewIntValidators(dummyTr{})
	fn, err := BuildIntValidator(iv, []string{
		"int64", "min=0",
	}, "int64")
	if err != nil {
		t.Fatalf("build err %v", err)
	}
	if err := fn(int(1)); err == nil {
		t.Fatalf("int should not be accepted by int64 validator")
	}
	if err := fn(int64(1)); err != nil {
		t.Fatalf("int64 should be accepted, got %v", err)
	}
}

func TestInt_toInt64_AllIntKinds_And_Error(t *testing.T) {
	iv := NewIntValidators(nil)
	fn := iv.WithInt() // no rules, just type coercion

	// All signed int kinds accepted.
	if err := fn(int8(1)); err != nil {
		t.Fatalf("int8 should pass: %v", err)
	}
	if err := fn(int16(1)); err != nil {
		t.Fatalf("int16 should pass: %v", err)
	}
	if err := fn(int32(1)); err != nil {
		t.Fatalf("int32 should pass: %v", err)
	}
	if err := fn(int64(1)); err != nil {
		t.Fatalf("int64 should pass: %v", err)
	}

	// A non-int should fail the type conversion branch.
	if err := fn(float64(1)); err == nil {
		t.Fatalf("float64 must fail toInt64")
	}
}
