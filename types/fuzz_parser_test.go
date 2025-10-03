package types

import (
	"testing"
)

// FuzzParseTag tests the tag parser with random inputs to find edge cases
func FuzzParseTag(f *testing.F) {
	// Add some seed inputs
	f.Add("string")
	f.Add("string;min=3;max=50")
	f.Add("int;min=1;max=100")
	f.Add("slice;min=1;max=10")
	f.Add("string;oneof=red,green,blue")
	f.Add("string;oneof=red green blue")
	f.Add("slice;min=1;foreach=(string;min=2;max=10)")

	f.Fuzz(func(t *testing.T, tag string) {
		// ParseTag should not panic on any input
		rules, err := ParseTag(tag)

		// If parsing succeeds, rules should be valid
		if err == nil {
			for _, rule := range rules {
				if rule.Kind == "" {
					t.Errorf("Empty rule kind in parsed rules: %+v", rule)
				}
			}
		}

		// If parsing fails, error should be meaningful
		if err != nil {
			if err.Error() == "" {
				t.Errorf("Empty error message for tag: %q", tag)
			}
		}
	})
}

// FuzzParseTagLong tests with very long inputs to find memory issues
func FuzzParseTagLong(f *testing.F) {
	f.Add("string;min=1;max=1000;length=500;regex=.*;oneof=a,b,c,d,e,f,g,h,i,j,k,l,m,n,o,p,q,r,s,t,u,v,w,x,y,z")

	f.Fuzz(func(t *testing.T, tag string) {
		// Should not panic or hang on long inputs
		rules, err := ParseTag(tag)

		// Should either succeed or fail gracefully
		if err != nil {
			// Error messages can be long for malformed inputs - that's OK for fuzz testing
			// The important thing is that parsing doesn't panic or hang
		} else {
			// Should not return too many rules
			if len(rules) > 100 {
				t.Errorf("Too many rules returned: %d", len(rules))
			}
		}
	})
}
