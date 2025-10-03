package translator

import "fmt"

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

// NewSimpleTranslator creates a new SimpleTranslator.
//
// Parameters:
//   - messages: Map of message keys to localized strings.
//
// Returns:
//   - *SimpleTranslator: A new SimpleTranslator instance.
func NewSimpleTranslator(messages map[string]string) *SimpleTranslator {
	return &SimpleTranslator{
		messages: messages,
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
	if msg, ok := st.messages[key]; ok {
		return fmt.Sprintf(msg, params...)
	}
	// Fallback: use key as the format string.
	return fmt.Sprintf(key, params...)
}

// DefaultEnglishTranslations returns a map of default English messages.
//
// Returns:
//   - map[string]string: A map containing default English error messages
//     for validation failures.
func DefaultEnglishTranslations() map[string]string {
	return map[string]string{
		// Type errors
		"bool.type":   "expected boolean",
		"int.type":    "expected integer",
		"int64.type":  "expected int64",
		"string.type": "expected string",
		"slice.type":  "expected slice",

		// String validation
		"string.length":               "must be exactly %d characters long",
		"string.min":                  "minimum length is %d",
		"string.max":                  "maximum length is %d",
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
		"slice.forEach":             "element validation failed",
		"slice.element":             "element %d: %s",
		"slice.invalidLenParameter": "invalid parameter for len",
		"slice.invalidMinParameter": "invalid parameter for min",
		"slice.invalidMaxParameter": "invalid parameter for max",
		"slice.unknownValidator":    "unknown slice validator: %s",
		"slice.notSlice":            "value is not a slice",

		// Legacy compatibility
		"bool.notBool": "value is not a boolean",
	}
}
