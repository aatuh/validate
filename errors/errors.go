package errors

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// FieldError represents one validation failure at a specific path.
// Path example: "User.Addresses[2].Zip"
//
// Fields:
//   - Path: The field path where the validation failed.
//   - Code: Stable machine-readable identifier (e.g., "string.min", "int.max").
//   - Param: Rule parameter (e.g., 3 for min length).
//   - Msg: Translated, human-readable message if a Translator is set.
type FieldError struct {
	Path string `json:"path"`
	// Code is a stable machine-readable identifier, e.g. "string.min",
	// "int.max", "slice.unique". Prefer stable codes in UIs and tests.
	Code string `json:"code"`
	// Param carries rule parameter, e.g. 3 for min length. Keep it simple.
	Param any `json:"param,omitempty"`
	// Msg is the translated, human-readable message if a Translator is set.
	Msg string `json:"message,omitempty"`
}

// String returns a concise string for logs.
//
// Returns:
//   - string: A formatted string representation of the field error.
func (e FieldError) String() string {
	p := ""
	if e.Param != nil {
		p = fmt.Sprintf(" param=%v", e.Param)
	}
	if e.Msg != "" {
		return fmt.Sprintf("%s [%s]%s: %s", e.Path, e.Code, p, e.Msg)
	}
	return fmt.Sprintf("%s [%s]%s", e.Path, e.Code, p)
}

// Errors is a collection of FieldError that implements error.
//
// The Error() message is a single line intended for logs. For structured
// consumption prefer AsMap or JSON marshaling.
//
// This type represents a collection of validation errors that can be used
// as a single error value.
type Errors []FieldError

// Error joins all error strings into one compact line.
//
// Returns:
//   - string: A single line containing all error messages.
func (es Errors) Error() string {
	if len(es) == 0 {
		return ""
	}
	var buf bytes.Buffer
	for i, e := range es {
		if i > 0 {
			buf.WriteString("; ")
		}
		buf.WriteString(e.String())
	}
	return buf.String()
}

// Has reports whether any error targets the exact path.
//
// Parameters:
//   - path: The field path to check for errors.
//
// Returns:
//   - bool: True if any error exists for the given path.
func (es Errors) Has(path string) bool {
	for _, e := range es {
		if e.Path == path {
			return true
		}
	}
	return false
}

// Filter returns errors whose Path has the given prefix. Useful for forms
// where fields are grouped, e.g. "User.Addresses".
//
// Parameters:
//   - prefix: The path prefix to filter by.
//
// Returns:
//   - Errors: A new Errors collection containing only matching errors.
func (es Errors) Filter(prefix string) Errors {
	out := make(Errors, 0, len(es))
	for _, e := range es {
		if strings.HasPrefix(e.Path, prefix) {
			out = append(out, e)
		}
	}
	return out
}

// AsMap groups errors by exact field path. The slice per key preserves
// original order (stable).
//
// Returns:
//   - map[string][]FieldError: A map where keys are field paths and values
//     are slices of errors for that path.
func (es Errors) AsMap() map[string][]FieldError {
	m := make(map[string][]FieldError, len(es))
	for _, e := range es {
		m[e.Path] = append(m[e.Path], e)
	}
	return m
}

// MarshalJSON ensures deterministic key ordering for better diffs.
//
// Returns:
//   - []byte: JSON representation of the errors.
//   - error: An error if JSON marshaling fails.
func (es Errors) MarshalJSON() ([]byte, error) {
	if len(es) == 0 {
		return []byte("[]"), nil
	}
	type fe FieldError
	cp := make([]fe, len(es))
	for i := range es {
		cp[i] = fe(es[i])
	}
	// No custom order within fields, but we can keep stable overall.
	return json.Marshal(cp)
}

// Unwrap allows using errors.Is/As when you wrap Errors with fmt.Errorf.
// Returns nil because there is no single underlying error to unwrap.
//
// Returns:
//   - error: Always returns nil.
func (es Errors) Unwrap() error { return nil }

// Join concatenates multiple error values into an Errors slice.
// It flattens nested Errors and ignores nils.
//
// Parameters:
//   - errs: Variable number of error values to join.
//
// Returns:
//   - Errors: A new Errors collection containing all non-nil errors.
func Join(errs ...error) Errors {
	var out Errors
	for _, err := range errs {
		if err == nil {
			continue
		}
		var es Errors
		if errors.As(err, &es) {
			out = append(out, es...)
			continue
		}
		// Wrap unknown error as generic at path "".
		out = append(out, FieldError{
			Path: "",
			Code: CodeUnknown,
			Msg:  err.Error(),
		})
	}
	return out
}

// SortByPath then Code to provide stable presentation when needed.
func (es Errors) Sort() {
	sort.SliceStable(es, func(i, j int) bool {
		if es[i].Path == es[j].Path {
			return es[i].Code < es[j].Code
		}
		return es[i].Path < es[j].Path
	})
}
