package examples

import (
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3"
)

// Test_errorHandling demonstrates error handling patterns
// for validation errors, including structured error inspection.
func Test_errorHandling(t *testing.T) {
	v := validate.New()

	// Single value validation error
	validator := v.String().MinLength(5).Build()
	if err := validator("hi"); err != nil {
		fmt.Println("single error:", err)
	}

	// Struct validation with structured errors
	type User struct {
		Name string `validate:"string;min=3;max=50"`
		Age  int    `validate:"int;min=18;max=120"`
	}

	user := User{Name: "Al", Age: 16}
	if err := v.ValidateStruct(user); err != nil {
		if es, ok := err.(validate.Errors); ok {
			fmt.Println("has Name error:", len(es.Filter("Name")) > 0)
			fmt.Println("has Age error:", len(es.Filter("Age")) > 0)
			fmt.Println("error map:", es.AsMap())
		}
	}

	// Output:
	// single error:  [string.min]: minimum length is 5
	// has Name error: true
	// has Age error: true
	// error map: map[Age.:[Age. [int.min]: minimum value is 18] Name.:[Name. [string.min]: minimum length is 3]]
}
