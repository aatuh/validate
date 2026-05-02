# Backlog

Project: github.com/aatuh/validate/v3

Status legend:

- [ ] not done
- [x] done

## Epic E1 - Nested Custom Type Registry Propagation [x]

Description: Finish the per-instance custom type work by making nested tag parsing use the validator's local type registry consistently.

### Ticket E1-T1 - Add Nested Local Custom Type Regression Tests [x]

Description: Add failing tests for `WithTypeValidator` with local-only custom types inside `slice;foreach=(...)`, `map;keys=(...)`, and `map;values=(...)`.

Implementation rules:

- implement the ticket in the smallest sensible step
- do not change production code in this test-only ticket except for required test helpers
- run `gofmt` on modified Go files
- run `go test ./types ./core ./glue -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Cover tag parsing through `Validate.FromTag` or `Validate.FromRules`.
- Include at least one builder/manual-rule comparison that already works, so the failure is clearly parser-scoped.

### Ticket E1-T2 - Thread TypeRegistry Through Nested Parsers [x]

Description: Update nested `foreach`, map key, and map value parsing so `ParseTagWithRegistry` passes the optional registry through all recursive tag parses.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve global custom type fallback and all existing valid tag behavior
- avoid ad hoc tag splitting; keep nested parentheses behavior covered
- run `gofmt` on modified Go files
- run `go test ./types ./core ./glue -count=1`
- run `scripts/fuzz.sh` because parser and tag handling change
- run `go test ./... -race -count=1` because compiler and registry behavior are affected
- do not invent Makefile targets; this repository has no Makefile
- do not create git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Existing top-level `ParseTag` should keep using global registrations for backward compatibility.

### Ticket E1-T3 - Document Nested Custom Type Support [x]

Description: Update README or examples to show that per-instance custom types work in nested collection tags once parser propagation is fixed.

Implementation rules:

- keep documentation concise and aligned with executable examples
- prefer an example under `examples/` if the behavior is public-facing
- run `gofmt` on modified Go example files
- run `go test ./examples -v -count=1`
- run `go test ./... -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Mention global fallback only where it helps users avoid accidental process-wide registrations.

## Epic E2 - Structured Path Privacy And Size Caps [x]

Description: Prevent caller-controlled map keys from creating oversized or sensitive structured error paths while preserving useful short-key diagnostics.

### Ticket E2-T1 - Define Map Key Path Rendering Policy [x]

Description: Decide how map keys should appear in `FieldError.Path`, including length caps, escaping, sensitive-marker redaction, and compatibility expectations.

Implementation rules:

- write the policy in README, package docs, or a short audit-linked design note before changing behavior
- preserve short, ordinary map keys such as `items[id]` where practical
- explicitly identify any backward compatibility risk for path strings
- run documentation checks through `go test ./examples -v -count=1` if examples change
- run `go test ./... -count=1` before marking the ticket complete
- do not invent Makefile targets; this repository has no Makefile
- do not create git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Treat `Path` as structured output, not just display text.

### Ticket E2-T2 - Add Map Key Privacy Regression Tests [x]

Description: Add tests for long and sensitive-looking map keys in compiler map validation and struct recursion.

Implementation rules:

- implement the ticket in the smallest sensible step
- assert structured paths, codes, and params rather than English text
- include both `types` map key/value validation and `structvalidator` recursive map paths
- run `gofmt` on modified Go files
- run `go test ./types ./structvalidator -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Include short-key tests to protect current useful path behavior.

### Ticket E2-T3 - Implement Bounded Map Key Paths [x]

Description: Route compiler and struct-validator map path construction through a deterministic formatter that applies the approved policy.

Implementation rules:

- implement the ticket in the smallest sensible step
- keep path formatting deterministic across map traversal
- avoid exposing raw long inputs, secrets, tokens, or private caller data
- run `gofmt` on modified Go files
- run `go test ./types ./structvalidator ./errors ./examples -count=1`
- run `go test ./... -race -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- If the helper must be shared, place it where it does not create an import cycle.

### Ticket E2-T4 - Document Path Privacy Behavior [x]

Description: Update README error guidance to clarify that messages avoid submitted values and paths use bounded map key previews.

Implementation rules:

- keep README wording short and exact
- update examples only if expected output changes
- run `go test ./examples -v -count=1`
- run `go test ./... -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Do not claim privacy behavior that custom validators or translators can bypass.

## Epic E3 - Conditional Validation [ ]

Description: Complete the remaining v1 open work by adding a small, additive conditional validation feature without changing existing defaults.

### Ticket E3-T1 - Select The First Additive Conditional Rule [ ]

Description: Define the first conditional rule beyond `requiredWith`, including tag syntax, supported construction paths, error code, and compatibility behavior.

Implementation rules:

- keep the first rule small and independently testable
- prefer same-level struct field references unless a broader compiler-level design is justified
- preserve existing `requiredWith`, `eqField`, and `neField` behavior
- record the decision in README, package docs, or the backlog before implementation
- run `go test ./... -count=1` before marking the ticket complete
- do not invent Makefile targets; this repository has no Makefile
- do not create git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Candidate rules should be evaluated for parser complexity and public API compatibility before implementation.

### Ticket E3-T2 - Add Conditional Rule Tests First [ ]

Description: Add failing tests for the selected conditional rule across valid, invalid, missing-reference, and malformed-tag cases.

Implementation rules:

- implement the ticket in the smallest sensible step
- test tag parsing and struct validation behavior before production changes
- include builder or manual-rule coverage if the selected design exposes those paths
- assert stable error codes and paths rather than full English text
- run `gofmt` on modified Go files
- run `go test ./types ./structvalidator ./glue -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Include a typo or inaccessible referenced field case if the rule references another field.

### Ticket E3-T3 - Implement The Conditional Rule [ ]

Description: Implement the selected rule across parser, compiler or structvalidator ownership, root re-exports if needed, README, and executable examples.

Implementation rules:

- implement the ticket in the smallest sensible step
- keep parser, builder, manual rule, and struct behavior aligned with the selected design
- preserve existing public APIs and error codes unless an additive code is documented
- run `gofmt` on modified Go files
- run `go test ./types ./structvalidator ./glue ./examples -count=1`
- run `go test ./... -race -count=1`
- run `scripts/fuzz.sh` because parser and tag behavior change
- do not invent Makefile targets; this repository has no Makefile
- do not create git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Update README supported tags and error-code documentation if a new public code is added.

## Epic E4 - Field Error Aggregation And Context Scope [ ]

Description: Make two known extensibility limits explicit and add only backward-compatible APIs where the value justifies the extra surface area.

### Ticket E4-T1 - Design All-Errors-Per-Field Collection [ ]

Description: Decide whether all-errors-per-field belongs in compiler options, `core.ValidateOpts`, or separate compile/build APIs while preserving current fail-fast defaults.

Implementation rules:

- keep the design additive; do not change default `CompileRules`, `FromTag`, builder, or struct behavior
- define how structured paths should be produced when several rules fail on one field
- identify the narrowest first implementation path
- run documentation checks through `go test ./examples -v -count=1` if examples change
- run `go test ./... -count=1` before marking the ticket complete
- do not invent Makefile targets; this repository has no Makefile
- do not create git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- The current fail-fast behavior is visible in `types.Compiler.CompileE` and should stay compatible.

### Ticket E4-T2 - Implement A Minimal All-Errors Opt-In [ ]

Description: Add the smallest approved API for collecting multiple failures on a single value or field.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve current default fail-fast behavior
- add tests for success, first failure, multiple failures, `omitempty`, and `required`
- run `gofmt` on modified Go files
- run `go test ./types ./core ./glue ./structvalidator -count=1`
- run `go test ./... -race -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Avoid changing English message assertions; use codes and structured fields.

### Ticket E4-T3 - Clarify Context-Aware Validation Scope [ ]

Description: Document whether context support remains adapter-only or define explicit context-aware compiler/builder/struct APIs for a future implementation.

Implementation rules:

- keep existing `func(any) error` APIs stable
- avoid adding context parameters to existing public methods
- if no implementation is chosen, document the adapter-only scope clearly
- run `go test ./core ./examples -count=1` if docs or examples change
- run `go test ./... -count=1` before marking the ticket complete
- do not invent Makefile targets; this repository has no Makefile
- do not create git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Context APIs are useful only if cancellation or request-scoped validators can observe the context.

## Epic E5 - Cache Semantics And Public Error-Code Reference [ ]

Description: Reduce ambiguity in custom extension behavior and improve downstream compatibility documentation.

### Ticket E5-T1 - Define Cache Policy For Opaque AST Args [ ]

Description: Decide how `SerializeRules` should treat unsupported non-function arguments used by custom rule compilers.

Implementation rules:

- document whether opaque pointer/interface args are cacheable, skipped, or require caller-provided stable representations
- preserve deterministic cache behavior for all built-in rule arguments
- add focused tests for supported scalar, time, nested rule, function, and opaque pointer args
- run `gofmt` on modified Go files
- run `go test ./core ./types -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- This should not reopen the fixed `Rule.Elem` cache collision.

### Ticket E5-T2 - Implement Opaque Arg Cache Handling If Needed [ ]

Description: If E5-T1 finds unsafe cache behavior, update cache skipping or serialization so custom rule args cannot collide or depend on unstable pointer formatting.

Implementation rules:

- implement the ticket in the smallest sensible step
- preserve existing cache keys for built-in deterministic args where practical
- run `gofmt` on modified Go files
- run `go test ./core ./types -count=1`
- run `go test ./... -race -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Prefer skipping cache for unsafe opaque args over inventing unstable string keys.

### Ticket E5-T3 - Add Complete Error-Code Documentation [ ]

Description: Add a concise public reference for stable built-in error codes from `errors/codes.go`.

Implementation rules:

- keep the code list aligned with `errors/codes.go`
- do not change existing codes
- prefer concise grouped tables in README or package docs
- run `go test ./examples -v -count=1` if README examples change
- run `go test ./errors ./translator -count=1`
- do not invent Makefile targets; this repository has no Makefile
- do not create git commits unless the operator explicitly asks for commits
- update the ticket checkmark from `[ ]` to `[x]` only after the ticket is actually complete and checks pass
- update the epic checkmark from `[ ]` to `[x]` only when all child tickets are complete

Notes:

- Treat codes as a public compatibility contract.
