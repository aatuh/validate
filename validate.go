package validate

import (
	"github.com/aatuh/validate/v3/glue"
	"github.com/aatuh/validate/v3/translator"

	// Ensure built-in plugin validators register themselves.
	_ "github.com/aatuh/validate/v3/validators/email"
	_ "github.com/aatuh/validate/v3/validators/ulid"
	_ "github.com/aatuh/validate/v3/validators/uuid"
)

// Re-export glue types for a developer-friendly root facade.
type Validate = glue.Validate
type StringBuilder = glue.StringBuilder
type IntBuilder = glue.IntBuilder
type BoolBuilder = glue.BoolBuilder
type SliceBuilder = glue.SliceBuilder

// New returns a Validate configured with sensible defaults.
//
// Defaults:
// - Installs default English translations via SimpleTranslator.
// - Registers built-in plugins (email, ulid, uuid) via blank imports.
func New() *Validate {
	v := glue.New()
	tr := translator.NewSimpleTranslator(
		translator.DefaultEnglishTranslations(),
	)
	return v.WithTranslator(tr)
}

// NewWithTranslator returns a Validate configured with the provided
// translator while keeping other defaults.
func NewWithTranslator(tr translator.Translator) *Validate {
	return glue.NewWithTranslator(tr)
}

// NewBare returns a Validate without installing a default translator.
// Useful for advanced setups that manage translations differently.
func NewBare() *Validate { return glue.NewBare() }

// FromTag compiles a single tag string using v (or a fresh instance).
func FromTag(v *Validate, tag string) (func(any) error, error) {
	if v == nil {
		v = New()
	}
	return v.FromTag(tag)
}
