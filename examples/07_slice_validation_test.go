package examples

import (
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3"
)

// Test_sliceValidation demonstrates slice validation with element rules
// using different methods for cache-friendly validation.
func Test_sliceValidation(t *testing.T) {
	v := validate.New()

	// Method 1: ForEach with function (not cache-friendly)
	tagElem := v.String().MinLength(2).Build()
	tagsV := v.Slice().MinLength(1).ForEach(tagElem).Build()

	// Method 2: ForEachRules (cache-friendly, better performance)
	tagsV2 := v.Slice().MinLength(1).ForEachRules(
		validate.NewRule(validate.KString, nil),
		validate.NewRule(validate.KMinLength, map[string]any{"n": int64(2)}),
	).Build()

	// Method 3: ForEachStringBuilder (convenience for string elements)
	stringBuilder := v.String().MinLength(2)
	tagsV3 := v.Slice().MinLength(1).ForEachStringBuilder(stringBuilder).Build()

	// Test with valid data
	if err := tagsV([]string{"go", "lib"}); err != nil {
		fmt.Println("validation failed:", err)
	}

	// Test with invalid data (too short element)
	if err := tagsV2([]string{"g", "lib"}); err != nil {
		fmt.Println("validation failed:", err)
	}

	// Test with empty slice
	if err := tagsV3([]string{}); err != nil {
		fmt.Println("validation failed:", err)
	}

	// Output:
	// validation failed: [0] [string.min]: minimum length is 2
	// validation failed:  [slice.min]: minimum length is 1
}
