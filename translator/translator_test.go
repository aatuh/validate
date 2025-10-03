package translator

import "testing"

func TestSimpleTranslator_LookupAndFallback(t *testing.T) {
	tr := NewSimpleTranslator(map[string]string{
		"hi %s": "hello %s",
	})
	if got := tr.T("hi %s", "world"); got != "hello world" {
		t.Fatalf("lookup format failed: %q", got)
	}
	// Fallback uses the key as the format string when not found.
	if got := tr.T("x=%d", 7); got != "x=7" {
		t.Fatalf("fallback format failed: %q", got)
	}
}

func TestDefaultEnglishTranslations_KeysPresent(t *testing.T) {
	m := DefaultEnglishTranslations()
	keys := []string{
		"string.minLength",
		"string.regex.noMatch",
		"int.notInt64",
		"slice.element",
		"slice.notSlice",
	}
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			t.Fatalf("expected default key %q", k)
		}
	}
}
