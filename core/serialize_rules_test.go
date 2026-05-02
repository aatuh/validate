package core

import (
	"errors"
	"strings"
	"testing"

	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/types"
)

func TestCompileRules_CacheKeyIncludesRuleElem(t *testing.T) {
	tests := []struct {
		name string
		rule func(types.Rule) types.Rule
	}{
		{
			name: "pointer constructor",
			rule: func(elem types.Rule) types.Rule {
				return types.NewRuleWithElem(types.KForEach, nil, &elem)
			},
		},
		{
			name: "value constructor",
			rule: func(elem types.Rule) types.Rule {
				return types.NewRuleWithElemValue(types.KForEach, nil, elem)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			minTwo := tt.rule(types.NewRule(types.KMinLength, map[string]any{"n": 2}))
			maxOne := tt.rule(types.NewRule(types.KMaxLength, map[string]any{"n": 1}))

			if err := v.CompileRules([]types.Rule{minTwo})([]string{"aa"}); err != nil {
				t.Fatalf("min length validator should pass: %v", err)
			}

			err := v.CompileRules([]types.Rule{maxOne})([]string{"aa"})
			if err == nil {
				t.Fatalf("max length validator passed; cache likely reused different Elem rule")
			}
			var es verrs.Errors
			if !errors.As(err, &es) || len(es) == 0 || es[0].Code != verrs.CodeStringMax {
				t.Fatalf("error = %v, want first code %q", err, verrs.CodeStringMax)
			}
		})
	}
}

func TestSerializeRules_IncludesElemAndDetectsElemFunctions(t *testing.T) {
	ruleA := types.NewRuleWithElem(types.KForEach, nil, &types.Rule{
		Kind: types.KMinLength,
		Args: map[string]any{"n": 2},
	})
	ruleB := types.NewRuleWithElem(types.KForEach, nil, &types.Rule{
		Kind: types.KMaxLength,
		Args: map[string]any{"n": 1},
	})

	gotA := SerializeRules([]types.Rule{ruleA})
	gotB := SerializeRules([]types.Rule{ruleB})
	if gotA == gotB {
		t.Fatalf("SerializeRules returned the same key for different Elem rules: %q", gotA)
	}
	if !strings.Contains(gotA, "elem:{kind:minLength,args:{n:2}}") {
		t.Fatalf("SerializeRules missing Elem details: %q", gotA)
	}

	withFunc := types.NewRuleWithElem(types.KForEach, nil, &types.Rule{
		Kind: types.KForEach,
		Args: map[string]any{
			"validator": func(any) error { return nil },
		},
	})
	if !HasFuncArgs([]types.Rule{withFunc}) {
		t.Fatalf("HasFuncArgs did not inspect nested Elem rule")
	}
}
