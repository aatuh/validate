package glue

import (
	"errors"
	"math"
	"testing"
	"time"

	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/translator"
	"github.com/aatuh/validate/v3/types"
)

func TestExpandedBuilders(t *testing.T) {
	v := New()

	tests := []struct {
		name    string
		fn      func(any) error
		valid   any
		invalid any
		code    string
	}{
		{"string", v.String().Required().Contains("go").NotContains("java").Prefix("go").Suffix("lang").Build(), "golang", "", verrs.CodeRequired},
		{"float", v.Float().Required().Finite().Between(1, 10).Positive().Build(), 2.5, math.Inf(1), verrs.CodeNumberFinite},
		{"bool", v.Bool().True().Build(), true, false, verrs.CodeBoolTrue},
		{"slice", v.Slice().Required().Unique().Contains("a").Build(), []string{"a", "b"}, []string{"b", "c"}, verrs.CodeSliceContains},
		{"array", v.Array().Required().Unique().Contains("a").Build(), [2]string{"a", "b"}, [2]string{"b", "c"}, verrs.CodeArrayContains},
		{"map", v.Map().Required().MinKeys(1).KeysRules(types.NewRule(types.KString, nil)).ValuesRules(types.NewRule(types.KInt, nil)).Build(), map[string]int{"a": 1}, map[string]int{}, verrs.CodeRequired},
		{"time", v.Time().Required().After(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)).Build(), time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), time.Time{}, verrs.CodeRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(tt.valid); err != nil {
				t.Fatalf("valid input failed: %v", err)
			}
			err := tt.fn(tt.invalid)
			if err == nil {
				t.Fatalf("invalid input passed")
			}
			var es verrs.Errors
			if !errors.As(err, &es) {
				t.Fatalf("expected structured error, got %T %v", err, err)
			}
			if len(es) == 0 || es[0].Code != tt.code {
				t.Fatalf("code = %#v, want first code %q", es, tt.code)
			}
		})
	}
}

func TestBuilders_CustomRuleEscapeHatches(t *testing.T) {
	v := New().WithRuleCompiler("fail", func(c *types.Compiler, rule types.Rule) (func(any) error, error) {
		return func(any) error {
			return verrs.Errors{verrs.FieldError{Code: "custom.fail", Msg: "custom failure"}}
		}, nil
	})

	const customTypeName = "builderRuleEscapeHatchType"
	types.RegisterGlobalType(customTypeName, builderRuleTestFactory{})

	tests := []struct {
		name  string
		fn    func(any) error
		value any
	}{
		{"string", v.String().Rule("fail", nil).Build(), "value"},
		{"int", v.Int().Rule("fail", nil).Build(), 1},
		{"float", v.Float().Rule("fail", nil).Build(), 1.5},
		{"bool", v.Bool().Rule("fail", nil).Build(), true},
		{"slice", v.Slice().Rule("fail", nil).Build(), []string{"a"}},
		{"array", v.Array().Rule("fail", nil).Build(), [1]string{"a"}},
		{"map", v.Map().Rule("fail", nil).Build(), map[string]string{"a": "b"}},
		{"time", v.Time().Rule("fail", nil).Build(), time.Now()},
		{"custom type", v.CustomType(customTypeName).Rule("fail", nil).Build(), "value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn(tt.value)
			var es verrs.Errors
			if !errors.As(err, &es) || len(es) == 0 || es[0].Code != "custom.fail" {
				t.Fatalf("error = %#v, want custom.fail", err)
			}
		})
	}
}

type builderRuleTestFactory struct{}

func (builderRuleTestFactory) CreateValidator(_ translator.Translator) types.TypeValidator {
	return builderRuleTestValidator{}
}

type builderRuleTestValidator struct{}

func (builderRuleTestValidator) Validate(any) error { return nil }
