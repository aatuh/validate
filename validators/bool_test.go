package validators

import "testing"

type trConst struct{ s string }

func (t trConst) T(key string, params ...any) string { return t.s }

func TestBool_New_WithBool_TypeChecks_And_Translate(t *testing.T) {
	// Nil translator path.
	bNil := NewBoolValidators(nil).WithBool()
	if err := bNil(true); err != nil {
		t.Fatalf("bool true should pass: %v", err)
	}
	if err := bNil(false); err != nil {
		t.Fatalf("bool false should pass: %v", err)
	}
	if err := bNil(1); err == nil {
		t.Fatalf("non-bool should fail")
	}

	// Translator used for error message.
	bt := NewBoolValidators(trConst{"X"}).WithBool()
	if err := bt(123); err == nil {
		t.Fatalf("expect error for non-bool")
	} else if err.Error() != "X" {
		t.Fatalf("translator should drive message, got %q", err.Error())
	}
}
