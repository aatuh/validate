package validators

import "testing"

// FuzzBuildIntRules ensures int rule building is resilient.
func FuzzBuildIntRules(f *testing.F) {
	iv := NewIntValidators(dummyTr{})
	f.Add("min=0;max=10", int64(5))
	f.Add("min=-5", int64(-1))
	f.Fuzz(func(t *testing.T, tail string, v int64) {
		tokens := []string{"int"}
		for _, seg := range splitSemi(tail) {
			if seg != "" {
				tokens = append(tokens, seg)
			}
		}
		fn, err := BuildIntValidator(iv, tokens, "int")
		if err != nil {
			return
		}
		_ = fn(v) // must not panic
	})
}
