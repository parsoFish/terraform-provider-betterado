# CI green: gofmt, golangci-lint, and unit tests all pass

> _Derived from `demo.json` (ADR 021). Essence:_ Accumulated formatting drift and lint violations were blocking all PR merges via CI. After applying gofmt/gofumpt sweeps, golangci-lint auto-fix, and targeted nolint suppressions, all three CI checks (gofmt, golangci-lint, unit tests) now exit 0.

## go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... exits 0

- **Before:** Unit tests could not be confirmed passing against the unformatted/unlinted tree; golangci-lint and gofmt violations blocked CI, making the build state unreliable.
- **After:** All 45 unit tests across 3 packages (service/release: 11, service/taskagent: 33, service/taskagent/validate: 1) pass with exit 0. No regressions introduced by the formatting and nolint fixes.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| unit test exit code | unreliable (lint-blocked CI) | 0 (all pass) | — | lower-is-better |
| tests passing | unknown | 45 | — | higher-is-better |
| packages covered | 0 (CI blocked) | 3 | — | higher-is-better |

## golangci-lint run -v ./azuredevops/... exits 0 — all lint errors resolved

- **Before:** golangci-lint reported SA1019 (deprecated API usage), wastedassign, and formatting violations across resource_release_definition.go, resource_task_group.go, and provider.go. CI was failing on every PR.
- **After:** golangci-lint exits 0 with no errors. SA1019 usages suppressed with //nolint:staticcheck directives where deferred by operator decision; dead assignments removed; gofmt/gofumpt formatting corrected.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| golangci-lint exit code | 1 (lint errors) | 0 (clean) | — | lower-is-better |
| lint violations | >0 (SA1019 + wastedassign + format) | 0 | — | lower-is-better |

## scripts/gofmtcheck.sh exits 0 — no formatting drift

- **Before:** scripts/gofmtcheck.sh exited non-zero; multiple .go files had formatting drift that blocked CI on every push.
- **After:** scripts/gofmtcheck.sh exits 0 with no diff output. All Go source files are correctly formatted per gofmt/gofumpt.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| gofmtcheck exit code | 1 (files differ) | 0 (clean) | — | lower-is-better |

## Acceptance criteria

- scripts/gofmtcheck.sh exits 0 with no diff output
- make terrafmt-check exits 0 with no HCL formatting errors
- golangci-lint run -v ./azuredevops/... exits 0
- go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... exits 0

## Changed files

```
 AGENT.md                                           |  33 +++++++
 .../resource_release_definition_test.go            |   8 +-
 .../service/release/resource_release_definition.go |  54 +++++-----
 .../service/taskagent/resource_task_group.go       |  52 ++--------
 azuredevops/provider.go                            |   2 +-
 azuredevops/provider_test.go                       |   1 +
 demo/INIT-2026-06-01-ci-green/DEMO.html            | 109 +++++++++++++++++++++
 demo/INIT-2026-06-01-ci-green/DEMO.md              |  54 ++++++++++
 demo/INIT-2026-06-01-ci-green/demo.json            |  86 ++++++++++++++++
 fix_plan.md                                        |   9 ++
 10 files changed, 334 insertions(+), 74 deletions(-)
```
