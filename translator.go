package validate

import "fmt"

// Translator is an interface for obtaining localized messages.
type Translator interface {
	// T returns a localized message for the given key and parameters.
	T(key string, params ...any) string
}

// SimpleTranslator is a basic implementation of Translator using a map.
type SimpleTranslator struct {
	messages map[string]string
}

// NewSimpleTranslator creates a new SimpleTranslator.
func NewSimpleTranslator(messages map[string]string) *SimpleTranslator {
	return &SimpleTranslator{
		messages: messages,
	}
}

// Translate returns the translated message or the default if not found.
func (st *SimpleTranslator) Translate(key string, params ...any) string {
	if msg, ok := st.messages[key]; ok {
		return fmt.Sprintf(msg, params...)
	}
	// Fallback: use key as the format string.
	return fmt.Sprintf(key, params...)
}

// DefaultEnglishTranslations returns a map of default English messages.
func DefaultEnglishTranslations() map[string]string {
	return map[string]string{
		"bool.notBool":                "value is not a boolean",
		"int.min":                     "must be at least %d",
		"int.max":                     "must be at most %d",
		"int.invalidMinParameter":     "invalid parameter for min",
		"int.invalidMaxParameter":     "invalid parameter for max",
		"int.unknownIntValidator":     "unknown int validator: %s",
		"int.unknownInt64Validator":   "unknown int64 validator: %s",
		"int.notInteger":              "value is not an integer",
		"int.notInt64":                "value is not an int64",
		"string.length":               "must be exactly %d characters long",
		"string.minLength":            "must be at least %d characters long",
		"string.maxLength":            "must be at most %d characters long",
		"string.email.invalid":        "must be a valid email address",
		"string.regex.invalidPattern": "invalid regex pattern: %s",
		"string.regex.noMatch":        "must match pattern %s",
		"slice.length":                "must have exactly %d elements",
		"slice.min":                   "must have at least %d elements",
		"slice.max":                   "must have at most %d elements",
		"slice.element":               "element %d: %s",
		"slice.invalidLenParameter":   "invalid parameter for len",
		"slice.invalidMinParameter":   "invalid parameter for min",
		"slice.invalidMaxParameter":   "invalid parameter for max",
		"slice.unknownValidator":      "unknown slice validator: %s",
		"slice.notSlice":              "value is not a slice",
	}
}
