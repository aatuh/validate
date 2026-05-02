package types

import (
	"regexp"
	"strings"

	"github.com/aatuh/validate/v3/errors"
)

const maxRegexPatternMessageRunes = 100

/*
compileRegexSafe prepares a regexp for a pattern, ensuring it is anchored and
that invalid pattern errors can use a sanitized pattern for translation.
*/
func (c *Compiler) compileRegexSafe(pattern string) (*regexp.Regexp, error) {
	pattern = normalizeRegexPattern(pattern)
	return regexp.Compile(pattern)
}

func normalizeRegexPattern(pattern string) string {
	if len(pattern) > 0 && pattern[0] != '^' {
		pattern = "^" + pattern
	}
	if n := len(pattern); n > 0 && pattern[n-1] != '$' {
		pattern = pattern + "$"
	}
	return pattern
}

func (c *Compiler) invalidRegexPatternError(pattern string) error {
	msg := c.translateMessage(
		errors.CodeStringRegexInvalidPattern,
		"invalid regex pattern: %s",
		[]any{regexPatternForMessage(pattern)},
	)
	return errors.Errors{errors.FieldError{
		Path: "", Code: errors.CodeStringRegexInvalidPattern, Msg: msg,
	}}
}

func regexPatternForMessage(pattern string) string {
	pattern = normalizeRegexPattern(pattern)
	if containsSensitiveMarker(pattern) {
		return "[redacted]"
	}
	runes := []rune(pattern)
	if len(runes) <= maxRegexPatternMessageRunes {
		return pattern
	}
	return string(runes[:maxRegexPatternMessageRunes]) + "..."
}

func containsSensitiveMarker(s string) bool {
	lower := strings.ToLower(s)
	for _, marker := range []string{
		"authorization",
		"bearer",
		"api_key",
		"apikey",
		"credential",
		"password",
		"passwd",
		"private_key",
		"secret",
		"token",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}
