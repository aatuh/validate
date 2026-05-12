# Maturity Criteria

`validate` is mature when v3 is boring to adopt and maintain: contracts are
clear, behavior is deterministic, checks are green, and new work is mostly
compatibility, docs, and narrow validator maintenance.

## Stable v3 Guarantees

- The root package stays a thin facade over package-owned behavior.
- Existing public APIs, tag grammar, rule names, and stable error codes remain
  backward compatible unless a future major version explicitly changes them.
- Structured errors use `Path`, `Code`, `Param`, and `Msg`; program logic should
  depend on `Path`, `Code`, and `Param`, not English messages.
- Builders, tags, manual AST rules, struct validation, custom rules, custom
  types, and translation stay aligned where they expose the same behavior.
- Regex validation keeps full-match semantics, invalid-pattern handling, input
  length caps, and sanitized diagnostics.
- Map key paths preserve short ordinary keys and redact long, sensitive-looking,
  private-looking, or escaping-sensitive keys.

## Intentionally Out Of Scope

`validate` should not grow into an API framework. The following belong in
application code, optional adapters, or separate packages:

- HTTP middleware, request binding, routing, and response writing
- persistence, migrations, repositories, or schema export
- authentication, authorization, sessions, or CSRF handling
- framework-specific packages
- authoritative business checks such as deliverability, ownership, DNS lookup,
  national ID databases, phone metadata, or payment-card brand databases

## Support Policy

The module keeps `go 1.23` in `go.mod` as the language compatibility floor.
Release CI and vulnerability scanning use a patched Go toolchain, currently Go
`1.25.10`. Consumers should run patched Go versions for security-sensitive
deployments because standard-library vulnerabilities are fixed by the Go
toolchain.

## Done Criteria

The active maturity pass is complete when:

- `make ci` passes locally and in GitHub Actions.
- public contracts are documented in README and `docs/`.
- all built-in error codes in `errors/codes.go` are represented in
  `docs/error-codes.md`.
- direct tests cover map-key path policy and cache serialization policy.
- public builder behavior exposed through the root facade has representative
  success/failure coverage.
- benchmark baselines exist without making performance claims.
