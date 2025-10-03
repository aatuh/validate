package core

import (
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
	customRules map[string]func(any) error
	translator  translator.Translator
	pathSep     string

	// compiled caches compiled validators.
	// Keys are compiledKey values with ckTag or ckAST prefixes.
	compiled sync.Map // map[compiledKey]types.ValidatorFunc
}

// NewEngine creates a new Engine with sane defaults.
func NewEngine() *Engine {
	return &Engine{
		customRules: make(map[string]func(any) error),
		pathSep:     ".",
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
		customRules: make(map[string]func(any) error),
		translator:  e.translator,
		pathSep:     e.pathSep,
		// Note: compiled cache is intentionally not copied (new empty cache)
	}

	// Copy custom rules
	for k, v := range e.customRules {
		newEngine.customRules[k] = v
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
		customRules: newCustom,
		translator:  e.translator,
		pathSep:     e.pathSep,
		// Note: compiled cache is intentionally not copied (new empty cache)
	}
}

// WithTranslator returns a new Engine with a translator.
func (e *Engine) WithTranslator(t translator.Translator) *Engine {
	return &Engine{
		customRules: e.customRules,
		translator:  t,
		pathSep:     e.pathSep,
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
		customRules: e.customRules,
		translator:  e.translator,
		pathSep:     newPathSep,
		// Note: compiled cache is intentionally not copied (new empty cache)
	}
}

// Translator exposes the configured translator.
func (e *Engine) Translator() translator.Translator { return e.translator }

// GetPathSeparator exposes the configured path separator.
func (e *Engine) GetPathSeparator() string { return e.pathSep }

// FromRules compiles validators from rule tokens (e.g. "string","min=2").
func (e *Engine) FromRules(tokens []string) (func(any) error, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty rules")
	}

	// Custom single-token rule?
	if rule, ok := e.customRules[tokens[0]]; ok && len(tokens) == 1 {
		return rule, nil
	}

	// Normalize tokens to a tag string and cache by it.
	tag := strings.Join(tokens, ";")
	key := compiledKey(ckTag + tag)

	if v, ok := e.compiled.Load(key); ok {
		return v.(types.ValidatorFunc), nil
	}

	ast, err := types.ParseTag(tag)
	if err != nil {
		return nil, fmt.Errorf("parse rules: %w", err)
	}
	fn := types.NewCompiler(e.translator).Compile(ast)

	if existing, loaded := e.compiled.LoadOrStore(key, fn); loaded {
		return existing.(types.ValidatorFunc), nil
	}
	return fn, nil
}

// CompileRules compiles AST rules. We cache deterministically unless any
// rule carries a function argument (non-deterministic).
func (e *Engine) CompileRules(rules []types.Rule) func(any) error {
	// If any arg is a func (directly or nested), skip cache by design.
	if HasFuncArgs(rules) {
		return types.NewCompiler(e.translator).Compile(rules)
	}

	serialized := SerializeRules(rules) // canonical, deterministic
	key := compiledKey(ckAST + serialized)

	if v, ok := e.compiled.Load(key); ok {
		return v.(types.ValidatorFunc)
	}

	fn := types.NewCompiler(e.translator).Compile(rules)
	if existing, loaded := e.compiled.LoadOrStore(key, fn); loaded {
		return existing.(types.ValidatorFunc)
	}
	return fn
}
