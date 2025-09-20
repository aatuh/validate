# validate

Composable validation for Go with fluent builders, rule tags, struct
validation, and optional message translation.

## Packages

- `github.com/aatuh/validate`: main API (`Validate`, builders)
- `github.com/aatuh/validate/validators`: type-specific rules
- `github.com/aatuh/validate/errors`: error types and codes
- `github.com/aatuh/validate/structvalidator`: struct validation
- `github.com/aatuh/validate/translator`: i18n helpers

## Quick start

### Basic validation

```go
package main

import (
    "fmt"
    "github.com/aatuh/validate"
)

func main() {
    v := validate.New()

    // String: Email() is terminal and returns func(any) error.
    emailV := v.String().MinLength(3).MaxLength(50).Email()
    if err := emailV("user@example.com"); err != nil {
        fmt.Println("validation failed:", err)
    }

    // Int: call Build() to obtain func(any) error.
    ageV := v.Int().MinInt(18).MaxInt(120).Build()
    if err := ageV(25); err != nil {
        fmt.Println("validation failed:", err)
    }
}
```

### Slice validation with element rules

```go
v := validate.New()
// tags must be non-empty strings, at least 2 chars each
tagElem := v.String().MinLength(2).Build()
tagsV := v.Slice().MinSliceLength(1).ForEach(tagElem).Build()
if err := tagsV([]string{"go", "lib"}); err != nil {
    fmt.Println("validation failed:", err)
}
```

### Struct validation

```go
package main

import (
    "fmt"
    "github.com/aatuh/validate"
    verrs "github.com/aatuh/validate/errors"
    "github.com/aatuh/validate/structvalidator"
)

type User struct {
    Name  string `validate:"string;min=3;max=50"`
    Email string `validate:"string;email"`
    Age   int    `validate:"int;min=18;max=120"`
}

func main() {
    v := validate.New()
    sv := structvalidator.NewStructValidator(v)

    u := User{Name: "John Doe", Email: "john@example.com", Age: 25}
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
    "github.com/aatuh/validate"
    "github.com/aatuh/validate/translator"
)

func main() {
    msgs := map[string]string{
        "string.minLength": "doit contenir au moins %d caractères",
        "string.email.invalid": "adresse email invalide",
    }
    tr := translator.NewSimpleTranslator(msgs)

    v := validate.New().WithTranslator(tr)

    check := v.String().MinLength(5).Email()
    if err := check("ab"); err != nil {
        fmt.Println("fr:", err)
    }
}
```

Tip: `translator.DefaultEnglishTranslations()` provides sensible defaults.

## API reference

### Validate

```go
v := validate.New()
v = v.WithTranslator(tr)
v = v.PathSeparator(".")

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

// Bool
boolV := v.Bool() // type-check only
```

### Struct tags

```go
type Example struct {
    Name  string   `validate:"string;min=3;max=50"`
    Email string   `validate:"string;email"`
    Age   int      `validate:"int;min=18;max=120"`
    Tags  []string `validate:"slice;min=1;max=5"`
    Flag  bool     `validate:"bool"`
}
```

Supported tokens:

- string: `len`, `min`, `max`, `oneof=a b c`, `regex=...`, `email`
- int/int64: `min`, `max` (use `int64;...` to require int64)
- slice: `len`, `min`, `max`
- bool: type only

### Errors

Struct validation returns `errors.Errors` which offers helpers:

```go
if es, ok := err.(verrs.Errors); ok {
    _ = es.Has("Email")
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
