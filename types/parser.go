package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// truncateForError truncates a string for use in error messages to prevent
// extremely long error messages from fuzz testing.
func truncateForError(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// SplitTag splits a tag string by semicolons, respecting parentheses.
func SplitTag(tag string) []string {
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

func splitTagSafely(tag string) []string { return SplitTag(tag) }

// ParseTag parses a struct tag string into a slice of rules using global
// custom type registrations.
func ParseTag(tag string) ([]Rule, error) {
	return ParseTagWithRegistry(tag, nil)
}

// ParseTagWithRegistry parses a struct tag string with an optional per-instance
// custom type registry. Per-instance types are checked before global types.
// Example: "string;min=3;max=50" -> []Rule
func ParseTagWithRegistry(tag string, registry *TypeRegistry) ([]Rule, error) {
	if tag == "" {
		return nil, nil
	}

	parts := SplitTag(tag)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty tag")
	}

	var rules []Rule
	baseType := parts[0]
	if isGenericRuleToken(baseType) {
		for _, part := range parts {
			rule, err := parseGenericRule(part)
			if err != nil {
				return nil, err
			}
			if rule != nil {
				rules = append(rules, *rule)
			}
		}
		return rules, nil
	}

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
	case "float":
		rules = append(rules, NewRule(KFloat, nil))
		for _, part := range parts[1:] {
			rule, err := parseNumberRule(part)
			if err != nil {
				return nil, fmt.Errorf("invalid float rule %q: %w", truncateForError(part, 50), err)
			}
			if rule != nil {
				rules = append(rules, *rule)
			}
		}
	case "slice":
		rules = append(rules, NewRule(KSlice, nil))
		for _, part := range parts[1:] {
			rule, err := parseSliceRule(part, registry)
			if err != nil {
				return nil, fmt.Errorf("invalid slice rule %q: %w", truncateForError(part, 50), err)
			}
			if rule != nil {
				rules = append(rules, *rule)
			}
		}
	case "array":
		rules = append(rules, NewRule(KArray, nil))
		for _, part := range parts[1:] {
			rule, err := parseArrayRule(part, registry)
			if err != nil {
				return nil, fmt.Errorf("invalid array rule %q: %w", truncateForError(part, 50), err)
			}
			if rule != nil {
				rules = append(rules, *rule)
			}
		}
	case "map":
		rules = append(rules, NewRule(KMap, nil))
		for _, part := range parts[1:] {
			rule, err := parseMapRule(part, registry)
			if err != nil {
				return nil, fmt.Errorf("invalid map rule %q: %w", truncateForError(part, 50), err)
			}
			if rule != nil {
				rules = append(rules, *rule)
			}
		}
	case "bool":
		rules = append(rules, NewRule(KBool, nil))
		for _, part := range parts[1:] {
			rule, err := parseBoolRule(part)
			if err != nil {
				return nil, fmt.Errorf("invalid bool rule %q: %w", truncateForError(part, 20), err)
			}
			if rule != nil {
				rules = append(rules, *rule)
			}
		}
	case "time":
		rules = append(rules, NewRule(KTime, nil))
		for _, part := range parts[1:] {
			rule, err := parseTimeRule(part)
			if err != nil {
				return nil, fmt.Errorf("invalid time rule %q: %w", truncateForError(part, 50), err)
			}
			if rule != nil {
				rules = append(rules, *rule)
			}
		}
	default:
		// Check if it's a custom type
		if isTypeRegistered(baseType, registry) {
			// Create a custom type rule
			rules = append(rules, NewRule(Kind(baseType), nil))
			// Parse any additional rules for the custom type
			for _, part := range parts[1:] {
				rule, err := parseCustomTypeRule(part)
				if err != nil {
					return nil, fmt.Errorf("invalid %s rule %q: %w", baseType, truncateForError(part, 20), err)
				}
				if rule != nil {
					rules = append(rules, *rule)
				}
			}
		} else {
			return nil, fmt.Errorf("unknown type: %s", truncateForError(baseType, 50))
		}
	}

	return rules, nil
}

func isTypeRegistered(name string, registry *TypeRegistry) bool {
	if registry != nil && registry.IsTypeRegistered(name) {
		return true
	}
	return IsGlobalTypeRegistered(name)
}

func parseStringRule(part string) (*Rule, error) {
	if part == "" {
		return nil, nil
	}
	if rule, ok, err := parseGenericRuleMaybe(part); ok || err != nil {
		return rule, err
	}

	switch {
	case strings.HasPrefix(part, "length="), strings.HasPrefix(part, "len="):
		_, value, _ := strings.Cut(part, "=")
		n, err := strconv.Atoi(value)
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
	case strings.HasPrefix(part, "minRunes="):
		n, err := strconv.Atoi(strings.TrimPrefix(part, "minRunes="))
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KMinRunes, Args: map[string]any{"n": n}}, nil
	case strings.HasPrefix(part, "maxRunes="):
		n, err := strconv.Atoi(strings.TrimPrefix(part, "maxRunes="))
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KMaxRunes, Args: map[string]any{"n": n}}, nil
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
	case part == "nonempty":
		return &Rule{Kind: KNonEmpty, Args: nil}, nil
	case strings.HasPrefix(part, "contains="):
		return &Rule{Kind: KContains, Args: map[string]any{"value": strings.TrimPrefix(part, "contains=")}}, nil
	case strings.HasPrefix(part, "notContains="):
		return &Rule{Kind: KNotContains, Args: map[string]any{"value": strings.TrimPrefix(part, "notContains=")}}, nil
	case strings.HasPrefix(part, "prefix="):
		return &Rule{Kind: KPrefix, Args: map[string]any{"value": strings.TrimPrefix(part, "prefix=")}}, nil
	case strings.HasPrefix(part, "suffix="):
		return &Rule{Kind: KSuffix, Args: map[string]any{"value": strings.TrimPrefix(part, "suffix=")}}, nil
	case part == "url":
		return &Rule{Kind: KURL, Args: nil}, nil
	case part == "hostname":
		return &Rule{Kind: KHostname, Args: nil}, nil
	case part == "ip":
		return &Rule{Kind: KIP, Args: nil}, nil
	case part == "ipv4":
		return &Rule{Kind: KIPv4, Args: nil}, nil
	case part == "ipv6":
		return &Rule{Kind: KIPv6, Args: nil}, nil
	case part == "cidr":
		return &Rule{Kind: KCIDR, Args: nil}, nil
	case part == "ascii":
		return &Rule{Kind: KASCII, Args: nil}, nil
	case part == "alpha":
		return &Rule{Kind: KAlpha, Args: nil}, nil
	case part == "alnum":
		return &Rule{Kind: KAlnum, Args: nil}, nil
	default:
		return parseCustomRuleToken(part)
	}
}

func parseIntRule(part string) (*Rule, error) {
	if part == "" {
		return nil, nil
	}
	if rule, ok, err := parseGenericRuleMaybe(part); ok || err != nil {
		return rule, err
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
	case strings.HasPrefix(part, "gt="):
		return parseFloatArgRule(KGreaterThan, part, "gt=")
	case strings.HasPrefix(part, "gte="):
		return parseFloatArgRule(KGreaterThanEqual, part, "gte=")
	case strings.HasPrefix(part, "lt="):
		return parseFloatArgRule(KLessThan, part, "lt=")
	case strings.HasPrefix(part, "lte="):
		return parseFloatArgRule(KLessThanEqual, part, "lte=")
	case strings.HasPrefix(part, "between="):
		return parseBetweenRule(part)
	case part == "positive":
		return &Rule{Kind: KPositive, Args: nil}, nil
	case part == "nonnegative":
		return &Rule{Kind: KNonNegative, Args: nil}, nil
	default:
		return parseCustomRuleToken(part)
	}
}

func parseNumberRule(part string) (*Rule, error) {
	if part == "" {
		return nil, nil
	}
	if rule, ok, err := parseGenericRuleMaybe(part); ok || err != nil {
		return rule, err
	}
	switch {
	case part == "finite":
		return &Rule{Kind: KFinite, Args: nil}, nil
	case strings.HasPrefix(part, "min="):
		return parseFloatArgRule(KMinNumber, part, "min=")
	case strings.HasPrefix(part, "max="):
		return parseFloatArgRule(KMaxNumber, part, "max=")
	case strings.HasPrefix(part, "gt="):
		return parseFloatArgRule(KGreaterThan, part, "gt=")
	case strings.HasPrefix(part, "gte="):
		return parseFloatArgRule(KGreaterThanEqual, part, "gte=")
	case strings.HasPrefix(part, "lt="):
		return parseFloatArgRule(KLessThan, part, "lt=")
	case strings.HasPrefix(part, "lte="):
		return parseFloatArgRule(KLessThanEqual, part, "lte=")
	case strings.HasPrefix(part, "between="):
		return parseBetweenRule(part)
	case part == "positive":
		return &Rule{Kind: KPositive, Args: nil}, nil
	case part == "nonnegative":
		return &Rule{Kind: KNonNegative, Args: nil}, nil
	default:
		return parseCustomRuleToken(part)
	}
}

func parseSliceRule(part string, registry *TypeRegistry) (*Rule, error) {
	if part == "" {
		return nil, nil
	}
	if rule, ok, err := parseGenericRuleMaybe(part); ok || err != nil {
		return rule, err
	}

	switch {
	case strings.HasPrefix(part, "length="), strings.HasPrefix(part, "len="):
		_, value, _ := strings.Cut(part, "=")
		n, err := strconv.Atoi(value)
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
		innerRules, err := ParseTagWithRegistry(inner, registry)
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
	case part == "unique":
		return &Rule{Kind: KSliceUnique, Args: nil}, nil
	case strings.HasPrefix(part, "contains="):
		return &Rule{Kind: KSliceContains, Args: map[string]any{"value": strings.TrimPrefix(part, "contains=")}}, nil
	default:
		return parseCustomRuleToken(part)
	}
}

func parseArrayRule(part string, registry *TypeRegistry) (*Rule, error) {
	if part == "" {
		return nil, nil
	}
	if rule, ok, err := parseGenericRuleMaybe(part); ok || err != nil {
		return rule, err
	}

	switch {
	case strings.HasPrefix(part, "length="), strings.HasPrefix(part, "len="):
		_, value, _ := strings.Cut(part, "=")
		n, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KArrayLength, Args: map[string]any{"n": n}}, nil
	case strings.HasPrefix(part, "min="):
		n, err := strconv.Atoi(strings.TrimPrefix(part, "min="))
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KMinArrayLength, Args: map[string]any{"n": n}}, nil
	case strings.HasPrefix(part, "max="):
		n, err := strconv.Atoi(strings.TrimPrefix(part, "max="))
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KMaxArrayLength, Args: map[string]any{"n": n}}, nil
	case strings.HasPrefix(part, "foreach="):
		inner := strings.TrimPrefix(part, "foreach=")
		if !strings.HasPrefix(inner, "(") || !strings.HasSuffix(inner, ")") {
			return nil, fmt.Errorf("foreach must be wrapped in parentheses: %s", truncateForError(inner, 50))
		}
		inner = strings.TrimPrefix(inner, "(")
		inner = strings.TrimSuffix(inner, ")")

		innerRules, err := ParseTagWithRegistry(inner, registry)
		if err != nil {
			return nil, fmt.Errorf("invalid foreach rules: %w", err)
		}
		if len(innerRules) == 0 {
			return nil, fmt.Errorf("foreach must have at least one rule")
		}

		return &Rule{
			Kind: KArrayForEach,
			Args: map[string]any{"rules": innerRules},
			Elem: &innerRules[0],
		}, nil
	case part == "unique":
		return &Rule{Kind: KArrayUnique, Args: nil}, nil
	case strings.HasPrefix(part, "contains="):
		return &Rule{Kind: KArrayContains, Args: map[string]any{"value": strings.TrimPrefix(part, "contains=")}}, nil
	default:
		return parseCustomRuleToken(part)
	}
}

func parseMapRule(part string, registry *TypeRegistry) (*Rule, error) {
	if part == "" {
		return nil, nil
	}
	if rule, ok, err := parseGenericRuleMaybe(part); ok || err != nil {
		return rule, err
	}
	switch {
	case strings.HasPrefix(part, "length="), strings.HasPrefix(part, "len="):
		_, value, _ := strings.Cut(part, "=")
		n, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KMapLength, Args: map[string]any{"n": n}}, nil
	case strings.HasPrefix(part, "minKeys="), strings.HasPrefix(part, "min="):
		_, value, _ := strings.Cut(part, "=")
		n, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KMinMapKeys, Args: map[string]any{"n": n}}, nil
	case strings.HasPrefix(part, "maxKeys="), strings.HasPrefix(part, "max="):
		_, value, _ := strings.Cut(part, "=")
		n, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KMaxMapKeys, Args: map[string]any{"n": n}}, nil
	case strings.HasPrefix(part, "keys="):
		return parseNestedRulesRule(KMapKeys, part, "keys=", registry)
	case strings.HasPrefix(part, "values="):
		return parseNestedRulesRule(KMapValues, part, "values=", registry)
	default:
		return parseCustomRuleToken(part)
	}
}

func parseTimeRule(part string) (*Rule, error) {
	if part == "" {
		return nil, nil
	}
	if rule, ok, err := parseGenericRuleMaybe(part); ok || err != nil {
		return rule, err
	}
	switch {
	case part == "notzero":
		return &Rule{Kind: KTimeNotZero, Args: nil}, nil
	case strings.HasPrefix(part, "before="):
		t, err := parseRFC3339(strings.TrimPrefix(part, "before="))
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KTimeBefore, Args: map[string]any{"time": t}}, nil
	case strings.HasPrefix(part, "after="):
		t, err := parseRFC3339(strings.TrimPrefix(part, "after="))
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KTimeAfter, Args: map[string]any{"time": t}}, nil
	case strings.HasPrefix(part, "between="):
		raw := strings.TrimPrefix(part, "between=")
		values := strings.SplitN(raw, ",", 2)
		if len(values) != 2 {
			return nil, fmt.Errorf("between requires start,end")
		}
		start, err := parseRFC3339(values[0])
		if err != nil {
			return nil, err
		}
		end, err := parseRFC3339(values[1])
		if err != nil {
			return nil, err
		}
		return &Rule{Kind: KTimeBetween, Args: map[string]any{"start": start, "end": end}}, nil
	default:
		return parseCustomRuleToken(part)
	}
}

func parseCustomTypeRule(part string) (*Rule, error) {
	if part == "" {
		return nil, nil
	}
	if rule, ok, err := parseGenericRuleMaybe(part); ok || err != nil {
		return rule, err
	}

	return parseCustomRuleToken(part)
}

func parseBoolRule(part string) (*Rule, error) {
	if part == "" {
		return nil, nil
	}
	if rule, ok, err := parseGenericRuleMaybe(part); ok || err != nil {
		return rule, err
	}
	switch part {
	case "true":
		return &Rule{Kind: KBoolTrue, Args: nil}, nil
	case "false":
		return &Rule{Kind: KBoolFalse, Args: nil}, nil
	}
	return parseCustomRuleToken(part)
}

func parseCustomRuleToken(part string) (*Rule, error) {
	if strings.HasPrefix(part, "custom:") {
		raw := strings.TrimPrefix(part, "custom:")
		name, value, hasValue := strings.Cut(raw, "=")
		if err := validateCustomRuleName(name); err != nil {
			return nil, err
		}
		var args map[string]any
		if hasValue {
			args = map[string]any{"value": value}
		}
		return &Rule{Kind: Kind(name), Args: args}, nil
	}
	if strings.Contains(part, "=") {
		return nil, fmt.Errorf("unknown custom rule %q; use custom:name=value for parameterized custom rules", truncateForError(part, 50))
	}
	if err := validateCustomRuleName(part); err != nil {
		return nil, err
	}
	return &Rule{Kind: Kind(part), Args: nil}, nil
}

func validateCustomRuleName(name string) error {
	if name == "" {
		return fmt.Errorf("custom rule name cannot be empty")
	}
	for i, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r == '_' || r == '-' || r == '.':
		case i > 0 && r >= '0' && r <= '9':
		default:
			return fmt.Errorf("invalid custom rule name: %s", truncateForError(name, 50))
		}
	}
	return nil
}

func isGenericRuleToken(part string) bool {
	return part == "required" || part == "omitempty"
}

func parseGenericRuleMaybe(part string) (*Rule, bool, error) {
	if !isGenericRuleToken(part) {
		return nil, false, nil
	}
	rule, err := parseGenericRule(part)
	return rule, true, err
}

func parseGenericRule(part string) (*Rule, error) {
	switch part {
	case "":
		return nil, nil
	case "required":
		return &Rule{Kind: KRequired, Args: nil}, nil
	case "omitempty":
		return &Rule{Kind: KOmitempty, Args: nil}, nil
	default:
		return nil, fmt.Errorf("unknown generic rule: %s", truncateForError(part, 50))
	}
}

func parseFloatArgRule(kind Kind, part, prefix string) (*Rule, error) {
	n, err := strconv.ParseFloat(strings.TrimPrefix(part, prefix), 64)
	if err != nil {
		return nil, err
	}
	return &Rule{Kind: kind, Args: map[string]any{"n": n}}, nil
}

func parseBetweenRule(part string) (*Rule, error) {
	raw := strings.TrimPrefix(part, "between=")
	values := strings.SplitN(raw, ",", 2)
	if len(values) != 2 {
		return nil, fmt.Errorf("between requires min,max")
	}
	min, err := strconv.ParseFloat(strings.TrimSpace(values[0]), 64)
	if err != nil {
		return nil, err
	}
	max, err := strconv.ParseFloat(strings.TrimSpace(values[1]), 64)
	if err != nil {
		return nil, err
	}
	return &Rule{Kind: KBetween, Args: map[string]any{"min": min, "max": max}}, nil
}

func parseNestedRulesRule(kind Kind, part, prefix string, registry *TypeRegistry) (*Rule, error) {
	inner := strings.TrimPrefix(part, prefix)
	if !strings.HasPrefix(inner, "(") || !strings.HasSuffix(inner, ")") {
		return nil, fmt.Errorf("%s must be wrapped in parentheses: %s", strings.TrimSuffix(prefix, "="), truncateForError(inner, 50))
	}
	inner = strings.TrimPrefix(inner, "(")
	inner = strings.TrimSuffix(inner, ")")
	innerRules, err := ParseTagWithRegistry(inner, registry)
	if err != nil {
		return nil, err
	}
	if len(innerRules) == 0 {
		return nil, fmt.Errorf("nested rules must have at least one rule")
	}
	return &Rule{Kind: kind, Args: map[string]any{"rules": innerRules}}, nil
}

func parseRFC3339(value string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, strings.TrimSpace(value))
}
