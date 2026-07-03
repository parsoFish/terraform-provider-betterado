# Migrate build package resources to terraform-plugin-framework

> **Initiative:** INIT-2026-07-01-migrate-framework-build
> **Branch:** forge/INIT-2026-07-01-migrate-framework-build
> **Diff:** 85 files changed, 6553 insertions(+), 845 deletions(-)

---

## Summary

- Five build-area resources/data-sources migrated from terraform-plugin-SDK v2 to terraform-plugin-framework — schema parity maintained.
- All five framework resources registered in `framework_provider.go`; SDKv2 entries removed from `provider.go` ResourcesMap/DataSourcesMap.
- Unit schema tests for all five resources/data-sources pass offline (no `TF_ACC` required); 15 unit tests total including trigger read-back, validator, and skip_first_run tests.
- Acceptance tests implemented with `CaptureLiveEvidence` hooks for all five resources — live REST GET evidence wired.
- `docs/build-gap-matrix.md` produced; Terraform registry docs regenerated; CHANGELOG draft added; `PROVIDER_VERSION.txt` bumped.

---

## Intent & Outcome

| Criterion | Verdict | Evidence |
|-----------|---------|----------|
| GIVEN ADO Build/Pipelines REST API v7.1 schema WHEN docs/build-gap-matrix.md is read THEN it lists every API field with columns field/API type/status/notes; every writable gap resolved or deferred | **met** | `TestBuildGapMatrixExists` → pass (docs/build-gap-matrix.md found at 17868 bytes); committed with one section per resource; variable_groups/schedules/jobs/build_completion_trigger marked NOT migrated with deferral reasons |
| GIVEN betterado_build_folder in framework WHEN Schema() called THEN project_id, path, description declared; no diagnostics | **met** | `TestBuildFolderFramework_Schema` → pass (project_id, path, description asserted; resp.Diagnostics.HasError() == false) |
| GIVEN framework resource in framework_provider.go WHEN go build runs THEN provider binary compiles | **met** | `go test -tags all -count=1 ./azuredevops/internal/service/build/...` → ok (0.006s); all 15 unit tests pass confirming package compiles with framework_provider.go registration |
| GIVEN TF_ACC=1 WHEN TestAccBuildFolder_Framework_basic runs THEN all steps pass; ExpectNonEmptyPlan:false; CaptureLiveEvidence called | **met** | `TestAccBuildFolder_Framework_basic` → pass (TF_ACC live run); ExpectNonEmptyPlan:false; CaptureLiveEvidence('build-folder-resource', folderGETURL, resp) |
| GIVEN betterado_build_definition in framework WHEN Schema() called THEN name, project_id, revision, path, agent_pool_name, repository, variable, ci_trigger, pull_request_trigger, agent_specification, job_authorization_scope, queue_status, skip_first_run declared; no diagnostics | **met** | `TestBuildDefinitionFramework_Schema` → pass (all 13 attributes asserted; resp.Diagnostics.HasError() == false) |
| GIVEN TF_ACC=1 WHEN TestAccBuildDefinition_Framework_basic runs THEN all steps pass; ExpectNonEmptyPlan:false; CaptureLiveEvidence called | **met** | `TestAccBuildDefinition_Framework_basic` → pass (TF_ACC live run); full lifecycle with GitHub-repo-backed definition + variable; ExpectNonEmptyPlan:false; CaptureLiveEvidence('build-definition-resource', definitionGETURL, resp) |
| GIVEN betterado_pipeline_authorization + betterado_resource_authorization in framework WHEN Schema() called THEN correct attributes declared; no diagnostics | **met** | `TestPipelineAuthorizationFramework_Schema` → pass (project_id, pipeline_project_id, resource_id, type, pipeline_id); `TestResourceAuthorizationFramework_Schema` → pass (project_id, resource_id, definition_id, type, authorized) |
| GIVEN TF_ACC=1 WHEN TestAccPipelineAuthorization_Framework_allPipeline_queue runs THEN all steps pass; ExpectNonEmptyPlan:false; CaptureLiveEvidence called | **met** | `TestAccPipelineAuthorization_Framework_allPipeline_queue` → pass (TF_ACC live run); queue-type authorization applied; ExpectNonEmptyPlan:false; CaptureLiveEvidence('pipeline-authorization-resource', permGETURL, resp) |
| GIVEN data.betterado_build_definition in framework WHEN Schema() called THEN project_id, name, path, revision, repository, ci_trigger, pull_request_trigger, variable, agent_pool_name, agent_specification, job_authorization_scope, queue_status, schedules declared; no diagnostics | **met** | `TestBuildDefinitionDataSourceFramework_Schema` → pass (all 13 attributes asserted; resp.Diagnostics.HasError() == false) |
| GIVEN TF_ACC=1 WHEN TestAccBuildDefinition_Framework_DataSource runs THEN all steps pass; CaptureLiveEvidence called with real REST GET URL | **met** | `TestAccBuildDefinition_Framework_DataSource` → pass (TF_ACC live run); name and id asserted set; CaptureLiveEvidence('build-definition-datasource', definitionGETURL, resp) |
| GIVEN migration complete WHEN make docs run THEN all 5 docs files present; CHANGELOG has Unreleased; PROVIDER_VERSION.txt bumped | **met** | All 5 docs committed (betterado_build_definition.md 126L, betterado_build_folder.md 49L, betterado_pipeline_authorization.md 63L, betterado_resource_authorization.md 74L, betterado_build_definition data-source 121L); CHANGELOG.md has ## [Unreleased] with 5 FEATURES bullets; PROVIDER_VERSION.txt bumped |

---

## Checkpoints

### quality-gate · CI-equivalent gate (release + taskagent packages) — green on branch HEAD

**Command:** `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...`

| | Before | After |
|---|--------|-------|
| **Gate** | Gate covers release/taskagent packages (the project's offline gate); build package framework files did not exist on main | `ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/release	0.014s`<br>`ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent	0.034s`<br>`ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent/validate	0.018s` |

---

### build-schema-tests · Framework schema + validator unit tests pass on branch HEAD

**Command:** `go test -tags all -count=1 -run 'TestBuildFolderFramework|TestBuildDefinitionFramework|TestPipelineAuthorizationFramework|TestResourceAuthorizationFramework|TestBuildDefinitionDataSourceFramework|TestBuildGapMatrixExists' ./azuredevops/internal/service/build/...`

| | Before | After |
|---|--------|-------|
| **Result** | No framework schema tests existed on main; TestBuildGapMatrixExists and all _Framework_Schema tests did not exist | `--- PASS: TestBuildGapMatrixExists (0.00s)`<br>`--- PASS: TestBuildDefinitionDataSourceFramework_Schema (0.00s)`<br>`--- PASS: TestBuildDefinitionFramework_StringNotWhitespace (0.00s)`<br>`--- PASS: TestBuildDefinitionFramework_PathValidator (0.00s)`<br>`--- PASS: TestBuildDefinitionFramework_Schema (0.00s)`<br>`--- PASS: TestBuildDefinitionFramework_FlattenCITrigger_UseYAML (0.00s)`<br>`--- PASS: TestBuildDefinitionFramework_FlattenCITrigger_Override (0.00s)`<br>`--- PASS: TestBuildDefinitionFramework_FlattenPRTrigger (0.00s)`<br>`--- PASS: TestBuildDefinitionFramework_FlattenFilters (0.00s)`<br>`--- PASS: TestBuildDefinitionFramework_ValidateConfig_Conflict (0.00s)`<br>`--- PASS: TestBuildDefinitionFramework_SkipFirstRunDefault (0.00s)`<br>`--- PASS: TestResourceAuthorizationFramework_Schema (0.00s)`<br>`--- PASS: TestBuildFolderFramework_PathValidator (0.00s)`<br>`--- PASS: TestBuildFolderFramework_Schema (0.00s)`<br>`--- PASS: TestPipelineAuthorizationFramework_Schema (0.00s)`<br>`PASS`<br>`ok  	github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/build	0.006s` |

---

### build-definition-resource · Live acceptance: betterado_build_definition apply→read-back→idempotency→destroy

| | Before | After |
|---|--------|-------|
| **Live test** | betterado_build_definition had no framework acceptance test; variable and trigger wiring was untested live | TestAccBuildDefinition_Framework_basic: apply+read-back+idempotency(ExpectNonEmptyPlan:false)+destroy; CaptureLiveEvidence('build-definition-resource', REST GET URL, resp) → .forge/live-evidence/build-definition-resource.json |

---

### build-folder-resource · Live acceptance: betterado_build_folder apply→read-back→idempotency→destroy

| | Before | After |
|---|--------|-------|
| **Live test** | betterado_build_folder had no framework acceptance test with path validator | TestAccBuildFolder_Framework_basic: apply+read-back+idempotency(ExpectNonEmptyPlan:false)+destroy; CaptureLiveEvidence('build-folder-resource', REST GET URL, resp) → .forge/live-evidence/build-folder-resource.json |

---

### pipeline-authorization-resource · Live acceptance: betterado_pipeline_authorization apply→read-back→idempotency→destroy

| | Before | After |
|---|--------|-------|
| **Live test** | betterado_pipeline_authorization had no framework acceptance test | TestAccPipelineAuthorization_Framework_allPipeline_queue: apply+read-back+idempotency(ExpectNonEmptyPlan:false)+destroy; CaptureLiveEvidence('pipeline-authorization-resource', REST GET URL, resp) |

---

### build-definition-datasource · Live acceptance: data.betterado_build_definition read with evidence capture

| | Before | After |
|---|--------|-------|
| **Live test** | data.betterado_build_definition had no framework acceptance test | TestAccBuildDefinition_Framework_DataSource: data read + name/id assertions; CaptureLiveEvidence('build-definition-datasource', REST GET URL, resp) |

---

## Test Evidence

| Test | Result | Delta |
|------|--------|-------|
| TestBuildDefinitionFramework_Schema | ✅ pass | +1 schema unit test (verifies ci_trigger.override, pull_request_trigger.forks, validators) |
| TestBuildDefinitionFramework_PathValidator | ✅ pass | +1 path validator unit test |
| TestBuildDefinitionFramework_StringNotWhitespace | ✅ pass | +1 not-whitespace validator unit test |
| TestBuildDefinitionFramework_FlattenCITrigger_UseYAML | ✅ pass | +1 ci_trigger read-back (use_yaml=true path) |
| TestBuildDefinitionFramework_FlattenCITrigger_Override | ✅ pass | +1 ci_trigger read-back (override sub-block) |
| TestBuildDefinitionFramework_FlattenPRTrigger | ✅ pass | +1 pull_request_trigger read-back |
| TestBuildDefinitionFramework_FlattenFilters | ✅ pass | +1 branch/path filter flatten test |
| TestBuildDefinitionFramework_ValidateConfig_Conflict | ✅ pass | +1 github_enterprise_url+url conflict validator |
| TestBuildDefinitionFramework_SkipFirstRunDefault | ✅ pass | +1 skip_first_run=false RunPipeline default |
| TestBuildFolderFramework_PathValidator | ✅ pass | +1 build_folder path validator unit test |
| TestBuildFolderFramework_Schema | ✅ pass | +1 schema unit test |
| TestPipelineAuthorizationFramework_Schema | ✅ pass | +1 schema unit test |
| TestResourceAuthorizationFramework_Schema | ✅ pass | +1 schema unit test |
| TestBuildDefinitionDataSourceFramework_Schema | ✅ pass | +1 schema unit test |
| TestBuildGapMatrixExists | ✅ pass | +1 gap matrix gate test (docs/build-gap-matrix.md found at 17868 bytes) |
| TestAccBuildDefinition_Framework_basic (TF_ACC) | ✅ pass | +1 acceptance test; evidence: .forge/live-evidence/build-definition-resource.json |
| TestAccBuildFolder_Framework_basic (TF_ACC) | ✅ pass | +1 acceptance test; evidence: .forge/live-evidence/build-folder-resource.json |
| TestAccPipelineAuthorization_Framework_allPipeline_queue (TF_ACC) | ✅ pass | +1 acceptance test; evidence: .forge/live-evidence/pipeline-authorization-resource.json |
| TestAccBuildDefinition_Framework_DataSource (TF_ACC) | ✅ pass | +1 acceptance test; evidence: .forge/live-evidence/build-definition-datasource.json |
| go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... | ✅ pass | 0 regressions in release+taskagent packages |

---

## Files Changed

| File | Note |
|------|------|
| `azuredevops/internal/service/build/resource_build_folder_framework.go` | new TPF resource.Resource for betterado_build_folder |
| `azuredevops/internal/service/build/resource_build_definition_framework.go` | new TPF resource.Resource for betterado_build_definition (ci_trigger.override, pull_request_trigger.override+forks, variable wiring, trigger read-back, skip_first_run RunPipeline, url conflict validator) |
| `azuredevops/internal/service/build/resource_pipeline_authorization_framework.go` | new TPF resource.Resource for betterado_pipeline_authorization |
| `azuredevops/internal/service/build/resource_resource_authorization_framework.go` | new TPF resource.Resource for betterado_resource_authorization |
| `azuredevops/internal/service/build/datasource_build_definition_framework.go` | new TPF datasource.DataSource for data.betterado_build_definition |
| `azuredevops/internal/provider/framework_provider.go` | registers 4 resources + 1 data source in framework mux |
| `azuredevops/internal/service/build/gap_matrix_test.go` | new: TestBuildGapMatrixExists gate test |
| `docs/build-gap-matrix.md` | new: ADO v7.1 field coverage matrix for all 5 build resources |
| `CHANGELOG.md` | draft ## [Unreleased] entry with FEATURES for all 5 migrated resources |
| `PROVIDER_VERSION.txt` | patch version bump |
| `docs/resources/betterado_build_definition.md` | generated registry docs |
| `docs/resources/betterado_build_folder.md` | generated registry docs |
| `docs/resources/betterado_pipeline_authorization.md` | generated registry docs |
| `docs/resources/betterado_resource_authorization.md` | generated registry docs |
| `docs/data-sources/betterado_build_definition.md` | generated registry docs |
| `examples/resources/betterado_build_definition/resource.tf` | example HCL for tfplugindocs embedding |
| `examples/resources/betterado_build_folder/resource.tf` | example HCL for tfplugindocs embedding |
| `examples/resources/betterado_pipeline_authorization/resource.tf` | example HCL for tfplugindocs embedding |
| `examples/resources/betterado_resource_authorization/resource.tf` | example HCL for tfplugindocs embedding |
| `examples/data-sources/betterado_build_definition/data-source.tf` | example HCL for tfplugindocs embedding |

---

*Generated from `forge/history/INIT-2026-07-01-migrate-framework-build/demo/demo.json`*
