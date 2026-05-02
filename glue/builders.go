package glue

import (
	"time"

	"github.com/aatuh/validate/v3/core"
	"github.com/aatuh/validate/v3/types"
)

// StringBuilder accumulates string validation rules.
type StringBuilder struct {
	rules  []types.Rule
	engine *core.Engine
}

func (b *StringBuilder) Length(n int) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KLength, map[string]any{"n": int64(n)}))
	return b
}

func (b *StringBuilder) Required() *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KRequired, nil))
	return b
}

func (b *StringBuilder) MinLength(n int) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMinLength, map[string]any{"n": int64(n)}))
	return b
}

func (b *StringBuilder) MaxLength(n int) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMaxLength, map[string]any{"n": int64(n)}))
	return b
}

func (b *StringBuilder) OneOf(vals ...string) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KOneOf, map[string]any{"values": vals}))
	return b
}

func (b *StringBuilder) MinRunes(n int) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMinRunes, map[string]any{"n": int64(n)}))
	return b
}

func (b *StringBuilder) MaxRunes(n int) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMaxRunes, map[string]any{"n": int64(n)}))
	return b
}

func (b *StringBuilder) Regex(pat string) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KRegex, map[string]any{"pattern": pat}))
	return b
}

func (b *StringBuilder) NonEmpty() *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KNonEmpty, nil))
	return b
}

func (b *StringBuilder) Contains(value string) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KContains, map[string]any{"value": value}))
	return b
}

func (b *StringBuilder) NotContains(value string) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KNotContains, map[string]any{"value": value}))
	return b
}

func (b *StringBuilder) Prefix(value string) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KPrefix, map[string]any{"value": value}))
	return b
}

func (b *StringBuilder) Suffix(value string) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KSuffix, map[string]any{"value": value}))
	return b
}

func (b *StringBuilder) URL() *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KURL, nil))
	return b
}

func (b *StringBuilder) Hostname() *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KHostname, nil))
	return b
}

func (b *StringBuilder) IP() *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KIP, nil))
	return b
}

func (b *StringBuilder) IPv4() *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KIPv4, nil))
	return b
}

func (b *StringBuilder) IPv6() *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KIPv6, nil))
	return b
}

func (b *StringBuilder) CIDR() *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KCIDR, nil))
	return b
}

func (b *StringBuilder) ASCII() *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KASCII, nil))
	return b
}

func (b *StringBuilder) Alpha() *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KAlpha, nil))
	return b
}

func (b *StringBuilder) Alnum() *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KAlnum, nil))
	return b
}

func (b *StringBuilder) Slug() *StringBuilder {
	return b.Rule("slug", nil)
}

func (b *StringBuilder) SemVer() *StringBuilder {
	return b.Rule("semver", nil)
}

func (b *StringBuilder) JSON() *StringBuilder {
	return b.Rule("json", nil)
}

func (b *StringBuilder) JWT() *StringBuilder {
	return b.Rule("jwt", nil)
}

func (b *StringBuilder) Base64() *StringBuilder {
	return b.Rule("base64", nil)
}

func (b *StringBuilder) Base64URL() *StringBuilder {
	return b.Rule("base64url", nil)
}

func (b *StringBuilder) Hex() *StringBuilder {
	return b.Rule("hex", nil)
}

func (b *StringBuilder) MAC() *StringBuilder {
	return b.Rule("mac", nil)
}

func (b *StringBuilder) E164() *StringBuilder {
	return b.Rule("e164", nil)
}

func (b *StringBuilder) FQDN() *StringBuilder {
	return b.Rule("fqdn", nil)
}

func (b *StringBuilder) Date() *StringBuilder {
	return b.Rule("date", nil)
}

func (b *StringBuilder) RFC3339() *StringBuilder {
	return b.Rule("rfc3339", nil)
}

func (b *StringBuilder) Luhn() *StringBuilder {
	return b.Rule("luhn", nil)
}

func (b *StringBuilder) UUIDv1() *StringBuilder {
	return b.Rule("uuidv1", nil)
}

func (b *StringBuilder) UUIDv3() *StringBuilder {
	return b.Rule("uuidv3", nil)
}

func (b *StringBuilder) UUIDv4() *StringBuilder {
	return b.Rule("uuidv4", nil)
}

func (b *StringBuilder) UUIDv5() *StringBuilder {
	return b.Rule("uuidv5", nil)
}

func (b *StringBuilder) UUIDv6() *StringBuilder {
	return b.Rule("uuidv6", nil)
}

func (b *StringBuilder) UUIDv7() *StringBuilder {
	return b.Rule("uuidv7", nil)
}

func (b *StringBuilder) UUIDv8() *StringBuilder {
	return b.Rule("uuidv8", nil)
}

func (b *StringBuilder) Rule(kind types.Kind, args map[string]any) *StringBuilder {
	b.rules = append(b.rules, types.NewRule(kind, args))
	return b
}

func (b *StringBuilder) OmitEmpty() *StringBuilder {
	b.rules = append(b.rules, types.NewRule(types.KOmitempty, nil))
	return b
}

func (b *StringBuilder) Build() func(any) error {
	return b.engine.CompileRules(b.rules)
}

func (b *StringBuilder) BuildWithOpts(opts types.CompileOpts) func(any) error {
	return b.engine.CompileRulesWithOpts(b.rules, opts)
}

func (b *StringBuilder) BuildAll() func(any) error {
	return b.BuildWithOpts(types.CompileOpts{CollectAll: true})
}

func (b *StringBuilder) BuildContext() types.ContextValidatorFunc {
	return b.engine.CompileRulesContext(b.rules)
}

func (b *StringBuilder) BuildContextWithOpts(opts types.CompileOpts) types.ContextValidatorFunc {
	return b.engine.CompileRulesContextWithOpts(b.rules, opts)
}

// IntBuilder accumulates integer validation rules.
type IntBuilder struct {
	rules  []types.Rule
	exact  bool
	engine *core.Engine
}

// NewIntBuilder creates a new IntBuilder with the base type rule.
func NewIntBuilder(exact bool, engine *core.Engine) *IntBuilder {
	builder := &IntBuilder{
		rules:  []types.Rule{},
		exact:  exact,
		engine: engine,
	}

	// Set the base type rule
	if exact {
		builder.rules = append(builder.rules, types.NewRule(types.KInt64, map[string]any{}))
	} else {
		builder.rules = append(builder.rules, types.NewRule(types.KInt, map[string]any{}))
	}

	return builder
}

func (b *IntBuilder) MinInt(n int64) *IntBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMinInt, map[string]any{"n": n}))
	return b
}

func (b *IntBuilder) Required() *IntBuilder {
	b.rules = append(b.rules, types.NewRule(types.KRequired, nil))
	return b
}

func (b *IntBuilder) MaxInt(n int64) *IntBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMaxInt, map[string]any{"n": n}))
	return b
}

func (b *IntBuilder) GreaterThan(n int64) *IntBuilder {
	b.rules = append(b.rules, types.NewRule(types.KGreaterThan, map[string]any{"n": float64(n)}))
	return b
}

func (b *IntBuilder) GreaterThanEqual(n int64) *IntBuilder {
	b.rules = append(b.rules, types.NewRule(types.KGreaterThanEqual, map[string]any{"n": float64(n)}))
	return b
}

func (b *IntBuilder) LessThan(n int64) *IntBuilder {
	b.rules = append(b.rules, types.NewRule(types.KLessThan, map[string]any{"n": float64(n)}))
	return b
}

func (b *IntBuilder) LessThanEqual(n int64) *IntBuilder {
	b.rules = append(b.rules, types.NewRule(types.KLessThanEqual, map[string]any{"n": float64(n)}))
	return b
}

func (b *IntBuilder) Between(min, max int64) *IntBuilder {
	b.rules = append(b.rules, types.NewRule(types.KBetween, map[string]any{"min": float64(min), "max": float64(max)}))
	return b
}

func (b *IntBuilder) Positive() *IntBuilder {
	b.rules = append(b.rules, types.NewRule(types.KPositive, nil))
	return b
}

func (b *IntBuilder) NonNegative() *IntBuilder {
	b.rules = append(b.rules, types.NewRule(types.KNonNegative, nil))
	return b
}

func (b *IntBuilder) Rule(kind types.Kind, args map[string]any) *IntBuilder {
	b.rules = append(b.rules, types.NewRule(kind, args))
	return b
}

func (b *IntBuilder) OmitEmpty() *IntBuilder {
	b.rules = append(b.rules, types.NewRule(types.KOmitempty, nil))
	return b
}

func (b *IntBuilder) Build() func(any) error {
	return b.engine.CompileRules(b.rules)
}

func (b *IntBuilder) BuildWithOpts(opts types.CompileOpts) func(any) error {
	return b.engine.CompileRulesWithOpts(b.rules, opts)
}

func (b *IntBuilder) BuildAll() func(any) error {
	return b.BuildWithOpts(types.CompileOpts{CollectAll: true})
}

func (b *IntBuilder) BuildContext() types.ContextValidatorFunc {
	return b.engine.CompileRulesContext(b.rules)
}

func (b *IntBuilder) BuildContextWithOpts(opts types.CompileOpts) types.ContextValidatorFunc {
	return b.engine.CompileRulesContextWithOpts(b.rules, opts)
}

// FloatBuilder accumulates floating-point validation rules.
type FloatBuilder struct {
	rules  []types.Rule
	engine *core.Engine
}

func NewFloatBuilder(engine *core.Engine) *FloatBuilder {
	return &FloatBuilder{
		rules:  []types.Rule{types.NewRule(types.KFloat, nil)},
		engine: engine,
	}
}

func (b *FloatBuilder) Required() *FloatBuilder {
	b.rules = append(b.rules, types.NewRule(types.KRequired, nil))
	return b
}

func (b *FloatBuilder) Min(n float64) *FloatBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMinNumber, map[string]any{"n": n}))
	return b
}

func (b *FloatBuilder) Max(n float64) *FloatBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMaxNumber, map[string]any{"n": n}))
	return b
}

func (b *FloatBuilder) GreaterThan(n float64) *FloatBuilder {
	b.rules = append(b.rules, types.NewRule(types.KGreaterThan, map[string]any{"n": n}))
	return b
}

func (b *FloatBuilder) GreaterThanEqual(n float64) *FloatBuilder {
	b.rules = append(b.rules, types.NewRule(types.KGreaterThanEqual, map[string]any{"n": n}))
	return b
}

func (b *FloatBuilder) LessThan(n float64) *FloatBuilder {
	b.rules = append(b.rules, types.NewRule(types.KLessThan, map[string]any{"n": n}))
	return b
}

func (b *FloatBuilder) LessThanEqual(n float64) *FloatBuilder {
	b.rules = append(b.rules, types.NewRule(types.KLessThanEqual, map[string]any{"n": n}))
	return b
}

func (b *FloatBuilder) Between(min, max float64) *FloatBuilder {
	b.rules = append(b.rules, types.NewRule(types.KBetween, map[string]any{"min": min, "max": max}))
	return b
}

func (b *FloatBuilder) Positive() *FloatBuilder {
	b.rules = append(b.rules, types.NewRule(types.KPositive, nil))
	return b
}

func (b *FloatBuilder) NonNegative() *FloatBuilder {
	b.rules = append(b.rules, types.NewRule(types.KNonNegative, nil))
	return b
}

func (b *FloatBuilder) Finite() *FloatBuilder {
	b.rules = append(b.rules, types.NewRule(types.KFinite, nil))
	return b
}

func (b *FloatBuilder) Rule(kind types.Kind, args map[string]any) *FloatBuilder {
	b.rules = append(b.rules, types.NewRule(kind, args))
	return b
}

func (b *FloatBuilder) OmitEmpty() *FloatBuilder {
	b.rules = append(b.rules, types.NewRule(types.KOmitempty, nil))
	return b
}

func (b *FloatBuilder) Build() func(any) error {
	return b.engine.CompileRules(b.rules)
}

func (b *FloatBuilder) BuildWithOpts(opts types.CompileOpts) func(any) error {
	return b.engine.CompileRulesWithOpts(b.rules, opts)
}

func (b *FloatBuilder) BuildAll() func(any) error {
	return b.BuildWithOpts(types.CompileOpts{CollectAll: true})
}

func (b *FloatBuilder) BuildContext() types.ContextValidatorFunc {
	return b.engine.CompileRulesContext(b.rules)
}

func (b *FloatBuilder) BuildContextWithOpts(opts types.CompileOpts) types.ContextValidatorFunc {
	return b.engine.CompileRulesContextWithOpts(b.rules, opts)
}

// BoolBuilder accumulates boolean validation rules.
type BoolBuilder struct {
	rules  []types.Rule
	engine *core.Engine
}

// NewBoolBuilder creates a new BoolBuilder with the base type rule.
func NewBoolBuilder(engine *core.Engine) *BoolBuilder {
	return &BoolBuilder{
		rules:  []types.Rule{types.NewRule(types.KBool, nil)},
		engine: engine,
	}
}

func (b *BoolBuilder) Build() func(any) error {
	return b.engine.CompileRules(b.rules)
}

func (b *BoolBuilder) BuildWithOpts(opts types.CompileOpts) func(any) error {
	return b.engine.CompileRulesWithOpts(b.rules, opts)
}

func (b *BoolBuilder) BuildAll() func(any) error {
	return b.BuildWithOpts(types.CompileOpts{CollectAll: true})
}

func (b *BoolBuilder) BuildContext() types.ContextValidatorFunc {
	return b.engine.CompileRulesContext(b.rules)
}

func (b *BoolBuilder) BuildContextWithOpts(opts types.CompileOpts) types.ContextValidatorFunc {
	return b.engine.CompileRulesContextWithOpts(b.rules, opts)
}

func (b *BoolBuilder) Required() *BoolBuilder {
	b.rules = append(b.rules, types.NewRule(types.KRequired, nil))
	return b
}

func (b *BoolBuilder) True() *BoolBuilder {
	b.rules = append(b.rules, types.NewRule(types.KBoolTrue, nil))
	return b
}

func (b *BoolBuilder) False() *BoolBuilder {
	b.rules = append(b.rules, types.NewRule(types.KBoolFalse, nil))
	return b
}

func (b *BoolBuilder) OmitEmpty() *BoolBuilder {
	b.rules = append(b.rules, types.NewRule(types.KOmitempty, nil))
	return b
}

func (b *BoolBuilder) Rule(kind types.Kind, args map[string]any) *BoolBuilder {
	b.rules = append(b.rules, types.NewRule(kind, args))
	return b
}

// SliceBuilder accumulates slice validation rules.
type SliceBuilder struct {
	engine *core.Engine
	rules  []types.Rule
}

func (b *SliceBuilder) Length(n int) *SliceBuilder {
	b.rules = append(b.rules, types.NewRule(types.KSliceLength, map[string]any{"n": int64(n)}))
	return b
}

func (b *SliceBuilder) Required() *SliceBuilder {
	b.rules = append(b.rules, types.NewRule(types.KRequired, nil))
	return b
}

func (b *SliceBuilder) MinLength(n int) *SliceBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMinSliceLength, map[string]any{"n": int64(n)}))
	return b
}

func (b *SliceBuilder) MaxLength(n int) *SliceBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMaxSliceLength, map[string]any{"n": int64(n)}))
	return b
}

func (b *SliceBuilder) ForEach(elemValidator func(any) error) *SliceBuilder {
	b.rules = append(b.rules, types.NewRule(types.KForEach, map[string]any{"validator": elemValidator}))
	return b
}

func (b *SliceBuilder) Unique() *SliceBuilder {
	b.rules = append(b.rules, types.NewRule(types.KSliceUnique, nil))
	return b
}

func (b *SliceBuilder) Contains(value any) *SliceBuilder {
	b.rules = append(b.rules, types.NewRule(types.KSliceContains, map[string]any{"value": value}))
	return b
}

// ForEachRules applies inner rules to each slice element.
// This form is cache-friendly (no function args).
func (b *SliceBuilder) ForEachRules(inner ...types.Rule) *SliceBuilder {
	if len(inner) == 0 {
		return b
	}
	// Convert to []types.Rule slice for the compiler
	innerRules := make([]types.Rule, len(inner))
	copy(innerRules, inner)
	r := types.NewRule(types.KForEach, map[string]any{"rules": innerRules})
	b.rules = append(b.rules, r)
	return b
}

// ForEachStringBuilder copies rules from a StringBuilder as element rules.
func (b *SliceBuilder) ForEachStringBuilder(sb *StringBuilder) *SliceBuilder {
	if sb == nil {
		return b
	}
	cp := append([]types.Rule(nil), sb.rules...)
	return b.ForEachRules(cp...)
}

func (b *SliceBuilder) Rule(kind types.Kind, args map[string]any) *SliceBuilder {
	b.rules = append(b.rules, types.NewRule(kind, args))
	return b
}

func (b *SliceBuilder) OmitEmpty() *SliceBuilder {
	b.rules = append(b.rules, types.NewRule(types.KOmitempty, nil))
	return b
}

func (b *SliceBuilder) Build() func(any) error {
	return b.engine.CompileRules(b.rules)
}

func (b *SliceBuilder) BuildWithOpts(opts types.CompileOpts) func(any) error {
	return b.engine.CompileRulesWithOpts(b.rules, opts)
}

func (b *SliceBuilder) BuildAll() func(any) error {
	return b.BuildWithOpts(types.CompileOpts{CollectAll: true})
}

func (b *SliceBuilder) BuildContext() types.ContextValidatorFunc {
	return b.engine.CompileRulesContext(b.rules)
}

func (b *SliceBuilder) BuildContextWithOpts(opts types.CompileOpts) types.ContextValidatorFunc {
	return b.engine.CompileRulesContextWithOpts(b.rules, opts)
}

// ArrayBuilder accumulates array validation rules.
type ArrayBuilder struct {
	engine *core.Engine
	rules  []types.Rule
}

func NewArrayBuilder(engine *core.Engine) *ArrayBuilder {
	return &ArrayBuilder{
		engine: engine,
		rules:  []types.Rule{types.NewRule(types.KArray, nil)},
	}
}

func (b *ArrayBuilder) Length(n int) *ArrayBuilder {
	b.rules = append(b.rules, types.NewRule(types.KArrayLength, map[string]any{"n": int64(n)}))
	return b
}

func (b *ArrayBuilder) Required() *ArrayBuilder {
	b.rules = append(b.rules, types.NewRule(types.KRequired, nil))
	return b
}

func (b *ArrayBuilder) MinLength(n int) *ArrayBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMinArrayLength, map[string]any{"n": int64(n)}))
	return b
}

func (b *ArrayBuilder) MaxLength(n int) *ArrayBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMaxArrayLength, map[string]any{"n": int64(n)}))
	return b
}

func (b *ArrayBuilder) ForEach(elemValidator func(any) error) *ArrayBuilder {
	b.rules = append(b.rules, types.NewRule(types.KArrayForEach, map[string]any{"validator": elemValidator}))
	return b
}

func (b *ArrayBuilder) Unique() *ArrayBuilder {
	b.rules = append(b.rules, types.NewRule(types.KArrayUnique, nil))
	return b
}

func (b *ArrayBuilder) Contains(value any) *ArrayBuilder {
	b.rules = append(b.rules, types.NewRule(types.KArrayContains, map[string]any{"value": value}))
	return b
}

// ForEachRules applies inner rules to each array element.
// This form is cache-friendly (no function args).
func (b *ArrayBuilder) ForEachRules(inner ...types.Rule) *ArrayBuilder {
	if len(inner) == 0 {
		return b
	}
	innerRules := make([]types.Rule, len(inner))
	copy(innerRules, inner)
	b.rules = append(b.rules, types.NewRule(types.KArrayForEach, map[string]any{"rules": innerRules}))
	return b
}

// ForEachStringBuilder copies rules from a StringBuilder as element rules.
func (b *ArrayBuilder) ForEachStringBuilder(sb *StringBuilder) *ArrayBuilder {
	if sb == nil {
		return b
	}
	cp := append([]types.Rule(nil), sb.rules...)
	return b.ForEachRules(cp...)
}

func (b *ArrayBuilder) Rule(kind types.Kind, args map[string]any) *ArrayBuilder {
	b.rules = append(b.rules, types.NewRule(kind, args))
	return b
}

func (b *ArrayBuilder) OmitEmpty() *ArrayBuilder {
	b.rules = append(b.rules, types.NewRule(types.KOmitempty, nil))
	return b
}

func (b *ArrayBuilder) Build() func(any) error {
	return b.engine.CompileRules(b.rules)
}

func (b *ArrayBuilder) BuildWithOpts(opts types.CompileOpts) func(any) error {
	return b.engine.CompileRulesWithOpts(b.rules, opts)
}

func (b *ArrayBuilder) BuildAll() func(any) error {
	return b.BuildWithOpts(types.CompileOpts{CollectAll: true})
}

func (b *ArrayBuilder) BuildContext() types.ContextValidatorFunc {
	return b.engine.CompileRulesContext(b.rules)
}

func (b *ArrayBuilder) BuildContextWithOpts(opts types.CompileOpts) types.ContextValidatorFunc {
	return b.engine.CompileRulesContextWithOpts(b.rules, opts)
}

// MapBuilder accumulates map validation rules.
type MapBuilder struct {
	engine *core.Engine
	rules  []types.Rule
}

func NewMapBuilder(engine *core.Engine) *MapBuilder {
	return &MapBuilder{
		engine: engine,
		rules:  []types.Rule{types.NewRule(types.KMap, nil)},
	}
}

func (b *MapBuilder) Required() *MapBuilder {
	b.rules = append(b.rules, types.NewRule(types.KRequired, nil))
	return b
}

func (b *MapBuilder) Length(n int) *MapBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMapLength, map[string]any{"n": int64(n)}))
	return b
}

func (b *MapBuilder) MinKeys(n int) *MapBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMinMapKeys, map[string]any{"n": int64(n)}))
	return b
}

func (b *MapBuilder) MaxKeys(n int) *MapBuilder {
	b.rules = append(b.rules, types.NewRule(types.KMaxMapKeys, map[string]any{"n": int64(n)}))
	return b
}

func (b *MapBuilder) KeysRules(inner ...types.Rule) *MapBuilder {
	if len(inner) == 0 {
		return b
	}
	cp := append([]types.Rule(nil), inner...)
	b.rules = append(b.rules, types.NewRule(types.KMapKeys, map[string]any{"rules": cp}))
	return b
}

func (b *MapBuilder) ValuesRules(inner ...types.Rule) *MapBuilder {
	if len(inner) == 0 {
		return b
	}
	cp := append([]types.Rule(nil), inner...)
	b.rules = append(b.rules, types.NewRule(types.KMapValues, map[string]any{"rules": cp}))
	return b
}

func (b *MapBuilder) Rule(kind types.Kind, args map[string]any) *MapBuilder {
	b.rules = append(b.rules, types.NewRule(kind, args))
	return b
}

func (b *MapBuilder) OmitEmpty() *MapBuilder {
	b.rules = append(b.rules, types.NewRule(types.KOmitempty, nil))
	return b
}

func (b *MapBuilder) Build() func(any) error {
	return b.engine.CompileRules(b.rules)
}

func (b *MapBuilder) BuildWithOpts(opts types.CompileOpts) func(any) error {
	return b.engine.CompileRulesWithOpts(b.rules, opts)
}

func (b *MapBuilder) BuildAll() func(any) error {
	return b.BuildWithOpts(types.CompileOpts{CollectAll: true})
}

func (b *MapBuilder) BuildContext() types.ContextValidatorFunc {
	return b.engine.CompileRulesContext(b.rules)
}

func (b *MapBuilder) BuildContextWithOpts(opts types.CompileOpts) types.ContextValidatorFunc {
	return b.engine.CompileRulesContextWithOpts(b.rules, opts)
}

// TimeBuilder accumulates time.Time validation rules.
type TimeBuilder struct {
	engine *core.Engine
	rules  []types.Rule
}

func NewTimeBuilder(engine *core.Engine) *TimeBuilder {
	return &TimeBuilder{
		engine: engine,
		rules:  []types.Rule{types.NewRule(types.KTime, nil)},
	}
}

func (b *TimeBuilder) Required() *TimeBuilder {
	b.rules = append(b.rules, types.NewRule(types.KRequired, nil))
	return b
}

func (b *TimeBuilder) NotZero() *TimeBuilder {
	b.rules = append(b.rules, types.NewRule(types.KTimeNotZero, nil))
	return b
}

func (b *TimeBuilder) Before(t time.Time) *TimeBuilder {
	b.rules = append(b.rules, types.NewRule(types.KTimeBefore, map[string]any{"time": t}))
	return b
}

func (b *TimeBuilder) After(t time.Time) *TimeBuilder {
	b.rules = append(b.rules, types.NewRule(types.KTimeAfter, map[string]any{"time": t}))
	return b
}

func (b *TimeBuilder) Between(start, end time.Time) *TimeBuilder {
	b.rules = append(b.rules, types.NewRule(types.KTimeBetween, map[string]any{"start": start, "end": end}))
	return b
}

func (b *TimeBuilder) Rule(kind types.Kind, args map[string]any) *TimeBuilder {
	b.rules = append(b.rules, types.NewRule(kind, args))
	return b
}

func (b *TimeBuilder) OmitEmpty() *TimeBuilder {
	b.rules = append(b.rules, types.NewRule(types.KOmitempty, nil))
	return b
}

func (b *TimeBuilder) Build() func(any) error {
	return b.engine.CompileRules(b.rules)
}

func (b *TimeBuilder) BuildWithOpts(opts types.CompileOpts) func(any) error {
	return b.engine.CompileRulesWithOpts(b.rules, opts)
}

func (b *TimeBuilder) BuildAll() func(any) error {
	return b.BuildWithOpts(types.CompileOpts{CollectAll: true})
}

func (b *TimeBuilder) BuildContext() types.ContextValidatorFunc {
	return b.engine.CompileRulesContext(b.rules)
}

func (b *TimeBuilder) BuildContextWithOpts(opts types.CompileOpts) types.ContextValidatorFunc {
	return b.engine.CompileRulesContextWithOpts(b.rules, opts)
}

// CustomTypeBuilder accumulates custom type validation rules.
type CustomTypeBuilder struct {
	engine   *core.Engine
	typeName string
	rules    []types.Rule
}

func (b *CustomTypeBuilder) Build() func(any) error {
	return b.engine.CompileRules(b.rules)
}

func (b *CustomTypeBuilder) BuildWithOpts(opts types.CompileOpts) func(any) error {
	return b.engine.CompileRulesWithOpts(b.rules, opts)
}

func (b *CustomTypeBuilder) BuildAll() func(any) error {
	return b.BuildWithOpts(types.CompileOpts{CollectAll: true})
}

func (b *CustomTypeBuilder) BuildContext() types.ContextValidatorFunc {
	return b.engine.CompileRulesContext(b.rules)
}

func (b *CustomTypeBuilder) BuildContextWithOpts(opts types.CompileOpts) types.ContextValidatorFunc {
	return b.engine.CompileRulesContextWithOpts(b.rules, opts)
}

func (b *CustomTypeBuilder) Required() *CustomTypeBuilder {
	b.rules = append(b.rules, types.NewRule(types.KRequired, nil))
	return b
}

func (b *CustomTypeBuilder) OmitEmpty() *CustomTypeBuilder {
	b.rules = append(b.rules, types.NewRule(types.KOmitempty, nil))
	return b
}

func (b *CustomTypeBuilder) Rule(kind types.Kind, args map[string]any) *CustomTypeBuilder {
	b.rules = append(b.rules, types.NewRule(kind, args))
	return b
}
