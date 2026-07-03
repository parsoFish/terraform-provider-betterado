# Fix Plan

> Checklist for WI-5. Tick items as you complete them; add items as you discover sub-problems.

- [x] Create `azuredevops/internal/acceptancetests/data_pipeline_run_test.go` with `TestAccDataPipelineRun`
- [ ] AC1: GIVEN betterado_pipeline resource exists and a TestAccPipeline test exists that creates a live pipeline and its run WHEN a data betterado_pipeline_run block references the pipeline's ID and a known run_id THEN TestAccDataPipelineRun passes live: data source reads state, result, created_date; ExpectNonEmptyPlan:false; destroy step leaves the run record in ADO (runs are immutable)
  - [x] Test file created with build tag `(all || data_source_pipeline) && !exclude_data_source_pipeline`
  - [x] Creates pipeline via `betterado_pipeline` resource
  - [x] Triggers a run via `RunPipeline` + polls until completed (3-min timeout)
  - [x] Reads run via `betterado_pipeline_run` data source
  - [x] Asserts `state` + `created_date` are set
  - [x] `ExpectNonEmptyPlan:false` in Step 3
  - [x] No `CheckDestroy` (runs are immutable in ADO)
  - [x] Fixed `pipeline_id = ` / `run_id = ` empty-value bug (iter 2: ConfigFile + temp file pattern)
  - [ ] **Live gate must pass** (awaiting forge live TF_ACC run)
- [ ] AC2: GIVEN the live evidence requirement WHEN CaptureLiveEvidence is called THEN .forge/live-evidence/acceptance-resource.json is populated
  - [x] `capturePipelineRunEvidence` calls `testutils.CaptureLiveEvidence("acceptance-resource", url, r)`
  - [ ] **Live gate must run to actually write the file**
