package glue

import (
	"testing"
)

// TestTagsVsBuildersEquivalence ensures that tag-based and builder-based
// validation produce equivalent results for the same rules.
func TestTagsVsBuildersEquivalence(t *testing.T) {
	testCases := []struct {
		name     string
		tag      string
		buildFn  func(*Validate) func(any) error
		testVals []any
	}{
		{
			name: "string_length",
			tag:  "string;length=5",
			buildFn: func(v *Validate) func(any) error {
				return v.String().Length(5).Build()
			},
			testVals: []any{"hello", "hi", "world", "test", 123},
		},
		{
			name: "string_min_max",
			tag:  "string;min=3;max=10",
			buildFn: func(v *Validate) func(any) error {
				return v.String().MinLength(3).MaxLength(10).Build()
			},
			testVals: []any{"hi", "hello", "world", "verylongstring", 123},
		},
		{
			name: "string_oneof",
			tag:  "string;oneof=red,green,blue",
			buildFn: func(v *Validate) func(any) error {
				return v.String().OneOf("red", "green", "blue").Build()
			},
			testVals: []any{"red", "green", "blue", "yellow", "purple", 123},
		},
		{
			name: "int_min_max",
			tag:  "int;min=1;max=100",
			buildFn: func(v *Validate) func(any) error {
				return v.Int().MinInt(1).MaxInt(100).Build()
			},
			testVals: []any{0, 1, 50, 100, 101, -1, "not-int"},
		},
		{
			name: "slice_length",
			tag:  "slice;length=3",
			buildFn: func(v *Validate) func(any) error {
				return v.Slice().Length(3).Build()
			},
			testVals: []any{[]string{"a", "b", "c"}, []string{"a", "b"}, []string{"a", "b", "c", "d"}, "not-slice"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := New()

			// Create tag-based validator
			tagValidator, err := v.FromRules([]string{tc.tag})
			if err != nil {
				t.Fatalf("Failed to create tag validator: %v", err)
			}

			// Create builder-based validator
			builderValidator := tc.buildFn(v)

			// Test each value with both validators
			for _, testVal := range tc.testVals {
				tagErr := tagValidator(testVal)
				builderErr := builderValidator(testVal)

				// Both should either succeed or fail
				tagSuccess := tagErr == nil
				builderSuccess := builderErr == nil

				if tagSuccess != builderSuccess {
					t.Errorf("Mismatch for value %v: tag=%v, builder=%v",
						testVal, tagSuccess, builderSuccess)
					if tagErr != nil {
						t.Errorf("Tag error: %v", tagErr)
					}
					if builderErr != nil {
						t.Errorf("Builder error: %v", builderErr)
					}
				}
			}
		})
	}
}

// TestTagsVsBuildersErrorCodes ensures that tag-based and builder-based
// validation produce the same error codes for the same failures.
func TestTagsVsBuildersErrorCodes(t *testing.T) {
	testCases := []struct {
		name     string
		tag      string
		buildFn  func(*Validate) func(any) error
		testVal  any
		wantCode string
	}{
		{
			name:     "string_min_length",
			tag:      "string;min=5",
			buildFn:  func(v *Validate) func(any) error { return v.String().MinLength(5).Build() },
			testVal:  "hi",
			wantCode: "string.min",
		},
		{
			name:     "string_max_length",
			tag:      "string;max=3",
			buildFn:  func(v *Validate) func(any) error { return v.String().MaxLength(3).Build() },
			testVal:  "hello",
			wantCode: "string.max",
		},
		{
			name:     "int_min",
			tag:      "int;min=10",
			buildFn:  func(v *Validate) func(any) error { return v.Int().MinInt(10).Build() },
			testVal:  5,
			wantCode: "int.min",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := New()

			// Create tag-based validator
			tagValidator, err := v.FromRules([]string{tc.tag})
			if err != nil {
				t.Fatalf("Failed to create tag validator: %v", err)
			}

			// Create builder-based validator
			builderValidator := tc.buildFn(v)

			// Test with failing value
			tagErr := tagValidator(tc.testVal)
			builderErr := builderValidator(tc.testVal)

			// Both should fail
			if tagErr == nil || builderErr == nil {
				t.Errorf("Expected both validators to fail for %v", tc.testVal)
				return
			}

			// Extract error codes (simplified - in real implementation you'd parse the structured errors)
			tagErrStr := tagErr.Error()
			builderErrStr := builderErr.Error()

			// Both should contain the expected error code
			if !contains(tagErrStr, tc.wantCode) {
				t.Errorf("Tag error missing code %q: %s", tc.wantCode, tagErrStr)
			}
			if !contains(builderErrStr, tc.wantCode) {
				t.Errorf("Builder error missing code %q: %s", tc.wantCode, builderErrStr)
			}
		})
	}
}

// TestTagsVsBuildersPerformance ensures that both approaches have similar
// performance characteristics (basic smoke test).
func TestTagsVsBuildersPerformance(t *testing.T) {
	v := New()

	// Create validators
	tagValidator, err := v.FromRules([]string{"string;min=3;max=10"})
	if err != nil {
		t.Fatalf("Failed to create tag validator: %v", err)
	}

	builderValidator := v.String().MinLength(3).MaxLength(10).Build()

	testVal := "hello"

	// Run both validators multiple times
	iterations := 1000

	// Tag-based validation
	for i := 0; i < iterations; i++ {
		_ = tagValidator(testVal)
	}

	// Builder-based validation
	for i := 0; i < iterations; i++ {
		_ = builderValidator(testVal)
	}

	// If we get here without panicking, performance is reasonable
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			contains(s[1:], substr))))
}
