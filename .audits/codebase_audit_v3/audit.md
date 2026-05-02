## Executive summary
This v3 audit covers the current dirty working tree of `github.com/aatuh/validate/v3` after the v2 remediation work. I treated the current worktree as the audit target and did not attempt to revert or attribute unrelated changes.

The two material v2 findings appear resolved. Nested collection tag parsing now threads the optional per-instance `TypeRegistry` through `foreach`, map key, and map value rule parsing (`types/parser.go:132`, `types/parser.go:143`, `types/parser.go:402`, `types/parser.go:629`), and regression tests cover local-only custom types in `slice;foreach=(...)`, `map;keys=(...)`, and `map;values=(...)` (`glue/type_registry_test.go:38`). Compiler and struct recursion map paths now use the shared bounded formatter in `internal/pathutil`, preserving short ordinary keys and redacting long or sensitive-looking keys (`types/compiler.go:1145`, `structvalidator/struct.go:124`, `internal/pathutil/map_key.go:14`).

The codebase is now in a good public-library state. The root facade remains thin, rule parsing and compilation are centralized in `types`, reflection behavior is isolated in `structvalidator`, and extension APIs are increasingly per-instance. The remaining risks are mostly API-shape and contract completeness issues rather than acute correctness defects.

The biggest remaining product and robustness gaps are conditional validation breadth, explicit context-aware validation through compiler/builder/struct paths, an opt-in all-errors-per-field mode, global process-wide registries, clearer cache semantics for opaque custom rule arguments, and a complete stable error-code reference. These are meaningful for a validation library, but none block the recently remediated v2 behavior.

Checks run in this audit passed: uncached `go test ./... -count=1`, `go vet ./...`, `go test ./... -race -count=1`, `scripts/fuzz.sh`, and cross-package coverage via `go test ./... -count=1 -coverpkg=./... -coverprofile=/tmp/validate-v3-coverpkg.out`. The cross-package coverage summary reported 70.9% total statement coverage.

## Scorecard
| Dimension                              | Score | Notes |
|----------------------------------------|------:|-------|
| Architecture & boundaries              | 8.6/10 | Clear package boundaries remain: root facade in `validate.go`, engine/cache in `core`, parser/compiler in `types`, builders in `glue`, reflection in `structvalidator`, and plugins in `validators/...`. The new `internal/pathutil` helper is an appropriate shared internal policy point and does not invert package ownership. |
| SOLID / cohesion / coupling            | 8.1/10 | Most packages are cohesive and extension APIs are additive. `types/compiler.go` remains a large switch-heavy compiler/runtime, and `validators/` still exposes a parallel legacy helper surface with documented semantic differences. |
| Correctness & robustness               | 8.4/10 | V1 and v2 material correctness issues are covered by tests: cache keys include nested element rules, malformed manual slice rules return structured errors, nested local custom types work, and map paths are bounded. Remaining gaps are mostly feature-level: conditional rules, context-aware validation, and all-errors-per-field behavior. |
| Security                               | 8.5/10 | Regex diagnostics are capped/redacted, regex input length is capped, global registries are synchronized, and map key paths now redact long/sensitive/private-looking keys. Residual risk is mostly from caller-provided custom validators/translators and global extension state. |
| Test effectiveness                     | 8.3/10 | The suite has 52 test files, examples, fuzz targets, race-oriented tests, parser/compiler privacy tests, and v2 regression coverage. Some builder methods, plugin type-validator paths, context scope, and future conditional/error-aggregation behavior remain thin or unimplemented. |
| Change safety & backward compatibility | 8.3/10 | Public facade aliases, additive APIs, stable codes, examples, and docs lower downstream risk. The map-path privacy policy intentionally changes long/sensitive key rendering while preserving short ordinary keys; complete code documentation would further improve compatibility work. |
| Operability & observability            | 8.0/10 | For a library, deterministic structured errors, paths, examples, CI checks, and fuzz smoke give good diagnostics. There is no service logging/metrics expectation, but error-code discovery and cache behavior are still not fully self-documenting. |
| Clarity & developer experience         | 8.2/10 | README and executable examples now explain supported tags, struct rules, custom rules, custom types, and path privacy. DX remains uneven around context helpers, fail-fast defaults, global vs per-instance registration choices, and full error-code discovery. |
| Extensibility                          | 8.0/10 | Per-instance rule compilers, struct rule compilers, type validators, custom tags, builder escape hatches, and nested custom type support are strong. The next extensibility frontier is conditional validation, context propagation, all-errors collection, and clearer global/cache contracts. |
| Overall                                | 8.3/10 | Strong public-library foundation with v2 issues resolved and mostly lower-priority API, documentation, and extensibility gaps remaining. |

Confidence: high

Scorecard mean: 8.2667/10 across the nine dimensions excluding Overall.

## Findings by severity
### Critical
- None found.

### High
- None found.

### Medium
- Conditional validation is still narrow and struct-only. The README documents `eqField`, `neField`, and `requiredWith` as struct-only rules (`README.md:125`), and implementation recognizes those built-ins plus custom `struct:` rules in `structvalidator/struct.go:263`. This is compatible with current behavior, but there is no broader additive design for common conditional validation such as required-if, excluded-if, or conditionally applying nested rule sets across tag, builder, manual-rule, and struct paths.
- Context-aware validation remains adapter-only. `core.CheckFuncCtx`, `WithContext`, and `WithoutContext` exist (`core/ctx.go:8`), but compiler, builder, and struct validation APIs still compile and execute `func(any) error` validators (`types/compiler.go:111`, `glue/builders.go:136`, `structvalidator/struct.go:41`). Callers needing request-scoped cancellation or context-dependent external checks must use closures or adapters outside the main rule pipeline.
- Single-value compiled validators remain fail-fast per value. `types.Compiler.CompileE` returns the first failing rule for a single value (`types/compiler.go:136`), while slice/map and struct paths can aggregate across elements or fields (`types/compiler.go:985`, `types/compiler.go:1136`, `structvalidator/struct.go:78`). This preserves compatibility, but there is still no opt-in all-errors-per-field mode for form-style consumers who need every failing rule on a field.

### Low
- Global extension registries remain process-wide. Global rule compilers are copied into new compilers (`types/compiler.go:28`, `types/compiler.go:50`), global custom type registrations overwrite by name (`types/type_registry.go:89`), and default translations are process-wide (`translator/translator.go:25`). Per-instance APIs now mitigate most normal use cases, but tests and plugins can still collide when using global registration.
- AST cache serialization has clearer function handling but still opaque fallback semantics for unsupported non-function arguments. Function arguments are detected recursively and skip caching (`core/serialize_rules.go:42`, `core/serialize_rules.go:92`), and built-in args serialize deterministically. Unsupported custom arg values still fall back to `fmt.Sprintf("%v", v)` (`core/serialize_rules.go:217`), which is hard to reason about for pointer-like or stateful custom rule arguments.
- Error-code documentation is representative rather than complete. `errors/codes.go` is the real source of truth (`errors/codes.go:3`), and the README advises consumers to use `Code`, `Path`, and `Param` (`README.md:149`), but there is no complete table mapping every built-in code to its rule/tag, parameter, and path behavior.
- The map key path formatter is shared correctly, but direct policy coverage is mostly indirect. Compiler and struct-validator tests cover short, long, and sensitive string keys (`types/map_path_privacy_test.go:11`, `structvalidator/map_path_privacy_test.go:12`), while `internal/pathutil` itself has no direct tests for nil, numeric, boolean, complex, and escaping-sensitive keys. That is a small confidence gap around a public-output policy.
- Several builder and plugin helper paths remain weakly exercised despite improved cross-package coverage. The coverage pass shows many builder methods and legacy validator helper methods at 0% while core parser/compiler and struct paths are better covered. This matters mainly because the root facade exposes these builders as public API.

## Hexagonal architecture verdict
This repository is not a classic hexagonal application because it is a Go library with no HTTP transport, persistence, deployment, or vendor SDK adapter layer. It is best described as a well-layered validation library with ports-and-adapters-style extension points.

What is clean: `validate.go` stays a thin root facade and plugin registration point (`validate.go:17`), `core` owns engine configuration and cache behavior (`core/engine.go:20`), `types` owns rule representation, parsing, and compilation (`types/parser.go:63`, `types/compiler.go:42`), `glue` integrates builders with the engine (`glue/validate.go:10`), and `structvalidator` owns reflection traversal (`structvalidator/struct.go:17`). The dependency direction is predictable and there is no framework leakage.

What leaks across boundaries: process-wide extension state still exists for rule compilers, custom types, and default translations. The legacy `validators` package still provides direct helper APIs in parallel with the main compiler path. Those are now documented and tested enough to be manageable, but they are still a source of conceptual duplication.

The new `internal/pathutil` package is an acceptable boundary compromise. It holds output-formatting policy shared by `types` and `structvalidator` without depending on either package or pulling reflection traversal into the compiler.

Verdict: partially hexagonal/layered. The package structure is sound for a public validation library; the main remaining boundary concerns are global registries, legacy helper surfaces, and future context/conditional features needing careful ownership.

## Test verdict
Covered well: nested local custom type support in collection tags (`glue/type_registry_test.go:38`), map key path privacy in compiler and struct recursion (`types/map_path_privacy_test.go:11`, `structvalidator/map_path_privacy_test.go:12`), parser fuzzing (`types/fuzz_parser_test.go:16`), cache-key regressions (`core/serialize_rules_test.go:12`), manual slice/foreach malformed-input behavior (`types/manual_rules_test.go:10`), invalid regex privacy (`types/regex_privacy_test.go:12`), struct cross-field references (`structvalidator/advanced_struct_test.go:95`), global/per-instance extension behavior (`core/extensibility_test.go:15`), examples, and CI workflow coverage (`.github/workflows/.ci.yml:23`).

Weak: future-facing behavior is not tested because it is not implemented: broader conditional validation, first-class context-aware rule execution, and all-errors-per-field collection. Some builder methods, plugin type-validator paths, and `internal/pathutil` edge cases have limited direct coverage. The coverage run reported 70.9% total statement coverage with several 0% helper methods visible in builders and legacy validator adapters.

Checks run in this audit:

- `go test ./...` passed, using cache.
- `go test ./... -count=1` passed.
- `go vet ./...` passed.
- `scripts/fuzz.sh` passed for `FuzzParseTag`, `FuzzParseTagLong`, `FuzzBuildIntRules`, and `FuzzBuildStringRules`.
- `go test ./... -race -count=1` passed.
- `go test ./... -count=1 -coverpkg=./... -coverprofile=/tmp/validate-v3-coverpkg.out` passed.
- `go tool cover -func=/tmp/validate-v3-coverpkg.out` reported `total: (statements) 70.9%`.

I did not run `go mod tidy` because this audit phase made no dependency or module-file changes. I did not run `govulncheck` because no dependency surface changed and the requested audit checks were satisfied without installing additional tooling.

## Best next fixes
1. Add a complete public error-code reference generated from or manually synchronized with `errors/codes.go`.
2. Add direct tests for `internal/pathutil` map-key rendering policy, including nil, bool, numeric, complex, long, sensitive, and escaping-sensitive keys.
3. Document and test cache semantics for custom rule args that are not built-in primitive/map/slice/rule/function values.
4. Design the first additive conditional validation rule beyond `requiredWith`, with explicit ownership across parser, struct validation, builders/manual rules, docs, and examples.
5. Design an opt-in all-errors-per-field API that preserves current fail-fast defaults.
6. Decide whether context-aware validation should remain adapter-only or become a first-class compile/build/struct API.
7. Fill public builder and plugin helper coverage gaps where those methods are part of the root facade contract.

## Optional follow-up
- Targeted remediation plan
- Package-by-package review
- Security-focused pass for structured output, custom validators, and translation behavior
- Test-gap plan for builders, plugin helpers, and extension APIs
