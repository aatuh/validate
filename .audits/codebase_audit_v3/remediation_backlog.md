# Backlog

Project: github.com/aatuh/validate/v3

Status legend:

- [ ] not done
- [x] done

## Epic E1 - Public Contracts And Documentation [ ]

Description: Make stable public behavior easier for downstream users to rely on, especially error codes, cache behavior, global registrations, and currently unsupported extension modes.

### Ticket E1-T1 - Add A Complete Error-Code Reference [ ]

Description: Document every built-in error code from `errors/codes.go`, including the related tag/rule, parameter shape, and whether the path may include collection segments.

Implementation rules:

- implement the ticket in the smallest sensible step
- keep the reference synchronized with `errors/codes.go`
- prefer README or package docs; add generation only if it stays simple and maintainable
- run `go test ./... -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Avoid asserting English message text as the compatibility contract.

### Ticket E1-T2 - Document Cache Semantics For Custom Args [ ]

Description: Clarify how `CompileRules` caches AST rules, when function arguments skip caching, and what callers should avoid with opaque pointer-like custom rule arguments.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve current cache behavior unless a later implementation ticket changes it deliberately
- include at least one concise example or warning for custom rule authors
- run `go test ./examples -v -count=1` if examples change
- run `go test ./... -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Keep the default behavior backward compatible.

### Ticket E1-T3 - Clarify Global Registry Scope [ ]

Description: Strengthen docs around process-wide rule, type, and translation registration so users prefer per-instance APIs unless they are intentionally writing plugins.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve global registration APIs
- mention duplicate-name overwrite behavior and test/plugin collision risk
- run `go test ./examples -v -count=1` if examples change
- run `go test ./... -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Do not remove global APIs; they are public compatibility surface.

### Ticket E1-T4 - State Current Context And Error Aggregation Scope [ ]

Description: Document that context-aware validation is currently adapter-only and single-value validators are fail-fast by default.

Implementation rules:

- implement the ticket in the smallest sensible step
- describe current behavior without promising future APIs
- keep wording concise and aligned with existing examples
- run `go test ./examples -v -count=1` if examples change
- run `go test ./... -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Preserve current fail-fast defaults.

## Epic E2 - Existing Behavior Test Hardening [ ]

Description: Add focused regression coverage for public behavior that already exists but has thin direct tests.

### Ticket E2-T1 - Add Direct Map Key Formatter Tests [ ]

Description: Add tests for `internal/pathutil` covering short string keys, long keys, sensitive-looking keys, nil, booleans, numbers, non-string complex keys, and escaping-sensitive strings.

Implementation rules:

- implement the ticket in the smallest sensible step
- assert deterministic formatted segments rather than English messages
- preserve short ordinary string keys and numeric/bool readability
- run `gofmt` on modified Go files
- run `go test ./internal/pathutil ./types ./structvalidator -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Keep this as policy coverage; avoid changing behavior in the test ticket.

### Ticket E2-T2 - Expand Builder Method Contract Tests [ ]

Description: Cover public builder methods that are exposed through the root facade but currently have little or no direct test coverage.

Implementation rules:

- implement the ticket in the smallest sensible step
- prioritize methods with 0% coverage from the latest coverage output
- assert stable error codes and success/failure behavior
- run `gofmt` on modified Go files
- run `go test ./glue ./validators ./... -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Keep tests table-driven and avoid duplicating the entire compiler suite.

### Ticket E2-T3 - Add Cache Fallback Regression Tests [ ]

Description: Add tests that pin current cache behavior for function args, nested function args, `Rule.Elem`, and unsupported non-function custom args.

Implementation rules:

- implement the ticket in the smallest sensible step
- do not change production cache behavior in this ticket
- assert deterministic cache keys where supported and skipped caching for function args
- run `gofmt` on modified Go files
- run `go test ./core -count=1`
- run `go test ./... -race -count=1` if shared cache behavior is touched
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- This ticket should make future cache hardening safer.

## Epic E3 - Conditional Validation [ ]

Description: Add the next conditional validation capability in a backward-compatible way with clear parser, struct, builder, manual-rule, docs, and test ownership.

### Ticket E3-T1 - Choose The First Additive Conditional Rule [ ]

Description: Decide the first rule beyond `requiredWith`, including tag syntax, construction paths, error code, missing-reference behavior, and compatibility expectations.

Implementation rules:

- implement the ticket in the smallest sensible step
- prefer a same-level struct-field rule unless a compiler-level design is explicitly justified
- preserve existing `eqField`, `neField`, and `requiredWith` behavior
- record the decision in README, package docs, or this backlog before implementation
- run `go test ./... -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Avoid designing an open-ended conditional language in the first ticket.

### Ticket E3-T2 - Add Conditional Rule Failing Tests [ ]

Description: Add tests for the selected conditional rule before production implementation, covering valid, invalid, zero-value, missing-reference, and malformed-tag cases.

Implementation rules:

- implement the ticket in the smallest sensible step
- assert stable error codes, paths, and params rather than English text
- include parser and struct-validator coverage
- include builder or manual-rule coverage only if the selected design exposes those paths
- run `gofmt` on modified Go files
- run `go test ./types ./structvalidator ./glue -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- This ticket should fail until E3-T3 is implemented.

### Ticket E3-T3 - Implement The Conditional Rule [ ]

Description: Implement the selected conditional rule across its approved construction paths and update docs/examples.

Implementation rules:

- implement the ticket in the smallest sensible step
- keep parser, struct validation, builders, manual rules, docs, and examples aligned
- preserve existing defaults and public APIs
- add or document any new error code
- run `gofmt` on modified Go files
- run `go test ./types ./structvalidator ./glue ./examples -count=1`
- run `scripts/fuzz.sh` because parser and tag behavior change
- run `go test ./... -race -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Keep the implementation additive; do not change `requiredWith` semantics.

## Epic E4 - Error Aggregation [ ]

Description: Provide an opt-in way to collect multiple failures for one value or field while preserving current fail-fast defaults.

### Ticket E4-T1 - Design All-Errors-Per-Field API [ ]

Description: Decide whether aggregation belongs in compiler options, `ValidateOpts`, a separate compile API, or a builder option.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve default fail-fast behavior for existing APIs
- define ordering, path handling, `omitempty`, `required`, and nested collection semantics
- record the decision in README, package docs, or this backlog before implementation
- run `go test ./... -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Treat this as an additive public contract, not an internal refactor.

### Ticket E4-T2 - Add Aggregation Failing Tests [ ]

Description: Add tests for multiple failures on a single value or struct field using the chosen API.

Implementation rules:

- implement the ticket in the smallest sensible step
- cover success, first failure compatibility, multiple failures, `omitempty`, `required`, and nested collection paths
- assert structured codes and paths
- run `gofmt` on modified Go files
- run `go test ./types ./core ./glue ./structvalidator -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- This ticket should fail until E4-T3 is implemented.

### Ticket E4-T3 - Implement Aggregation Opt-In [ ]

Description: Implement the approved all-errors-per-field API and update docs/examples.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve current fail-fast defaults
- keep aggregation deterministic
- update README or examples for the new opt-in behavior
- run `gofmt` on modified Go files
- run `go test ./types ./core ./glue ./structvalidator ./examples -count=1`
- run `go test ./... -race -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Avoid changing `CompileE` behavior unless the selected API explicitly does so under opt-in.

## Epic E5 - Context-Aware Validation [ ]

Description: Decide and implement the smallest useful context-aware validation story without forcing context into simple callers.

### Ticket E5-T1 - Decide Context Scope [ ]

Description: Decide whether context-aware validation remains adapter-only or gains first-class compile, builder, and struct APIs.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve existing `CheckFunc`, `CheckFuncCtx`, `WithContext`, and `WithoutContext`
- define cancellation behavior and nil-context handling if first-class APIs are chosen
- record the decision in README, package docs, or this backlog before implementation
- run `go test ./... -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Avoid adding context parameters to existing APIs in a breaking way.

### Ticket E5-T2 - Add Context API Tests Or Documentation Tests [ ]

Description: Add tests for the chosen context behavior, or executable documentation if the decision is to keep context adapter-only.

Implementation rules:

- implement the ticket in the smallest sensible step
- cover cancellation or adapter behavior according to the selected design
- avoid tests that depend on timing-sensitive sleeps
- run `gofmt` on modified Go files
- run `go test ./core ./glue ./structvalidator ./examples -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- This ticket should define observable behavior before any implementation ticket.

### Ticket E5-T3 - Implement Context-Aware APIs If Approved [ ]

Description: Implement the smallest approved first-class context-aware validation API, or close the epic after documenting adapter-only scope.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve all existing context-free APIs
- keep context handling deterministic and race-safe
- update README or examples for the approved behavior
- run `gofmt` on modified Go files
- run `go test ./core ./glue ./structvalidator ./examples -count=1`
- run `go test ./... -race -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- If adapter-only scope is retained, this ticket can be completed with docs/tests only.

## Epic E6 - Global Registry And Cache Hardening [ ]

Description: Reduce collision and determinism risk around process-wide registries and compiled validator caches without breaking existing users.

### Ticket E6-T1 - Add Global Registry Collision Tests [ ]

Description: Add tests that make overwrite behavior and per-instance precedence explicit for global rule compilers, global type validators, and default translations.

Implementation rules:

- implement the ticket in the smallest sensible step
- avoid tests that require resetting global state
- use unique names to prevent pollution across packages
- run `gofmt` on modified Go files
- run `go test ./types ./core ./translator ./glue -count=1`
- run `go test ./... -race -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Do not introduce unregister/reset APIs unless explicitly approved by a later design.

### Ticket E6-T2 - Decide Cache Handling For Unsupported Args [ ]

Description: Decide whether unsupported non-function custom rule arguments should keep the current `fmt.Sprintf` fallback, skip caching, or require an explicit stable-key interface.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve built-in rule cache behavior
- assess backward compatibility for custom rule authors
- record the decision in README, package docs, or this backlog before implementation
- run `go test ./... -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Keep cache keys deterministic and safe under concurrent use.

### Ticket E6-T3 - Implement Approved Cache Hardening [ ]

Description: Implement the approved cache behavior for unsupported custom args and add regression tests.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve current behavior for built-in rules and function-arg cache skipping
- add tests for deterministic keys, skipped caching, and custom arg behavior
- run `gofmt` on modified Go files
- run `go test ./core ./types ./glue -count=1`
- run `go test ./... -race -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create or require git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Run `scripts/fuzz.sh` if parser, compiler, builder, or tag handling changes as part of the solution.
