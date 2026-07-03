# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (initial implementation)

**What I did:**
1. Created `resource_pipeline_test.go` + `data_pipeline_test.go`.
2. Fixed `resource_pipeline_authorization_test.go` to use `data "betterado_project"` (org at 1000-project cap).

**Outcome:** Live gate failed with `Error creating pipeline — Value cannot be null. Parameter name: Path`.

### Iteration 2 (fix create body)

**Root cause of last-gate-failure.md:**
- `POST _apis/pipelines` requires `configuration.path` and `configuration.repository` for YAML pipelines.
- The SDK's `CreatePipelineConfigurationParameters` only has `Type`, so those fields were not sent.
- The resource schema also had no `repo_id` or `yaml_path` attributes.

**What I did:**
1. Added `repo_id` and `yaml_path` schema attributes to `resource_pipeline_framework.go`.
2. Added `createPipelineRaw()` — posts full JSON body via `impl.Client.Send` with:
   - `configuration.type = "yaml"`
   - `configuration.path = yaml_path`
   - `configuration.repository.id = repo_id`
   - `configuration.repository.type = "azureReposGit"`
3. Replaced the SDK `CreatePipeline` call in `Create()` with `createPipelineRaw()`.
4. Removed the now-unused `expandPipeline()` and `strPtr()`.
5. Updated HCL fixtures in both `resource_pipeline_test.go` and `data_pipeline_test.go` to look up the `betterado-standing-demo` default git repo via `data "betterado_git_repository"` and pass `repo_id` + `yaml_path = "/azure-pipelines.yml"`.
6. Updated `examples/resources/betterado_pipeline/resource.tf` and ran `make docs`.

**Outcome:** Live gate failed with `Error deleting pipeline — The requested resource does not support http method 'DELETE'`.

### Iteration 3 (fix delete — use Build Definitions API)

**Root cause of last-gate-failure.md:**
- The Pipelines v2 API (`_apis/pipelines`) does NOT support HTTP DELETE.
- The original `deletePipeline()` was sending `http.MethodDelete` to the `_apis/pipelines/{pipelineId}` endpoint, which the ADO REST API does not handle.
- Both `TestAccPipeline_basic` and `TestAccDataPipeline_basic` failed in the destroy phase with `Error: Error deleting pipeline / The requested resource does not support http method 'DELETE'`.

**Fix:**
- Every YAML pipeline created via `_apis/pipelines` is backed by a Build Definition with the same integer ID.
- The Build Definitions API (`_apis/build/definitions/{id}`) DOES support DELETE.
- Replaced `deletePipeline()` raw HTTP DELETE with `agg.BuildClient.DeleteDefinition(ctx, adoBuild.DeleteDefinitionArgs{...})`.
- Added import: `adoBuild "github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"`.
- `go build -tags all ./azuredevops/...` — clean.
- `golangci-lint run --new-from-rev=main ./azuredevops/...` — 0 issues.
- `make test` — no FAIL.

**Outcome:** Pending live gate.

## What worked

- Raw HTTP POST via `impl.Client.Send` to bypass SDK struct limitations and send the full configuration body.
- Using `data "betterado_git_repository"` with `name = SharedFixtureProjectName` to get the default repo GUID.
- `yaml_path = "/azure-pipelines.yml"` — the standing demo project's default repo has this file.
- `flattenPipeline` intentionally does NOT touch `repo_id` / `yaml_path` because the GET API doesn't return them — they stay preserved from state, giving idempotent plan.
- **Delete via `BuildClient.DeleteDefinition`** — the Pipelines v2 API doesn't support DELETE; route through Build Definitions API (same integer ID).

## What didn't work

- Using the SDK `CreatePipeline` without path/repo → `Value cannot be null. Parameter name: Path`.
- HTTP DELETE on `_apis/pipelines/{pipelineId}` → `The requested resource does not support http method 'DELETE'`.

## Key patterns

- ADO `POST _apis/pipelines?api-version=7.1-preview.1` body for YAML pipelines:
  ```json
  {
    "name": "...",
    "folder": "\\...",
    "configuration": {
      "type": "yaml",
      "path": "/azure-pipelines.yml",
      "repository": {
        "id": "<repo-guid>",
        "type": "azureReposGit"
      }
    }
  }
  ```
- The GET/LIST response does NOT echo back `configuration.path` or `configuration.repository` in the basic `Pipeline` struct — only `name`, `folder`, `revision`, `id`, `url`, `configuration.type`.
- Therefore: `repo_id` and `yaml_path` are `Optional` (not `Computed`) in state, and `flattenPipeline` skips them. State values persist from the user config + initial create, giving clean idempotency.
- **DELETE**: Must use `_apis/build/definitions/{id}` (Build Definitions API) NOT `_apis/pipelines/{id}`. Pipeline ID == Build Definition ID. Use `agg.BuildClient.DeleteDefinition`.
- `checkPipelineDestroyed` calls `PipelinesClient.GetPipeline` after `BuildClient.DeleteDefinition` — this works because deleting the build definition removes the underlying object, making the pipeline 404 on the pipelines API too.
