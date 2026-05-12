# Why use validate instead of go-playground/validator?

`validate` is intentionally narrow: it validates values and structs, then
returns structured failures. It does not bind HTTP requests, manage routes,
own response formats, or replace an application framework.

## Choose validate when

- You want structured errors with stable machine-readable codes, paths, and
  params instead of relying on English strings.
- You want fluent builders and tag parsing to share the same rule model.
- You need deterministic tag parsing for nested rules such as `foreach`,
  `keys`, and `values`.
- You want regex validation to use full-match behavior, invalid-pattern
  handling, and input length caps.
- You need optional translation while keeping code/path/param as the stable
  contract.
- You want custom rules and custom type validators that can return your own
  stable error codes.

## Choose go-playground/validator when

- You need the largest existing ecosystem and broad tag vocabulary.
- You want drop-in compatibility with projects or frameworks already built
  around `go-playground/validator`.
- You prefer its conventions for tag names, field errors, translations, and
  community integrations.

## Design boundary

`validate` should stay small enough to understand as a library. Recipes may
show adapters for HTTP, JSON, and API problem responses, but those adapters
belong in application code. The library should not grow middleware, request
binding, routing, persistence, auth, or framework-specific packages.
