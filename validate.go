package validate

import (
	"github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/glue"
	"github.com/aatuh/validate/v3/translator"
	"github.com/aatuh/validate/v3/types"

	// Ensure built-in plugin validators register themselves.
	_ "github.com/aatuh/validate/v3/validators/email"
	_ "github.com/aatuh/validate/v3/validators/ulid"
	_ "github.com/aatuh/validate/v3/validators/uuid"
)

// Re-export types for a developer-friendly root facade.
type Validate = glue.Validate
type StringBuilder = glue.StringBuilder
type IntBuilder = glue.IntBuilder
type BoolBuilder = glue.BoolBuilder
type SliceBuilder = glue.SliceBuilder
type CustomTypeBuilder = glue.CustomTypeBuilder
type Errors = errors.Errors

// Re-export types package for manual rule construction
type Rule = types.Rule
type Kind = types.Kind
type ValidatorFunc = types.ValidatorFunc

// Re-export commonly used rule kinds
const (
	// String validation kinds
	KString    = types.KString
	KLength    = types.KLength
	KMinLength = types.KMinLength
	KMaxLength = types.KMaxLength
	KRegex     = types.KRegex
	KOneOf     = types.KOneOf
	KMinRunes  = types.KMinRunes
	KMaxRunes  = types.KMaxRunes

	// Generic modifiers
	KOmitempty = types.KOmitempty

	// Integer validation kinds
	KInt    = types.KInt
	KInt64  = types.KInt64
	KMinInt = types.KMinInt
	KMaxInt = types.KMaxInt

	// Slice validation kinds
	KSlice          = types.KSlice
	KSliceLength    = types.KSliceLength
	KMinSliceLength = types.KMinSliceLength
	KMaxSliceLength = types.KMaxSliceLength
	KForEach        = types.KForEach

	// Boolean validation kinds
	KBool = types.KBool
)

// Re-export translator package
type Translator = translator.Translator
type SimpleTranslator = translator.SimpleTranslator

// Re-export translator functions
var (
	NewSimpleTranslator        = translator.NewSimpleTranslator
	DefaultEnglishTranslations = translator.DefaultEnglishTranslations
)

// Re-export types functions
var (
	NewRule = types.NewRule
)

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

// ValidateStruct validates a struct using v (or a fresh instance).
func ValidateStruct(v *Validate, s any) error {
	if v == nil {
		v = New()
	}
	return v.ValidateStruct(s)
}
