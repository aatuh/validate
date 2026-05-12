package validate

import (
	"errors"
	"testing"
	"time"
)

func TestRootFacade_PublicBuilderContracts(t *testing.T) {
	v := New()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		check   func(any) error
		valid   any
		invalid any
		code    string
	}{
		{"string min runes", v.String().MinRunes(2).Build(), "åb", "å", "string.minRunes"},
		{"string max runes", v.String().MaxRunes(2).Build(), "åb", "åbc", "string.maxRunes"},
		{"string nonempty", v.String().NonEmpty().Build(), "go", "", "string.nonempty"},
		{"string url", v.String().URL().Build(), "https://example.com", "example.com", "string.url"},
		{"string hostname", v.String().Hostname().Build(), "example.com", "-bad.example", "string.hostname"},
		{"string ip", v.String().IP().Build(), "127.0.0.1", "999.1.1.1", "string.ip"},
		{"string cidr", v.String().CIDR().Build(), "192.0.2.0/24", "192.0.2.1", "string.cidr"},
		{"string ascii", v.String().ASCII().Build(), "abc123", "åbc", "string.ascii"},
		{"string alpha", v.String().Alpha().Build(), "abc", "abc1", "string.alpha"},
		{"string alnum", v.String().Alnum().Build(), "abc123", "abc-123", "string.alnum"},
		{"int greater than", v.Int().GreaterThan(10).Build(), 11, 10, "number.gt"},
		{"int less than equal", v.Int().LessThanEqual(10).Build(), 10, 11, "number.lte"},
		{"int between", v.Int().Between(1, 3).Build(), 2, 4, "number.between"},
		{"int positive", v.Int().Positive().Build(), 1, 0, "number.positive"},
		{"float finite", v.Float().Finite().Build(), 1.5, 1.0 / zeroFloat(), "number.finite"},
		{"bool false", v.Bool().False().Build(), false, true, "bool.false"},
		{"slice min", v.Slice().MinLength(2).Build(), []string{"a", "b"}, []string{"a"}, "slice.min"},
		{"slice foreach rules", v.Slice().ForEachRules(NewRule(KString, nil), NewRule(KMinLength, map[string]any{"n": int64(2)})).Build(), []string{"go"}, []string{"g"}, "string.min"},
		{"array max", v.Array().MaxLength(1).Build(), [1]string{"a"}, [2]string{"a", "b"}, "array.max"},
		{"array foreach rules", v.Array().ForEachRules(NewRule(KString, nil), NewRule(KMinLength, map[string]any{"n": int64(2)})).Build(), [1]string{"go"}, [1]string{"g"}, "string.min"},
		{"map max keys", v.Map().MaxKeys(1).Build(), map[string]int{"a": 1}, map[string]int{"a": 1, "b": 2}, "map.maxkeys"},
		{"map values", v.Map().ValuesRules(NewRule(KInt, nil), NewRule(KPositive, nil)).Build(), map[string]int{"id": 1}, map[string]int{"id": 0}, "number.positive"},
		{"time not zero", v.Time().NotZero().Build(), start, time.Time{}, "time.notzero"},
		{"time between", v.Time().Between(start, end).Build(), start.Add(time.Hour), end.Add(time.Hour), "time.between"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.check(tt.valid); err != nil {
				t.Fatalf("valid value rejected: %v", err)
			}
			requireRootCode(t, tt.check(tt.invalid), tt.code)
		})
	}
}

func zeroFloat() float64 { return 0 }

func requireRootCode(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("got nil error, want code %q", want)
	}
	var es Errors
	if !errors.As(err, &es) || len(es) == 0 {
		t.Fatalf("got %T %v, want structured errors", err, err)
	}
	if es[0].Code != want {
		t.Fatalf("code = %q, want %q; errors=%#v", es[0].Code, want, es)
	}
}
