package examples

import (
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3"
)

// Test_fromTag demonstrates FromTag convenience methods
// for single tag string validation.
func Test_fromTag(t *testing.T) {
	v := validate.New()

	// More convenient than FromRules([]string{"string;min=3;max=50"})
	validator, _ := v.FromTag("string;min=3;max=50")
	if err := validator("hello"); err != nil {
		fmt.Println("validation failed:", err)
	}

	// Root-level FromTag (works with nil Validate)
	validator2, _ := validate.FromTag(nil, "int;min=1;max=100")
	if err := validator2(50); err != nil {
		fmt.Println("validation failed:", err)
	}

	// Test with invalid data
	if err := validator("hi"); err != nil {
		fmt.Println("validation failed:", err)
	}

	if err := validator2(150); err != nil {
		fmt.Println("validation failed:", err)
	}

	// Output:
	// validation failed:  [string.min]: minimum length is 3
	// validation failed:  [int.max]: maximum value is 100
}
