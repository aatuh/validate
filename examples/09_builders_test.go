package examples

import (
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3"
)

// Test_builders demonstrates all available builder types
// and their configuration options.
func Test_builders(t *testing.T) {
	v := validate.New()

	// String builder with multiple rules
	strV := v.String().MinLength(3).MaxLength(50).Regex(`^[a-z0-9_]+$`).Build()
	fmt.Println("string valid:", strV("hello_world") == nil)
	fmt.Println("string invalid:", strV("Hello-World") != nil)

	// Int builder (accepts any Go int type at call time)
	intV := v.Int().MinInt(0).MaxInt(100).Build()
	fmt.Println("int valid:", intV(50) == nil)
	fmt.Println("int invalid:", intV(150) != nil)

	// Int64 builder (requires exactly int64 at call time)
	int64V := v.Int64().MinInt(0).MaxInt(100).Build()
	fmt.Println("int64 valid:", int64V(int64(50)) == nil)

	// Slice builder with ForEach
	elem := v.String().MinLength(2).Build()
	sliceV := v.Slice().MinLength(1).ForEach(elem).Build()
	fmt.Println("slice valid:", sliceV([]string{"go", "lib"}) == nil)
	fmt.Println("slice invalid:", sliceV([]string{"g"}) != nil)

	// Cache-friendly slice validation
	sliceV2 := v.Slice().MinLength(1).ForEachRules(
		validate.NewRule(validate.KString, nil),
		validate.NewRule(validate.KMinLength, map[string]any{"n": int64(2)}),
	).Build()
	fmt.Println("cache-friendly slice valid:", sliceV2([]string{"go", "lib"}) == nil)

	// Convenience method for string elements
	stringBuilder := v.String().MinLength(2)
	sliceV3 := v.Slice().MinLength(1).ForEachStringBuilder(stringBuilder).Build()
	fmt.Println("convenience slice valid:", sliceV3([]string{"go", "lib"}) == nil)

	// Bool builder (type-check only)
	boolV := v.Bool().Build()
	fmt.Println("bool valid:", boolV(true) == nil)
	fmt.Println("bool invalid:", boolV("not-bool") != nil)

	// Output:
	// string valid: true
	// string invalid: true
	// int valid: true
	// int invalid: true
	// int64 valid: true
	// slice valid: true
	// slice invalid: true
	// cache-friendly slice valid: true
	// convenience slice valid: true
	// bool valid: true
	// bool invalid: true
}
