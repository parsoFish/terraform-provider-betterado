# Agent Memory — WI-5

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (WI-5)

- **Gate failure:** `go test -tags all -run TestAccDataPipelineRun ./azuredevops/internal/acceptancetests/` exited with `[no tests to run]` — the test file didn't exist yet.
- **Action:** Created `azuredevops/internal/acceptancetests/data_pipeline_run_test.go` with `TestAccDataPipelineRun`.
- **Commit:** `feat(WI-5): add TestAccDataPipelineRun live acceptance test for betterado_pipeline_run data source`
- **Build tag used:** `//go:build (all || data_source_pipeline) && !exclude_data_source_pipeline`
- **Test strategy:** Option B (trigger a run imperatively). Steps:
  1. Step 1: create `betterado_pipeline` resource + trigger run in `triggerAndWaitForRun` Check func (polls up to 3 min)
  2. Step 2: read completed run via `betterado_pipeline_run` data source; asserts state/created_date; captures live evidence
  3. Step 3: idempotency (`ExpectNonEmptyPlan:false`, `PlanOnly:true`)
  - No `CheckDestroy` — runs are immutable in ADO
- **HCL design:** Data source config in Step 2 is parameterized via pointer dereference (`hclPipelineRunDataSource(name, &pipelineIDStr, &runIDStr)`) — pointers are set by `triggerAndWaitForRun` in Step 1's Check, then lazily read in Step 2's Config call.
- **Lint:** 0 issues from golangci-lint `--new-from-rev=main`.
- **make test:** All offline tests pass, no FAILs.
- **terrafmt:** Passes.

## What worked

- Build tag `(all || data_source_pipeline) && !exclude_data_source_pipeline` matches the WI spec.
- Using pointer variables (`*string`) to pass pipelineID/runID from Step 1's Check to Step 2's Config (evaluated lazily at test-step execution time).
- `adoPipelines.RunPipelineParameters{}` (empty struct) is sufficient to trigger a run.
- `adoPipelines.RunStateValues.Completed` for polling termination condition.

## What didn't work

_(none — first iteration created the file cleanly)_

## Open questions

- Will the azure-pipelines.yml in `betterado-standing-demo` succeed quickly enough (under 3 min) for the gate to pass?
- If the run is still `inProgress` after 3 min, the data source step may fail because the `result` attribute is empty (which is expected for in-progress runs) — but `state` and `created_date` are always set.

## Notes for reflection

- The `hclPipelineRunDataSource` function uses pointer dereferencing which is a pattern not seen elsewhere in the test suite — worth documenting in brain if it recurs.
