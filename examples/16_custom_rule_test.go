package examples

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aatuh/validate/v3"
	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/types"
)

func Test_customRuleCompiler(t *testing.T) {
	v := validate.New().
		WithRuleCompiler("even", compileEven).
		WithRuleCompiler("mod", compileMod)

	check := v.CompileRules([]validate.Rule{
		validate.NewRule(validate.KInt, nil),
		validate.NewRule("even", nil),
	})
	builder := v.Int().Rule("even", nil).Build()

	fmt.Println("manual ok:", check(2) == nil)
	fmt.Println("tag ok:", v.CheckTag("int;even", 2) == nil)
	fmt.Println("custom arg ok:", v.CheckTag("int;custom:mod=2", 4) == nil)
	fmt.Println("builder ok:", builder(3) == nil)

	// Output:
	// manual ok: true
	// tag ok: true
	// custom arg ok: true
	// builder ok: false
}

func Test_customStructRuleCompiler(t *testing.T) {
	v := validate.New().WithStructRuleCompiler("matchesField", func(rule validate.Rule) (validate.StructRuleFunc, error) {
		fieldName, _ := rule.Args["value"].(string)
		return func(ctx validate.StructRuleContext) error {
			other, _ := ctx.FieldValue(fieldName)
			if ctx.Value != other {
				return verrs.Errors{verrs.FieldError{Code: "field.matches", Msg: "must match"}}
			}
			return nil
		}, nil
	})

	type Signup struct {
		Password string `validate:"string;required"`
		Confirm  string `validate:"string;struct:matchesField=Password"`
	}

	fmt.Println("matching ok:", v.ValidateStruct(Signup{Password: "alpha12345", Confirm: "alpha12345"}) == nil)
	fmt.Println("mismatch ok:", v.ValidateStruct(Signup{Password: "alpha12345", Confirm: "different"}) == nil)

	// Output:
	// matching ok: true
	// mismatch ok: false
}

func compileEven(c *types.Compiler, rule types.Rule) (func(any) error, error) {
	return func(value any) error {
		n, ok := value.(int)
		if !ok || n%2 != 0 {
			return verrs.Errors{verrs.FieldError{
				Code: "number.even",
				Msg:  c.T("number.even", "must be even", nil),
			}}
		}
		return nil
	}, nil
}

func compileMod(c *types.Compiler, rule types.Rule) (func(any) error, error) {
	raw, _ := rule.Args["value"].(string)
	mod, err := strconv.Atoi(raw)
	if err != nil || mod == 0 {
		return nil, fmt.Errorf("invalid mod parameter")
	}
	return func(value any) error {
		n, ok := value.(int)
		if !ok || n%mod != 0 {
			return verrs.Errors{verrs.FieldError{Code: "number.mod", Msg: c.T("number.mod", "invalid modulus", nil)}}
		}
		return nil
	}, nil
}
