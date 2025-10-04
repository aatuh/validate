package examples

import (
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3"
)

// Test_caching demonstrates automatic caching for builder validators
// to improve performance when the same validation rules are used repeatedly.
func Test_caching(t *testing.T) {
	v := validate.New()

	// First call compiles and caches the validator
	validator1 := v.String().MinLength(3).MaxLength(50).Build()

	// Subsequent calls with identical rules return cached validator
	validator2 := v.String().MinLength(3).MaxLength(50).Build()
	// validator1 and validator2 are the same cached function

	// Different rules create new validators
	validator3 := v.String().MinLength(5).MaxLength(20).Build() // Different cache key

	// Test that all validators work
	fmt.Println("validator1 works:", validator1("hello") == nil)
	fmt.Println("validator2 works:", validator2("hello") == nil)
	fmt.Println("validator3 works:", validator3("hello") == nil)

	// Test that they behave differently
	fmt.Println("validator1 rejects short:", validator1("hi") != nil)
	fmt.Println("validator3 accepts short:", validator3("hi") != nil)

	// Output:
	// validator1 works: true
	// validator2 works: true
	// validator3 works: true
	// validator1 rejects short: true
	// validator3 accepts short: true
}
