# validate

Composable validation for Go with fluent builders, rule tags, struct
validation, structured errors, safe regex handling, caching, universal domain
format validators, custom rules, and optional message translation.

## Packages

- `github.com/aatuh/validate/v3`: main API (`New`, builders, `FromTag`, struct validation)
- `github.com/aatuh/validate/v3/core`: cache-aware validation engine
- `github.com/aatuh/validate/v3/glue`: builder and engine integration
- `github.com/aatuh/validate/v3/types`: rule AST, tag parser, compiler, and extension hooks
- `github.com/aatuh/validate/v3/errors`: structured errors and stable codes
- `github.com/aatuh/validate/v3/structvalidator`: reflection-based struct validation
- `github.com/aatuh/validate/v3/translator`: message translation helpers
- `github.com/aatuh/validate/v3/validators/...`: root and optional plugin validators

## Boundaries And Docs

`validate` is a validation library, not an API framework. It does not bind
HTTP requests, manage routes, own response formats, or replace application
transport code.

Further docs:

- [Why use validate?](docs/why-validate.md)
- [Recipes](docs/recipes.md)
- [Docs index](docs/README.md)

## Quick Start

```go
v := validate.New()

checkName := v.String().Required().MinRunes(3).MaxRunes(40).Build()
_ = checkName("gopher")

type User struct {
    Email string `validate:"string;required;email"`
    Age   int    `validate:"int;min=18"`
}

_ = v.ValidateStruct(User{Email: "user@example.com", Age: 30})
```

Run examples:

```bash
go test ./examples -v -count 1
```

## Supported Tags

Tags keep the existing v3 grammar: the first token is usually the base type,
followed by semicolon-separated rules. Existing tags remain supported.

Generic rules:

| Tag       | Meaning |
|-----------|---------|
| required  | Value must be non-zero/non-empty |
| omitempty | Skip validation for zero, nil, empty string, empty slice, or empty map |

Custom rule tags:

| Tag | Meaning |
|-----|---------|
| `ruleName` | Apply a registered custom compiler named `ruleName` |
| `custom:ruleName` | Apply a registered custom compiler explicitly |
| `custom:ruleName=raw` | Apply a registered custom compiler with `Args["value"] == "raw"` |

Bare custom rule names and `custom:` rules are supported after every base type.
Malformed built-in rule arguments, such as `int;min=bad`, still fail during
tag parsing. Unregistered custom rules fail during compilation or validation
with code `unknown`.

String rules:

| Tag | Meaning |
|-----|---------|
| len=N or length=N | Exact byte length |
| min=N / max=N | Minimum / maximum byte length |
| minRunes=N / maxRunes=N | Minimum / maximum Unicode rune count |
| nonempty | String must not be empty |
| oneof=a,b,c | Value must be one listed value |
| regex=PATTERN | Full-match regexp; anchors are added and input length is capped |
| contains=X / notContains=X | Required/prohibited substring |
| prefix=X / suffix=X | Required prefix/suffix |
| url / hostname | Absolute URL or hostname |
| ip / ipv4 / ipv6 / cidr | IP address or CIDR prefix |
| ascii / alpha / alnum | Character class checks |
| email / uuid / ulid | Built-in string plugins imported by the root package |
| slug / semver / json / jwt | Universal zero-dependency format validators |
| base64 / base64url / hex / mac | Encoding and identifier format validators |
| e164 / fqdn / date / rfc3339 / luhn | Phone, DNS, date/time, and checksum format validators |
| uuidv1 / uuidv3 / uuidv4 / uuidv5 / uuidv6 / uuidv7 / uuidv8 | Canonical UUID with version and RFC variant checks |

Number rules:

| Type | Tags |
|------|------|
| int / int64 | `min=N`, `max=N`, `gt=N`, `gte=N`, `lt=N`, `lte=N`, `between=A,B`, `positive`, `nonnegative` |
| float | `finite`, `min=N`, `max=N`, `gt=N`, `gte=N`, `lt=N`, `lte=N`, `between=A,B`, `positive`, `nonnegative` |

Collection and other rules:

| Type | Tags |
|------|------|
| bool | `true`, `false` |
| slice | `len=N`, `length=N`, `min=N`, `max=N`, `unique`, `contains=X`, `foreach=(...)` |
| array | `len=N`, `length=N`, `min=N`, `max=N`, `unique`, `contains=X`, `foreach=(...)` |
| map | `len=N`, `length=N`, `min=N`, `max=N`, `minKeys=N`, `maxKeys=N`, `keys=(...)`, `values=(...)` |
| time | `notzero`, `before=RFC3339`, `after=RFC3339`, `between=RFC3339,RFC3339` |

Examples:

```go
_ = v.CheckTag("slice;min=1;foreach=(string;minRunes=2)", []string{"go"})
_ = v.CheckTag("array;len=2;foreach=(string;slug)", [2]string{"api", "docs"})
_ = v.CheckTag("map;keys=(string;min=2);values=(int;positive)", map[string]int{"id": 1})
_ = v.CheckTag("time;after=2026-01-01T00:00:00Z", time.Now().UTC())
```

Domain validators are conservative format checks. They do not verify ownership,
deliverability, country-specific numbering plans, payment-card brands, JWT
signatures, JWT claims, DNS resolution, or registry authority.

```go
_ = v.String().SemVer().Build()("1.2.3")
_ = v.CheckTag("string;e164", "+358401234567")
jwtToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiIxMjMifQ.c2lnbmF0dXJl"
_ = v.CheckTag("slice;foreach=(string;jwt)", []string{jwtToken})
```

## Struct Validation

Struct validation uses `validate` tags, skips unexported fields, recurses into
nested structs, slices, arrays, and maps, and returns deterministic structured
errors.

```go
type Signup struct {
    Email    *string `json:"email" validate:"string;omitempty;email"`
    Password string  `json:"password" validate:"string;required;min=8"`
    Confirm  string  `json:"confirm" validate:"string;eqField=Password"`
    Token    string  `json:"token" validate:"string;requiredWith=Password"`
}

err := v.ValidateStructWithOpts(input, validate.ValidateOpts{
    FieldNameFunc: validate.JSONFieldName,
})
```

Struct-only cross-field rules:

| Tag | Meaning |
|-----|---------|
| eqField=FieldName | Value must equal another field on the same struct |
| neField=FieldName | Value must differ from another field on the same struct |
| requiredWith=FieldName | Value is required when the referenced field is non-zero |
| requiredIf=FieldName,value | Value is required when the referenced field equals value |
| requiredUnless=FieldName,value | Value is required unless the referenced field equals value |
| struct:ruleName or struct:ruleName=raw | Apply a registered struct-level rule to the current field |

Cross-field tags reference same-level Go field names. A missing or
inaccessible referenced field returns `field.reference` on the current field.
Conditional values are compared with exact string formatting and do not support
escaping commas in this version.

## Compile Options And Context

Existing validators are fail-fast by default. Opt in to collecting all rule
failures for a single value with `CompileOpts{CollectAll: true}`:

```go
err := v.CheckTagWithOpts("string;min=5;max=2", "abc", validate.CompileOpts{
    CollectAll: true,
})
```

`omitempty` still skips zero values. Failed requiredness rules, including
`required`, `requiredWith`, `requiredIf`, and `requiredUnless`, short-circuit
later same-field rules with only the requiredness code.

Context-aware APIs are additive. Built-in rules check cancellation before rule
execution; custom context compilers can read request-scoped context values:

```go
err := v.CheckTagContext(ctx, "string;required", value)
```

## Errors And Translation

Validation failures return `errors.Errors`, a stable slice of field errors:

```go
var es validate.Errors
if errors.As(err, &es) {
    _ = es.AsMap()
    _ = es.Filter("email")
}
```

Each field error contains:

- `Path`: deterministic field path
- `Code`: stable machine-readable code such as `string.min`, `required`, or `map.minkeys`
- `Param`: optional simple rule parameter
- `Msg`: translated human-readable message

Prefer `Code`, `Path`, and `Param` for program logic. Built-in validation
messages do not echo submitted values; invalid regex pattern diagnostics use a
capped/redacted pattern preview. Map keys in `Path` preserve short ordinary
keys such as `items[id]`; long keys, sensitive-looking keys, private-looking
keys, and keys that would need escaping are rendered as `[<redacted>]`.
Custom validators and translators control their own messages, so avoid
including secrets, tokens, or private caller data there.

`validate.New()` installs default English translations and root-level plugin
translations for email, UUID, and ULID. Use `validate.NewWithTranslator` or
`WithTranslator` to provide custom messages.

### Error Code Reference

`errors/codes.go` is the source of truth for stable built-in codes. Program
logic should use codes rather than English messages.

| Code | Related rule or condition |
|------|---------------------------|
| `unknown` | Unknown rules, malformed struct rule tags, or non-structured errors |
| `required` | `required` |
| `required.with` | `requiredWith` |
| `required.if` | `requiredIf` |
| `required.unless` | `requiredUnless` |
| `omitempty` | Informational skipped empty value |
| `field.eq` | `eqField` |
| `field.ne` | `neField` |
| `field.reference` | Missing or inaccessible referenced struct field |
| `string.type` | Expected string |
| `string.length` | `len` / `length` |
| `string.min` | `min` byte length |
| `string.max` | `max` byte length |
| `string.nonempty` | `nonempty` |
| `string.pattern` | Legacy pattern code |
| `string.oneof` | `oneof` |
| `string.prefix` | `prefix` |
| `string.suffix` | `suffix` |
| `string.contains` | `contains` |
| `string.notContains` | `notContains` |
| `string.url` | `url` |
| `string.hostname` | `hostname` |
| `string.ip` | `ip`, `ipv4`, or `ipv6` |
| `string.cidr` | `cidr` |
| `string.ascii` | `ascii` |
| `string.alpha` | `alpha` |
| `string.alnum` | `alnum` |
| `string.regex.invalidPattern` | Invalid `regex` pattern |
| `string.regex.inputTooLong` | Regex input length cap |
| `string.regex.noMatch` | Regex mismatch |
| `string.minRunes` | `minRunes` |
| `string.maxRunes` | `maxRunes` |
| `int.type` | Expected integer |
| `int64.type` | Expected exact `int64` |
| `number.type` | Expected number |
| `int.min` | Integer `min` |
| `int.max` | Integer `max` |
| `number.min` | Float/number `min` |
| `number.max` | Float/number `max` |
| `number.positive` | `positive` |
| `number.nonnegative` | `nonnegative` |
| `number.between` | `between` |
| `number.gt` | `gt` |
| `number.gte` | `gte` |
| `number.lt` | `lt` |
| `number.lte` | `lte` |
| `number.finite` | `finite` |
| `float.type` | Expected float |
| `slice.type` | Expected slice |
| `slice.length` | Slice `len` / `length` |
| `slice.min` | Slice `min` |
| `slice.max` | Slice `max` |
| `slice.forEach` | Element validation wrapper |
| `slice.unique` | `unique` |
| `slice.contains` | `contains` |
| `array.type` | Expected array |
| `array.length` | Array `len` / `length` |
| `array.min` | Array `min` |
| `array.max` | Array `max` |
| `array.forEach` | Element validation wrapper |
| `array.unique` | `unique` |
| `array.contains` | `contains` |
| `map.type` | Expected map |
| `map.length` | Map `len` / `length` |
| `map.minkeys` | `min` / `minKeys` |
| `map.maxkeys` | `max` / `maxKeys` |
| `map.keys` | `keys=(...)` |
| `map.values` | `values=(...)` |
| `bool.type` | Expected bool |
| `bool.true` | `true` |
| `bool.false` | `false` |
| `time.type` | Expected `time.Time` |
| `time.notzero` | `notzero` |
| `time.before` | `before` |
| `time.after` | `after` |
| `time.between` | `between` |
| `string.slug.invalid` | `slug` |
| `string.semver.invalid` | `semver` |
| `string.json.invalid` | `json` |
| `string.jwt.invalid` | `jwt` |
| `string.base64.invalid` | `base64` |
| `string.base64url.invalid` | `base64url` |
| `string.hex.invalid` | `hex` |
| `string.mac.invalid` | `mac` |
| `string.e164.invalid` | `e164` |
| `string.fqdn.invalid` | `fqdn` |
| `string.date.invalid` | `date` |
| `string.rfc3339.invalid` | `rfc3339` |
| `string.luhn.invalid` | `luhn` |
| `string.uuid.version` | `uuidv1`, `uuidv3`, `uuidv4`, `uuidv5`, `uuidv6`, `uuidv7`, or `uuidv8` version/variant mismatch |

## Extensibility

Per-instance rule compilers work from tags, manual rules, and builder escape
hatches:

```go
v := validate.New().
    WithRuleCompiler("even", func(c *types.Compiler, rule types.Rule) (func(any) error, error) {
        return func(value any) error {
            n, ok := value.(int)
            if !ok || n%2 != 0 {
                return validate.Errors{{Code: "number.even", Msg: c.T("number.even", "must be even", nil)}}
            }
            return nil
        }, nil
    }).
    WithRuleCompiler("mod", func(c *types.Compiler, rule types.Rule) (func(any) error, error) {
        raw, _ := rule.Args["value"].(string)
        return func(value any) error {
            // parse raw and validate value
            _ = raw
            return nil
        }, nil
    })

_ = v.CheckTag("int;even", 2)
_ = v.CheckTag("int;custom:mod=2", 4)
_ = v.Int().Rule("even", nil).Build()(2)
```

Use `WithContextRuleCompiler` when a custom rule must observe cancellation or
request-scoped context values. Existing `WithRuleCompiler` rules continue to
work through context-aware APIs by ignoring the context.

Compile-error-aware callers can use `CompileRulesE`:

```go
check, err := v.CompileRulesE([]validate.Rule{
    validate.NewRule(validate.KInt, nil),
    validate.NewRule("even", nil),
})
_ = check
_ = err
```

Struct-level custom rules can inspect the current field and same-level fields:

```go
v := validate.New().WithStructRuleCompiler("matchesField", func(rule validate.Rule) (validate.StructRuleFunc, error) {
    fieldName, _ := rule.Args["value"].(string)
    return func(ctx validate.StructRuleContext) error {
        other, _ := ctx.FieldValue(fieldName)
        if ctx.Value != other {
            return validate.Errors{{Code: "field.matches", Msg: "must match"}}
        }
        return nil
    }, nil
})
```

Global plugin rule:

```go
func init() {
    types.RegisterRule("countryCode", compileCountryCode)
    translator.RegisterDefaultEnglishTranslations(map[string]string{
        "string.country.invalid": "invalid country code",
    })
}
```

Custom domain validators should return stable codes, avoid echoing submitted
values in messages, and document whether they are syntax-only or authoritative
business checks. The `custom:name=value` tag grammar passes a single raw string
argument in `rule.Args["value"]`; use a custom parser inside the compiler when
you need more structure.

Global rule, type, and translation registration is process-wide and intended
primarily for plugins. Duplicate names overwrite earlier registrations. For
application code and tests, prefer `WithRuleCompiler`, `WithContextRuleCompiler`,
`WithStructRuleCompiler`, `WithTypeValidator`, and `WithTranslator`.

Custom types can be registered per validator with `WithTypeValidator`, then
used with `v.CustomType("name")`, a matching tag on that validator, or nested
collection tags such as `slice;foreach=(name)`, `map;keys=(name)`, and
`map;values=(name)`. Global plugin-style registration with
`types.RegisterGlobalType` is still supported, but it is process-wide:
registration order matters, and duplicate names overwrite earlier factories. A
per-instance type registration takes precedence over a global type with the
same name.

`CompileRules` caches deterministic AST rules. Rules containing function
arguments skip the cache, including nested function arguments. Opaque custom
argument values are serialized with their string form for cache keys, so custom
rule authors should prefer simple stable arguments or use function arguments
when identity or mutable state matters.

The root package includes universal, zero-dependency format validators.
Regional, authoritative, or dependency-heavy validators such as postal-code
databases, national ID rules, phone-number metadata, currency registries, cron
interpreters, and JSON Schema export should live in custom validators or
optional plugin packs.

## Quality Gate

Use the Makefile-backed gate locally and in CI:

```bash
make finalize
```

`make ci` is the same full gate used by GitHub Actions. It runs module tidy
checks, `go vet ./...`, `go test ./...`, executable examples,
`govulncheck ./...`, race tests with atomic coverage, coverage summary, and
the parser/compiler fuzz smoke script.

Useful focused targets:

| Target | Purpose |
|--------|---------|
| `make test` | Run `go test ./...`; override with `PKG=./types` |
| `make examples` | Run executable examples with `-v -count 1` |
| `make race-cover` | Run race tests with `coverage.out` |
| `make coverage` | Run race coverage and print the coverage summary |
| `make fuzz` | Run `scripts/fuzz.sh` |
| `make vuln` | Install and run `govulncheck` |
