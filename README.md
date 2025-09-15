# validate

Composable validation helpers for Go with fluent builders, rule tags,
and struct validation. Optional message translation support.

## Install

```go
import "github.com/pureapi/pureapi-util/validate"
```

## Quick start

```go
v := validate.NewValidate(nil)

// Direct composition
check := v.WithString(v.MinLength(3), v.MaxLength(10))
if err := check("hello"); err != nil { /* handle */ }

// Builder style
rule := v.StringBuilder().
  WithMin(3).
  WithMax(10).
  WithEmail().
  Build()
_ = rule("user@example.com")
```

## Rules and builders

- **string**: `len`, `min`, `max`, `oneof`, `email`, `regex`
- **int/int64**: `min`, `max`
- **bool**: builder `MustBeTrue`, `MustBeFalse`
- **slice**: `len`, `min`, `max`, `ForEach(elemValidator)`

Builder entry points: `StringBuilder()`, `IntBuilder()`, `BoolBuilder()`,
`SliceBuilder()`.

## Struct tags

```go
type Input struct {
  Name  string `validate:"string;min=3;max=32"`
  Age   int    `validate:"int;min=0;max=130"`
  Emails []any `validate:"slice;min=1"`
}

v := validate.NewValidate(nil)
if err := v.ValidateStruct(Input{ Name: "Al", Age: 200 }); err != nil {
  // err includes per-field messages
}
```

## Translation

Provide a `Translator` with a `T(key, ...params)` method.

```go
type mapTranslator struct{ msgs map[string]string }
func (t mapTranslator) T(k string, p ...any) string {
  if m, ok := t.msgs[k]; ok { return fmt.Sprintf(m, p...) }
  return fmt.Sprintf(k, p...)
}

v := validate.NewValidate(nil)
msgs := validate.DefaultEnglishTranslations()
v.WithTranslator(mapTranslator{msgs: msgs})
```
