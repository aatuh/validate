package validators

import "testing"

func TestSlice_Len_Min_Max(t *testing.T) {
	sv := NewSliceValidators(dummyTr{})
	fn, err := BuildSliceValidator(sv, []string{
		"slice", "min=1", "max=3",
	})
	if err != nil {
		t.Fatalf("build err %v", err)
	}
	if err := fn([]int{}); err == nil {
		t.Fatalf("min should fail")
	}
	if err := fn([]int{1}); err != nil {
		t.Fatalf("min ok got %v", err)
	}
	if err := fn([]int{1, 2, 3}); err != nil {
		t.Fatalf("max ok got %v", err)
	}
	if err := fn([]int{1, 2, 3, 4}); err == nil {
		t.Fatalf("max should fail")
	}
}

func TestSlice_ForEach_ElementErrorsIncludeIndex(t *testing.T) {
	// Each element must be non-empty string.
	sv := NewSliceValidators(dummyTr{})
	str := NewStringValidators(dummyTr{})
	elem := str.WithString(str.MinLength(1))

	fn := sv.WithSlice(sv.ForEach(elem))

	err := fn([]string{"ok", ""})
	if err == nil {
		t.Fatalf("want element error at index 1")
	}
	got := err.Error()
	if want := "slice.element"; !contains(got, want) {
		t.Fatalf("want substr %q in %q", want, got)
	}
	if want := "1"; !contains(got, want) {
		t.Fatalf("want index 1 in %q", got)
	}
}

func TestSlice_NotSlice(t *testing.T) {
	sv := NewSliceValidators(dummyTr{})
	fn := sv.WithSlice()
	if err := fn(42); err == nil {
		t.Fatalf("not a slice should fail")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) > 0 &&
		indexOf(s, sub) >= 0))
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestSlice_SliceLength_Exact(t *testing.T) {
	// Nil translator to hit fallback branch in translate().
	sv := NewSliceValidators(nil)
	fn := sv.WithSlice(sv.SliceLength(2))

	if err := fn([]int{1}); err == nil {
		t.Fatalf("len=1 should fail for len=2 rule")
	}
	if err := fn([]int{1, 2}); err != nil {
		t.Fatalf("len=2 should pass: %v", err)
	}
	if err := fn([]int{1, 2, 3}); err == nil {
		t.Fatalf("len=3 should fail for len=2 rule")
	}
}
