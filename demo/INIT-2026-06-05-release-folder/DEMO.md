# Add betterado_release_folder resource for Azure DevOps release folder management

> _Derived from `demo.json` (ADR 021). Essence:_ Adds a new Terraform resource `betterado_release_folder` that manages Azure DevOps release folder hierarchy. Before this change, release folders could not be managed as Terraform desired-state. After this change, operators can create, read, update and delete ADO release folders via HCL; 5 canonical gomock unit tests are green and the full release+taskagent test suite passes.

## Summary

- New resource `betterado_release_folder` with project_id, path, description schema
- Full CRUD over ADO Folders REST API (CreateFolder POST, GetFolders, UpdateFolder, DeleteFolder)
- Registered in provider.go alongside existing release resources
- 5 unit tests covering expand/flatten, create-error, read-404-clears-id, update-args, delete-error

## Test Evidence

### Full release + taskagent test suite passes on branch tip

- **Before:** No betterado_release_folder resource existed; the release package had no folder CRUD tests.
- **After:** 5 new TestReleaseFolder unit tests all PASS; go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... exits 0 with ok on all three packages.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| TestReleaseFolder suite (5 cases) | N/A — resource did not exist | 5/5 PASS | — | incomplete |
| go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... | package had 0 folder tests; gate passed on pre-existing tests only | ok (release 0.020s, taskagent 0.007s, taskagent/validate 0.004s) | — | within |

## API / Behaviour Diff

### betterado_release_folder (added)

**Before:**
```
resource did not exist
```
**After:**
```
resource with project_id (Required), path (Required), description (Optional) — Create/Read/Update/Delete
```

## Test Evidence

| test | result | delta |
|---|---|---|
| TestReleaseFolderExpand | pass | +1 |
| TestReleaseFolderCreate_Error | pass | +1 |
| TestReleaseFolderRead_404ClearsID | pass | +1 |
| TestReleaseFolderUpdate_Args | pass | +1 |
| TestReleaseFolderDelete_Error | pass | +1 |

## Acceptance criteria

- betterado_release_folder resource implemented with schema fields project_id (Required), path (Required), description (Optional)
- CRUD functions wired to CreateFolder/GetFolders/UpdateFolder/DeleteFolder
- Resource registered as betterado_release_folder in provider.go
- HCL usage example added under examples/resources/betterado_release_folder/main.tf
- 5 canonical gomock unit tests pass: expand/flatten roundtrip, create-error, read-404-clears-id, update-args, delete-error
- CI gate green: go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...

## Files Changed

- `azuredevops/internal/service/release/resource_release_folder.go` — New resource: schema, expandReleaseFolder, flattenReleaseFolder, CRUD functions
- `azuredevops/internal/service/release/resource_release_folder_test.go` — 5 canonical gomock unit tests
- `azuredevops/provider.go` — One-line registration of betterado_release_folder in resource registry
- `examples/resources/betterado_release_folder/main.tf` — HCL usage example

```
azuredevops/internal/service/release/resource_release_folder.go    | 145 ++++++++++++
 azuredevops/internal/service/release/resource_release_folder_test.go | 198 +++++++++++++++++
 azuredevops/provider.go                                            |   1 +
 azuredevops/provider_test.go                                       |   1 +
 demo/INIT-2026-06-05-release-folder/DEMO.html                      | 243 +++++++++++++++++++++
 demo/INIT-2026-06-05-release-folder/DEMO.md                        |  90 ++++++++
 demo/INIT-2026-06-05-release-folder/demo.json                      | 109 +++++++++
 examples/resources/betterado_release_folder/main.tf                |   9 +
 10 files changed, 813 insertions(+), 56 deletions(-)
```

## Usage

```
```hcl
resource "betterado_release_folder" "example" {
  project_id  = data.betterado_project.example.id
  path        = "/MyApp/Production"
  description = "Production release folder for MyApp"
}
```
```

## Impact

- Operators can now manage Azure DevOps release folder hierarchy as Terraform desired-state — no more manual portal clicks
- Enables folder-scoped release organisation as code, consistent with the provider's existing resource model
- Prerequisite for downstream initiatives (release-data-sources, release-definition-permissions) that rely on stable, declarative folder paths
