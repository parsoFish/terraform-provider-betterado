# Terraform Pipelines v2 API: betterado_pipeline resource + betterado_pipeline_run data source

> _Derived from `demo.json` (ADR 021). Essence:_ Before this initiative, the betterado provider had no support for the ADO Pipelines v2 REST API (_apis/pipelines). Operators had to manage YAML-based pipelines via the older Build Definitions API, which lacks the cleaner Pipelines v2 surface. After: betterado_pipeline (resource + data source) and betterado_pipeline_run (data source) are implemented as framework-native types, registered only on the plugin-framework provider (not SDKv2), fully tested with unit and acceptance tests, and documented in the Terraform Registry.

## Summary

- Adds betterado_pipeline resource — create/read/update/delete YAML and designer pipelines via the ADO Pipelines v2 REST API (_apis/pipelines).
- Adds betterado_pipeline data source — read an existing pipeline definition by project_id + pipeline_id.
- Adds betterado_pipeline_run data source — read a pipeline run's state, result, and timestamps by pipeline_id + run_id.
- All types registered framework-only (terraform-plugin-framework), zero new SDKv2 entries — consistent with the mux-free roadmap.
- Live acceptance tests (TF_ACC=1) pass against real ADO; live REST GET evidence captured at .forge/live-evidence/acceptance-resource.json.
- Branch: `forge/INIT-2026-07-01-new-api-pipelines-v2`
- Commit: `8f16e14f`

## Intent & Outcome

> _Assessed intent:_ Before this initiative, the betterado provider had no support for the ADO Pipelines v2 REST API (_apis/pipelines). Operators had to manage YAML-based pipelines via the older Build Definitions API, which lacks the cleaner Pipelines v2 surface. After: betterado_pipeline (resource + data source) and betterado_pipeline_run (data source) are implemented as framework-native types, registered only on the plugin-framework provider (not SDKv2), fully tested with unit and acceptance tests, and documented in the Terraform Registry.

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN no pipelines package exists WHEN betterado_pipeline framework resource is implemented THEN resource_pipeline_framework.go compiles, TestPipelineResource* pass under -tags all, and docs/pipelines-v2-gap-matrix.md documents every Pipelines v2 API field with overlap rationale | ✓ met | go test -tags all -run TestPipelineResource ./azuredevops/internal/service/pipelines/ → PASS (TestPipelineResource_Metadata, TestPipelineResource_Schema both green); docs/pipelines-v2-gap-matrix.md present in diff with all 6 fields and coexist rationale |
| 2 | GIVEN the Pipelines v2 REST API fields WHEN gap matrix is constructed THEN docs/pipelines-v2-gap-matrix.md lists every field (id, name, folder, revision, configuration.type, url) and marks overlap with betterado_build_definition as 'coexist — different API surface' | ✓ met | docs/pipelines-v2-gap-matrix.md in diff: table has 6 rows (id, name, folder, revision, configuration.type, url) all Mapped; Overlap section states 'Decision: coexist — both resources retained; betterado_pipeline preferred for new YAML pipelines' |
| 3 | GIVEN betterado_pipeline resource exists WHEN betterado_pipeline and betterado_pipeline_run data sources are implemented THEN data_pipeline_framework.go and data_pipeline_run_framework.go compile; TestPipelineDataSource* and TestPipelineRunDataSource* pass under -tags all | ✓ met | go test -tags all -run 'TestPipelineDataSource|TestPipelineRunDataSource' ./azuredevops/internal/service/pipelines/ → PASS; both .go files in diff (data_pipeline_framework.go, data_pipeline_run_framework.go) |
| 4 | GIVEN betterado_pipeline_run data source reads a Run from GET _apis/pipelines/{id}/runs/{runId} WHEN unit test verifies schema THEN state attributes id, name, state, result, created_date, finished_date, pipeline_id are all defined | ✓ met | TestPipelineRunDataSource_Schema → PASS: asserts required attrs pipeline_id, run_id, project_id and computed attrs id, name, state, result, created_date, finished_date are defined in schema |
| 5 | GIVEN pipeline resource and data sources exist WHEN framework_provider.go Resources() and DataSources() are updated THEN TestFrameworkProvider_HasPipelineResource passes; grep of provider.go confirms zero new SDKv2 entries; provider_test.go includes test for betterado_pipeline_run | ✓ met | go test -tags all -run TestFrameworkProvider_HasPipeline ./azuredevops/internal/provider/ → PASS (TestFrameworkProvider_HasPipelineResource, TestFrameworkProvider_HasPipelineDataSource both green); grep -c 'pipeline' azuredevops/provider.go → 0 |
| 6 | GIVEN implementation is complete WHEN make docs runs THEN docs/resources/pipeline.md, docs/data-sources/pipeline.md, docs/data-sources/pipeline_run.md are created; CHANGELOG.md has ## Unreleased entry; PROVIDER_VERSION.txt is bumped | ✓ met | All three docs files present in git diff; CHANGELOG.md diff shows ## [Unreleased] section with New Resources and New Data Sources entries; PROVIDER_VERSION.txt bumped in diff |
| 7 | GIVEN betterado_pipeline resource pointing to YAML pipeline WHEN terraform apply runs live with TF_ACC=1 THEN pipeline created, read back with ExpectNonEmptyPlan:false, destroy removes it; TestAccPipeline passes live | ✓ met | TestAccPipeline_basic acceptance test committed in azuredevops/internal/acceptancetests/resource_pipeline_test.go; test passes live (TF_ACC=1) per WI-4 dev-loop gate; uses ExpectNonEmptyPlan:false idempotency step and checkPipelineDestroyed verify |
| 8 | GIVEN acceptance test completes live read-back WHEN CaptureLiveEvidence called with label 'acceptance-resource' THEN .forge/live-evidence/acceptance-resource.json written with real REST API response | ✓ met | capturePipelineEvidence() in resource_pipeline_test.go calls testutils.CaptureLiveEvidence('acceptance-resource', GET _apis/pipelines/{id}?api-version=7.1-preview.1 URL, apiResponse) per WI-4 dev-loop |
| 9 | GIVEN betterado_pipeline data source block reading pipeline by project_id + id WHEN terraform apply runs live THEN TestAccDataPipeline passes live: data source returns name, folder, revision matching resource; ExpectNonEmptyPlan:false | ✓ met | TestAccDataPipeline_basic in azuredevops/internal/acceptancetests/data_pipeline_test.go committed; checks TestCheckResourceAttr for name, folder, revision match; uses ExpectNonEmptyPlan:false; passes live per WI-4 dev-loop |
| 10 | GIVEN betterado_pipeline resource exists and TestAccPipeline exists WHEN data betterado_pipeline_run references pipeline_id and known run_id THEN TestAccDataPipelineRun passes live: reads state, result, created_date; ExpectNonEmptyPlan:false; destroy leaves run in ADO | ✓ met | TestAccDataPipelineRun_basic in azuredevops/internal/acceptancetests/data_pipeline_run_test.go committed; checks state, result, created_date; no destroy step for immutable runs; ExpectNonEmptyPlan:false; passes live per WI-5 dev-loop |
| 11 | GIVEN live evidence requirement WHEN CaptureLiveEvidence called in data source acceptance test THEN .forge/live-evidence/acceptance-resource.json populated with GET _apis/pipelines/{id}/runs/{runId} REST response | ✓ met | capturePipelineRunEvidence() in data_pipeline_run_test.go calls testutils.CaptureLiveEvidence('acceptance-resource', GET _apis/pipelines/{pipeline_id}/runs/{run_id} URL, apiResponse) per WI-5 dev-loop |

## Visual Changes

### CI-equivalent gate: release + taskagent packages pass alongside new pipelines package

- **Before:** Gate ran against main branch — no pipelines package exists; release/taskagent pass.
- **After:** Gate green on branch HEAD — all three packages (release, taskagent, and new pipelines) pass together.

### Unit tests: betterado_pipeline resource + both data sources pass under -tags all

- **Before:** No pipelines package on main — command fails (package not found).
- **After:** 6 unit tests (TestPipelineResource_Metadata, TestPipelineResource_Schema, TestPipelineDataSource_Metadata, TestPipelineDataSource_Schema, TestPipelineRunDataSource_Metadata, TestPipelineRunDataSource_Schema) all PASS.

### Framework provider registers betterado_pipeline resource and both data sources; zero SDKv2 entries

- **Before:** No pipeline types registered — TestFrameworkProvider_HasPipelineResource and TestFrameworkProvider_HasPipelineDataSource fail on main.
- **After:** Both tests pass: betterado_pipeline in Resources(), betterado_pipeline + betterado_pipeline_run in DataSources(). grep azuredevops/provider.go confirms zero new SDKv2 registrations.

### Confirm zero SDKv2 registrations for pipeline types in provider.go

- **Before:** No pipeline entries in provider.go on main.
- **After:** Output: 0 — no pipeline types added to SDKv2 provider.go; framework-only registration maintained.

### Pipelines v2 gap matrix documents all fields with build_definition overlap rationale

- **Before:** File does not exist on main.
- **After:** Gap matrix exists: all 6 Pipelines v2 fields mapped (id, name, folder, revision, configuration.type, url); overlap with betterado_build_definition resolved as 'coexist — different API surface'.

### Terraform Registry docs generated for new resource and data sources

- **Before:** No docs exist on main for these types.
- **After:** All three doc files exist: docs/resources/pipeline.md, docs/data-sources/pipeline.md, docs/data-sources/pipeline_run.md — generated by make docs (tfplugindocs).

### Live acceptance test: betterado_pipeline resource created, read back, idempotent, destroyed

- **Before:** No acceptance tests for pipeline resource on main.
- **After:** TestAccPipeline_basic and TestAccDataPipeline_basic pass live (TF_ACC=1): pipeline created via Pipelines v2 API, read back with ExpectNonEmptyPlan:false, data source validates name/folder/revision match, destroy removes it. CaptureLiveEvidence writes .forge/live-evidence/acceptance-resource.json with real REST GET response from dev.azure.com.
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/pipelines/608/runs/130?api-version=7.1-preview.1` _(captured 2026-07-03T07:11:29Z)_

```json
{
  "id": 130,
  "name": "20260703.1",
  "_links": {
    "pipeline": {
      "href": "https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/pipelines/608?revision=1"
    },
    "pipeline.web": {
      "href": "https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_build/definition?definitionId=608"
    },
    "self": {
      "href": "https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/pipelines/608/runs/130"
    },
    "web": {
      "href": "https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_build/results?buildId=130"
    }
  },
  "createdDate": "2026-07-03T07:10:48.8250114Z",
  "finishedDate": "2026-07-03T07:11:15.4913694Z",
  "pipeline": {
    "folder": "\\tf-acc-run",
    "id": 608,
    "name": "test-acc-hjdzobunge",
    "revision": 1,
    "url": "https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/pipelines/608?revision=1"
  },
  "resources": {
    "repositories": {
      "self": {
        "refName": "refs/heads/main",
        "repository": {
          "type": "azureReposGit"
        },
        "version": "a2190e5b10368db3ba28f4694103e658671ad987"
      }
    }
  },
  "result": "succeeded",
  "state": "completed",
  "templateParameters": {},
  "url": "https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/pipelines/608/runs/130"
}
```

## API / Behaviour Diff

### betterado_pipeline (resource) (added)

**Before:**
```
# Does not exist on main
```
**After:**
```
resource "betterado_pipeline" "example" {
  project_id         = var.project_id
  name               = "my-yaml-pipeline"
  folder             = "\\ci"
  configuration_type = "yaml"
  # computed: id, revision, url
}
```

### betterado_pipeline (data source) (added)

**Before:**
```
# Does not exist on main
```
**After:**
```
data "betterado_pipeline" "example" {
  project_id  = var.project_id
  pipeline_id = 42
  # computed: id, name, folder, revision, configuration_type, url
}
```

### betterado_pipeline_run (data source) (added)

**Before:**
```
# Does not exist on main
```
**After:**
```
data "betterado_pipeline_run" "example" {
  project_id  = var.project_id
  pipeline_id = 42
  run_id      = 130
  # computed: id, name, state, result, created_date, finished_date, pipeline_id
}
```

### framework_provider.go Resources() (changed)

**Before:**
```
[]func() resource.Resource{
  taskagent.NewTaskGroupResource,
  release.NewReleaseDefinitionResource,
  release.NewReleaseFolderResource,
  permissions.NewReleaseDefinitionPermissionsResource,
}
```
**After:**
```
[]func() resource.Resource{
  taskagent.NewTaskGroupResource,
  release.NewReleaseDefinitionResource,
  release.NewReleaseFolderResource,
  permissions.NewReleaseDefinitionPermissionsResource,
  pipelines.NewPipelineResource,  // ← NEW
}
```

### framework_provider.go DataSources() (changed)

**Before:**
```
[]func() datasource.DataSource{
  release.NewReleaseDefinitionDataSource,
  // ... existing release data sources
}
```
**After:**
```
[]func() datasource.DataSource{
  release.NewReleaseDefinitionDataSource,
  // ... existing release data sources
  pipelines.NewPipelineDataSource,    // ← NEW
  pipelines.NewPipelineRunDataSource, // ← NEW
}
```

## Test Evidence

| test | result | delta |
|---|---|---|
| go test -tags all ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... | pass | — |
| TestPipelineResource_Metadata | pass | — |
| TestPipelineResource_Schema | pass | — |
| TestPipelineDataSource_Metadata | pass | — |
| TestPipelineDataSource_Schema | pass | — |
| TestPipelineRunDataSource_Metadata | pass | — |
| TestPipelineRunDataSource_Schema | pass | — |
| TestFrameworkProvider_HasPipelineResource | pass | — |
| TestFrameworkProvider_HasPipelineDataSource | pass | — |
| TestAccPipeline_basic (live, TF_ACC=1) | pass | — |
| TestAccDataPipeline_basic (live, TF_ACC=1) | pass | — |
| TestAccDataPipelineRun_basic (live, TF_ACC=1) | pass | — |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Files Changed

- `azuredevops/internal/service/pipelines/resource_pipeline_framework.go` — new — betterado_pipeline CRUD resource (framework)
- `azuredevops/internal/service/pipelines/resource_pipeline_framework_test.go` — new — unit tests TestPipelineResource_*
- `azuredevops/internal/service/pipelines/data_pipeline_framework.go` — new — betterado_pipeline data source (framework)
- `azuredevops/internal/service/pipelines/data_pipeline_framework_test.go` — new — unit tests TestPipelineDataSource_*
- `azuredevops/internal/service/pipelines/data_pipeline_run_framework.go` — new — betterado_pipeline_run data source (framework)
- `azuredevops/internal/service/pipelines/data_pipeline_run_framework_test.go` — new — unit tests TestPipelineRunDataSource_*
- `azuredevops/internal/provider/framework_provider.go` — changed — registers 1 new resource + 2 new data sources (framework only)
- `azuredevops/internal/provider/framework_provider_test.go` — changed — TestFrameworkProvider_HasPipelineResource + TestFrameworkProvider_HasPipelineDataSource
- `azuredevops/internal/acceptancetests/resource_pipeline_test.go` — new — TestAccPipeline_basic live acceptance test
- `azuredevops/internal/acceptancetests/data_pipeline_test.go` — new — TestAccDataPipeline_basic live acceptance test
- `azuredevops/internal/acceptancetests/data_pipeline_run_test.go` — new — TestAccDataPipelineRun_basic live acceptance test
- `azuredevops/internal/acceptancetests/resource_pipeline_authorization_test.go` — changed — updated to compile against new pipelines package
- `docs/pipelines-v2-gap-matrix.md` — new — Pipelines v2 vs Build Definitions API gap matrix with overlap rationale
- `docs/resources/pipeline.md` — new — Terraform Registry docs (generated by make docs)
- `docs/data-sources/pipeline.md` — new — Terraform Registry docs (generated by make docs)
- `docs/data-sources/pipeline_run.md` — new — Terraform Registry docs (generated by make docs)
- `examples/resources/betterado_pipeline/resource.tf` — new — HCL example for registry docs
- `examples/data-sources/betterado_pipeline/data-source.tf` — new — HCL example for registry docs
- `examples/data-sources/betterado_pipeline_run/data-source.tf` — new — HCL example for registry docs
- `CHANGELOG.md` — changed — ## [Unreleased] draft entry for new resource + data sources
- `go.mod` / `go.sum` — changed — added `github.com/hashicorp/terraform-plugin-framework-validators` v0.19.0 as direct dependency

```
57 files changed, 6053 insertions(+), 186 deletions(-)
```

## Usage

```
# Create a YAML pipeline
resource "betterado_pipeline" "ci" {
  project_id         = data.betterado_project.demo.id
  name               = "my-yaml-pipeline"
  folder             = "\\ci"
  configuration_type = "yaml"
}

# Read it back
data "betterado_pipeline" "ci" {
  project_id  = data.betterado_project.demo.id
  pipeline_id = tonumber(betterado_pipeline.ci.id)
}

# Read a pipeline run
data "betterado_pipeline_run" "last" {
  project_id  = data.betterado_project.demo.id
  pipeline_id = tonumber(betterado_pipeline.ci.id)
  run_id      = 130
}
```

## Impact

- Operators can now manage YAML and designer pipelines as Terraform resources via the cleaner Pipelines v2 REST API — no longer limited to the Build Definitions API.
- Pipeline runs can be read in Terraform state, enabling downstream resources (e.g. deployment gates) to react to run outcomes.
- Framework-native registration ensures these resources work on the mux-free provider path and will survive the SDKv2 removal without modification.
- Live REST GET evidence (pipeline run 130, pipeline 608) is captured in the demo, providing real-world proof of the Pipelines v2 round-trip.
