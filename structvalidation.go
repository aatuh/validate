package validate

import (
	"fmt"
	"reflect"
	"strings"
)

// FieldError represents an error for a specific struct field.
type FieldError struct {
	Field string
	Err   error
}

func (fe FieldError) Error() string {
	return fmt.Sprintf("%s: %v", fe.Field, fe.Err)
}

// StructValidationError aggregates multiple FieldError values.
type StructValidationError struct {
	Errors []FieldError
}

func (sve StructValidationError) Error() string {
	var parts []string
	for _, e := range sve.Errors {
		parts = append(parts, e.Error())
	}
	return strings.Join(parts, "; ")
}

// ValidateStruct validates a struct based on its `validate` tag.
// Expected tag format for a field: `validate:"string;min=3;max=10"`
// This function iterates over exported struct fields, builds validators from
// the tag, and applies them to the field value.
func (v *Validate) ValidateStruct(s any) error {
	var errs []FieldError
	val := reflect.ValueOf(s)
	typ := reflect.TypeOf(s)

	// Dereference pointer if necessary.
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("ValidateStruct: expected a struct but got %T", s)
	}

	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i)
		fieldType := typ.Field(i)
		tag := fieldType.Tag.Get("validate")
		if tag == "" {
			continue
		}
		// Skip unexported fields.
		if !fieldVal.CanInterface() {
			continue
		}
		rules := strings.Split(tag, ";")
		validator, err := v.FromRules(rules)
		if err != nil {
			errs = append(errs, FieldError{Field: fieldType.Name, Err: err})
			continue
		}
		if err := validator(fieldVal.Interface()); err != nil {
			errs = append(errs, FieldError{Field: fieldType.Name, Err: err})
		}
	}

	if len(errs) > 0 {
		return StructValidationError{Errors: errs}
	}
	return nil
}
