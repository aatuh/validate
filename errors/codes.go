package errors

const (
	// Generic
	CodeUnknown   = "unknown"
	CodeRequired  = "required"
	CodeOmitEmpty = "omitempty" // informational when skipped

	// String
	CodeStringType                = "string.type"
	CodeStringLength              = "string.length"
	CodeStringMin                 = "string.min"
	CodeStringMax                 = "string.max"
	CodeStringNonEmpty            = "string.nonempty"
	CodeStringPattern             = "string.pattern"
	CodeStringOneOf               = "string.oneof"
	CodeStringPrefix              = "string.prefix"
	CodeStringSuffix              = "string.suffix"
	CodeStringURL                 = "string.url"
	CodeStringHost                = "string.hostname"
	CodeStringRegexInvalidPattern = "string.regex.invalidPattern"
	CodeStringRegexInputTooLong   = "string.regex.inputTooLong"
	CodeStringRegexNoMatch        = "string.regex.noMatch"
	CodeStringMinRunes            = "string.minRunes"
	CodeStringMaxRunes            = "string.maxRunes"

	// Number (covers ints and floats)
	CodeIntType        = "int.type"
	CodeInt64Type      = "int64.type"
	CodeIntMin         = "int.min"
	CodeIntMax         = "int.max"
	CodeNumberMin      = "number.min"
	CodeNumberMax      = "number.max"
	CodeNumberPositive = "number.positive"
	CodeNumberNonNeg   = "number.nonnegative"
	CodeNumberBetween  = "number.between"

	// Slice
	CodeSliceType     = "slice.type"
	CodeSliceLength   = "slice.length"
	CodeSliceMin      = "slice.min"
	CodeSliceMax      = "slice.max"
	CodeSliceForEach  = "slice.forEach"
	CodeSliceUnique   = "slice.unique"
	CodeSliceContains = "slice.contains"

	// Map
	CodeMapMinKeys = "map.minkeys"
	CodeMapMaxKeys = "map.maxkeys"

	// Bool
	CodeBoolType = "bool.type"

	// Time
	CodeTimeNotZero = "time.notzero"
	CodeTimeBefore  = "time.before"
	CodeTimeAfter   = "time.after"
	CodeTimeBetween = "time.between"
)
