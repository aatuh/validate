package types

import (
	"strings"
	"testing"

	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/translator"
)

func TestGlobalTypeRegistry_OverwriteUsesLatestFactory(t *testing.T) {
	name := uniqueTypeName(t)
	RegisterGlobalType(name, registryTestFactory{code: "type.first"})
	RegisterGlobalType(name, registryTestFactory{code: "type.second"})

	validator, ok := GetGlobalTypeValidator(name, nil)
	if !ok {
		t.Fatalf("global type %q was not registered", name)
	}
	requireErrorsWithCode(t, validator.Validate("value"), "type.second")
}

func TestTypeRegistry_CloneIsIsolated(t *testing.T) {
	name := uniqueTypeName(t)
	registry := NewTypeRegistry()
	registry.RegisterType(name, registryTestFactory{code: "type.original"})

	clone := registry.Clone()
	registry.RegisterType(name, registryTestFactory{code: "type.changed"})

	validator, ok := clone.GetTypeValidator(name, nil)
	if !ok {
		t.Fatalf("cloned type %q was not registered", name)
	}
	requireErrorsWithCode(t, validator.Validate("value"), "type.original")
}

func uniqueTypeName(t *testing.T) string {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	return "audit_" + name
}

type registryTestFactory struct {
	code string
}

func (f registryTestFactory) CreateValidator(_ translator.Translator) TypeValidator {
	return registryTestValidator{code: f.code}
}

type registryTestValidator struct {
	code string
}

func (v registryTestValidator) Validate(any) error {
	return verrs.Errors{verrs.FieldError{Code: v.code}}
}
