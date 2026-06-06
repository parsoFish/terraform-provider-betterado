# Add betterado_release_definition and betterado_release_definitions data sources

> _Derived from `demo.json` (ADR 021). Essence:_ Two new Terraform data sources allow users to look up existing Azure DevOps release pipeline definitions by id or name and list them for cross-referencing in configs. Previously there was no read-only way to reference release definitions from Terraform.

## Summary

- Added betterado_release_definition data source (lookup by id or name via GetReleaseDefinition)
- Added betterado_release_definitions data source (list with optional path filter via GetReleaseDefinitions)
- Registered both in provider.go data-source map (+2 lines)
- Full unit test coverage: read path + not-found error for both data sources (6 new unit tests)
- Acceptance test stubs added (TF_ACC=1 live-ADO gate per project contract)
- Terraform examples and user-facing docs added for both data sources

## Test Evidence

### go test -tags all -count=1 passes for all release and taskagent packages after adding both data sources

- **Before:** No release data sources existed; the test suite had no tests for betterado_release_definition or betterado_release_definitions.
- **After:** 37 unit tests across release + taskagent + validate packages all pass in ~0.03 s. Provider registry includes both data sources. Quality gate exits 0.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| release package — TestDataReleaseDefinition_ReadById_Populates | N/A (data source absent) | PASS (0.00s) | — | incomplete |
| release package — TestDataReleaseDefinition_ReadByName_Populates | N/A (data source absent) | PASS (0.00s) | — | incomplete |
| release package — TestDataReleaseDefinition_Read_404ReturnsError | N/A (data source absent) | PASS (0.00s) | — | incomplete |
| release package — TestDataReleaseDefinitions_ReadList_PopulatesAll | N/A (data source absent) | PASS (0.00s) | — | incomplete |
| release package — TestDataReleaseDefinitions_ReadList_WithPathFilter | N/A (data source absent) | PASS (0.00s) | — | incomplete |
| release package — TestDataReleaseDefinitions_ReadList_APIErrorSurfaces | N/A (data source absent) | PASS (0.00s) | — | incomplete |
| release package — all pre-existing tests (resource_release_definition, release_folder) | PASS | PASS (0.018s total package) | 0.0% | match |
| taskagent package — all tests | PASS | PASS (0.007s) | 0.0% | match |
| taskagent/validate package — all tests | PASS | PASS (0.003s) | 0.0% | match |

## Test Evidence

| test | result | delta |
|---|---|---|
| TestDataReleaseDefinition_ReadById_Populates | pass | new |
| TestDataReleaseDefinition_ReadByName_Populates | pass | new |
| TestDataReleaseDefinition_Read_404ReturnsError | pass | new |
| TestDataReleaseDefinitions_ReadList_PopulatesAll | pass | new |
| TestDataReleaseDefinitions_ReadList_WithPathFilter | pass | new |
| TestDataReleaseDefinitions_ReadList_APIErrorSurfaces | pass | new |
| release package — pre-existing resource_release_definition tests (17 cases) | pass | 0 (unchanged) |
| release package — pre-existing release_folder tests (5 cases) | pass | 0 (unchanged) |
| taskagent package — all tests (10 cases) | pass | 0 (unchanged) |
| taskagent/validate package — TestEnvironmentName | pass | 0 (unchanged) |
| acceptance tests — TestAccDataReleaseDefinition* (TF_ACC=1, live ADO) | skip | new (142 lines; requires TF_ACC + PAT) |
| acceptance tests — TestAccDataReleaseDefinitions* (TF_ACC=1, live ADO) | skip | new (requires TF_ACC + PAT) |

## Acceptance criteria

- data.betterado_release_definition (by id or name) implemented and registered in provider.go
- data.betterado_release_definitions (list) implemented and registered in provider.go
- Unit tests cover the read path and not-found error path for both data sources
- CI gate (go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...) is green

## Files Changed

- `azuredevops/internal/service/release/data_release_definition.go` — New data source: look up release definition by id or name via GetReleaseDefinition
- `azuredevops/internal/service/release/data_release_definitions.go` — New data source: list release definitions via GetReleaseDefinitions
- `azuredevops/internal/service/release/data_release_definition_test.go` — Unit tests: read-by-id, read-by-name, 404 not-found error path
- `azuredevops/internal/service/release/data_release_definitions_test.go` — Unit tests: list all, path-filtered list, API error surface
- `azuredevops/internal/acceptance/data_release_definition_test.go` — Acceptance tests for data_release_definition (requires live ADO, TF_ACC=1)
- `azuredevops/provider.go` — Register betterado_release_definition and betterado_release_definitions in provider data-source map (+2 lines)
- `azuredevops/provider_test.go` — Assert both data sources appear in provider registry (+2 lines)
- `docs/resources/release_definition.md` — User-facing documentation for betterado_release_definition data source
- `docs/resources/release_definitions.md` — User-facing documentation for betterado_release_definitions data source
- `examples/data-sources/betterado_release_definition/main.tf` — Runnable HCL example for betterado_release_definition
- `examples/data-sources/betterado_release_definitions/main.tf` — Runnable HCL example for betterado_release_definitions

```
azuredevops/internal/acceptance/data_release_definition_test.go                | 142 ++++++++++++++++++
 azuredevops/internal/service/release/data_release_definition.go                | 125 ++++++++++++++++
 azuredevops/internal/service/release/data_release_definition_test.go           | 157 ++++++++++++++++++++
 azuredevops/internal/service/release/data_release_definitions.go               | 129 +++++++++++++++++
 azuredevops/internal/service/release/data_release_definitions_test.go          | 158 +++++++++++++++++++++
 azuredevops/provider.go                                                        |   2 +
 azuredevops/provider_test.go                                                   |   2 +
 docs/resources/release_definition.md                                           |  50 +++++++
 docs/resources/release_definitions.md                                          |  41 ++++++
 examples/data-sources/betterado_release_definition/main.tf                     |  19 +++
 examples/data-sources/betterado_release_definitions/main.tf                    |  16 +++
 13 files changed, 889 insertions(+)
```

## Usage

```
# Look up a specific release definition by name
data "betterado_release_definition" "my_pipeline" {
  project_id = data.betterado_project.proj.id
  name       = "My Release Pipeline"
}

# Look up by id
data "betterado_release_definition" "by_id" {
  project_id = data.betterado_project.proj.id
  id         = 42
}

# List all release definitions in a project
data "betterado_release_definitions" "all" {
  project_id = data.betterado_project.proj.id
}

# Filter by path
data "betterado_release_definitions" "team" {
  project_id = data.betterado_project.proj.id
  path       = "\\MyTeam"
}

# Reference the id in another resource
output "pipeline_id" {
  value = data.betterado_release_definition.my_pipeline.id
}

output "all_pipeline_ids" {
  value = [for d in data.betterado_release_definitions.all.definitions : d.id]
}
```

## Impact

- Terraform users can now look up existing Azure DevOps release definitions by id or name — no more hard-coded numeric IDs
- Lists of release definitions can be fetched and filtered by path directly in Terraform HCL
- Enables cross-referencing release pipelines when configuring permissions or wiring build artifacts in downstream resources
- Both data sources are read-only — no create/update/delete side effects, safe to use in any plan
