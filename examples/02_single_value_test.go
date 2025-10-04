package examples

import (
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3"
)

// Test_singleValue demonstrates ad-hoc validation of single values
// using a fluent builder and FromTag.
func Test_singleValue(t *testing.T) {
	v := validate.New()

	// Builder: slug with lowercase letters, digits, underscore
	isSlug := v.String().Regex(`^[a-z0-9_]+$`).Build()
	fmt.Println(isSlug("hello_world") == nil)
	fmt.Println(isSlug("Hello-World") == nil)

	// FromTag: integer range
	check, _ := v.FromTag("int;min=1;max=10")
	fmt.Println(check(5) == nil)
	fmt.Println(check(0) == nil)

	// Output:
	// true
	// false
	// true
	// false
}
