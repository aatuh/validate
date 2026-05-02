## Executive summary
This v2 audit covers the current dirty working tree of `github.com/aatuh/validate/v3` as a Go validation library. The repository already contained remediation changes from v1; this audit treats those changes as the target state and does not attribute findings to a specific author or diff.

The v1 high-risk areas are materially improved. `Rule.Elem` now participates in AST cache serialization and nested function-argument cache skipping (`core/serialize_rules.go:51`, `core/serialize_rules.go:84`), manual slice and `forEach` rules now return structured slice type errors instead of panicking on nil or wrong-type inputs (`types/compiler.go:942`, `types/compiler.go:978`, `types/compiler.go:1012`), invalid regex diagnostics are capped/redacted (`types/regex.go:31`, `types/regex.go:42`), missing cross-field references return `field.reference` (`structvalidator/struct.go:363`, `structvalidator/struct.go:416`), and legacy validator semantic differences are documented and tested (`validators/doc.go:1`, `validators/string_test.go:112`).

The biggest remaining correctness issue is a partial fix in per-instance custom type support. Top-level tags and manual rules can use `WithTypeValidator`, but nested tag parsing still calls global-only `ParseTag` inside `foreach`, map key rules, and map value rules (`types/parser.go:392`, `types/parser.go:402`, `types/parser.go:622`, `types/parser.go:629`). A local-only custom type therefore cannot be used in nested collection tags even though the top-level parser accepts the per-instance registry.

Security and trust-boundary posture is better than v1, especially around regex, malformed AST paths, and global registry locking. The main remaining privacy/DoS risk is that map keys are embedded directly into error paths without length caps or redaction in both compiler map validation and struct recursion (`types/compiler.go:1136`, `types/compiler.go:1144`, `structvalidator/struct.go:120`, `structvalidator/struct.go:123`). That can expose caller data through structured error paths even when messages avoid echoing submitted values.

Tests are now substantially more confidence-building. I ran `go test ./...`, `go vet ./...`, non-cached coverage, `go test ./... -coverpkg=./...`, `scripts/fuzz.sh`, and `go test ./... -race -count=1`; all passed. The cross-package coverage pass reported 70.4% total statement coverage. I did not run `go mod tidy` or `govulncheck` in this audit phase because no dependencies or module files were changed and the request asked for reasonable audit checks rather than repeating the full CI gate.

## Scorecard
| Dimension                              | Score | Notes |
|----------------------------------------|------:|-------|
| Architecture & boundaries              | 8.4/10 | Clear library layering remains: root facade in `validate.go`, engine/cache in `core`, parser/compiler in `types`, builders in `glue`, reflection in `structvalidator`, and plugin validators under `validators/...`. Boundary pressure remains around global registries and nested parser paths that do not receive per-instance registry context. |
| SOLID / cohesion / coupling            | 8.0/10 | Packages are mostly cohesive and the root stays thin. `types/compiler.go` is still a large switch-heavy compiler/runtime, and `validators/` remains a parallel legacy helper surface with different semantics from the main compiler path. |
| Correctness & robustness               | 8.0/10 | V1 cache, regex, manual AST, and cross-field bugs are mostly fixed. Remaining correctness risk is nested per-instance custom type parsing plus fail-fast field behavior where consumers may expect aggregate field errors. |
| Security                               | 7.8/10 | Regex pattern output is capped/redacted, regex input length is capped, malformed manual AST paths avoid panics, and registries are synchronized. Map key path rendering can still leak long or sensitive caller-controlled key values. |
| Test effectiveness                     | 8.0/10 | 49 test files, examples, parser and legacy builder fuzz targets, race tests, cache regression tests, regex privacy tests, and struct reference tests now exist. Gaps remain around nested per-instance custom types, map-key path privacy, conditional rules, context-aware APIs, and all-errors-per-field behavior. |
| Change safety & backward compatibility | 8.0/10 | Public facade aliases, stable codes, additive APIs, and docs/examples reduce compatibility risk. The nested custom type gap means the new per-instance API is not uniformly supported across tag forms. |
| Operability & observability            | 7.6/10 | For a library, structured errors, deterministic traversal, examples, and CI give good feedback. There is no logging/metrics expectation, but very long map-key paths can make diagnostics noisy or unsafe. |
| Clarity & developer experience         | 8.0/10 | README and examples are much stronger, and legacy validator semantics are clearer. DX remains uneven around adapter-only context helpers, first-error-per-field behavior, and custom type behavior in nested tags. |
| Extensibility                          | 7.6/10 | Per-instance rule compilers, struct rules, type validators, custom tags, and builder escape hatches are useful. Conditional validation, context-aware validation, all-errors-per-field collection, and nested custom type parsing still need design/implementation work. |
| Overall                                | 7.9/10 | Good public-library foundation after v1 remediation, with one material partial-resolution bug and several scoped extensibility/privacy gaps left. |

Confidence: high

## Findings by severity
### Critical
- None found.

### High
- None found.

### Medium
- Per-instance custom type registration is only partially resolved for tag parsing. `ParseTagWithRegistry` accepts an optional registry and checks it before global types (`types/parser.go:60`, `types/parser.go:175`, `types/parser.go:196`), and the compiler prefers per-instance type validators at runtime (`types/compiler.go:1293`, `types/compiler.go:1308`). However, nested parsers still call global-only `ParseTag`: `foreach=(...)` calls `ParseTag(inner)` (`types/parser.go:392`, `types/parser.go:402`), and map `keys=(...)` / `values=(...)` call `ParseTag(inner)` via `parseNestedRulesRule` (`types/parser.go:622`, `types/parser.go:629`). Observed impact: a validator configured with `WithTypeValidator("local", ...)` can compile `local` as a top-level tag, but `slice;foreach=(local)` or `map;values=(local)` cannot parse unless `local` is registered globally. This means the v1 custom type finding is resolved for top-level and manual paths, but not for nested collection tags.
- Map keys are embedded directly into structured error paths without caps or redaction. Compiler map key/value validation builds `pathPrefix := "[" + fmt.Sprint(key.Interface()) + "]"` (`types/compiler.go:1136`, `types/compiler.go:1144`), and reflection recursion for untagged maps builds paths with `fmt.Sprint(mk.Interface())` (`structvalidator/struct.go:120`, `structvalidator/struct.go:123`). Observed fact: short keys produce useful paths and tests assert this (`structvalidator/advanced_struct_test.go:61`). Inferred risk: long, token-like, or secret-bearing map keys can appear in `FieldError.Path`, JSON output, and `Error()` strings, bypassing the message-level privacy hardening documented in `README.md:156`.
- Conditional validation remains a narrow struct-only feature. The current public docs list `requiredWith` plus equality/inequality field rules (`README.md:125`), and code only recognizes `eqField`, `neField`, and `requiredWith` as built-in struct rules (`structvalidator/struct.go:263`, `structvalidator/struct.go:269`). The v1 backlog still has E5-T2 and E5-T3 open, and there is no additive conditional rule design covering tag, builder/manual, and struct behavior.
- Context-aware validation remains adapter-only. `core.CheckFuncCtx`, `WithContext`, and `WithoutContext` exist (`core/ctx.go:5`, `core/ctx.go:11`, `core/ctx.go:19`), but the compiler, builders, and struct validator still produce and consume `func(any) error` (`types/compiler.go:99`, `glue/builders.go:136`, `structvalidator/struct.go:40`). This is acceptable as a documented current scope, but request-scoped or cancellable custom validators still require closure-based out-of-band state.
- Single-value compiled validators remain first-error-only. `CompileE` stops at the first failing compiled rule for a value (`types/compiler.go:142`), while struct validation aggregates across fields (`structvalidator/struct.go:77`, `structvalidator/struct.go:177`). This preserves compatibility, but there is still no opt-in all-errors-per-field mode for form-style workflows.

### Low
- Global extension registries remain process-wide. Global rule compilers are copied into new compilers (`types/compiler.go:27`, `types/compiler.go:49`), global custom types have no unregister/reset API (`types/type_registry.go:89`, `types/type_registry.go:92`), and default translations are process-wide (`translator/translator.go:25`, `translator/translator.go:66`). Per-instance APIs mitigate much of this, but plugin/test collisions remain possible when callers use globals.
- AST cache serialization still has an implicit fallback policy for unsupported non-function args. Functions are detected and skip caching (`core/serialize_rules.go:37`, `core/serialize_rules.go:92`), and built-in rule args serialize deterministically. Unsupported values fall back to `fmt.Sprintf("%v", v)` (`core/serialize_rules.go:217`), which can be unstable for pointer-like custom rule arguments. This is low risk for built-in rules, but custom compilers that depend on pointer identity or opaque args need clearer cache semantics.
- The README documents representative stable codes but not a complete error-code table. `errors/codes.go` is the source of truth (`errors/codes.go:3`), and README guidance points consumers to `Code`, `Path`, and `Param` (`README.md:149`). A complete generated or manually maintained code table would improve downstream compatibility work.

## Hexagonal architecture verdict
This repository is not a classic hexagonal application because it is a library with no HTTP transport, persistence, vendor SDK adapter, or deployment layer. It is best described as a well-layered Go library with some ports-and-adapters-style extension points.

What is clean: the root package remains a thin facade and plugin registration point (`validate.go:11`, `validate.go:17`), `core` owns engine configuration and cache behavior (`core/engine.go:20`), `types` owns AST parsing/compilation (`types/parser.go:54`, `types/compiler.go:41`), `glue` integrates builders and the engine (`glue/validate.go:10`), and `structvalidator` owns reflection traversal (`structvalidator/struct.go:16`). The dependency direction is mostly predictable and there is no framework leakage.

What leaks across boundaries: process-wide extension state still exists for global rule compilers, global custom types, and default translations. Per-instance type support is present, but registry context is not threaded into nested parser calls. The `validators` package still exposes a legacy direct-helper API in parallel with the main compiler path, although its compatibility status and semantic mismatch are now documented.

Verdict: partially hexagonal/layered. The package structure is sound for a public validation library; global registries, nested parser context gaps, and duplicate validator surfaces are the main reasons it is not cleaner.

## Test verdict
Covered well: `Rule.Elem` cache regression behavior (`core/serialize_rules_test.go:12`), nested function-argument cache skipping (`core/serialize_rules_test.go:53`), manual slice/foreach panic resistance (`types/manual_rules_test.go:10`), invalid regex privacy (`types/regex_privacy_test.go:12`), struct cross-field missing-reference behavior (`structvalidator/advanced_struct_test.go:95`), per-instance custom type top-level behavior (`glue/type_registry_test.go:13`), per-instance/global type precedence (`core/extensibility_test.go:44`), legacy oneof semantic mismatch (`validators/string_test.go:112`), parser fuzz seeds (`types/fuzz_parser_test.go:16`), and CI quality gates (`.github/workflows/.ci.yml:23`).

Weak: nested per-instance custom type tags are not covered, which is why the parser registry propagation bug remains. Map-key privacy/size behavior is not covered. Conditional validation beyond `requiredWith`, context-aware compiler/builder APIs, and all-errors-per-field behavior are not covered because they are not implemented. Some builder methods and plugin type-validator paths still have relatively low direct coverage, though cross-package coverage is improved.

Checks run in this audit:

- `go test ./...` passed.
- `go vet ./...` passed.
- `go test ./... -count=1 -coverprofile=/tmp/validate-v2-coverage.out` passed; default per-package total was 55.4%.
- `go test ./... -count=1 -coverpkg=./... -coverprofile=/tmp/validate-v2-coverpkg.out` passed; cross-package total was 70.4%.
- `scripts/fuzz.sh` passed for `FuzzParseTag`, `FuzzParseTagLong`, `FuzzBuildIntRules`, and `FuzzBuildStringRules`.
- `go test ./... -race -count=1` passed.

The tests are now confidence-building for the v1 remediation targets and documented happy paths, but the remaining gaps are real behavioral surfaces rather than just coverage cosmetics.

## Best next fixes
1. Thread the optional `TypeRegistry` through nested tag parsing for `foreach`, map keys, and map values, then add regression tests for local-only nested custom types.
2. Define and implement a bounded/redacted map-key path rendering policy for compiler map errors and struct recursion.
3. Complete the additive conditional validation design and implement the first small rule across parser, struct validation, builders/manual rules, docs, and examples.
4. Decide whether all-errors-per-field belongs in compiler options, `ValidateOpts`, or a separate API, then implement it without changing fail-fast defaults.
5. Decide whether context-aware validation should remain documented as adapter-only or gain explicit compiler/builder/struct APIs.
6. Clarify cache semantics for unsupported non-function AST args used by custom rule compilers.
7. Add a complete stable error-code reference in README or package docs.

## Optional follow-up
- Targeted remediation plan
- Package-by-package review
- Security-focused pass for structured output and path rendering
- Test-gap plan for extension APIs
