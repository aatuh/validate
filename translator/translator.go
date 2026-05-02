package translator

import (
	"fmt"
	"sync"
)

// Translator is an interface for obtaining localized messages.
//
// This interface provides a way to get localized error messages for
// validation failures.
type Translator interface {
	// T returns a localized message for the given key and parameters.
	T(key string, params ...any) string
}

// SimpleTranslator is a basic implementation of Translator using a map.
//
// Fields:
//   - messages: Map of message keys to localized strings.
type SimpleTranslator struct {
	messages map[string]string
}

var (
	defaultMu           sync.RWMutex
	defaultTranslations = map[string]string{}
)

// NewSimpleTranslator creates a new SimpleTranslator.
//
// Parameters:
//   - messages: Map of message keys to localized strings.
//
// Returns:
//   - *SimpleTranslator: A new SimpleTranslator instance.
func NewSimpleTranslator(messages map[string]string) *SimpleTranslator {
	cp := make(map[string]string, len(messages))
	for k, v := range messages {
		cp[k] = v
	}
	return &SimpleTranslator{
		messages: cp,
	}
}

// T returns the translated message or the default if not found.
//
// Parameters:
//   - key: The message key to look up.
//   - params: Variable number of parameters for message formatting.
//
// Returns:
//   - string: The localized message or the key as fallback.
func (st *SimpleTranslator) T(key string, params ...any) string {
	if st == nil {
		return ""
	}
	if msg, ok := st.messages[key]; ok {
		return fmt.Sprintf(msg, params...)
	}
	// Fallback: use key as the format string.
	return fmt.Sprintf(key, params...)
}

// RegisterDefaultEnglishTranslations adds process-wide default English
// translations. Plugin packages call this from init.
func RegisterDefaultEnglishTranslations(messages map[string]string) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	for k, v := range messages {
		defaultTranslations[k] = v
	}
}

// MergeTranslations returns a new map with later maps overriding earlier maps.
func MergeTranslations(base map[string]string, overlays ...map[string]string) map[string]string {
	size := len(base)
	for _, overlay := range overlays {
		size += len(overlay)
	}
	out := make(map[string]string, size)
	for k, v := range base {
		out[k] = v
	}
	for _, overlay := range overlays {
		for k, v := range overlay {
			out[k] = v
		}
	}
	return out
}

// DefaultEnglishTranslations returns a map of default English messages.
//
// Returns:
//   - map[string]string: A map containing default English error messages
//     for validation failures.
func DefaultEnglishTranslations() map[string]string {
	base := map[string]string{
		// Type errors
		"bool.type":   "expected boolean",
		"int.type":    "expected integer",
		"int64.type":  "expected int64",
		"float.type":  "expected finite floating-point number",
		"number.type": "expected number",
		"string.type": "expected string",
		"slice.type":  "expected slice",
		"map.type":    "expected map",
		"time.type":   "expected time.Time",

		// Generic validation
		"required":        "value is required",
		"required.with":   "value is required",
		"required.if":     "value is required",
		"required.unless": "value is required",
		"field.eq":        "must match the referenced field",
		"field.ne":        "must differ from the referenced field",
		"field.reference": "invalid referenced field",

		// String validation
		"string.length":               "must be exactly %d characters long",
		"string.min":                  "minimum length is %d",
		"string.max":                  "maximum length is %d",
		"string.nonempty":             "must not be empty",
		"string.contains":             "must contain required text",
		"string.notContains":          "must not contain prohibited text",
		"string.prefix":               "must have required prefix",
		"string.suffix":               "must have required suffix",
		"string.url":                  "must be a valid absolute URL",
		"string.hostname":             "must be a valid hostname",
		"string.ip":                   "must be a valid IP address",
		"string.cidr":                 "must be a valid CIDR prefix",
		"string.ascii":                "must contain only ASCII characters",
		"string.alpha":                "must contain only letters",
		"string.alnum":                "must contain only letters and digits",
		"string.minLength":            "must be at least %d characters long",
		"string.maxLength":            "must be at most %d characters long",
		"string.minRunes":             "minimum rune count is %d",
		"string.maxRunes":             "maximum rune count is %d",
		"string.oneof":                "must be one of: %s",
		"string.regex.invalidPattern": "invalid regex pattern: %s",
		"string.regex.inputTooLong":   "input too long for regex validation",
		"string.regex.noMatch":        "does not match required pattern",

		// Integer validation
		"int.min":                   "minimum value is %d",
		"int.max":                   "maximum value is %d",
		"number.min":                "minimum value is %g",
		"number.max":                "maximum value is %g",
		"number.gt":                 "must be greater than %g",
		"number.gte":                "must be greater than or equal to %g",
		"number.lt":                 "must be less than %g",
		"number.lte":                "must be less than or equal to %g",
		"number.between":            "must be between %g and %g",
		"number.positive":           "must be positive",
		"number.nonnegative":        "must be nonnegative",
		"number.finite":             "must be finite",
		"int.invalidMinParameter":   "invalid parameter for min",
		"int.invalidMaxParameter":   "invalid parameter for max",
		"int.unknownIntValidator":   "unknown int validator: %s",
		"int.unknownInt64Validator": "unknown int64 validator: %s",
		"int.notInteger":            "value is not an integer",
		"int.notInt64":              "value is not an int64",

		// Slice validation
		"slice.length":              "must have exactly %d elements",
		"slice.min":                 "minimum length is %d",
		"slice.max":                 "maximum length is %d",
		"slice.unique":              "must contain unique elements",
		"slice.contains":            "must contain required element",
		"slice.forEach":             "element validation failed",
		"slice.element":             "element %d: %s",
		"slice.invalidLenParameter": "invalid parameter for len",
		"slice.invalidMinParameter": "invalid parameter for min",
		"slice.invalidMaxParameter": "invalid parameter for max",
		"slice.unknownValidator":    "unknown slice validator: %s",
		"slice.notSlice":            "value is not a slice",

		// Array validation
		"array.type":     "expected array",
		"array.length":   "must have exactly %d elements",
		"array.min":      "minimum length is %d",
		"array.max":      "maximum length is %d",
		"array.unique":   "must contain unique elements",
		"array.contains": "must contain required element",
		"array.forEach":  "element validation failed",

		// Map validation
		"map.length":  "must have exactly %d keys",
		"map.minkeys": "minimum key count is %d",
		"map.maxkeys": "maximum key count is %d",
		"map.keys":    "map key validation failed",
		"map.values":  "map value validation failed",

		// Bool validation
		"bool.true":  "must be true",
		"bool.false": "must be false",

		// Time validation
		"time.notzero": "must not be zero",
		"time.before":  "must be before %s",
		"time.after":   "must be after %s",
		"time.between": "must be between %s and %s",

		// Legacy compatibility
		"bool.notBool": "value is not a boolean",
	}
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return MergeTranslations(base, defaultTranslations)
}
