package validate

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	verrs "github.com/aatuh/validate/v3/errors"
)

func TestRootDomainValidators_WorkAcrossPublicAPIs(t *testing.T) {
	v := New()

	tests := []struct {
		tag     string
		valid   string
		invalid string
		code    string
		empty   string
		build   func(*Validate) func(any) error
	}{
		{"slug", "alpha-123", "SECRET-token-123", "string.slug.invalid", "string.slug.invalid", func(v *Validate) func(any) error { return v.String().Slug().Build() }},
		{"semver", "1.2.3-alpha.1+build.5", "SECRET-token-123", "string.semver.invalid", "string.semver.invalid", func(v *Validate) func(any) error { return v.String().SemVer().Build() }},
		{"json", `{"ok":true}`, "SECRET-token-123", "string.json.invalid", "string.json.invalid", func(v *Validate) func(any) error { return v.String().JSON().Build() }},
		{"jwt", "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiIxMjMifQ.c2lnbmF0dXJl", "SECRET-token-123", "string.jwt.invalid", "string.jwt.invalid", func(v *Validate) func(any) error { return v.String().JWT().Build() }},
		{"base64", "dmFsaWQ=", "SECRET-token-123", "string.base64.invalid", "string.base64.invalid", func(v *Validate) func(any) error { return v.String().Base64().Build() }},
		{"base64url", "dmFsaWQ", "SECRET/token/123", "string.base64url.invalid", "string.base64url.invalid", func(v *Validate) func(any) error { return v.String().Base64URL().Build() }},
		{"hex", "deadBEEF", "SECRET-token-123", "string.hex.invalid", "string.hex.invalid", func(v *Validate) func(any) error { return v.String().Hex().Build() }},
		{"mac", "01:23:45:67:89:ab", "SECRET-token-123", "string.mac.invalid", "string.mac.invalid", func(v *Validate) func(any) error { return v.String().MAC().Build() }},
		{"e164", "+358401234567", "SECRET-token-123", "string.e164.invalid", "string.e164.invalid", func(v *Validate) func(any) error { return v.String().E164().Build() }},
		{"fqdn", "api.example.com", "SECRET-token-123", "string.fqdn.invalid", "string.fqdn.invalid", func(v *Validate) func(any) error { return v.String().FQDN().Build() }},
		{"date", "2026-05-08", "SECRET-token-123", "string.date.invalid", "string.date.invalid", func(v *Validate) func(any) error { return v.String().Date().Build() }},
		{"rfc3339", "2026-05-08T10:30:00Z", "SECRET-token-123", "string.rfc3339.invalid", "string.rfc3339.invalid", func(v *Validate) func(any) error { return v.String().RFC3339().Build() }},
		{"luhn", "79927398713", "SECRET-token-123", "string.luhn.invalid", "string.luhn.invalid", func(v *Validate) func(any) error { return v.String().Luhn().Build() }},
		{"uuidv1", "6ba7b810-9dad-11d1-80b4-00c04fd430c8", "550e8400-e29b-41d4-a716-446655440000", "string.uuid.version", "string.uuid.invalid", func(v *Validate) func(any) error { return v.String().UUIDv1().Build() }},
		{"uuidv3", "6fa459ea-ee8a-3ca4-894e-db77e160355e", "550e8400-e29b-41d4-a716-446655440000", "string.uuid.version", "string.uuid.invalid", func(v *Validate) func(any) error { return v.String().UUIDv3().Build() }},
		{"uuidv4", "550e8400-e29b-41d4-a716-446655440000", "6ba7b810-9dad-11d1-80b4-00c04fd430c8", "string.uuid.version", "string.uuid.invalid", func(v *Validate) func(any) error { return v.String().UUIDv4().Build() }},
		{"uuidv5", "2ed6657d-e927-568b-95e1-2665a8aea6a2", "550e8400-e29b-41d4-a716-446655440000", "string.uuid.version", "string.uuid.invalid", func(v *Validate) func(any) error { return v.String().UUIDv5().Build() }},
		{"uuidv6", "1ef21d2f-1207-6660-8c4f-419efbd44d48", "550e8400-e29b-41d4-a716-446655440000", "string.uuid.version", "string.uuid.invalid", func(v *Validate) func(any) error { return v.String().UUIDv6().Build() }},
		{"uuidv7", "01890f13-a93c-7cc2-98e5-9f8c7e2b8a6f", "550e8400-e29b-41d4-a716-446655440000", "string.uuid.version", "string.uuid.invalid", func(v *Validate) func(any) error { return v.String().UUIDv7().Build() }},
		{"uuidv8", "01890f13-a93c-8cc2-98e5-9f8c7e2b8a6f", "550e8400-e29b-41d4-a716-446655440000", "string.uuid.version", "string.uuid.invalid", func(v *Validate) func(any) error { return v.String().UUIDv8().Build() }},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			tag := "string;" + tt.tag
			if err := v.CheckTag(tag, tt.valid); err != nil {
				t.Fatalf("CheckTag valid failed: %v", err)
			}
			fn, err := v.FromTag(tag)
			if err != nil {
				t.Fatalf("FromTag failed: %v", err)
			}
			if err := fn(tt.valid); err != nil {
				t.Fatalf("FromTag validator failed: %v", err)
			}
			if err := tt.build(v)(tt.valid); err != nil {
				t.Fatalf("builder valid failed: %v", err)
			}
			manual := v.CompileRules([]Rule{
				NewRule(KString, nil),
				NewRule(Kind(tt.tag), nil),
			})
			if err := manual(tt.valid); err != nil {
				t.Fatalf("manual rules valid failed: %v", err)
			}
			requireRootDomainCode(t, v.CheckTag(tag, tt.invalid), tt.code)
			requireRootDomainCode(t, v.CheckTag(tag, 123), verrs.CodeStringType)
			requireRootDomainCode(t, v.CheckTag(tag, ""), tt.empty)
			requireRootDomainCode(t, v.CheckTag("string;required;"+tt.tag, ""), verrs.CodeRequired)
			requireNoRootDomainLeak(t, v.CheckTag(tag, tt.invalid), tt.invalid)

			requireRootDomainStructValid(t, v, tag, tt.valid)
			if err := v.CheckTag("slice;foreach=("+tag+")", []string{tt.valid}); err != nil {
				t.Fatalf("nested slice valid failed: %v", err)
			}
			if err := v.CheckTag("map;values=("+tag+")", map[string]string{"id": tt.valid}); err != nil {
				t.Fatalf("nested map values valid failed: %v", err)
			}
		})
	}
}

func TestRootArrayValidation_TagsBuilderAndPaths(t *testing.T) {
	v := New()

	if err := v.CheckTag("array;len=2;foreach=(string;slug)", [2]string{"alpha", "beta"}); err != nil {
		t.Fatalf("array tag valid failed: %v", err)
	}
	requireRootDomainCode(t, v.CheckTag("array;len=2", []string{"alpha", "beta"}), "array.type")
	requireRootDomainCode(t, v.CheckTag("array;len=3", [2]string{"alpha", "beta"}), "array.length")
	requireRootDomainCode(t, v.CheckTag("array;unique", [2]string{"alpha", "alpha"}), "array.unique")
	requireRootDomainCode(t, v.CheckTag("array;contains=alpha", [2]string{"beta", "gamma"}), "array.contains")
	requireRootDomainPathCode(t, v.CheckTag("array;foreach=(string;slug)", [2]string{"alpha", "bad_slug"}), "[1]", "string.slug.invalid")

	if _, err := v.FromTag("array;foreach=string"); err == nil {
		t.Fatalf("malformed array foreach tag compiled")
	}
	builder := v.Array().Length(2).ForEachRules(NewRule(KString, nil), NewRule(Kind("slug"), nil)).Build()
	if err := builder([2]string{"alpha", "beta"}); err != nil {
		t.Fatalf("array builder valid failed: %v", err)
	}
	manual := v.CompileRules([]Rule{
		NewRule(KArray, nil),
		NewRule(KArrayLength, map[string]any{"n": 2}),
	})
	if err := manual([2]string{"alpha", "beta"}); err != nil {
		t.Fatalf("array manual rules valid failed: %v", err)
	}
}

func requireRootDomainStructValid(t *testing.T, v *Validate, tag, value string) {
	t.Helper()
	st := reflect.StructOf([]reflect.StructField{{
		Name: "Value",
		Type: reflect.TypeOf(""),
		Tag:  reflect.StructTag(`validate:"` + tag + `"`),
	}})
	rv := reflect.New(st).Elem()
	rv.Field(0).SetString(value)
	if err := v.ValidateStruct(rv.Interface()); err != nil {
		t.Fatalf("ValidateStruct(%s) failed: %v", tag, err)
	}
}

func requireRootDomainCode(t *testing.T, err error, want string) {
	t.Helper()
	requireRootDomainPathCode(t, err, "", want)
}

func requireRootDomainPathCode(t *testing.T, err error, path, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("got nil error, want code %q", want)
	}
	var es Errors
	if !errors.As(err, &es) || len(es) == 0 {
		t.Fatalf("got %T %v, want structured errors", err, err)
	}
	if es[0].Path != path || es[0].Code != want {
		t.Fatalf("first error = %#v, want path %q code %q", es[0], path, want)
	}
}

func requireNoRootDomainLeak(t *testing.T, err error, value string) {
	t.Helper()
	var es Errors
	if !errors.As(err, &es) {
		t.Fatalf("got %T %v, want structured errors", err, err)
	}
	for _, fe := range es {
		if strings.Contains(fe.Msg, value) {
			t.Fatalf("message leaked submitted value: %#v", fe)
		}
	}
}
