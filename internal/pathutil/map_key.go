package pathutil

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	maxOrdinaryMapKeyBytes = 64
	redactedMapKey         = "<redacted>"
)

// MapKeySegment formats a map key for a validation path segment.
func MapKeySegment(key any) string {
	return "[" + MapKey(key) + "]"
}

// MapKey returns a bounded, privacy-aware map key representation.
func MapKey(key any) string {
	if key == nil {
		return "<nil>"
	}

	rv := reflect.ValueOf(key)
	switch rv.Kind() {
	case reflect.String:
		return stringMapKey(rv.String())
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return fmt.Sprint(key)
	default:
		return redactedMapKey
	}
}

func stringMapKey(key string) string {
	if isOrdinaryMapKey(key) && !hasSensitiveMarker(key) {
		return key
	}
	return redactedMapKey
}

func isOrdinaryMapKey(key string) bool {
	if len(key) > maxOrdinaryMapKeyBytes {
		return false
	}
	for _, r := range key {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-' || r == '.' || r == ':':
		default:
			return false
		}
	}
	return true
}

func hasSensitiveMarker(key string) bool {
	lower := strings.ToLower(key)
	if strings.Contains(lower, "@") || strings.Contains(lower, "://") {
		return true
	}

	normalized := strings.NewReplacer(
		"_", "",
		"-", "",
		".", "",
		":", "",
		"=", "",
	).Replace(lower)

	for _, marker := range []string{
		"password",
		"passwd",
		"secret",
		"token",
		"apikey",
		"accesskey",
		"auth",
		"bearer",
		"session",
		"cookie",
		"credential",
		"private",
		"jwt",
	} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}
