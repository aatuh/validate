// Package core holds builder utilities for the validate library.
package core

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/aatuh/validate/v3/types"
)

/*
SerializeRules returns a deterministic, canonical string for a rule set.
Use it as a cache key for compiled validators. It avoids embedding
function addresses (which are process-specific and non-deterministic)
by emitting a stable "fn" marker for function arguments.
*/
func SerializeRules(rules []types.Rule) string {
	var b strings.Builder
	b.Grow(256)

	var writeRule func(r types.Rule)
	writeRule = func(r types.Rule) {
		b.WriteString("{")
		// Kind is stable.
		b.WriteString("kind:")
		b.WriteString(string(r.Kind))

		// Args are serialized with sorted keys for determinism.
		if r.Args != nil && len(r.Args) > 0 {
			b.WriteString(",args:{")
			keys := make([]string, 0, len(r.Args))
			for k := range r.Args {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for i, k := range keys {
				if i > 0 {
					b.WriteByte(',')
				}
				b.WriteString(k)
				b.WriteByte(':')
				serializeArg(&b, r.Args[k])
			}
			b.WriteByte('}')
		}

		b.WriteByte('}')
	}

	b.WriteByte('[')
	for i, r := range rules {
		if i > 0 {
			b.WriteByte(',')
		}
		writeRule(r)
	}
	b.WriteByte(']')

	return b.String()
}

/*
HasFuncArgs returns true if any rule argument is a function (directly or
nested inside maps/slices). This is used to decide whether to skip
caching, because function pointer addresses are not deterministic.
*/
func HasFuncArgs(rules []types.Rule) bool {
	for _, r := range rules {
		if argHasFunc(r.Args) {
			return true
		}
	}
	return false
}

func argHasFunc(v any) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Func:
		return true
	case reflect.Pointer, reflect.Interface:
		if rv.IsNil() {
			return false
		}
		return argHasFunc(rv.Elem().Interface())
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			if argHasFunc(rv.Index(i).Interface()) {
				return true
			}
		}
		return false
	case reflect.Map:
		iter := rv.MapRange()
		for iter.Next() {
			if argHasFunc(iter.Value().Interface()) {
				return true
			}
		}
		return false
	case reflect.Struct:
		// Best-effort: iterate exported fields only.
		rt := rv.Type()
		for i := 0; i < rv.NumField(); i++ {
			if rt.Field(i).IsExported() &&
				argHasFunc(rv.Field(i).Interface()) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// serializeArg emits a deterministic representation of a rule argument.
func serializeArg(b *strings.Builder, v any) {
	if v == nil {
		b.WriteString("nil")
		return
	}

	switch x := v.(type) {
	case string:
		b.WriteString(strconv.Quote(x))
	case bool:
		if x {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case int:
		b.WriteString(strconv.FormatInt(int64(x), 10))
	case int8:
		b.WriteString(strconv.FormatInt(int64(x), 10))
	case int16:
		b.WriteString(strconv.FormatInt(int64(x), 10))
	case int32:
		b.WriteString(strconv.FormatInt(int64(x), 10))
	case int64:
		b.WriteString(strconv.FormatInt(x, 10))
	case uint:
		b.WriteString(strconv.FormatUint(uint64(x), 10))
	case uint8:
		b.WriteString(strconv.FormatUint(uint64(x), 10))
	case uint16:
		b.WriteString(strconv.FormatUint(uint64(x), 10))
	case uint32:
		b.WriteString(strconv.FormatUint(uint64(x), 10))
	case uint64:
		b.WriteString(strconv.FormatUint(x, 10))
	case float32:
		b.WriteString(strconv.FormatFloat(float64(x), 'g', -1, 32))
	case float64:
		b.WriteString(strconv.FormatFloat(x, 'g', -1, 64))

	case []string:
		cp := append([]string(nil), x...)
		sort.Strings(cp)
		b.WriteByte('[')
		for i, s := range cp {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(strconv.Quote(s))
		}
		b.WriteByte(']')

	case []types.Rule:
		b.WriteString(SerializeRules(x))

	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		b.WriteByte('{')
		for i, k := range keys {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(k)
			b.WriteByte(':')
			serializeArg(b, x[k])
		}
		b.WriteByte('}')

	case *types.Rule:
		if x == nil {
			b.WriteString("nil")
		} else {
			// Serialize a single rule.
			b.WriteString("{")
			b.WriteString("kind:")
			b.WriteString(string(x.Kind))
			if x.Args != nil && len(x.Args) > 0 {
				b.WriteString(",args:{")
				keys := make([]string, 0, len(x.Args))
				for k := range x.Args {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for i, k := range keys {
					if i > 0 {
						b.WriteByte(',')
					}
					b.WriteString(k)
					b.WriteByte(':')
					serializeArg(b, x.Args[k])
				}
				b.WriteByte('}')
			}
			b.WriteByte('}')
		}

	default:
		// Avoid unstable serialization for funcs, pointers, etc.
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Func {
			b.WriteString(`"fn"`)
			return
		}
		// Fallback to fmt for other simple cases.
		b.WriteString(strconv.Quote(fmt.Sprintf("%v", v)))
	}
}
