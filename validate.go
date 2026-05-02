package validate

import (
	"context"

	"github.com/aatuh/validate/v3/core"
	"github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/glue"
	"github.com/aatuh/validate/v3/structvalidator"
	"github.com/aatuh/validate/v3/translator"
	"github.com/aatuh/validate/v3/types"

	// Ensure built-in plugin validators register themselves.
	_ "github.com/aatuh/validate/v3/validators/domain"
	_ "github.com/aatuh/validate/v3/validators/email"
	_ "github.com/aatuh/validate/v3/validators/ulid"
	_ "github.com/aatuh/validate/v3/validators/uuid"
)

// Re-export types for a developer-friendly root facade.
type Validate = glue.Validate
type StringBuilder = glue.StringBuilder
type IntBuilder = glue.IntBuilder
type FloatBuilder = glue.FloatBuilder
type BoolBuilder = glue.BoolBuilder
type SliceBuilder = glue.SliceBuilder
type ArrayBuilder = glue.ArrayBuilder
type MapBuilder = glue.MapBuilder
type TimeBuilder = glue.TimeBuilder
type CustomTypeBuilder = glue.CustomTypeBuilder
type Errors = errors.Errors
type ValidateOpts = core.ValidateOpts

// Re-export types package for manual rule construction
type Rule = types.Rule
type Kind = types.Kind
type ValidatorFunc = types.ValidatorFunc
type ContextValidatorFunc = types.ContextValidatorFunc
type CompileOpts = types.CompileOpts
type RuleCompiler = types.RuleCompiler
type ContextRuleCompiler = types.ContextRuleCompiler
type TypeValidator = types.TypeValidator
type TypeValidatorFactory = types.TypeValidatorFactory
type StructRuleContext = core.StructRuleContext
type StructRuleFunc = core.StructRuleFunc
type StructRuleCompiler = core.StructRuleCompiler

// Re-export commonly used rule kinds
const (
	// String validation kinds
	KString      = types.KString
	KLength      = types.KLength
	KMinLength   = types.KMinLength
	KMaxLength   = types.KMaxLength
	KRegex       = types.KRegex
	KOneOf       = types.KOneOf
	KMinRunes    = types.KMinRunes
	KMaxRunes    = types.KMaxRunes
	KNonEmpty    = types.KNonEmpty
	KContains    = types.KContains
	KNotContains = types.KNotContains
	KPrefix      = types.KPrefix
	KSuffix      = types.KSuffix
	KURL         = types.KURL
	KHostname    = types.KHostname
	KIP          = types.KIP
	KIPv4        = types.KIPv4
	KIPv6        = types.KIPv6
	KCIDR        = types.KCIDR
	KASCII       = types.KASCII
	KAlpha       = types.KAlpha
	KAlnum       = types.KAlnum

	// Generic modifiers
	KOmitempty = types.KOmitempty
	KRequired  = types.KRequired

	// Integer validation kinds
	KInt              = types.KInt
	KInt64            = types.KInt64
	KMinInt           = types.KMinInt
	KMaxInt           = types.KMaxInt
	KFloat            = types.KFloat
	KMinNumber        = types.KMinNumber
	KMaxNumber        = types.KMaxNumber
	KGreaterThan      = types.KGreaterThan
	KGreaterThanEqual = types.KGreaterThanEqual
	KLessThan         = types.KLessThan
	KLessThanEqual    = types.KLessThanEqual
	KBetween          = types.KBetween
	KPositive         = types.KPositive
	KNonNegative      = types.KNonNegative
	KFinite           = types.KFinite

	// Slice validation kinds
	KSlice          = types.KSlice
	KSliceLength    = types.KSliceLength
	KMinSliceLength = types.KMinSliceLength
	KMaxSliceLength = types.KMaxSliceLength
	KForEach        = types.KForEach
	KSliceUnique    = types.KSliceUnique
	KSliceContains  = types.KSliceContains

	// Array validation kinds
	KArray          = types.KArray
	KArrayLength    = types.KArrayLength
	KMinArrayLength = types.KMinArrayLength
	KMaxArrayLength = types.KMaxArrayLength
	KArrayForEach   = types.KArrayForEach
	KArrayUnique    = types.KArrayUnique
	KArrayContains  = types.KArrayContains

	// Map validation kinds
	KMap        = types.KMap
	KMapLength  = types.KMapLength
	KMinMapKeys = types.KMinMapKeys
	KMaxMapKeys = types.KMaxMapKeys
	KMapKeys    = types.KMapKeys
	KMapValues  = types.KMapValues

	// Boolean validation kinds
	KBool      = types.KBool
	KBoolTrue  = types.KBoolTrue
	KBoolFalse = types.KBoolFalse

	// Time validation kinds
	KTime        = types.KTime
	KTimeNotZero = types.KTimeNotZero
	KTimeBefore  = types.KTimeBefore
	KTimeAfter   = types.KTimeAfter
	KTimeBetween = types.KTimeBetween
)

// Re-export translator package
type Translator = translator.Translator
type SimpleTranslator = translator.SimpleTranslator

// Re-export translator functions
var (
	NewSimpleTranslator                = translator.NewSimpleTranslator
	DefaultEnglishTranslations         = translator.DefaultEnglishTranslations
	MergeTranslations                  = translator.MergeTranslations
	RegisterDefaultEnglishTranslations = translator.RegisterDefaultEnglishTranslations
	JSONFieldName                      = structvalidator.JSONFieldName
)

// Re-export types functions
var (
	NewRule            = types.NewRule
	RegisterRule       = types.RegisterRule
	RegisterGlobalType = types.RegisterGlobalType
)

// New returns a Validate configured with sensible defaults.
//
// Defaults:
// - Installs default English translations via SimpleTranslator.
// - Registers built-in plugins (domain, email, ulid, uuid) via blank imports.
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

// FromTagWithOpts compiles a single tag string using v (or a fresh instance)
// with compile options.
func FromTagWithOpts(v *Validate, tag string, opts CompileOpts) (func(any) error, error) {
	if v == nil {
		v = New()
	}
	return v.FromTagWithOpts(tag, opts)
}

// FromTagContext compiles a single tag string into a context-aware validator
// using v (or a fresh instance).
func FromTagContext(v *Validate, tag string) (ContextValidatorFunc, error) {
	if v == nil {
		v = New()
	}
	return v.FromTagContext(tag)
}

// FromTagContextWithOpts compiles a single tag string into a context-aware
// validator using v (or a fresh instance) with compile options.
func FromTagContextWithOpts(v *Validate, tag string, opts CompileOpts) (ContextValidatorFunc, error) {
	if v == nil {
		v = New()
	}
	return v.FromTagContextWithOpts(tag, opts)
}

// CheckTagWithOpts compiles a tag and validates a single value with compile
// options using v (or a fresh instance).
func CheckTagWithOpts(v *Validate, tag string, value any, opts CompileOpts) error {
	if v == nil {
		v = New()
	}
	return v.CheckTagWithOpts(tag, value, opts)
}

// CheckTagContext compiles a tag and validates a single value with context.
func CheckTagContext(ctx context.Context, v *Validate, tag string, value any) error {
	if v == nil {
		v = New()
	}
	return v.CheckTagContext(ctx, tag, value)
}

// CheckTagContextWithOpts compiles a tag and validates a single value with
// context and compile options using v (or a fresh instance).
func CheckTagContextWithOpts(ctx context.Context, v *Validate, tag string, value any, opts CompileOpts) error {
	if v == nil {
		v = New()
	}
	return v.CheckTagContextWithOpts(ctx, tag, value, opts)
}

// ValidateStruct validates a struct using v (or a fresh instance).
func ValidateStruct(v *Validate, s any) error {
	if v == nil {
		v = New()
	}
	return v.ValidateStruct(s)
}

// ValidateStructWithOpts validates a struct using v (or a fresh instance) with
// struct validation options.
func ValidateStructWithOpts(v *Validate, s any, opts ValidateOpts) error {
	if v == nil {
		v = New()
	}
	return v.ValidateStructWithOpts(s, opts)
}

// ValidateStructContext validates a struct with context using v (or a fresh instance).
func ValidateStructContext(ctx context.Context, v *Validate, s any) error {
	if v == nil {
		v = New()
	}
	return v.ValidateStructContext(ctx, s)
}

// ValidateStructContextWithOpts validates a struct with context using v (or a
// fresh instance) and struct validation options.
func ValidateStructContextWithOpts(ctx context.Context, v *Validate, s any, opts ValidateOpts) error {
	if v == nil {
		v = New()
	}
	return v.ValidateStructContextWithOpts(ctx, s, opts)
}
