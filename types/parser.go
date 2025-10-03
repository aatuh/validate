package types

import (
	"fmt"
	"strconv"
	"strings"
)

// truncateForError truncates a string for use in error messages to prevent
// extremely long error messages from fuzz testing.
func truncateForError(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ParseTag parses a struct tag string into a slice of rules.
// splitTagSafely splits a tag string by semicolons, respecting parentheses.
func splitTagSafely(tag string) []string {
	var parts []string
	var current strings.Builder
	parenDepth := 0

	for _, char := range tag {
		switch char {
		case ';':
			if parenDepth == 0 {
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		case '(':
			parenDepth++
			current.WriteRune(char)
		case ')':
			parenDepth--
			current.WriteRune(char)
		default:
			current.WriteRune(char)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// Example: "string;min=3;max=50" -> []Rule
func ParseTag(tag string) ([]Rule, error) {
	if tag == "" {
		return nil, nil
	}

	parts := splitTagSafely(tag)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty tag")
	}

	var rules []Rule
	baseType := parts[0]

	switch baseType {
	case "string":
		rules = append(rules, NewRule(KString, nil))
		for _, part := range parts[1:] {
			rule, err := parseStringRule(part)
			if err != nil {
				return nil, fmt.Errorf("invalid string rule %q: %w", truncateForError(part, 20), err)
			}
			if rule != nil {
				rules = append(rules, *rule)
			}
		}
	case "int", "int64":
		kind := KInt
		if baseType == "int64" {
			kind = KInt64
		}
		rules = append(rules, NewRule(kind, nil))
		for _, part := range parts[1:] {
			rule, err := parseIntRule(part)
			if err != nil {
				return nil, fmt.Errorf("invalid int rule %q: %w", truncateForError(part, 50), err)
			}
			if rule != nil {
				rules = append(rules, *rule)
			}
		}
	case "slice":
		rules = append(rules, NewRule(KSlice, nil))
		for _, part := range parts[1:] {
			rule, err := parseSliceRule(part)
			if err != nil {
				return nil, fmt.Errorf("invalid slice rule %q: %w", truncateForError(part, 50), err)
			}
			if rule != nil {
				rules = append(rules, *rule)
			}
		}
	case "bool":
		rules = append(rules, NewRule(KBool, nil))
	default:
		return nil, fmt.Errorf("unknown type: %s", truncateForError(baseType, 50))
	}

	return rules, nil
}

func parseStringRule(part string) (*Rule, error) {
	if part == "" {
		return nil, nil
	}

	switch {
	case strings.HasPrefix(part, "length="):
		n, err := strconv.Atoi(strings.TrimPrefix(part, "length="))
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KLength, Args: map[string]any{"n": n}}, nil
	case strings.HasPrefix(part, "min="):
		n, err := strconv.Atoi(strings.TrimPrefix(part, "min="))
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KMinLength, Args: map[string]any{"n": n}}, nil
	case strings.HasPrefix(part, "max="):
		n, err := strconv.Atoi(strings.TrimPrefix(part, "max="))
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KMaxLength, Args: map[string]any{"n": n}}, nil
	case strings.HasPrefix(part, "regex="):
		pattern := strings.TrimPrefix(part, "regex=")
		return &Rule{Kind: KRegex, Args: map[string]any{"pattern": pattern}}, nil
	case strings.HasPrefix(part, "oneof="):
		valueStr := strings.TrimPrefix(part, "oneof=")
		// Support both comma and space delimited values
		var values []string
		if strings.Contains(valueStr, ",") {
			// Comma delimited: red,green,blue
			values = strings.Split(valueStr, ",")
		} else {
			// Space delimited: red green blue
			values = strings.Fields(valueStr)
		}
		return &Rule{Kind: KOneOf, Args: map[string]any{"values": values}}, nil
	default:
		// Allow unknown rules to be passed through as custom rules
		// This enables plugin-based validation (email, uuid, etc.)
		return &Rule{Kind: Kind(part), Args: nil}, nil
	}
}

func parseIntRule(part string) (*Rule, error) {
	if part == "" {
		return nil, nil
	}

	switch {
	case strings.HasPrefix(part, "min="):
		n, err := strconv.ParseInt(strings.TrimPrefix(part, "min="), 10, 64)
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KMinInt, Args: map[string]any{"n": n}}, nil
	case strings.HasPrefix(part, "max="):
		n, err := strconv.ParseInt(strings.TrimPrefix(part, "max="), 10, 64)
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KMaxInt, Args: map[string]any{"n": n}}, nil
	default:
		return nil, fmt.Errorf("unknown int rule: %s", truncateForError(part, 50))
	}
}

func parseSliceRule(part string) (*Rule, error) {
	if part == "" {
		return nil, nil
	}

	switch {
	case strings.HasPrefix(part, "length="):
		n, err := strconv.Atoi(strings.TrimPrefix(part, "length="))
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KSliceLength, Args: map[string]any{"n": n}}, nil
	case strings.HasPrefix(part, "min="):
		n, err := strconv.Atoi(strings.TrimPrefix(part, "min="))
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KMinSliceLength, Args: map[string]any{"n": n}}, nil
	case strings.HasPrefix(part, "max="):
		n, err := strconv.Atoi(strings.TrimPrefix(part, "max="))
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KMaxSliceLength, Args: map[string]any{"n": n}}, nil
	case strings.HasPrefix(part, "foreach="):
		// Parse nested rules from foreach=(string;min=2;max=10)
		inner := strings.TrimPrefix(part, "foreach=")
		if !strings.HasPrefix(inner, "(") || !strings.HasSuffix(inner, ")") {
			return nil, fmt.Errorf("foreach must be wrapped in parentheses: %s", truncateForError(inner, 50))
		}
		inner = strings.TrimPrefix(inner, "(")
		inner = strings.TrimSuffix(inner, ")")

		// Parse the inner rules
		innerRules, err := ParseTag(inner)
		if err != nil {
			return nil, fmt.Errorf("invalid foreach rules: %w", err)
		}

		// Create a ForEach rule with all inner rules
		if len(innerRules) == 0 {
			return nil, fmt.Errorf("foreach must have at least one rule")
		}

		return &Rule{
			Kind: KForEach,
			Args: map[string]any{"rules": innerRules}, // Store all inner rules
			Elem: &innerRules[0],                      // Keep first rule for backward compatibility
		}, nil
	default:
		return nil, fmt.Errorf("unknown slice rule: %s", truncateForError(part, 50))
	}
}
