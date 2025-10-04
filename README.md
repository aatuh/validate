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

For examples, see the [examples package](./examples).

Run examples:

```bash
# Run all examples in learning order
go test ./examples -v -count 1

# Run specific examples
go test ./examples -v -count 1 -run Test_structTags
```

### Supported data types

| Type   | Description                            |
|--------|----------------------------------------|
| string | UTF-8 strings; rune-aware length rules |
| int    | Any Go int type accepted at call time  |
| int64  | Requires exactly int64 at call time    |
| slice  | Any slice type; supports per-element   |
| bool   | Boolean type-check only                |

Plugin rules that operate on strings (imported by default):

- uuid: validates canonical UUID string
- ulid: validates ULID string
- email: validates email address string

Note on plugin registration:

- `validate.New()` already registers these plugin validators via blank
  imports inside the root package.
- If you use `core`/`glue` directly, blank-import the plugins yourself:

```go
import (
    _ "github.com/aatuh/validate/v3/validators/email"
    _ "github.com/aatuh/validate/v3/validators/ulid"
    _ "github.com/aatuh/validate/v3/validators/uuid"
)
```

### Tags by type

Tags can be used in struct field tags or with `FromTag`.

#### string

| Tag           | Meaning                                    |
|---------------|--------------------------------------------|
| len=N         | Exact byte length equals N                 |
| min=N         | Minimum byte length                        |
| max=N         | Maximum byte length                        |
| minRunes=N    | Minimum Unicode rune count                 |
| maxRunes=N    | Maximum Unicode rune count                 |
| oneof=a,b,c   | Value must be one of the listed values     |
| regex=PATTERN | Full-match against PATTERN (anchors added) |
| uuid          | UUID format (via plugin)                   |
| ulid          | ULID format (via plugin)                   |
| email         | Email format (via plugin)                  |

Notes:

- `oneof` supports comma or space separated values.
- Regex is made safe (anchors added, input length capped).

#### int / int64

| Tag   | Meaning               |
|-------|-----------------------|
| min=N | Minimum numeric value |
| max=N | Maximum numeric value |

Use `int64;...` to require exactly `int64` at validation time; `int;...`
accepts any Go int type (`int`, `int8`, `int16`, `int32`, `int64`, and
unsigned variants for base type checks).

#### slice

| Tag           | Meaning                                         |
|---------------|-------------------------------------------------|
| len=N         | Exact slice length                              |
| min=N         | Minimum slice length                            |
| max=N         | Maximum slice length                            |
| foreach=(...) | Element rules in parentheses, e.g. string rules |

Example: `slice;min=1;foreach=(string;min=2)`.

#### bool

| Tag  | Meaning              |
|------|----------------------|
| bool | Type check (boolean) |

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
