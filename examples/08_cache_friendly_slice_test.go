package examples

import (
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3"
	"github.com/aatuh/validate/v3/types"
)

// Test_cacheFriendlySlice demonstrates cache-friendly slice validation
// using ForEachRules instead of ForEach for better performance.
func Test_cacheFriendlySlice(t *testing.T) {
	v := validate.New()

	// Cache-friendly: uses AST rules (cached)
	sliceValidator := v.Slice().MinLength(1).ForEachRules(
		types.NewRule(types.KString, nil),
		types.NewRule(types.KMinLength, map[string]any{"n": int64(2)}),
	).Build()

	// Not cache-friendly: uses function closure (not cached)
	elemValidator := v.String().MinLength(2).Build()
	sliceValidator2 := v.Slice().MinLength(1).ForEach(elemValidator).Build()

	// Test both validators
	validData := []string{"go", "lib"}
	invalidData := []string{"g", "lib"}

	fmt.Println("cache-friendly valid:", sliceValidator(validData) == nil)
	fmt.Println("cache-friendly invalid:", sliceValidator(invalidData) != nil)
	fmt.Println("function-based valid:", sliceValidator2(validData) == nil)
	fmt.Println("function-based invalid:", sliceValidator2(invalidData) != nil)

	// Output:
	// cache-friendly valid: true
	// cache-friendly invalid: true
	// function-based valid: true
	// function-based invalid: true
}
