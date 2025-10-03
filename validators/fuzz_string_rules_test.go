package validators

import "testing"

// FuzzBuildStringRules ensures the builder never panics on random input.
func FuzzBuildStringRules(f *testing.F) {
	sv := NewStringValidators(dummyTr{})
	f.Add("min=1;max=5;regex=^a.*z$", "abcz")
	f.Add("oneof=a b", "a")
	f.Fuzz(func(t *testing.T, tail string, val string) {
		// Split tail into tokens; tolerate empty or junk.
		tokens := []string{"string"}
		for _, seg := range splitSemi(tail) {
			if seg != "" {
				tokens = append(tokens, seg)
			}
		}
		fn, err := BuildStringValidator(sv, tokens)
		if err != nil {
			return // invalid tokens are fine
		}
		_ = fn(val) // must not panic
	})
}

func splitSemi(s string) []string {
	out := make([]string, 0, len(s)/2+1)
	cur := ""
	for i := 0; i < len(s); i++ {
		if s[i] == ';' {
			out = append(out, cur)
			cur = ""
			continue
		}
		cur += string(s[i])
	}
	out = append(out, cur)
	return out
}
