package types

import (
	"regexp"

	"github.com/aatuh/validate/v3/errors"
)

/*
compileRegexSafe prepares a regexp for a pattern, ensuring it is anchored and
that "invalid pattern" errors include the pattern for translation.
*/
func (c *Compiler) compileRegexSafe(pattern string) (*regexp.Regexp, error) {
	// Anchor if caller forgot to.
	if len(pattern) > 0 && pattern[0] != '^' {
		pattern = "^" + pattern
	}
	if n := len(pattern); n > 0 && pattern[n-1] != '$' {
		pattern = pattern + "$"
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		// Important: pass pattern as a param for translations.
		_ = c.translateMessage(
			errors.CodeStringRegexInvalidPattern,
			"invalid regex pattern: %s",
			[]any{pattern},
		)
		return nil, err
	}
	return re, nil
}
