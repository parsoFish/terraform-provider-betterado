# Fix Plan

> Checklist for WI-4. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a betterado_pipeline resource pointing to a YAML pipeline in the betterado-standing-demo project's existing git repo WHEN terraform apply runs live with TF_ACC=1 THEN the pipeline is created; provider reads it back with ExpectNonEmptyPlan:false; destroy removes it; TestAccPipeline passes live
  - Implemented in `resource_pipeline_test.go`: `TestAccPipeline_basic` uses `SharedFixtureProjectName`, references `betterado_git_repository` + `yaml_path=/azure-pipelines.yml`.
  - Fix (iter 2): added `repo_id` + `yaml_path` to resource schema; create now uses raw HTTP POST with extended JSON body to supply `configuration.path` and `configuration.repository`.
  - Fix (iter 3): delete now uses `BuildClient.DeleteDefinition` (Build Definitions API) instead of DELETE on _apis/pipelines which doesn't support it.
- [x] AC2: GIVEN the acceptance test completes its live read-back step before destroy WHEN CaptureLiveEvidence is called with label 'acceptance-resource' and the GET URL for the created pipeline THEN .forge/live-evidence/acceptance-resource.json is written with a real REST API response
  - `capturePipelineEvidence` calls `testutils.CaptureLiveEvidence("acceptance-resource", url, pipeline)` with GET URL `_apis/pipelines/{id}?api-version=7.1-preview.1`.
- [x] AC3: GIVEN a betterado_pipeline data source block reading the pipeline created by the betterado_pipeline resource in the same config (by project_id + id) WHEN terraform apply runs live with TF_ACC=1 THEN TestAccDataPipeline passes live: the data source returns name, folder and revision matching the resource; ExpectNonEmptyPlan:false
  - Implemented in `data_pipeline_test.go`: `TestAccDataPipeline_basic` creates the resource then reads via data source using `pipeline_id = tonumber(betterado_pipeline.test.id)`.
  - Fix (iter 2): HCL updated to include `repo_id` + `yaml_path`.
  - Fix (iter 3): delete fix applies here too — destroy was failing same way.

## Sub-tasks completed

- [x] Fix `resource_pipeline_authorization_test.go` — all HCL fixtures now use `data "betterado_project"` pointing to `SharedFixtureProjectName` instead of creating new projects.
- [x] Fix `resource_pipeline_framework.go` — `createPipelineRaw()` sends extended JSON body with `configuration.path` and `configuration.repository` required by ADO API.
- [x] Update HCL fixtures in resource + data pipeline tests to include `repo_id` + `yaml_path`.
- [x] Update docs + examples.
- [x] Fix `deletePipeline()` — route through `BuildClient.DeleteDefinition` instead of DELETE on _apis/pipelines (which returns 405).

## Outstanding

All ACs implemented. Awaiting live gate run to confirm pass.
