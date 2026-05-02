package core

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/aatuh/validate/v3/translator"
	"github.com/aatuh/validate/v3/types"
)

// compiledKey is a typed key so we never mix raw strings accidentally.
type compiledKey string

const (
	ckTag = "tag:" // cache-key prefix for tag token compiles
	ckAST = "ast:" // cache-key prefix for AST rule compiles
)

// Engine is the generic validation engine. It compiles tag tokens or AST
// rules into reusable validator functions and caches the results.
type Engine struct {
	customRules          map[string]func(any) error
	ruleCompilers        map[types.Kind]types.RuleCompiler
	contextRuleCompilers map[types.Kind]types.ContextRuleCompiler
	structRuleCompilers  map[types.Kind]StructRuleCompiler
	typeRegistry         *types.TypeRegistry
	translator           translator.Translator
	pathSep              string

	// compiled caches compiled validators.
	// Keys are compiledKey values with ckTag or ckAST prefixes.
	compiled        sync.Map // map[compiledKey]types.ValidatorFunc
	compiledContext sync.Map // map[compiledKey]types.ContextValidatorFunc
}

// NewEngine creates a new Engine with sane defaults.
func NewEngine() *Engine {
	return &Engine{
		customRules:          make(map[string]func(any) error),
		ruleCompilers:        make(map[types.Kind]types.RuleCompiler),
		contextRuleCompilers: make(map[types.Kind]types.ContextRuleCompiler),
		structRuleCompilers:  make(map[types.Kind]StructRuleCompiler),
		pathSep:              ".",
	}
}

// NewEngineWithCustomRules seeds the engine with custom rules.
func NewEngineWithCustomRules(custom map[string]func(any) error) *Engine {
	e := NewEngine()
	for k, fn := range custom {
		e.customRules[k] = fn
	}
	return e
}

// Copy returns a new Engine with the same configuration but separate cache.
// This mirrors prior behavior used in tests.
func (e *Engine) Copy() *Engine {
	if e == nil {
		return nil
	}
	// Create new Engine with same config but new cache
	newEngine := &Engine{
		customRules:          copyCustomRules(e.customRules),
		ruleCompilers:        copyRuleCompilers(e.ruleCompilers),
		contextRuleCompilers: copyContextRuleCompilers(e.contextRuleCompilers),
		structRuleCompilers:  copyStructRuleCompilers(e.structRuleCompilers),
		typeRegistry:         copyTypeRegistry(e.typeRegistry),
		translator:           e.translator,
		pathSep:              e.pathSep,
		// Note: compiled cache is intentionally not copied (new empty cache)
	}

	return newEngine
}

// WithCustomRule returns a new Engine with the rule registered.
func (e *Engine) WithCustomRule(name string, rule func(any) error) *Engine {
	newCustom := make(map[string]func(any) error, len(e.customRules)+1)
	for k, v := range e.customRules {
		newCustom[k] = v
	}
	newCustom[name] = rule

	return &Engine{
		customRules:          newCustom,
		ruleCompilers:        copyRuleCompilers(e.ruleCompilers),
		contextRuleCompilers: copyContextRuleCompilers(e.contextRuleCompilers),
		structRuleCompilers:  copyStructRuleCompilers(e.structRuleCompilers),
		typeRegistry:         copyTypeRegistry(e.typeRegistry),
		translator:           e.translator,
		pathSep:              e.pathSep,
		// Note: compiled cache is intentionally not copied (new empty cache)
	}
}

// WithRuleCompiler returns a new Engine with a per-instance rule compiler.
func (e *Engine) WithRuleCompiler(kind types.Kind, rc types.RuleCompiler) *Engine {
	newCompilers := copyRuleCompilers(e.ruleCompilers)
	newCompilers[kind] = rc
	return &Engine{
		customRules:          copyCustomRules(e.customRules),
		ruleCompilers:        newCompilers,
		contextRuleCompilers: copyContextRuleCompilers(e.contextRuleCompilers),
		structRuleCompilers:  copyStructRuleCompilers(e.structRuleCompilers),
		typeRegistry:         copyTypeRegistry(e.typeRegistry),
		translator:           e.translator,
		pathSep:              e.pathSep,
	}
}

// WithContextRuleCompiler returns a new Engine with a per-instance
// context-aware rule compiler.
func (e *Engine) WithContextRuleCompiler(kind types.Kind, rc types.ContextRuleCompiler) *Engine {
	newCompilers := copyContextRuleCompilers(e.contextRuleCompilers)
	newCompilers[kind] = rc
	return &Engine{
		customRules:          copyCustomRules(e.customRules),
		ruleCompilers:        copyRuleCompilers(e.ruleCompilers),
		contextRuleCompilers: newCompilers,
		structRuleCompilers:  copyStructRuleCompilers(e.structRuleCompilers),
		typeRegistry:         copyTypeRegistry(e.typeRegistry),
		translator:           e.translator,
		pathSep:              e.pathSep,
	}
}

// WithStructRuleCompiler returns a new Engine with a per-instance struct rule compiler.
func (e *Engine) WithStructRuleCompiler(kind types.Kind, compiler StructRuleCompiler) *Engine {
	newCompilers := copyStructRuleCompilers(e.structRuleCompilers)
	newCompilers[kind] = compiler
	return &Engine{
		customRules:          copyCustomRules(e.customRules),
		ruleCompilers:        copyRuleCompilers(e.ruleCompilers),
		contextRuleCompilers: copyContextRuleCompilers(e.contextRuleCompilers),
		structRuleCompilers:  newCompilers,
		typeRegistry:         copyTypeRegistry(e.typeRegistry),
		translator:           e.translator,
		pathSep:              e.pathSep,
	}
}

// WithTypeValidator returns a new Engine with a per-instance custom type validator.
func (e *Engine) WithTypeValidator(name string, factory types.TypeValidatorFactory) *Engine {
	newRegistry := copyTypeRegistry(e.typeRegistry)
	if newRegistry == nil {
		newRegistry = types.NewTypeRegistry()
	}
	newRegistry.RegisterType(name, factory)
	return &Engine{
		customRules:          copyCustomRules(e.customRules),
		ruleCompilers:        copyRuleCompilers(e.ruleCompilers),
		contextRuleCompilers: copyContextRuleCompilers(e.contextRuleCompilers),
		structRuleCompilers:  copyStructRuleCompilers(e.structRuleCompilers),
		typeRegistry:         newRegistry,
		translator:           e.translator,
		pathSep:              e.pathSep,
	}
}

// WithTranslator returns a new Engine with a translator.
func (e *Engine) WithTranslator(t translator.Translator) *Engine {
	return &Engine{
		customRules:          copyCustomRules(e.customRules),
		ruleCompilers:        copyRuleCompilers(e.ruleCompilers),
		contextRuleCompilers: copyContextRuleCompilers(e.contextRuleCompilers),
		structRuleCompilers:  copyStructRuleCompilers(e.structRuleCompilers),
		typeRegistry:         copyTypeRegistry(e.typeRegistry),
		translator:           t,
		pathSep:              e.pathSep,
		// Note: compiled cache is intentionally not copied (new empty cache)
	}
}

// PathSeparator returns a new Engine with a different path separator.
func (e *Engine) PathSeparator(sep string) *Engine {
	newPathSep := e.pathSep
	if sep != "" {
		newPathSep = sep
	}
	return &Engine{
		customRules:          copyCustomRules(e.customRules),
		ruleCompilers:        copyRuleCompilers(e.ruleCompilers),
		contextRuleCompilers: copyContextRuleCompilers(e.contextRuleCompilers),
		structRuleCompilers:  copyStructRuleCompilers(e.structRuleCompilers),
		typeRegistry:         copyTypeRegistry(e.typeRegistry),
		translator:           e.translator,
		pathSep:              newPathSep,
		// Note: compiled cache is intentionally not copied (new empty cache)
	}
}

// Translator exposes the configured translator.
func (e *Engine) Translator() translator.Translator { return e.translator }

// GetPathSeparator exposes the configured path separator.
func (e *Engine) GetPathSeparator() string { return e.pathSep }

// StructRuleCompiler returns a registered per-instance struct rule compiler.
func (e *Engine) StructRuleCompiler(kind types.Kind) (StructRuleCompiler, bool) {
	compiler, ok := e.structRuleCompilers[kind]
	return compiler, ok
}

// FromRules compiles validators from rule tokens (e.g. "string","min=2").
func (e *Engine) FromRules(tokens []string) (func(any) error, error) {
	return e.FromRulesWithOpts(tokens, types.CompileOpts{})
}

// FromRulesWithOpts compiles validators from rule tokens with compile options.
func (e *Engine) FromRulesWithOpts(tokens []string, opts types.CompileOpts) (func(any) error, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty rules")
	}

	// Custom single-token rule?
	if rule, ok := e.customRules[tokens[0]]; ok && len(tokens) == 1 {
		return rule, nil
	}

	// Normalize tokens to a tag string and cache by it.
	tag := strings.Join(tokens, ";")
	key := compiledKey(ckTag + compileOptsKeyPart(opts) + tag)

	if v, ok := e.compiled.Load(key); ok {
		return v.(types.ValidatorFunc), nil
	}

	ast, err := types.ParseTagWithRegistry(tag, e.typeRegistry)
	if err != nil {
		return nil, fmt.Errorf("parse rules: %w", err)
	}
	fn, err := e.newCompiler().CompileWithOptsE(ast, opts)
	if err != nil {
		return nil, err
	}

	if existing, loaded := e.compiled.LoadOrStore(key, fn); loaded {
		return existing.(types.ValidatorFunc), nil
	}
	return fn, nil
}

// FromRulesContext compiles a context-aware validator from rule tokens.
func (e *Engine) FromRulesContext(tokens []string) (types.ContextValidatorFunc, error) {
	return e.FromRulesContextWithOpts(tokens, types.CompileOpts{})
}

// FromRulesContextWithOpts compiles a context-aware validator from rule tokens
// with compile options.
func (e *Engine) FromRulesContextWithOpts(tokens []string, opts types.CompileOpts) (types.ContextValidatorFunc, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty rules")
	}
	if rule, ok := e.customRules[tokens[0]]; ok && len(tokens) == 1 {
		return func(ctx context.Context, v any) error {
			if ctx == nil {
				ctx = context.Background()
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			return rule(v)
		}, nil
	}

	tag := strings.Join(tokens, ";")
	key := compiledKey(ckTag + "ctx:" + compileOptsKeyPart(opts) + tag)

	if v, ok := e.compiledContext.Load(key); ok {
		return v.(types.ContextValidatorFunc), nil
	}

	ast, err := types.ParseTagWithRegistry(tag, e.typeRegistry)
	if err != nil {
		return nil, fmt.Errorf("parse rules: %w", err)
	}
	fn, err := e.newCompiler().CompileContextWithOptsE(ast, opts)
	if err != nil {
		return nil, err
	}
	if existing, loaded := e.compiledContext.LoadOrStore(key, fn); loaded {
		return existing.(types.ContextValidatorFunc), nil
	}
	return fn, nil
}

// CompileRules compiles AST rules. We cache deterministically unless any
// rule carries a function argument (non-deterministic).
func (e *Engine) CompileRules(rules []types.Rule) func(any) error {
	fn, err := e.CompileRulesE(rules)
	if err != nil {
		return func(any) error { return err }
	}
	return fn
}

// CompileRulesE compiles AST rules and returns compile-time custom-rule errors.
func (e *Engine) CompileRulesE(rules []types.Rule) (func(any) error, error) {
	return e.CompileRulesWithOptsE(rules, types.CompileOpts{})
}

// CompileRulesWithOpts compiles AST rules with options.
func (e *Engine) CompileRulesWithOpts(rules []types.Rule, opts types.CompileOpts) func(any) error {
	fn, err := e.CompileRulesWithOptsE(rules, opts)
	if err != nil {
		return func(any) error { return err }
	}
	return fn
}

// CompileRulesWithOptsE compiles AST rules with options and returns compile
// errors.
func (e *Engine) CompileRulesWithOptsE(rules []types.Rule, opts types.CompileOpts) (func(any) error, error) {
	// If any arg is a func (directly or nested), skip cache by design.
	if HasFuncArgs(rules) {
		return e.newCompiler().CompileWithOptsE(rules, opts)
	}

	serialized := SerializeRules(rules) // canonical, deterministic
	key := compiledKey(ckAST + compileOptsKeyPart(opts) + serialized)

	if v, ok := e.compiled.Load(key); ok {
		return v.(types.ValidatorFunc), nil
	}

	fn, err := e.newCompiler().CompileWithOptsE(rules, opts)
	if err != nil {
		return nil, err
	}
	if existing, loaded := e.compiled.LoadOrStore(key, fn); loaded {
		return existing.(types.ValidatorFunc), nil
	}
	return fn, nil
}

// CompileRulesContext compiles AST rules into a context-aware validator.
func (e *Engine) CompileRulesContext(rules []types.Rule) types.ContextValidatorFunc {
	fn, err := e.CompileRulesContextE(rules)
	if err != nil {
		return func(context.Context, any) error { return err }
	}
	return fn
}

// CompileRulesContextE compiles AST rules into a context-aware validator and
// returns compile errors.
func (e *Engine) CompileRulesContextE(rules []types.Rule) (types.ContextValidatorFunc, error) {
	return e.CompileRulesContextWithOptsE(rules, types.CompileOpts{})
}

// CompileRulesContextWithOpts compiles AST rules into a context-aware validator
// with options.
func (e *Engine) CompileRulesContextWithOpts(rules []types.Rule, opts types.CompileOpts) types.ContextValidatorFunc {
	fn, err := e.CompileRulesContextWithOptsE(rules, opts)
	if err != nil {
		return func(context.Context, any) error { return err }
	}
	return fn
}

// CompileRulesContextWithOptsE compiles AST rules into a context-aware
// validator with options and returns compile errors.
func (e *Engine) CompileRulesContextWithOptsE(rules []types.Rule, opts types.CompileOpts) (types.ContextValidatorFunc, error) {
	if HasFuncArgs(rules) {
		return e.newCompiler().CompileContextWithOptsE(rules, opts)
	}

	serialized := SerializeRules(rules)
	key := compiledKey(ckAST + "ctx:" + compileOptsKeyPart(opts) + serialized)

	if v, ok := e.compiledContext.Load(key); ok {
		return v.(types.ContextValidatorFunc), nil
	}

	fn, err := e.newCompiler().CompileContextWithOptsE(rules, opts)
	if err != nil {
		return nil, err
	}
	if existing, loaded := e.compiledContext.LoadOrStore(key, fn); loaded {
		return existing.(types.ContextValidatorFunc), nil
	}
	return fn, nil
}

func (e *Engine) newCompiler() *types.Compiler {
	c := types.NewCompiler(e.translator)
	c.SetTypeRegistry(e.typeRegistry)
	for kind, rc := range e.ruleCompilers {
		c.RegisterRule(kind, rc)
	}
	for kind, rc := range e.contextRuleCompilers {
		c.RegisterContextRule(kind, rc)
	}
	return c
}

func compileOptsKeyPart(opts types.CompileOpts) string {
	if opts.CollectAll {
		return "all:"
	}
	return ""
}

func copyCustomRules(in map[string]func(any) error) map[string]func(any) error {
	out := make(map[string]func(any) error, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyRuleCompilers(in map[types.Kind]types.RuleCompiler) map[types.Kind]types.RuleCompiler {
	out := make(map[types.Kind]types.RuleCompiler, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyContextRuleCompilers(in map[types.Kind]types.ContextRuleCompiler) map[types.Kind]types.ContextRuleCompiler {
	out := make(map[types.Kind]types.ContextRuleCompiler, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyStructRuleCompilers(in map[types.Kind]StructRuleCompiler) map[types.Kind]StructRuleCompiler {
	out := make(map[types.Kind]StructRuleCompiler, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyTypeRegistry(in *types.TypeRegistry) *types.TypeRegistry {
	return in.Clone()
}
