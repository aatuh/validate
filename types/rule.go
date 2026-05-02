package types

import "context"

// Kind represents the type of validation rule.
//
// This type is used to identify different validation rule types in the AST.
type Kind string

const (
	// String validation kinds
	KString      Kind = "string"
	KLength      Kind = "length"
	KMinLength   Kind = "minLength"
	KMaxLength   Kind = "maxLength"
	KRegex       Kind = "regex"
	KOneOf       Kind = "oneOf"
	KMinRunes    Kind = "minRunes"
	KMaxRunes    Kind = "maxRunes"
	KNonEmpty    Kind = "nonEmpty"
	KContains    Kind = "contains"
	KNotContains Kind = "notContains"
	KPrefix      Kind = "prefix"
	KSuffix      Kind = "suffix"
	KURL         Kind = "url"
	KHostname    Kind = "hostname"
	KIP          Kind = "ip"
	KIPv4        Kind = "ipv4"
	KIPv6        Kind = "ipv6"
	KCIDR        Kind = "cidr"
	KASCII       Kind = "ascii"
	KAlpha       Kind = "alpha"
	KAlnum       Kind = "alnum"

	// Generic modifiers
	KOmitempty Kind = "omitempty"
	KRequired  Kind = "required"

	// Integer validation kinds
	KInt              Kind = "int"
	KInt64            Kind = "int64"
	KMinInt           Kind = "minInt"
	KMaxInt           Kind = "maxInt"
	KFloat            Kind = "float"
	KMinNumber        Kind = "minNumber"
	KMaxNumber        Kind = "maxNumber"
	KGreaterThan      Kind = "greaterThan"
	KGreaterThanEqual Kind = "greaterThanEqual"
	KLessThan         Kind = "lessThan"
	KLessThanEqual    Kind = "lessThanEqual"
	KBetween          Kind = "between"
	KPositive         Kind = "positive"
	KNonNegative      Kind = "nonNegative"
	KFinite           Kind = "finite"

	// Slice validation kinds
	KSlice          Kind = "slice"
	KSliceLength    Kind = "sliceLength"
	KMinSliceLength Kind = "minSliceLength"
	KMaxSliceLength Kind = "maxSliceLength"
	KForEach        Kind = "forEach"
	KSliceUnique    Kind = "sliceUnique"
	KSliceContains  Kind = "sliceContains"

	// Array validation kinds
	KArray          Kind = "array"
	KArrayLength    Kind = "arrayLength"
	KMinArrayLength Kind = "minArrayLength"
	KMaxArrayLength Kind = "maxArrayLength"
	KArrayForEach   Kind = "arrayForEach"
	KArrayUnique    Kind = "arrayUnique"
	KArrayContains  Kind = "arrayContains"

	// Map validation kinds
	KMap        Kind = "map"
	KMapLength  Kind = "mapLength"
	KMinMapKeys Kind = "minMapKeys"
	KMaxMapKeys Kind = "maxMapKeys"
	KMapKeys    Kind = "mapKeys"
	KMapValues  Kind = "mapValues"

	// Boolean validation kinds
	KBool      Kind = "bool"
	KBoolTrue  Kind = "boolTrue"
	KBoolFalse Kind = "boolFalse"

	// Time validation kinds
	KTime        Kind = "time"
	KTimeNotZero Kind = "timeNotZero"
	KTimeBefore  Kind = "timeBefore"
	KTimeAfter   Kind = "timeAfter"
	KTimeBetween Kind = "timeBetween"
)

// Rule represents a single validation rule with its arguments.
//
// Fields:
//   - Kind: The type of validation rule.
//   - Args: Map of rule-specific arguments (e.g., {"n": int64(3),
//     "pattern": ".*"}).
//   - Elem: For nested rules (e.g., slice element validation).
type Rule struct {
	Kind Kind
	Args map[string]any // e.g. {"n": int64(3), "pattern": ".*"}
	Elem *Rule          // For nested rules (e.g., slice element validation)
}

// NewRuleWithElem builds a Rule with an element sub-rule for nesting.
func NewRuleWithElem(kind Kind, args map[string]any, elem *Rule) Rule {
	return Rule{Kind: kind, Args: args, Elem: elem}
}

// NewRule creates a new rule with the given kind and arguments.
func NewRule(kind Kind, args map[string]any) Rule {
	return Rule{
		Kind: kind,
		Args: args,
	}
}

// Deprecated: use NewRuleWithElem(kind, args, &elem) directly when you need
// to pass a value. This variant was kept for compatibility and will be
// removed in a future version.
func NewRuleWithElemValue(kind Kind, args map[string]any, elem Rule) Rule {
	return Rule{Kind: kind, Args: args, Elem: &elem}
}

// ValidatorFunc represents a compiled validation function.
type ValidatorFunc func(v any) error

// ContextValidatorFunc represents a compiled context-aware validation function.
type ContextValidatorFunc func(ctx context.Context, v any) error

// CompileOpts tunes rule compilation without changing existing defaults.
type CompileOpts struct {
	CollectAll bool
}

// FieldValidator represents a field-specific validation function.
type FieldValidator func(field any) error

// ContextRuleCompiler compiles one Rule into a context-aware validator.
type ContextRuleCompiler func(c *Compiler, rule Rule) (ContextValidatorFunc, error)
