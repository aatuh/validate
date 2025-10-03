package core

// Validate is now an alias of Engine to avoid duplication while keeping
// the public API stable for existing imports and tests.
type Validate = Engine

// New returns a new Validate (Engine) with sane defaults.
func New() *Validate { return NewEngine() }

// NewWithCustomRules returns a new Validate (Engine) with custom rules.
func NewWithCustomRules(custom map[string]func(any) error) *Validate {
	return NewEngineWithCustomRules(custom)
}
