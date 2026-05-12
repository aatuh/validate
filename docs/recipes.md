# Recipes

These recipes show application-side adapters around `validate`. They are
examples of usage, not new package APIs.

## Validate an HTTP JSON request

Use the standard library to decode the request body, then pass the decoded
struct to `ValidateStructWithOpts`.

```go
type SignupRequest struct {
    Email    string `json:"email" validate:"string;required;email"`
    Password string `json:"password" validate:"string;required;min=12"`
}

func decodeAndValidate(r *http.Request) (SignupRequest, error) {
    var input SignupRequest
    if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
        return input, err
    }

    v := validate.New()
    err := v.ValidateStructWithOpts(input, validate.ValidateOpts{
        FieldNameFunc: validate.JSONFieldName,
    })
    return input, err
}
```

Keep transport choices in your application: status codes, response bodies,
request size limits, and unknown-field policy are not owned by `validate`.

## Use JSON field names

Pass `validate.JSONFieldName` when API clients should see JSON names instead
of Go field names.

```go
err := v.ValidateStructWithOpts(input, validate.ValidateOpts{
    FieldNameFunc: validate.JSONFieldName,
})
```

With `json:"email"` on `Email`, validation errors use path `email`.

## Validate nested structs and collections

Struct validation recurses into exported nested structs, slices, arrays, and
maps. Use collection tags when the collection itself or each element has rules.

```go
type LineItem struct {
    SKU string `json:"sku" validate:"string;required;slug"`
    Qty int    `json:"qty" validate:"int;min=1"`
}

type OrderRequest struct {
    Items []LineItem       `json:"items" validate:"slice;min=1"`
    Meta  map[string]int64 `json:"meta" validate:"map;keys=(string;slug);values=(int64;nonnegative)"`
}
```

Nested paths stay deterministic, for example `items[0].sku` or
`meta[count]`.

## Add a custom validator

Prefer per-instance registration for application rules so tests and packages
do not depend on process-wide registration order.

```go
import (
    "github.com/aatuh/validate/v3"
    verrs "github.com/aatuh/validate/v3/errors"
    "github.com/aatuh/validate/v3/types"
)

v := validate.New().WithRuleCompiler("even", func(_ *types.Compiler, _ validate.Rule) (func(any) error, error) {
    return func(value any) error {
        n, ok := value.(int)
        if !ok {
            return verrs.Errors{{Code: verrs.CodeIntType}}
        }
        if n%2 != 0 {
            return validate.Errors{{Code: "int.even"}}
        }
        return nil
    }, nil
})

err := v.CheckTag("int;even", 4)
```

Custom validators should return stable codes and avoid echoing submitted
values in `Msg`.

## Translate messages

Use translations for display text, but keep `Code`, `Path`, and `Param` as the
contract for program logic.

```go
tr := validate.NewSimpleTranslator(map[string]string{
    "string.min": "minimum length is %d",
})

v := validate.New().WithTranslator(tr)
```

Do not parse English messages in tests or clients. Assert stable codes instead.

## Return PII-safe validation responses

For public API responses, map structured validation errors to path, code, and
param. Avoid returning submitted values, raw regex patterns, secrets, tokens,
or custom validator internals.

```go
type InvalidParam struct {
    Name  string `json:"name"`
    Code  string `json:"code"`
    Param any    `json:"param,omitempty"`
}

func invalidParams(err error) []InvalidParam {
    var es validate.Errors
    if !errors.As(err, &es) {
        return nil
    }

    out := make([]InvalidParam, 0, len(es))
    for _, fe := range es {
        out = append(out, InvalidParam{
            Name:  fe.Path,
            Code:  fe.Code,
            Param: fe.Param,
        })
    }
    return out
}
```

See `examples/19_api_problem_response_test.go` for an executable RFC
7807-style mapping example.
