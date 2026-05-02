## Executive summary
This audit covers the current working tree of `github.com/aatuh/validate/v3` as a Go validation library. The workspace already contained uncommitted source, test, README, and `AGENTS.md` changes before the audit; findings below are current-worktree findings and are not attributed to a specific author or diff.

The repository is feature-rich and reasonably well layered for a library: the root package is a facade, `core` owns engine/cache configuration, `types` owns rules/parsing/compilation, `glue` owns builders, `structvalidator` owns reflection, and plugin validators register through `types` and `translator`. There is no HTTP, persistence, auth, payment, frontend, deployment, or gambling scope evidenced.

The biggest correctness risk is cache-key incompleteness for legacy nested `Rule.Elem` rules. The compiler still honors `Rule.Elem` for `forEach`, but the cache serializer omits it, so two different manually constructed nested validators can share one cached function. That is a real public API correctness risk because manual AST construction is documented and exported.

The strongest security/robustness work is around deterministic structured errors, synchronized global registries, regex anchoring, regex input length limits, parser fuzzing, race tests, and CI. The main remaining trust-boundary gaps are unbounded invalid-regex pattern echoing, panic-prone manual AST rule paths, global custom type state, and silent struct cross-field typos.

Tests are meaningful but uneven. `go test`, `go vet`, `govulncheck`, race/coverage, and fuzz smoke all pass, but total statement coverage is 50.3% and several newly exposed extension paths have little or no direct coverage.

## Scorecard
| Dimension                              | Score | Notes |
|----------------------------------------|------:|-------|
| Architecture & boundaries              |  8/10 | Clear package responsibilities: root facade in `validate.go`, engine/cache in `core`, parser/compiler in `types`, reflection in `structvalidator`, and plugin validators under `validators/...`. Boundary leaks remain through process-wide registries and duplicate legacy validator APIs. |
| SOLID / cohesion / coupling            |  7/10 | Most packages are cohesive, but `types/compiler.go` is a large switch-heavy compiler/runtime, `validators/` contains a second legacy-style validation stack, and custom type behavior is coupled to global state. |
| Correctness & robustness               |  7/10 | Many validators handle type errors and deterministic paths well, but `Rule.Elem` is excluded from cache keys, manual AST slice rules can panic on nil inputs, and struct cross-field typos can silently alter behavior. |
| Security                               |  7/10 | No network/auth/payment attack surface is evidenced. Regex input length is capped and registries use locks, but invalid regex errors can echo full caller patterns and several malformed manual rule paths rely on callers including base rules. |
| Test effectiveness                     |  7/10 | 44 test files, examples, fuzz targets, race tests, and property-style tag/builder checks exist. Coverage is uneven at 50.3% total, with gaps in `Rule.Elem`, custom type registry isolation, struct rule lookup failures, invalid regex privacy, and manual malformed AST paths. |
| Change safety & backward compatibility |  7/10 | Public facade aliases and stable error codes are explicit, and CI is strong. Manual rule serialization, global registries, and duplicate old/new validation paths increase compatibility risk for future changes. |
| Operability & observability            |  7/10 | For a library, structured errors, deterministic map traversal, examples, and CI give useful operational feedback. There is no logging or metrics expectation here, but failure diagnostics are sometimes raw strings or first-error-only. |
| Clarity & developer experience         |  8/10 | README, examples, package docs, and root re-exports make the library approachable. DX suffers where `core/ctx.go` suggests context support without integration and `validators/` exposes older APIs with different semantics from the main compiler. |
| Extensibility                          |  7/10 | Per-instance rule compilers, struct rule compilers, custom rules, custom types, and builder escape hatches are useful. Extensibility is limited by global custom type registration, lack of context-aware compiler flow, limited conditional validation, and no all-errors-per-field mode. |
| Overall                                |  7.3/10 | Good public-library foundation with meaningful correctness, trust-boundary, and test-depth gaps to prioritize before calling it production-grade. |

Confidence: high

## Findings by severity
### Critical
- None found.

### High
- `Rule.Elem` is not part of the compiled-validator cache key even though the compiler still executes it. `core.SerializeRules` writes only `Kind` and `Args` for each rule and never serializes `Rule.Elem` (`core/serialize_rules.go:25`, `core/serialize_rules.go:32`, `core/serialize_rules.go:51`). `core.HasFuncArgs` also checks only `Args` (`core/serialize_rules.go:71`). The compiler still uses `rule.Elem` as the backward-compatible fallback for `KForEach` (`types/compiler.go:358`). Public constructors keep `Elem` available (`types/rule.go:97`, `types/rule.go:110`). Two different manually built `forEach` rules with different `Elem` values can therefore collide under the same `ast:` cache key and reuse the wrong compiled validator.

### Medium
- Invalid regex errors can echo the full caller-supplied pattern despite a truncation attempt. `compileRegexSafe` truncates a long pattern only for a discarded translation call (`types/regex.go:24`, `types/regex.go:30`), then returns the raw compile error. The compiled invalid-regex closure formats `pattern` directly into the translated message (`types/compiler.go:259`, `types/compiler.go:265`, `types/compiler.go:268`), and `validateRegexWithPattern` has the same full-pattern fallback path (`types/compiler.go:586`, `types/compiler.go:588`). This can produce oversized error messages or expose sensitive caller data embedded in validation patterns.
- Manual AST slice rules can panic on nil input when callers omit the base `KSlice` rule. `validateSliceLength`, `validateMinSliceLength`, `validateMaxSliceLength`, and `validateForEach` call `reflect.ValueOf(v).Kind()` without first checking `IsValid` (`types/compiler.go:943`, `types/compiler.go:956`, `types/compiler.go:969`, `types/compiler.go:982`). Tag and builder paths usually prepend `KSlice`, but `CompileRules` is a public manual API and should be panic-resistant for malformed or partial rule sets.
- Struct cross-field rules silently tolerate missing referenced fields. `StructRuleContext.FieldValue` returns `(nil, false)` when a field is missing or inaccessible (`core/struct_rule.go:20`), but `eqField`, `neField`, and `requiredWith` ignore the boolean (`structvalidator/struct.go:363`, `structvalidator/struct.go:375`, `structvalidator/struct.go:387`). A misspelled field name can become a normal validation comparison instead of a compile/configuration error.
- Custom types are process-wide, not per validator instance. The parser accepts a custom base type only if it is in the global registry (`types/parser.go:167`), the compiler validates custom types through `GetGlobalTypeValidator` at runtime (`types/compiler.go:1285`), and the root facade re-exports `RegisterGlobalType` (`validate.go:129`). This is usable for plugins but creates cross-test and cross-application collision risk with no unregister, namespace, or per-instance override.
- The compiler returns only the first rule error for a single value. `CompileE` stops at the first failing compiled rule (`types/compiler.go:128`). `structvalidator` aggregates across fields and nested items, but not all failures for one field. That is a deliberate current behavior, but it limits UI/form use cases and leaves no explicit all-errors-per-field mode.

### Low
- Context-aware validation is adapter-only. `core.CheckFuncCtx`, `WithContext`, and `WithoutContext` exist (`core/ctx.go:8`), but builders, the compiler, and struct validation still expose `func(any) error` flows. Context cancellation or request-scoped custom validators are not supported without out-of-band closures.
- The legacy `validators` package overlaps the main `types` compiler path and has different semantics in places. For example, old `StringValidators.OneOf` is case-insensitive (`validators/string.go:162`), while the compiler's `validateOneOf` compares exact strings (`types/compiler.go:615`). This makes extension behavior harder to reason about.
- Test coverage is broad but not deep in new extension surfaces. The uncached race/coverage run reported 50.3% total statement coverage. The coverage report showed 0% for `types.NewRuleWithElem`, `types.NewRuleWithElemValue`, custom type registry accessors, `core.StructRuleContext.FieldValue`, and several builder methods, which aligns with the highest-risk uncovered paths.

## Hexagonal architecture verdict
The code is not a traditional hexagonal application because it is a library with no transport, persistence, or infrastructure adapters. It is best described as a partially layered library architecture.

What is clean: dependency direction is mostly inward and predictable. `validate.go` stays thin and imports plugins only for registration (`validate.go:11`). `glue` depends on `core`, `structvalidator`, `translator`, and `types` (`glue/validate.go:3`). `structvalidator` depends on `core`, `errors`, and `types` (`structvalidator/struct.go:11`). `types` owns parsing and compilation without depending on `glue` or `structvalidator`.

What leaks across boundaries: plugin registration and custom types are process-wide through global registries (`types/compiler.go:27`, `types/type_registry.go:75`), which bypasses per-instance configuration. The older `validators` package implements a parallel validator stack rather than being the only compiler backend. Struct validation calls into the core engine directly, which is acceptable here but not a strict port boundary.

Verdict: partially hexagonal/layered. The package layout is strong for a Go library, but global registries and duplicate validation paths keep it from being a clean ports-and-adapters design.

## Test verdict
Covered well: basic parser behavior, expanded tags, tag/builder equivalence, struct recursion, `StopOnFirst`, JSON field names, plugin validators, errors, translator, thread-safety, examples, fuzz parser smoke, and legacy int/string builder fuzzing. CI evidence includes `go mod tidy`, `go vet`, `govulncheck`, race coverage, and fuzz smoke in `.github/workflows/.ci.yml`.

Weak: cache serialization for `Rule.Elem`, manual AST panic resistance, invalid regex pattern redaction, custom type registry isolation/collision behavior, missing struct rule field references, all-errors-per-field behavior, and context-aware validation are not sufficiently covered. Several tests assert string contents or broad success/failure rather than structured codes.

The tests are confidence-building for the documented happy paths and common negative paths, but superficial around newer extensibility and trust-boundary surfaces.

## Best next fixes
1. Fix and test cache serialization for `Rule.Elem`, including nested function arguments and deterministic cache behavior.
2. Add panic-resistance tests for manual AST rules, then harden slice/map/foreach validators against nil and malformed values.
3. Redact or cap invalid regex patterns in all error-message paths and test with long sensitive-looking patterns.
4. Make missing struct cross-field references explicit structured errors, preserving compatibility with valid existing tags.
5. Define custom type registration semantics: document global-only behavior or add per-instance custom type registration.
6. Add an opt-in all-errors-per-field mode before changing default fail-fast behavior.
7. Rationalize or document the relationship between the legacy `validators` package and the main `types` compiler path.

## Optional follow-up
- Targeted remediation plan
- Package-by-package review
- Refactor roadmap
- Security-focused pass
- Test-gap plan
