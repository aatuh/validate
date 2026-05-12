package core

import (
	"testing"

	"github.com/aatuh/validate/v3/types"
)

func BenchmarkEngine_FromRulesCached_String(b *testing.B) {
	v := New()
	tokens := []string{"string", "min=3", "max=40"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := v.FromRules(tokens); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEngine_CompiledStringValidation(b *testing.B) {
	v := New()
	fn, err := v.FromRules([]string{"string", "min=3", "max=40"})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := fn("validation-library"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEngine_NestedCollectionValidation(b *testing.B) {
	v := New()
	fn, err := v.FromRules(types.SplitTag("map;values=(slice;foreach=(string;min=2))"))
	if err != nil {
		b.Fatal(err)
	}
	input := map[string][]string{
		"primary":   {"go", "api", "docs"},
		"secondary": {"v3", "tag", "rule"},
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err := fn(input); err != nil {
			b.Fatal(err)
		}
	}
}
