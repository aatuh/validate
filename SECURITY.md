# Security Policy

## Supported Versions

`validate` follows the active v3 line. The `go.mod` language compatibility
floor is `go 1.23`, but release CI and `govulncheck` run on a patched Go
toolchain. Current CI uses Go `1.25.10`.

Consumers that use `validate` in security-sensitive services should run a
patched Go toolchain for their supported Go line. Standard-library
vulnerabilities are fixed by the Go toolchain, not by this module.

## Reporting Vulnerabilities

Prefer a private GitHub Security Advisory for exploitable issues. If private
advisories are unavailable, contact the repository owner before opening a
public issue with exploit details.

Useful reports include:

- affected version or commit
- minimal reproducer
- expected and observed behavior
- impact, especially panic, denial of service, data exposure, or unsafe error
  output
- whether the issue involves custom validators, translators, tags, regexes,
  struct tags, or map keys

Avoid including real secrets, tokens, private data, or production payloads in
reports. Use synthetic examples.

## Security Boundaries

`validate` is a validation library. It does not provide HTTP routing, request
binding, authentication, authorization, persistence, or framework middleware.

Security-sensitive library behavior includes:

- no panics for malformed tags or unexpected input types
- bounded and redacted regex diagnostics
- bounded and privacy-aware map-key path rendering
- stable structured error codes for program logic
- no default messages that echo submitted values
- race-safe copied validators, registries, and translation maps

Custom validators and translators are caller-controlled code. They should
return stable codes and avoid echoing secrets or personal data in messages.
