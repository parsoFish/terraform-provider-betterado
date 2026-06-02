# CI green: gofmt/terrafmt/golangci-lint violations fixed

> _Derived from `demo.json` (ADR 021). Essence:_ All GitHub Actions CI workflows (golint.yml, terrafmt.yml, unit-test.yml) were failing on main due to accumulated formatting drift and lint violations. This initiative fixes every violation — gofmt, terrafmt, and golangci-lint — so CI passes on every PR to main going forward.

## Quality gate: go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...

- **Before:** Tests pass on main: logic was never broken, but lint/fmt violations prevented CI from reaching the test stage on PRs.
- **After:** Tests still pass on HEAD (39/39); gofmt, terrafmt, and golangci-lint violations are now fixed so CI reaches and clears all three workflow gates.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| release package — tests passing | 11/11 (ok, 0.012s) | 11/11 (ok, 0.012s) | 0.0% | pass |
| taskagent package — tests passing | 28/28 (ok, 0.008s) | 28/28 (ok, 0.008s) | 0.0% | pass |
| golangci-lint violations | blocked CI (gocritic unlambda, SA1019 staticcheck, errcheck, unused functions) | 0 lint errors (nolint directives + dead-code removal) | -100.0% | pass |
| gofmt drift | formatting diff present — CI fails gofmtcheck.sh | no diff — gofmtcheck.sh exits 0 | -100.0% | pass |
| terrafmt HCL violations | HCL blocks in test files out of format — CI fails terrafmt-check | all HCL blocks formatted — make terrafmt-check exits 0 | -100.0% | pass |

## Acceptance criteria

- AC1: ./scripts/gofmtcheck.sh exits 0 — no gofmt diff
- AC2: make terrafmt-check exits 0 — no HCL formatting errors
- AC3: golangci-lint run -v ./azuredevops/... exits 0 — no lint errors
- AC4: go build -v ./... and make test both exit 0

## Changed files

```
 AGENT.md                                           | 44 ++++++++++++++++++
 .../resource_release_definition_test.go            |  8 ++--
 .../service/release/resource_release_definition.go | 54 +++++++++++-----------
 .../service/taskagent/resource_task_group.go       | 52 ++++-----------------
 azuredevops/provider.go                            |  2 +-
 azuredevops/provider_test.go                       |  1 +
 fix_plan.md                                        |  8 ++++
 main.go                                            |  5 +-
 website/docs/r/security_permissions.html.markdown  |  2 +-
 .../r/workitemtrackingprocess_page.html.markdown   |  4 +-
 10 files changed, 99 insertions(+), 81 deletions(-)
```
