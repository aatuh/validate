package types

import (
	"testing"

	"github.com/aatuh/validate/v3/translator"
)

func TestParseTag_StringRules(t *testing.T) {
	tests := []struct {
		tag    string
		expect []Rule
	}{
		{
			tag: "string",
			expect: []Rule{
				{Kind: KString, Args: nil},
			},
		},
		{
			tag: "string;min=3;max=50",
			expect: []Rule{
				{Kind: KString, Args: nil},
				{Kind: KMinLength, Args: map[string]any{"n": 3}},
				{Kind: KMaxLength, Args: map[string]any{"n": 50}},
			},
		},
		{
			tag: "string;length=5;oneof=red,green,blue",
			expect: []Rule{
				{Kind: KString, Args: nil},
				{Kind: KLength, Args: map[string]any{"n": 5}},
				{Kind: KOneOf, Args: map[string]any{"values": []string{"red", "green", "blue"}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			rules, err := ParseTag(tt.tag)
			if err != nil {
				t.Fatalf("ParseTag failed: %v", err)
			}
			if len(rules) != len(tt.expect) {
				t.Fatalf("expected %d rules, got %d", len(tt.expect), len(rules))
			}
			for i, rule := range rules {
				expect := tt.expect[i]
				if rule.Kind != expect.Kind {
					t.Errorf("rule %d: expected kind %s, got %s", i, expect.Kind, rule.Kind)
				}
				// Compare args more carefully
				if len(rule.Args) != len(expect.Args) {
					t.Errorf("rule %d: expected %d args, got %d", i, len(expect.Args), len(rule.Args))
				}
			}
		})
	}
}

func TestParseTag_IntRules(t *testing.T) {
	tests := []struct {
		tag    string
		expect []Rule
	}{
		{
			tag: "int",
			expect: []Rule{
				{Kind: KInt, Args: nil},
			},
		},
		{
			tag: "int64",
			expect: []Rule{
				{Kind: KInt64, Args: nil},
			},
		},
		{
			tag: "int;min=1;max=100",
			expect: []Rule{
				{Kind: KInt, Args: nil},
				{Kind: KMinInt, Args: map[string]any{"n": int64(1)}},
				{Kind: KMaxInt, Args: map[string]any{"n": int64(100)}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			rules, err := ParseTag(tt.tag)
			if err != nil {
				t.Fatalf("ParseTag failed: %v", err)
			}
			if len(rules) != len(tt.expect) {
				t.Fatalf("expected %d rules, got %d", len(tt.expect), len(rules))
			}
		})
	}
}

func TestParseTag_SliceRules(t *testing.T) {
	tests := []struct {
		tag    string
		expect []Rule
	}{
		{
			tag: "slice",
			expect: []Rule{
				{Kind: KSlice, Args: nil},
			},
		},
		{
			tag: "slice;length=3",
			expect: []Rule{
				{Kind: KSlice, Args: nil},
				{Kind: KSliceLength, Args: map[string]any{"n": 3}},
			},
		},
		{
			tag: "slice;min=1;max=10",
			expect: []Rule{
				{Kind: KSlice, Args: nil},
				{Kind: KMinSliceLength, Args: map[string]any{"n": 1}},
				{Kind: KMaxSliceLength, Args: map[string]any{"n": 10}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			rules, err := ParseTag(tt.tag)
			if err != nil {
				t.Fatalf("ParseTag failed: %v", err)
			}
			if len(rules) != len(tt.expect) {
				t.Fatalf("expected %d rules, got %d", len(tt.expect), len(rules))
			}
		})
	}
}

func TestCompiler_Compile(t *testing.T) {
	tr := translator.NewSimpleTranslator(translator.DefaultEnglishTranslations())
	compiler := NewCompiler(tr)

	tests := []struct {
		name  string
		rules []Rule
		value any
		valid bool
	}{
		{
			name:  "string length",
			rules: []Rule{{Kind: KString}, {Kind: KLength, Args: map[string]any{"n": 3}}},
			value: "abc",
			valid: true,
		},
		{
			name:  "string length fail",
			rules: []Rule{{Kind: KString}, {Kind: KLength, Args: map[string]any{"n": 3}}},
			value: "ab",
			valid: false,
		},
		{
			name:  "int min/max",
			rules: []Rule{{Kind: KInt}, {Kind: KMinInt, Args: map[string]any{"n": int64(1)}}, {Kind: KMaxInt, Args: map[string]any{"n": int64(10)}}},
			value: int64(5),
			valid: true,
		},
		{
			name:  "int min/max fail",
			rules: []Rule{{Kind: KInt}, {Kind: KMinInt, Args: map[string]any{"n": int64(1)}}, {Kind: KMaxInt, Args: map[string]any{"n": int64(10)}}},
			value: int64(15),
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := compiler.Compile(tt.rules)
			err := validator(tt.value)
			if tt.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Errorf("expected invalid, got no error")
			}
		})
	}
}
