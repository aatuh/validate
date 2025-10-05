package examples

import (
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3"
)

// Test_manualRules demonstrates compiling a validator from AST rules
// without using struct tags or fluent builders.
func Test_manualRules(t *testing.T) {
	// Build a string validator: string;min=3;max=8
	rules := []validate.Rule{
		validate.NewRule(validate.KString, nil),
		validate.NewRule(validate.KMinLength, map[string]any{"n": int64(3)}),
		validate.NewRule(validate.KMaxLength, map[string]any{"n": int64(8)}),
	}

	v := validate.New()
	check := v.CompileRules(rules)

	for _, s := range []string{"hi", "hello", "this-is-too-long"} {
		if err := check(s); err != nil {
			fmt.Println("invalid:", s)
		} else {
			fmt.Println("ok:", s)
		}
	}

	// Output:
	// invalid: hi
	// ok: hello
	// invalid: this-is-too-long
}
