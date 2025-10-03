# validate

Composable validation for Go with fluent builders, rule tags, struct
validation, and optional message translation.

## Packages

- `github.com/aatuh/validate/v3`: main API (`Validate`, builders, `FromTag`)
- `github.com/aatuh/validate/v3/core`: generic validation engine (unified, cache-optimized)
- `github.com/aatuh/validate/v3/glue`: integration layer with builders
- `github.com/aatuh/validate/v3/validators`: type-specific rules
- `github.com/aatuh/validate/v3/errors`: error types and codes
- `github.com/aatuh/validate/v3/structvalidator`: struct validation
- `github.com/aatuh/validate/v3/translator`: i18n helpers

The package is architecturally separated into:
- **Core**: Generic validation engine, type-agnostic, with unified caching
- **Glue**: Integration layer that connects core with type-specific builders
- **Validators**: Type-specific validation implementations

## Quick start

### Basic validation

```go
package main

import (
    "fmt"
    "github.com/aatuh/validate/v3"
)

func main() {
    v := validate.New()

    // String: call Build() to get func(any) error.
    nameV := v.String().MinLength(3).MaxLength(50).Build()
    if err := nameV("John"); err != nil {
        fmt.Println("validation failed:", err)
    }

    // Int: call Build() to obtain func(any) error.
    ageV := v.Int().MinInt(18).MaxInt(120).Build()
    if err := ageV(25); err != nil {
        fmt.Println("validation failed:", err)
    }

    // FromTag: compile single tag strings (more convenient than FromRules)
    tagV, _ := v.FromTag("string;min=3;max=50")
    if err := tagV("John"); err != nil {
        fmt.Println("validation failed:", err)
    }
}
```

### Slice validation with element rules

```go
import (
    "github.com/aatuh/validate/v3"
    "github.com/aatuh/validate/v3/types"
)

v := validate.New()

// Method 1: ForEach with function (not cache-friendly)
tagElem := v.String().MinLength(2).Build()
tagsV := v.Slice().MinSliceLength(1).ForEach(tagElem).Build()

// Method 2: ForEachRules (cache-friendly, better performance)
tagsV2 := v.Slice().MinSliceLength(1).ForEachRules(
    types.NewRule(types.KString, nil),
    types.NewRule(types.KMinLength, map[string]any{"n": int64(2)}),
).Build()

// Method 3: ForEachStringBuilder (convenience for string elements)
stringBuilder := v.String().MinLength(2)
tagsV3 := v.Slice().MinSliceLength(1).ForEachStringBuilder(stringBuilder).Build()

if err := tagsV([]string{"go", "lib"}); err != nil {
    fmt.Println("validation failed:", err)
}
```

### Struct validation

```go
package main

import (
    "fmt"
    "github.com/aatuh/validate/v3"
    verrs "github.com/aatuh/validate/v3/errors"
    "github.com/aatuh/validate/v3/structvalidator"
)

type User struct {
    Name     string   `validate:"string;min=3;max=50"`
    Website  string   `validate:"string;min=5;max=100"`
    Age      int      `validate:"int;min=18;max=120"`
    ID       string   `validate:"string;min=5;max=20"`
    Tags     []string `validate:"slice;min=1;max=5;foreach=(string;min=2)"`
    Status   string   `validate:"string;oneof=active,inactive,pending"`
    Bio      string   `validate:"string;maxRunes=500"` // Unicode-aware length
}

func main() {
    v := validate.New()
    sv := structvalidator.NewStructValidator(v)

    u := User{
        Name:    "John Doe", 
        Website: "https://example.com", 
        Age:     25,
        ID:      "user123",
        Tags:    []string{"golang", "validation"},
        Status:  "active",
        Bio:     "Software developer",
    }
    if err := sv.ValidateStruct(u); err != nil {
        if es, ok := err.(verrs.Errors); ok {
            fmt.Println("errors:", es.AsMap())
        } else {
            fmt.Println("validation failed:", err)
        }
    }
}
```

### With translation

```go
package main

import (
    "fmt"
    "github.com/aatuh/validate/v3"
    "github.com/aatuh/validate/v3/translator"
)

func main() {
    msgs := map[string]string{
        "string.minLength": "doit contenir au moins %d caractères",
        "string.maxLength": "ne peut pas dépasser %d caractères",
    }
    tr := translator.NewSimpleTranslator(msgs)

    v := validate.New().WithTranslator(tr)

    check := v.String().MinLength(5).MaxLength(10).Build()
    if err := check("ab"); err != nil {
        fmt.Println("fr:", err)
    }
}
```

Tip: `translator.DefaultEnglishTranslations()` provides sensible defaults.

### FromTag convenience

For single tag strings, use `FromTag` instead of `FromRules`:

```go
v := validate.New()

// More convenient than FromRules([]string{"string;min=3;max=50"})
validator, _ := v.FromTag("string;min=3;max=50")
if err := validator("hello"); err != nil {
    fmt.Println("validation failed:", err)
}

// Root-level FromTag (works with nil Validate)
validator2, _ := validate.FromTag(nil, "int;min=1;max=100")
if err := validator2(50); err != nil {
    fmt.Println("validation failed:", err)
}
```

### Builder caching

The validation package includes automatic caching for builder validators to improve performance when the same validation rules are used repeatedly:

```go
v := validate.New()

// First call compiles and caches the validator
validator1 := v.String().MinLength(3).MaxLength(50).Build()

// Subsequent calls with identical rules return cached validator
validator2 := v.String().MinLength(3).MaxLength(50).Build()
// validator1 and validator2 are the same cached function

// Different rules create new validators
validator3 := v.String().MinLength(5).MaxLength(20).Build() // Different cache key
```

The cache uses rule serialization to create unique keys, so identical rule combinations are automatically cached for optimal performance.

### Cache-friendly slice validation

For better performance with slice validation, use `ForEachRules` instead of `ForEach`:

```go
v := validate.New()

// Cache-friendly: uses AST rules (cached)
sliceValidator := v.Slice().MinLength(1).ForEachRules(
    types.NewRule(types.KString, nil),
    types.NewRule(types.KMinLength, map[string]any{"n": int64(2)}),
).Build()

// Not cache-friendly: uses function closure (not cached)
elemValidator := v.String().MinLength(2).Build()
sliceValidator2 := v.Slice().MinLength(1).ForEach(elemValidator).Build()
```

## API reference

### Validate

```go
v := validate.New()
v = v.WithTranslator(tr)
v = v.PathSeparator(".")

// FromTag convenience
validator, _ := v.FromTag("string;min=3;max=50")

// Root-level FromTag
validator2, _ := validate.FromTag(nil, "int;min=1;max=100")

custom := map[string]func(any) error{
    "customRule": func(v any) error { return nil },
}
v2 := validate.NewWithCustomRules(custom)
```

### Builders

```go
// String
strV := v.String().MinLength(3).MaxLength(50).Regex(`^[a-z0-9_]+$`).Build()

// Int (accepts any Go int type at call time)
intV := v.Int().MinInt(0).MaxInt(100).Build()

// Int64 (requires exactly int64 at call time)
int64V := v.Int64().MinInt(0).MaxInt(100).Build()

// Slice
elem := v.String().MinLength(2).Build()
sliceV := v.Slice().MinSliceLength(1).ForEach(elem).Build()

// Cache-friendly slice validation
sliceV2 := v.Slice().MinSliceLength(1).ForEachRules(
    types.NewRule(types.KString, nil),
    types.NewRule(types.KMinLength, map[string]any{"n": int64(2)}),
).Build()

// Convenience method for string elements
stringBuilder := v.String().MinLength(2)
sliceV3 := v.Slice().MinSliceLength(1).ForEachStringBuilder(stringBuilder).Build()

// Bool
boolV := v.Bool() // type-check only
```

### Struct tags

```go
type Example struct {
    Name    string   `validate:"string;min=3;max=50"`
    Website string   `validate:"string;min=5;max=100"`
    Age     int      `validate:"int;min=18;max=120"`
    Tags    []string `validate:"slice;min=1;max=5"`
    Flag    bool     `validate:"bool"`
}
```

Supported tokens:

- string: `len`, `min`, `max`, `minRunes`, `maxRunes`, `oneof=a,b,c` (or space-separated), `regex=...`
- int/int64: `min`, `max` (use `int64;...` to require int64)
- slice: `len`, `min`, `max`, `foreach=(string;min=2)` (nested rules)
- bool: type only

### Errors

Struct validation returns `errors.Errors` which offers helpers:

```go
if es, ok := err.(verrs.Errors); ok {
    _ = es.Has("Name")
    _ = es.Filter("") // or a nested prefix like "Address."
    _ = es.AsMap()
}
```

Note: direct builder calls usually return a single `error`, not an
`errors.Errors` aggregate.

## Error codes

Stable constants for programmatic handling (see `errors/codes.go`):

```go
const (
    // String
    CodeStringMin = "string.min"
    CodeStringMax = "string.max"
    CodeStringNonEmpty = "string.nonempty"
    CodeStringPattern = "string.pattern"
    CodeStringOneOf = "string.oneof"

    // Number (ints/floats)
    CodeNumberMin = "number.min"
    CodeNumberMax = "number.max"

    // Slice
    CodeSliceMin = "slice.min"
    CodeSliceMax = "slice.max"
)
```
