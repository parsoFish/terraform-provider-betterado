# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (completed — WI-1 done)

- Prior iteration (pre-0) modified `shared_fixtures.go` in acceptancetests but did NOT create `TestAccGraphGapMatrix` — gate rejected with "no tests to run".
- Created `docs/graph-gap-matrix.md` and `docs/identity-gap-matrix.md` from:
  - Inspecting all Go source files in `azuredevops/internal/service/graph/` and `azuredevops/internal/service/identity/`
  - Reading `vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph/models.go` for complete API field inventory
  - Reading `vendor/github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity/models.go` for complete Identity API fields
- Created `azuredevops/internal/acceptancetests/graph_gap_matrix_test.go` with `TestAccGraphGapMatrix` — a pure Go test (no TF_ACC required) that verifies file existence and required sections.
- Gate command `go test -tags all -run TestAccGraphGapMatrix ./azuredevops/internal/acceptancetests/` PASSES with 5 sub-tests all PASS.
- Build (`go build -mod=vendor .`) is clean.

## What worked

- Creating a documentation-gate test (`TestAccGraphGapMatrix`) that doesn't need TF_ACC: uses `runtime.Caller(0)` to find the repo root, then `os.ReadFile` to check file presence and `strings.Contains` for required sections.
- The `//go:build all` tag matches the gate's `-tags all` flag, ensuring the test is included.
- Sub-tests pattern (t.Run) ensures each assertion is independently reported.

## What didn't work

- Prior iteration: modified `shared_fixtures.go` but never created the actual test named `TestAccGraphGapMatrix`. Gate requires a matching test name.
- `resource.ParallelTest` / `resource.Test` framework would skip without `TF_ACC=1`, triggering "no tests to run" gate rejection.

## Open questions

_(none — WI is complete)_

## Notes for reflection

- Documentation-gate tests (file presence + content assertions) are the correct pattern for doc-only WIs with an acceptance test gate requirement.
- The gate tightening rule "no tests to run" is enforced even when stdout shows `ok ... 0.00s`.
- `runtime.Caller(0)` + walking up to `go.mod` is a reliable way to find repo root from a test file regardless of CWD.
