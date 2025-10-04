package examples

import (
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3"
)

// Test_basicValidation demonstrates basic string and int validation
// using fluent builders and FromTag.
func Test_basicValidation(t *testing.T) {
	v := validate.New()

	// String: call Build() to get func(any) error.
	nameV := v.String().MinLength(3).MaxLength(50).Build()
	if err := nameV("John"); err != nil {
		fmt.Println("validation failed:", err)
	}

	// Int: call Build() to obtain func(any) error.
	ageV := v.Int().MinInt(18).MaxInt(120).Build()
	if err := ageV(25); err != nil {
		fmt.Println("validation failed:", err)
	}

	// FromTag: compile single tag strings (more convenient than FromRules)
	tagV, _ := v.FromTag("string;min=3;max=50")
	if err := tagV("John"); err != nil {
		fmt.Println("validation failed:", err)
	}

	// Output:
}
