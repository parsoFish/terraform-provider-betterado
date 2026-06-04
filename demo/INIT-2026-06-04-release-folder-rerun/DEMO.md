# Add betterado_release_folder resource with full CRUD and acceptance test

> _Derived from `demo.json` (ADR 021). Essence:_ Introduces the betterado_release_folder Terraform resource for managing Azure DevOps release-definition folders. Operators can now create, read, update, and destroy folder paths (e.g. \Production\Web) via Terraform. Includes 5 gomock unit tests, an acceptance test (live ADO, self-skips without TF_ACC), and reference docs.

## Test Evidence

### Full initiative quality gate: release + taskagent packages all pass

- **Before:** No betterado_release_folder resource existed; TestReleaseFolder_* tests did not exist.
- **After:** go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... exits 0: all 3 packages ok.

| metric | before | after | Δ | parity |
|---|---|---|---|---|
| github.com/.../service/release | FAIL (compile error — resource_release_folder.go absent) | ok (0.012s) | — | incomplete |
| github.com/.../service/taskagent | ok | ok (0.007s) | — | match |
| github.com/.../service/taskagent/validate | ok | ok (0.003s) | — | match |

## Acceptance criteria

- betterado_release_folder resource registered in provider.go, schema-valid, full CRUD implemented
- 5 canonical gomock unit tests pass: ExpandFlatten_Roundtrip, Create_Success, Create_Error, Read_NotFound, Delete_Error
- Delete returns helpful error 'Cannot delete folder ... contains release definitions' when folder is non-empty
- Acceptance test TestAccReleaseFolder_basic present (runs live with TF_ACC=1)
- Reference docs at website/docs/r/release_folder.html.markdown

## Files Changed

- `azuredevops/internal/service/release/resource_release_folder.go` — New resource: full CRUD for betterado_release_folder
- `azuredevops/internal/service/release/resource_release_folder_test.go` — 5 canonical gomock unit tests
- `azuredevops/internal/acceptancetests/resource_release_folder_test.go` — Acceptance test (requires TF_ACC=1 + live ADO creds)
- `azuredevops/provider.go` — Register betterado_release_folder in provider resource map
- `website/docs/r/release_folder.html.markdown` — Reference documentation

```
.../resource_release_folder_test.go                | 114 ++++++++++++
 .../service/release/resource_release_folder.go     | 207 +++++++++++++++++++++
 .../release/resource_release_folder_test.go        | 196 +++++++++++++++++++
 azuredevops/provider.go                            |   1 +
 .../INIT-2026-06-04-release-folder-rerun/DEMO.html | 199 ++++++++++++++++++++
 demo/INIT-2026-06-04-release-folder-rerun/DEMO.md  |  63 +++++++
 .../INIT-2026-06-04-release-folder-rerun/demo.json | 108 +++++++++++
 website/docs/r/release_folder.html.markdown        |  46 +++++
 8 files changed, 934 insertions(+)
```

## Usage

```
resource "azuredevops_project" "example" {
  name = "Example Project"
}

resource "betterado_release_folder" "example" {
  project_id  = azuredevops_project.example.id
  path        = "\\Production\\Web"
  description = "Release definitions for the production web tier"
}
```

## Impact

- Operators can now manage Azure DevOps release-definition folder hierarchies as Terraform resources
- Enables IaC-driven folder organisation for release pipelines without manual portal clicks
- Safe delete: provider refuses to cascade-delete folders containing definitions, preventing accidental data loss
