package examples

import (
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3"
)

// Test_translation demonstrates using custom translations
// for validation error messages.
func Test_translation(t *testing.T) {
	msgs := map[string]string{
		"string.min": "doit contenir au moins %d caractères",
		"string.max": "ne peut pas dépasser %d caractères",
	}
	tr := validate.NewSimpleTranslator(msgs)

	v := validate.New().WithTranslator(tr)

	check := v.String().MinLength(5).MaxLength(10).Build()
	if err := check("ab"); err != nil {
		fmt.Println("fr:", err)
	}

	// Output:
	// fr:  [string.min]: doit contenir au moins 5 caractères
}
