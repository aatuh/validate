# Backlog

Project: github.com/aatuh/validate/v3

Status legend:

- [ ] not done
- [x] done

## Epic E1 - Cache Determinism And Manual AST Safety [x]

Description: Fix correctness risks in compiled validator caching and public manual rule construction before adding larger behavior changes.

### Ticket E1-T1 - Add Rule.Elem Cache Regression Tests [x]

Description: Add failing tests proving that two `KForEach` rules with different `Rule.Elem` values do not share the wrong cached validator.

Implementation rules:

- implement the ticket in the smallest sensible step
- do not change production code in this test-only ticket except for required test helpers
- run `gofmt` on modified Go files
- run `go test ./core ./types -count=1`
- run `go test ./... -race` if cache or concurrency behavior is touched outside tests
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Cover both `types.NewRuleWithElem` and `types.NewRuleWithElemValue` compatibility constructors.

### Ticket E1-T2 - Include Rule.Elem In Cache Serialization [x]

Description: Update deterministic rule serialization and function-argument detection so `Rule.Elem` participates in cache keys and function-arg cache skipping.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve existing serialized forms where possible except where `Elem` must be represented
- run `gofmt` on modified Go files
- run `go test ./core ./types -count=1`
- run `go test ./... -race -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Keep serialization canonical and stable for maps, slices, `time.Time`, nested rules, and nil elements.

### Ticket E1-T3 - Add Manual AST Panic-Resistance Tests [x]

Description: Add tests for malformed or partial public manual rules, especially slice length, min/max length, and `forEach` rules applied to nil or wrong-type inputs without a base `KSlice` rule.

Implementation rules:

- implement the ticket in the smallest sensible step
- assert structured errors or returned errors, not panics
- run `gofmt` on modified Go files
- run `go test ./types ./core -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Public `CompileRules` and `CompileRulesE` are the relevant APIs.

### Ticket E1-T4 - Harden Manual Slice Rule Validators [x]

Description: Make slice length, min/max length, and `forEach` validators return stable structured type errors for nil or non-slice values even when used without a base `KSlice` rule.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve successful behavior for existing tag and builder paths
- run `gofmt` on modified Go files
- run `go test ./types ./core ./glue -count=1`
- run `go test ./... -race -count=1` because public compiler behavior changes
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Use existing `slice.type` error code semantics.

## Epic E2 - Regex And Error Trust Boundaries [x]

Description: Prevent untrusted caller inputs from creating oversized or sensitive error output while keeping stable error codes.

### Ticket E2-T1 - Add Invalid Regex Redaction Tests [x]

Description: Add tests for long invalid regex patterns and sensitive-looking pattern strings to prove error messages are capped and codes remain stable.

Implementation rules:

- implement the ticket in the smallest sensible step
- assert stable codes such as `string.regex.invalidPattern`; avoid depending on full English text
- run `gofmt` on modified Go files
- run `go test ./types ./glue -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Cover builder and tag paths if both can expose the message.

### Ticket E2-T2 - Cap Invalid Regex Pattern Messages [x]

Description: Route invalid regex pattern formatting through one truncation/redaction helper so translated messages cannot echo full caller patterns.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve `string.regex.invalidPattern` as the error code
- avoid exposing raw long inputs, secrets, tokens, or private caller data in messages
- run `gofmt` on modified Go files
- run `go test ./types ./glue -count=1`
- run `scripts/fuzz.sh` because regex and tag handling are affected
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Keep truncation deterministic and documented by tests.

### Ticket E2-T3 - Audit Structured Error Output For Raw Caller Data [x]

Description: Review validation errors, translator defaults, examples, and README snippets for raw caller data exposure and add focused tests or docs where behavior is intentional.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve existing stable error codes
- prefer structured code assertions over English string assertions
- run `gofmt` on modified Go files if tests or examples change
- run `go test ./errors ./translator ./examples ./... -count=1` when public examples or docs are changed
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- This is a security/privacy hardening pass for a library, not an HTTP/API audit.

## Epic E3 - Struct Validation Contract Clarity [x]

Description: Make cross-field and aggregation behavior predictable without breaking existing valid struct tags.

### Ticket E3-T1 - Add Missing Cross-Field Reference Tests [x]

Description: Add tests for `eqField`, `neField`, `requiredWith`, and custom struct rules when the referenced field is missing, unexported, or inaccessible.

Implementation rules:

- implement the ticket in the smallest sensible step
- assert stable structured codes and paths rather than full English text
- run `gofmt` on modified Go files
- run `go test ./structvalidator ./core -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Include pointer owner structs and JSON field-name options where relevant.

### Ticket E3-T2 - Return Explicit Errors For Invalid Cross-Field References [x]

Description: Change struct cross-field rule execution so missing or inaccessible referenced fields produce deterministic structured configuration errors instead of silent comparisons.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve behavior for valid `eqField`, `neField`, and `requiredWith` tags
- avoid changing existing public error codes unless compatibility is explicitly justified in docs and tests
- run `gofmt` on modified Go files
- run `go test ./structvalidator ./core ./examples -count=1`
- run `go test ./... -race -count=1` because reflection behavior changes
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- If a new error code is needed, update README error-code documentation in the same ticket.

### Ticket E3-T3 - Design All-Errors-Per-Field Behavior [x]

Description: Define an opt-in behavior for collecting all failures on a single field while preserving current fail-fast defaults.

Implementation rules:

- implement the ticket in the smallest sensible step
- produce a short design note in the backlog, README, or package docs before implementation if API changes are required
- preserve existing default behavior and public APIs unless an additive option is introduced
- run documentation checks through `go test ./examples -v -count=1` if examples change
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- This ticket should decide whether the option belongs in `core.ValidateOpts`, compiler options, or a separate API.
- Design note from remediation v1: preserve the current fail-fast default for
  compiled field validators. The future opt-in should be additive and rooted in
  compiler behavior, then surfaced through struct validation with a
  `core.ValidateOpts` flag such as `CollectAllFieldErrors`. Single-value builder
  APIs should either remain fail-fast or gain explicit all-error variants so
  existing `Build`, `CompileRules`, and `FromTag` behavior does not change.

## Epic E4 - Custom Type Isolation And Extensibility [x]

Description: Reduce global-state risks while preserving plugin compatibility for custom rules and custom types.

### Ticket E4-T1 - Document Current Global Custom Type Semantics [x]

Description: Clarify in README or package docs that `RegisterGlobalType` is process-wide, registration order matters, and duplicate names overwrite existing factories.

Implementation rules:

- implement the ticket in the smallest sensible step
- keep docs concise and aligned with actual code
- run `go test ./examples -v -count=1` if examples change
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- This can land before any API additions to reduce immediate user confusion.

### Ticket E4-T2 - Add Custom Type Collision And Isolation Tests [x]

Description: Add tests that capture current global custom type overwrite behavior and verify per-instance rule compilers remain isolated.

Implementation rules:

- implement the ticket in the smallest sensible step
- avoid tests that depend on global registry cleanup unless cleanup support is added first
- use unique test names to prevent cross-test pollution
- run `gofmt` on modified Go files
- run `go test ./types ./core ./glue -count=1`
- run `go test ./... -race -count=1` because global registries are involved
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Use generated unique names in tests to avoid polluting unrelated packages.

### Ticket E4-T3 - Add Per-Instance Custom Type Registration [x]

Description: Introduce an additive per-instance custom type registration path so validators can isolate custom type factories while still falling back to global plugin registrations.

Implementation rules:

- implement the ticket in the smallest sensible step
- keep `RegisterGlobalType` backward compatible
- avoid breaking existing tag parsing for globally registered plugin types
- run `gofmt` on modified Go files
- run `go test ./types ./core ./glue ./validators/uuid -count=1`
- run `go test ./... -race -count=1`
- run `scripts/fuzz.sh` if tag parsing or compiler resolution changes
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Prefer additive APIs such as `WithTypeCompiler` or `WithTypeValidator` over changing global registration behavior.

## Epic E5 - Context And Conditional Validation Roadmap [ ]

Description: Close extensibility gaps that are useful for real applications while keeping the default library surface stable.

### Ticket E5-T1 - Define Context-Aware Validation Scope [x]

Description: Decide whether context should remain an adapter-only helper or become part of builders, compiler output, and custom validator APIs.

Implementation rules:

- implement the ticket in the smallest sensible step
- document the chosen scope before adding new public APIs
- preserve existing `func(any) error` APIs
- run `go test ./core -count=1` if only docs and context helpers change
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Context support should be additive and optional.
- Design note from remediation v1: keep context support adapter-only for the
  current public `func(any) error` APIs. Future context-aware validation should
  be additive through explicit context-bearing APIs or wrappers, and should not
  alter `Build`, `FromTag`, `CompileRules`, or struct-validation defaults.

### Ticket E5-T2 - Add Conditional Validation Design And Tests [ ]

Description: Identify the smallest additive conditional validation rules beyond `requiredWith`, then add failing tests for tag, builder, and struct paths before implementation.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve existing tag grammar and avoid ad hoc splitting that breaks nested parentheses
- run `gofmt` on modified Go files
- run `go test ./types ./structvalidator ./glue -count=1`
- run `scripts/fuzz.sh` if parser or tag splitting behavior changes
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Keep the first increment small, for example one well-specified conditional presence rule.

### Ticket E5-T3 - Implement The First Additive Conditional Rule [ ]

Description: Implement the selected conditional validation rule across parser, compiler or structvalidator, builder escape hatch if needed, README, and executable examples.

Implementation rules:

- implement the ticket in the smallest sensible step
- keep parser, builder, and manual rule behavior aligned
- preserve existing public APIs and error codes unless an additive code is documented
- run `gofmt` on modified Go files
- run `go test ./types ./structvalidator ./glue ./examples -count=1`
- run `go test ./... -race -count=1`
- run `scripts/fuzz.sh` because parser/tag behavior changes
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Update README supported tags and stable error-code documentation if new public behavior is added.

## Epic E6 - Test Quality And Developer Experience [x]

Description: Improve confidence and reduce ambiguity in public APIs after correctness and trust-boundary fixes land.

### Ticket E6-T1 - Expand Tag Versus Builder Equivalence Coverage [x]

Description: Add table-driven equivalence tests for expanded string, number, bool, slice, map, time, custom rule, and custom type paths.

Implementation rules:

- implement the ticket in the smallest sensible step
- assert structured error codes where possible
- run `gofmt` on modified Go files
- run `go test ./glue ./types ./core -count=1`
- run `scripts/fuzz.sh` if parser or builder behavior changes
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Avoid testing translated English as the primary contract.

### Ticket E6-T2 - Clarify Legacy Validators Package Status [x]

Description: Decide whether `validators/` legacy builders are supported public APIs, internal compatibility helpers, or candidates for deprecation, then align docs and tests.

Implementation rules:

- implement the ticket in the smallest sensible step
- do not remove exported APIs without an explicit breaking-change decision
- document semantic differences from the main compiler path if they remain
- run `gofmt` on modified Go files if tests change
- run `go test ./validators ./glue ./examples -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Pay attention to `oneof` case sensitivity and string token parsing differences.

### Ticket E6-T3 - Run Final CI-Equivalent Quality Gate [x]

Description: After backlog implementation work is complete, run the repository quality gate and capture results in the final report.

Implementation rules:

- run `go mod tidy`
- run `go vet ./...`
- run `govulncheck ./...`
- run `go test ./... -race -covermode=atomic -coverprofile=coverage.out`
- run `go tool cover -func=coverage.out`
- run `scripts/fuzz.sh`
- do not invent Makefile targets; this repository has no Makefile
- do not commit `coverage.out` or local fuzz artifacts unless the operator explicitly asks
- do not create a git commit unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after all checks pass or failures are documented
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- If checks are too expensive for a single execution pass, record exactly which checks were skipped and why.
