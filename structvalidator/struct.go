package structvalidator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/aatuh/validate"
	verrs "github.com/aatuh/validate/errors"
)

// StructValidator provides struct validation functionality.
type StructValidator struct {
	validator *validate.Validate
}

// NewStructValidator creates a new StructValidator instance.
func NewStructValidator(v *validate.Validate) *StructValidator {
	return &StructValidator{validator: v}
}

// ValidateStruct keeps backward compatibility and uses default options.
func (sv *StructValidator) ValidateStruct(s any) error {
	return sv.ValidateStructWithOpts(s, validate.ValidateOpts{})
}

// ValidateStructWithOpts validates s, honoring StopOnFirst and PathSep.
// Expected tag example: `validate:"string;min=3;max=10"`.
func (sv *StructValidator) ValidateStructWithOpts(
	s any, opts validate.ValidateOpts,
) error {
	opts = validate.ApplyOpts(sv.validator, opts)

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
				switch fv.Kind() {
				case reflect.Struct:
					if !walkStruct(fv, fv.Type(), fieldPath) &&
						opts.StopOnFirst {
						return false
					}
					continue
				case reflect.Slice, reflect.Array:
					for j := 0; j < fv.Len(); j++ {
						ep := fieldPath + "[" + strconv.Itoa(j) + "]"
						ev := fv.Index(j)
						if ev.Kind() == reflect.Struct {
							if !walkStruct(ev, ev.Type(), ep) &&
								opts.StopOnFirst {
								return false
							}
						}
					}
					continue
				case reflect.Map:
					for _, mk := range fv.MapKeys() {
						ev := fv.MapIndex(mk)
						ep := fieldPath + "[" + fmt.Sprint(
							mk.Interface(),
						) + "]"
						if ev.Kind() == reflect.Struct {
							if !walkStruct(ev, ev.Type(), ep) &&
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
				// Until validators return FieldError, wrap as generic.
				errs = append(errs, verrs.FieldError{
					Path: fieldPath, Code: verrs.CodeUnknown,
					Msg: err.Error(),
				})
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

// fieldPathJoin joins path parts with a custom separator.
func fieldPathJoin(base, name, sep string) string {
	if base == "" {
		return name
	}
	if sep == "" {
		sep = "."
	}
	return base + sep + name
}
