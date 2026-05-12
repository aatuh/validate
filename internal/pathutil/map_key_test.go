package pathutil

import "testing"

func TestMapKeySegmentPolicy(t *testing.T) {
	tests := []struct {
		name string
		key  any
		want string
	}{
		{"nil", nil, "[<nil>]"},
		{"short string", "user_id", "[user_id]"},
		{"ordinary punctuation", "items.v1:sku-1", "[items.v1:sku-1]"},
		{"bool", true, "[true]"},
		{"int", int64(42), "[42]"},
		{"negative int", -7, "[-7]"},
		{"uint", uint(9), "[9]"},
		{"float", 3.5, "[3.5]"},
		{"long string", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "[<redacted>]"},
		{"password marker", "PasswordHash", "[<redacted>]"},
		{"token marker", "api_token", "[<redacted>]"},
		{"email marker", "user@example.com", "[<redacted>]"},
		{"url marker", "https://example.test/id", "[<redacted>]"},
		{"escaping sensitive", "user/name", "[<redacted>]"},
		{"complex key", struct{ ID string }{ID: "abc"}, "[<redacted>]"},
		{"complex number", complex(1, 2), "[<redacted>]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapKeySegment(tt.key); got != tt.want {
				t.Fatalf("MapKeySegment(%#v) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}
