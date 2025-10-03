package structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/aatuh/validate/v3/core"
	verrs "github.com/aatuh/validate/v3/errors"
)

// StructValidator provides struct validation functionality.
//
// Fields:
//   - validator: The underlying Validate instance for validation rules.
type StructValidator struct{ validator *core.Validate }

// NewStructValidator creates a new StructValidator instance.
//
// Parameters:
//   - v: The Validate instance to use for validation rules.
//
// Returns:
//   - *StructValidator: A new StructValidator instance.
func NewStructValidator(v *core.Validate) *StructValidator {
	return &StructValidator{validator: v}
}

// ValidateStruct keeps backward compatibility and uses default options.
//
// Parameters:
//   - s: The struct to validate.
//
// Returns:
//   - error: Validation errors if any, nil if valid.
func (sv *StructValidator) ValidateStruct(s any) error {
	return sv.ValidateStructWithOpts(s, core.ValidateOpts{})
}

// ValidateStructWithOpts validates s, honoring StopOnFirst and PathSep.
// Expected tag example: `validate:"string;min=3;max=10"`.
//
// Parameters:
//   - s: The struct to validate.
//   - opts: Validation options including StopOnFirst and PathSep.
//
// Returns:
//   - error: Validation errors if any, nil if valid.
func (sv *StructValidator) ValidateStructWithOpts(
	s any, opts core.ValidateOpts,
) error {
	opts = core.ApplyOpts(sv.validator, opts)

	val := reflect.ValueOf(s)
	typ := reflect.TypeOf(s)

	// Dereference pointer if necessary.
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("ValidateStruct: expected struct, got %T", s)
	}

	var errs verrs.Errors

	// walkStruct returns true to continue, false to stop early.
	var walkStruct func(v reflect.Value, t reflect.Type, path string) bool
	walkStruct = func(v reflect.Value, t reflect.Type, path string) bool {
		for i := 0; i < v.NumField(); i++ {
			ft := t.Field(i)
			fv := v.Field(i)

			// Skip unexported fields.
			if ft.PkgPath != "" {
				continue
			}

			fieldPath := fieldPathJoin(path, ft.Name, opts.PathSep)

			// Recurse into structs/slices/maps when no tag is present.
			tag := ft.Tag.Get("validate")
			if tag == "" {
				// Dereference pointer before checking kind
				derefFv := derefPointer(fv)
				switch derefFv.Kind() {
				case reflect.Struct:
					if !walkStruct(derefFv, derefFv.Type(), fieldPath) &&
						opts.StopOnFirst {
						return false
					}
					continue
				case reflect.Slice, reflect.Array:
					for j := 0; j < derefFv.Len(); j++ {
						ep := fieldPath + "[" + strconv.Itoa(j) + "]"
						ev := derefFv.Index(j)
						// Dereference pointer in slice elements
						derefEv := derefPointer(ev)
						if derefEv.Kind() == reflect.Struct {
							if !walkStruct(derefEv, derefEv.Type(), ep) &&
								opts.StopOnFirst {
								return false
							}
						}
					}
					continue
				case reflect.Map:
					for _, mk := range derefFv.MapKeys() {
						ev := derefFv.MapIndex(mk)
						ep := fieldPath + "[" + fmt.Sprint(
							mk.Interface(),
						) + "]"
						// Dereference pointer in map values
						derefEv := derefPointer(ev)
						if derefEv.Kind() == reflect.Struct {
							if !walkStruct(derefEv, derefEv.Type(), ep) &&
								opts.StopOnFirst {
								return false
							}
						}
					}
					continue
				default:
					continue
				}
			}

			// Validate with rules from tag.
			rules := strings.Split(tag, ";")
			fn, err := sv.validator.FromRules(rules)
			if err != nil {
				errs = append(errs, verrs.FieldError{
					Path: fieldPath, Code: verrs.CodeUnknown,
					Msg: err.Error(),
				})
				if opts.StopOnFirst {
					return false
				}
				continue
			}
			if err := fn(fv.Interface()); err != nil {
				// Check if the error is already a structured FieldError
				var fieldErrors verrs.Errors
				if errors.As(err, &fieldErrors) {
					// Preserve structured errors and update their paths
					for _, fe := range fieldErrors {
						fe.Path = fieldPathJoin(fieldPath, fe.Path, opts.PathSep)
						errs = append(errs, fe)
					}
				} else {
					// Fallback for non-structured errors
					errs = append(errs, verrs.FieldError{
						Path: fieldPath, Code: verrs.CodeUnknown,
						Msg: err.Error(),
					})
				}
				if opts.StopOnFirst {
					return false
				}
			}
		}
		return true
	}

	// Start the walk from the root.
	walkStruct(val, typ, "")

	if len(errs) > 0 {
		return errs
	}
	return nil
}

// derefPointer dereferences a pointer value recursively until it reaches a non-pointer type.
func derefPointer(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v
}

// fieldPathJoin joins path parts with a custom separator.
// Handles bracket-prefixed paths (e.g., "[0]", "[key]") by concatenating without separator.
func fieldPathJoin(base, name, sep string) string {
	if base == "" {
		return name
	}
	if sep == "" {
		sep = "."
	}
	// If the child path starts with a bracket, concatenate directly without separator
	if len(name) > 0 && name[0] == '[' {
		return base + name
	}
	return base + sep + name
}
