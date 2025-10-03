package types

// Kind represents the type of validation rule.
//
// This type is used to identify different validation rule types in the AST.
type Kind string

const (
	// String validation kinds
	KString    Kind = "string"
	KLength    Kind = "length"
	KMinLength Kind = "minLength"
	KMaxLength Kind = "maxLength"
	KRegex     Kind = "regex"
	KOneOf     Kind = "oneOf"
	KMinRunes  Kind = "minRunes"
	KMaxRunes  Kind = "maxRunes"

	// Integer validation kinds
	KInt    Kind = "int"
	KInt64  Kind = "int64"
	KMinInt Kind = "minInt"
	KMaxInt Kind = "maxInt"

	// Slice validation kinds
	KSlice          Kind = "slice"
	KSliceLength    Kind = "sliceLength"
	KMinSliceLength Kind = "minSliceLength"
	KMaxSliceLength Kind = "maxSliceLength"
	KForEach        Kind = "forEach"

	// Boolean validation kinds
	KBool Kind = "bool"
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

// FieldValidator represents a field-specific validation function.
type FieldValidator func(field any) error
