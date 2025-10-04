package examples

import (
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3"
)

// ExampleStructTags demonstrates validating a struct using `validate` tags.
func Test_structTags(t *testing.T) {
	type Profile struct {
		Name   string   `validate:"string;min=3;max=50"`
		Age    int      `validate:"int;min=18;max=120"`
		Tags   []string `validate:"slice;min=1;max=5;foreach=(string;min=2)"`
		Email  string   `validate:"string;email"`
		UserID string   `validate:"string;uuid"`
	}

	v := validate.New()

	p := Profile{
		Name:   "Al",
		Age:    16,
		Tags:   []string{"g", "v"},
		Email:  "not-an-email",
		UserID: "bad-uuid",
	}

	if err := v.ValidateStruct(p); err != nil {
		if es, ok := err.(validate.Errors); ok {
			fmt.Println("has Name error:", len(es.Filter("Name")) > 0)
			fmt.Println("has Age error:", len(es.Filter("Age")) > 0)
			fmt.Println("has Tags error:", len(es.Filter("Tags")) > 0)
			fmt.Println("has Email error:", len(es.Filter("Email")) > 0)
			fmt.Println("has UserID error:", len(es.Filter("UserID")) > 0)
		}
	}

	// Output:
	// has Name error: true
	// has Age error: true
	// has Tags error: true
	// has Email error: true
	// has UserID error: true
}
