# Error Codes

`errors/codes.go` is the source of truth for built-in stable codes. Program
logic should use `Code`, `Path`, and `Param`; English messages are display text
and may be translated.

Path values may include struct fields, JSON field names when
`validate.JSONFieldName` is configured, slice/array indexes such as `[0]`, and
map key segments such as `[id]` or `[<redacted>]`.

## Built-In Codes

| Code | Rule or condition | Param | Path notes |
|------|-------------------|-------|------------|
| `unknown` | unknown rule, malformed struct rule, or non-structured error | none | any path |
| `required` | `required` | none | any field/value |
| `required.with` | `requiredWith` | none | struct fields |
| `required.if` | `requiredIf` | none | struct fields |
| `required.unless` | `requiredUnless` | none | struct fields |
| `omitempty` | skipped empty value | none | informational |
| `field.eq` | `eqField` | none | struct fields |
| `field.ne` | `neField` | none | struct fields |
| `field.reference` | missing or inaccessible referenced field | field name | struct fields |
| `string.type` | expected string | none | any path |
| `string.length` | `len` / `length` | expected length | any path |
| `string.min` | `min` byte length | minimum length | any path |
| `string.max` | `max` byte length | maximum length | any path |
| `string.nonempty` | `nonempty` | none | any path |
| `string.pattern` | legacy pattern code | pattern | any path |
| `string.oneof` | `oneof` | allowed values | any path |
| `string.prefix` | `prefix` | prefix | any path |
| `string.suffix` | `suffix` | suffix | any path |
| `string.contains` | `contains` | required substring | any path |
| `string.notContains` | `notContains` | prohibited substring | any path |
| `string.url` | `url` | none | any path |
| `string.hostname` | `hostname` | none | any path |
| `string.ip` | `ip`, `ipv4`, or `ipv6` | none | any path |
| `string.cidr` | `cidr` | none | any path |
| `string.ascii` | `ascii` | none | any path |
| `string.alpha` | `alpha` | none | any path |
| `string.alnum` | `alnum` | none | any path |
| `string.regex.invalidPattern` | invalid `regex` pattern | sanitized pattern preview | any path |
| `string.regex.inputTooLong` | regex input length cap | limit | any path |
| `string.regex.noMatch` | regex mismatch | none | any path |
| `string.minRunes` | `minRunes` | minimum rune count | any path |
| `string.maxRunes` | `maxRunes` | maximum rune count | any path |
| `string.slug.invalid` | `slug` | none | any path |
| `string.semver.invalid` | `semver` | none | any path |
| `string.json.invalid` | `json` | none | any path |
| `string.jwt.invalid` | `jwt` | none | any path |
| `string.base64.invalid` | `base64` | none | any path |
| `string.base64url.invalid` | `base64url` | none | any path |
| `string.hex.invalid` | `hex` | none | any path |
| `string.mac.invalid` | `mac` | none | any path |
| `string.e164.invalid` | `e164` | none | any path |
| `string.fqdn.invalid` | `fqdn` | none | any path |
| `string.date.invalid` | `date` | none | any path |
| `string.rfc3339.invalid` | `rfc3339` | none | any path |
| `string.luhn.invalid` | `luhn` | none | any path |
| `string.uuid.version` | UUID version-specific rules | expected version | any path |
| `int.type` | expected integer | none | any path |
| `int64.type` | expected exact `int64` | none | any path |
| `number.type` | expected number | none | any path |
| `int.min` | integer `min` | minimum value | any path |
| `int.max` | integer `max` | maximum value | any path |
| `number.min` | float/number `min` | minimum value | any path |
| `number.max` | float/number `max` | maximum value | any path |
| `number.positive` | `positive` | none | any path |
| `number.nonnegative` | `nonnegative` | none | any path |
| `number.between` | `between` | min/max values | any path |
| `number.gt` | `gt` | threshold | any path |
| `number.gte` | `gte` | threshold | any path |
| `number.lt` | `lt` | threshold | any path |
| `number.lte` | `lte` | threshold | any path |
| `number.finite` | `finite` | none | any path |
| `float.type` | expected float | none | any path |
| `slice.type` | expected slice | none | any path |
| `slice.length` | slice `len` / `length` | expected length | collection path |
| `slice.min` | slice `min` | minimum length | collection path |
| `slice.max` | slice `max` | maximum length | collection path |
| `slice.forEach` | element validation failed | none | may include `[index]` |
| `slice.unique` | `unique` | none | collection path |
| `slice.contains` | `contains` | required element | collection path |
| `array.type` | expected array | none | any path |
| `array.length` | array `len` / `length` | expected length | collection path |
| `array.min` | array `min` | minimum length | collection path |
| `array.max` | array `max` | maximum length | collection path |
| `array.forEach` | element validation failed | none | may include `[index]` |
| `array.unique` | `unique` | none | collection path |
| `array.contains` | `contains` | required element | collection path |
| `map.type` | expected map | none | any path |
| `map.length` | map `len` / `length` | expected key count | map path |
| `map.minkeys` | `minKeys` | minimum key count | map path |
| `map.maxkeys` | `maxKeys` | maximum key count | map path |
| `map.keys` | map key validation failed | none | may include key segment |
| `map.values` | map value validation failed | none | may include key segment |
| `bool.type` | expected boolean | none | any path |
| `bool.true` | `true` | none | any path |
| `bool.false` | `false` | none | any path |
| `time.type` | expected `time.Time` | none | any path |
| `time.notzero` | `notzero` | none | any path |
| `time.before` | `before` | timestamp | any path |
| `time.after` | `after` | timestamp | any path |
| `time.between` | `between` | start/end timestamps | any path |

## Root-Imported Plugin Codes

The root package blank-imports the domain, email, ULID, and UUID plugins.
Domain and UUID version codes above are shared with `errors/codes.go`.
Additional root-imported plugin code families are:

| Code | Rule or condition | Notes |
|------|-------------------|-------|
| `string.email.invalid` | `email` | conservative bare-address syntax check |
| `string.ulid.invalid` | `ulid` | canonical ULID syntax check |
| `string.uuid.invalid` | `uuid` | canonical UUID syntax check |

Plugin validators also use `string.type` when the input is not a string.
