package examples

import (
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3"
)

// Test_manualStructValidation demonstrates building validators
// programmatically and applying them to struct fields without using struct tags.
func Test_manualStructValidation(t *testing.T) {
	type User struct {
		Name string
		Age  int
		Tags []string
	}

	v := validate.New()

	// Build per-field validators using fluent builders
	nameV := v.String().MinLength(3).MaxLength(50).Build()
	ageV := v.Int().MinInt(18).MaxInt(120).Build()
	tagsV := v.Slice().MinLength(1).ForEachRules(
		validate.NewRule(validate.KString, nil),
		validate.NewRule(validate.KMinLength, map[string]any{"n": int64(2)}),
	).Build()

	u := User{Name: "Jane", Age: 20, Tags: []string{"go", "lib"}}

	// Demonstrate manual validation
	_ = nameV(u.Name)
	_ = ageV(u.Age)
	_ = tagsV(u.Tags)

	// Alternative: assemble rules directly and compile once
	rules := []validate.Rule{
		validate.NewRule(validate.KString, nil),
		validate.NewRule(validate.KMinLength, map[string]any{"n": int64(3)}),
	}
	nameV2 := v.CompileRules(rules)

	fmt.Println("nameV2 works:", nameV2("hello") == nil)
	fmt.Println("nameV2 rejects short:", nameV2("hi") != nil)

	// Output:
	// nameV2 works: true
	// nameV2 rejects short: true
}
