package errors

const (
	// Generic
	CodeUnknown   = "unknown"
	CodeRequired  = "required"
	CodeOmitEmpty = "omitempty" // informational when skipped

	// String
	CodeStringMin      = "string.min"
	CodeStringMax      = "string.max"
	CodeStringNonEmpty = "string.nonempty"
	CodeStringPattern  = "string.pattern"
	CodeStringOneOf    = "string.oneof"
	CodeStringPrefix   = "string.prefix"
	CodeStringSuffix   = "string.suffix"
	CodeStringURL      = "string.url"
	CodeStringUUID     = "string.uuid"
	CodeStringHost     = "string.hostname"

	// Number (covers ints and floats)
	CodeNumberMin      = "number.min"
	CodeNumberMax      = "number.max"
	CodeNumberPositive = "number.positive"
	CodeNumberNonNeg   = "number.nonnegative"
	CodeNumberBetween  = "number.between"

	// Slice
	CodeSliceMin      = "slice.min"
	CodeSliceMax      = "slice.max"
	CodeSliceUnique   = "slice.unique"
	CodeSliceContains = "slice.contains"

	// Map
	CodeMapMinKeys = "map.minkeys"
	CodeMapMaxKeys = "map.maxkeys"

	// Time
	CodeTimeNotZero = "time.notzero"
	CodeTimeBefore  = "time.before"
	CodeTimeAfter   = "time.after"
	CodeTimeBetween = "time.between"
)
