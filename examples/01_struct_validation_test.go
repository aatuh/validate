package examples

import (
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3"
)

// Test_structValidation demonstrates struct validation using tags
// with comprehensive field validation rules.
func Test_structValidation(t *testing.T) {
	type User struct {
		Name    string   `validate:"string;min=3;max=50"`
		Website string   `validate:"string;min=5;max=100"`
		Age     int      `validate:"int;min=18;max=120"`
		ID      string   `validate:"string;min=5;max=20"`
		Tags    []string `validate:"slice;min=1;max=5;foreach=(string;min=2)"`
		Status  string   `validate:"string;oneof=active,inactive,pending"`
		Bio     string   `validate:"string;max=500"` // Maximum length
	}

	v := validate.New()

	u := User{
		Name:    "John Doe",
		Website: "https://example.com",
		Age:     25,
		ID:      "user123",
		Tags:    []string{"golang", "validation"},
		Status:  "active",
		Bio:     "Software developer",
	}

	if err := v.ValidateStruct(u); err != nil {
		if es, ok := err.(validate.Errors); ok {
			fmt.Println("errors:", es.AsMap())
		} else {
			fmt.Println("validation failed:", err)
		}
	}

	// Output:
}
