package core

import (
	"context"
	"reflect"

	"github.com/aatuh/validate/v3/translator"
	"github.com/aatuh/validate/v3/types"
)

// StructRuleContext is the runtime input for struct-level field rules.
type StructRuleContext struct {
	Path       string
	Field      reflect.StructField
	Value      any
	Owner      reflect.Value
	Rule       types.Rule
	Context    context.Context
	Translator translator.Translator
}

// FieldValue returns an exported same-level field value by Go field name.
func (ctx StructRuleContext) FieldValue(name string) (any, bool) {
	if !ctx.Owner.IsValid() || ctx.Owner.Kind() != reflect.Struct {
		return nil, false
	}
	field := ctx.Owner.FieldByName(name)
	if !field.IsValid() {
		return nil, false
	}
	value, ok := structRuleValue(field)
	return value, ok
}

// StructRuleFunc validates one struct field with access to same-level fields.
type StructRuleFunc func(StructRuleContext) error

// StructRuleCompiler compiles a struct-level rule before validation.
type StructRuleCompiler func(types.Rule) (StructRuleFunc, error)

func structRuleValue(v reflect.Value) (any, bool) {
	if !v.IsValid() {
		return nil, false
	}
	for v.IsValid() && (v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface) {
		if v.IsNil() {
			return nil, true
		}
		v = v.Elem()
	}
	if !v.IsValid() || !v.CanInterface() {
		return nil, false
	}
	return v.Interface(), true
}
