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

### Iteration 2 (WI-5)

- **Gate failure:** Live gate produced `pipeline_id = ` / `run_id = ` (empty) in `terraform_plugin_test.tf` → `Invalid expression` error.
- **Root cause:** `resource.TestStep.Config` is a **string field** baked at initialization time. The `Steps` slice is built before any step runs. When `hclPipelineRunDataSource(name, &pipelineIDStr, &runIDStr)` is called to produce the string for Step 2's `Config`, the pointers point to their variables but `*pipelineIDStr == ""` and `*runIDStr == ""` at that moment. The string is captured with empty values and never updated.
- **Fix:** Changed Steps 2 and 3 to use `ConfigFile` instead of `Config`.
  - `ConfigFile` is a `func(config.TestStepConfigRequest) string` evaluated **lazily per-step** in the test runner loop (not at initialization time).
  - Step 1's Check (`triggerAndWaitForRun`) now writes the full HCL to a temp file (`os.CreateTemp`) and sets `*configFilePath`.
  - Steps 2 and 3 use `hclDataSourceConfigFile := func(_ config.TestStepConfigRequest) string { return configFilePath }`.
  - By the time Step 2's execution starts, Step 1's Check has fully run and `configFilePath` is set.
- **Provider block requirement:** `ConfigFile` bypasses the framework's `mergedConfig` injection (which adds `terraform { required_providers {} }` only for `Config` string steps). Added the block explicitly in `hclPipelineRunDataSourceContent` using `TF_ACC_PROVIDER_HOST`/`TF_ACC_PROVIDER_NAMESPACE` env vars (defaulting to `registry.terraform.io/hashicorp/betterado`). When `HasTerraformBlock = true`, `mergedConfig` returns `s.Config = ""`, so `confRequest.Raw = ""`, and `Configuration(confRequest)` uses File (our temp file) over Raw. ✓
- **Framework internals:**
  - `TestStep.Config string` → evaluated at struct build time.
  - `TestStep.ConfigFile TestStepConfigFunc` → evaluated lazily at step loop iteration time (before `PreConfig`, which is before `testStepNewConfig`).
  - `Configuration(req)` priority: Directory > File > Raw.
  - `mergedConfig` early-returns when `hasTerraformBlock = true`, returning `""`.
- **Lint:** 0 issues.
- **go vet:** passes.

## What didn't work

- Iteration 1: `Config` string with pointer dereference — string is baked at initialization time, pointers always empty.
- Iteration 1: `hclPipelineRunDataSource(name, &pipelineIDStr, &runIDStr)` — looks lazy but isn't; the return value is computed immediately.

## Open questions

- Will the azure-pipelines.yml in `betterado-standing-demo` succeed quickly enough (under 3 min) for the gate to pass?
- If the run is still `inProgress` after 3 min, the data source step may fail because the `result` attribute is empty (which is expected for in-progress runs) — but `state` and `created_date` are always set.

## Notes for reflection

- **NEVER use Config string for IDs that come from a prior step's Check.** Use `ConfigFile` with a closure + a temp file written in the prior step's Check. The `ConfigFile` func is lazy; `Config` string is not.
- The provider block (`terraform { required_providers {} }`) MUST be included in `ConfigFile` temp files because the framework's `mergedConfig` injection only applies to `Config` string steps.
- `TF_ACC_PROVIDER_HOST` and `TF_ACC_PROVIDER_NAMESPACE` env vars control the provider source address (defaults: `registry.terraform.io`/`hashicorp`).
