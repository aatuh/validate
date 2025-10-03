// Package types contains compiler helpers and rule primitives.
package types

import (
	"math"
	"strconv"
)

/*
toInt64 attempts to coerce supported integer representations to int64.
It rejects values that would overflow int64 and non-integer floats.
*/
func toInt64(v any) (int64, bool) {
	switch x := v.(type) {
	case int:
		return int64(x), true
	case int8:
		return int64(x), true
	case int16:
		return int64(x), true
	case int32:
		return int64(x), true
	case int64:
		return x, true

	case uint:
		u := uint64(x)
		if u > math.MaxInt64 {
			return 0, false
		}
		return int64(u), true
	case uint8:
		return int64(x), true
	case uint16:
		return int64(x), true
	case uint32:
		return int64(x), true
	case uint64:
		if x > math.MaxInt64 {
			return 0, false
		}
		return int64(x), true

	case string:
		// Only accept explicit base-10 integers.
		n, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	}

	// No float acceptance to avoid silent truncation.
	return 0, false
}
