package validators

import "testing"

func BenchmarkString_MinMax_10B(b *testing.B) {
	sv := NewStringValidators(nil)
	fn := sv.WithString(sv.MinLength(3), sv.MaxLength(12))
	in := "hello-go"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fn(in)
	}
}
