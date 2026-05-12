package types

import "testing"

func BenchmarkParseTag_StringRules(b *testing.B) {
	tag := "string;required;min=3;max=40;regex=[a-z0-9_-]+"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := ParseTag(tag); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseTag_NestedCollections(b *testing.B) {
	tag := "map;keys=(string;slug);values=(slice;min=1;foreach=(string;minRunes=2;maxRunes=20))"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := ParseTag(tag); err != nil {
			b.Fatal(err)
		}
	}
}
