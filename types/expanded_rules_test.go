package types

import (
	"errors"
	"reflect"
	"testing"
	"time"

	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/translator"
)

func TestParseTag_DocumentedAliasesAndExpandedRules(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want []Kind
	}{
		{
			name: "string aliases and predicates",
			tag:  "string;required;len=5;minRunes=2;maxRunes=5;contains=el;notContains=x;prefix=h;suffix=o;url;hostname;ip;ipv4;ipv6;cidr;ascii;alpha;alnum;nonempty",
			want: []Kind{
				KString, KRequired, KLength, KMinRunes, KMaxRunes, KContains,
				KNotContains, KPrefix, KSuffix, KURL, KHostname, KIP, KIPv4,
				KIPv6, KCIDR, KASCII, KAlpha, KAlnum, KNonEmpty,
			},
		},
		{
			name: "slice len alias and collection rules",
			tag:  "slice;len=2;unique;contains=a;foreach=(string;min=1)",
			want: []Kind{KSlice, KSliceLength, KSliceUnique, KSliceContains, KForEach},
		},
		{
			name: "array len alias and collection rules",
			tag:  "array;len=2;unique;contains=a;foreach=(string;min=1)",
			want: []Kind{KArray, KArrayLength, KArrayUnique, KArrayContains, KArrayForEach},
		},
		{
			name: "float rules",
			tag:  "float;finite;min=1.5;max=9.5;gt=1;gte=2;lt=10;lte=9;between=2,8;positive;nonnegative",
			want: []Kind{
				KFloat, KFinite, KMinNumber, KMaxNumber, KGreaterThan,
				KGreaterThanEqual, KLessThan, KLessThanEqual, KBetween,
				KPositive, KNonNegative,
			},
		},
		{
			name: "map rules",
			tag:  "map;len=2;minKeys=1;maxKeys=3;keys=(string;min=1);values=(int;min=1)",
			want: []Kind{KMap, KMapLength, KMinMapKeys, KMaxMapKeys, KMapKeys, KMapValues},
		},
		{
			name: "time rules",
			tag:  "time;notzero;after=2026-01-01T00:00:00Z;before=2027-01-01T00:00:00Z;between=2026-01-01T00:00:00Z,2027-01-01T00:00:00Z",
			want: []Kind{KTime, KTimeNotZero, KTimeAfter, KTimeBefore, KTimeBetween},
		},
		{
			name: "generic required only",
			tag:  "required",
			want: []Kind{KRequired},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTag(tt.tag)
			if err != nil {
				t.Fatalf("ParseTag(%q): %v", tt.tag, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("rule count = %d, want %d: %#v", len(got), len(tt.want), got)
			}
			for i, want := range tt.want {
				if got[i].Kind != want {
					t.Fatalf("rule[%d] = %s, want %s", i, got[i].Kind, want)
				}
			}
		})
	}
}

func TestCompiler_ExpandedRuleBehavior(t *testing.T) {
	tr := translator.NewSimpleTranslator(translator.DefaultEnglishTranslations())
	c := NewCompiler(tr)

	checkTag := func(t *testing.T, tag string) ValidatorFunc {
		t.Helper()
		rules, err := ParseTag(tag)
		if err != nil {
			t.Fatalf("ParseTag(%q): %v", tag, err)
		}
		return c.Compile(rules)
	}

	tests := []struct {
		name    string
		tag     string
		valid   any
		invalid any
		code    string
	}{
		{"string contains", "string;contains=go", "gopher", "java", verrs.CodeStringContains},
		{"string url", "string;url", "https://example.com/a", "not a url", verrs.CodeStringURL},
		{"string ipv4", "string;ipv4", "127.0.0.1", "::1", verrs.CodeStringIP},
		{"float finite", "float;finite;between=1,2", 1.5, 3.0, verrs.CodeNumberBetween},
		{"bool true", "bool;true", true, false, verrs.CodeBoolTrue},
		{"slice unique", "slice;unique", []string{"a", "b"}, []string{"a", "a"}, verrs.CodeSliceUnique},
		{"array unique", "array;unique", [2]string{"a", "b"}, [2]string{"a", "a"}, verrs.CodeArrayUnique},
		{"map min", "map;minKeys=1", map[string]int{"a": 1}, map[string]int{}, verrs.CodeMapMinKeys},
		{"time after", "time;after=2026-01-01T00:00:00Z", time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), verrs.CodeTimeAfter},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := checkTag(t, tt.tag)
			if err := fn(tt.valid); err != nil {
				t.Fatalf("valid input failed: %v", err)
			}
			err := fn(tt.invalid)
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

func TestParseTag_MapNestedRulesPreserveInnerSemicolons(t *testing.T) {
	rules, err := ParseTag("map;keys=(string;min=1);values=(slice;foreach=(string;min=2))")
	if err != nil {
		t.Fatalf("ParseTag: %v", err)
	}
	if rules[1].Kind != KMapKeys || rules[2].Kind != KMapValues {
		t.Fatalf("unexpected map rule kinds: %#v", rules)
	}
	valueRules, ok := rules[2].Args["rules"].([]Rule)
	if !ok {
		t.Fatalf("map values rules missing: %#v", rules[2].Args)
	}
	if !reflect.DeepEqual([]Kind{KSlice, KForEach}, []Kind{valueRules[0].Kind, valueRules[1].Kind}) {
		t.Fatalf("unexpected nested value rules: %#v", valueRules)
	}
}

func TestParseTag_CustomRulesAcrossBaseTypes(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		wantBase Kind
		wantRule Kind
		wantArg  string
	}{
		{"int bare", "int;even", KInt, "even", ""},
		{"float bare", "float;finite;money", KFloat, "money", ""},
		{"bool bare", "bool;truthy", KBool, "truthy", ""},
		{"slice bare", "slice;nonemptyElements", KSlice, "nonemptyElements", ""},
		{"array bare", "array;nonemptyElements", KArray, "nonemptyElements", ""},
		{"map bare", "map;customRule", KMap, "customRule", ""},
		{"time bare", "time;businessDay", KTime, "businessDay", ""},
		{"custom raw value", "int;custom:mod=2", KInt, "mod", "2"},
		{"custom raw empty", "string;custom:presence", KString, "presence", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, err := ParseTag(tt.tag)
			if err != nil {
				t.Fatalf("ParseTag(%q): %v", tt.tag, err)
			}
			if len(rules) < 2 {
				t.Fatalf("rules = %#v, want base plus custom rule", rules)
			}
			if rules[0].Kind != tt.wantBase || rules[len(rules)-1].Kind != tt.wantRule {
				t.Fatalf("kinds = %#v, want base %s and custom %s", rules, tt.wantBase, tt.wantRule)
			}
			if tt.wantArg != "" {
				if got := rules[len(rules)-1].Args["value"]; got != tt.wantArg {
					t.Fatalf("custom value arg = %#v, want %q", got, tt.wantArg)
				}
			}
		})
	}
}

func TestParseTag_CustomRulesRejectMalformedBuiltInArgs(t *testing.T) {
	for _, tag := range []string{
		"int;min=bad",
		"float;between=1",
		"slice;max=bad",
		"map;minKeys=bad",
		"time;after=not-rfc3339",
	} {
		t.Run(tag, func(t *testing.T) {
			if _, err := ParseTag(tag); err == nil {
				t.Fatalf("ParseTag(%q) succeeded, want malformed built-in arg error", tag)
			}
		})
	}
}
