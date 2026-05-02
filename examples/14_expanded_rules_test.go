package examples

import (
	"fmt"
	"testing"
	"time"

	"github.com/aatuh/validate/v3"
)

func Test_expandedRules(t *testing.T) {
	v := validate.New()

	username := v.String().
		Required().
		MinRunes(3).
		MaxRunes(20).
		ASCII().
		Alnum().
		Build()
	fmt.Println("username ok:", username("gopher123") == nil)

	score := v.Float().Finite().Between(0, 100).Build()
	fmt.Println("score ok:", score(98.5) == nil)

	labels := v.Map().
		MinKeys(1).
		KeysRules(validate.NewRule(validate.KString, nil), validate.NewRule(validate.KMinLength, map[string]any{"n": int64(2)})).
		ValuesRules(validate.NewRule(validate.KString, nil), validate.NewRule(validate.KNonEmpty, nil)).
		Build()
	fmt.Println("labels ok:", labels(map[string]string{"env": "prod"}) == nil)

	deadline := v.Time().After(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)).Build()
	fmt.Println("deadline ok:", deadline(time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)) == nil)

	// Output:
	// username ok: true
	// score ok: true
	// labels ok: true
	// deadline ok: true
}
