## Why

Azure DevOps release definitions live in folders (`/` by default), but until now the Terraform provider had no resource to manage that folder hierarchy as desired-state. Operators were forced to create and rename folders manually in the ADO portal, breaking the principle that all infrastructure is declared in HCL. This change closes that gap for the release-folder primitive — a prerequisite for the downstream `release-data-sources` and `release-definition-permissions` initiatives that rely on stable folder paths.

## What

- **New resource `betterado_release_folder`** (`azuredevops/internal/service/release/resource_release_folder.go`): schema with `project_id` (Required, UUID), `path` (Required, string), `description` (Optional, string); full CRUD wired to the ADO Folders REST API (`CreateFolder` POST, `GetFolders`, `UpdateFolder`, `DeleteFolder`).
- **Provider registration** (`azuredevops/provider.go`): one-line addition to the resource registry.
- **HCL usage example** (`examples/resources/betterado_release_folder/main.tf`): canonical snippet showing path and description usage.
- **5 unit tests** (`azuredevops/internal/service/release/resource_release_folder_test.go`): expand/flatten roundtrip, create-error, read-404-clears-id, update-args, delete-error — all passing under `go test -tags all -count=1`.

## How

The resource follows the existing `resource_release_definition.go` / `resource_task_group.go` pattern:

1. **Schema** declares `project_id`, `path`, `description` with the standard Terraform SDK `schema.Resource` helpers.
2. **`expandReleaseFolder`** maps the Terraform state to the ADO SDK `release.Folder` struct.
3. **`flattenReleaseFolder`** maps the SDK response back to Terraform state, setting `id` to `path` (the natural ADO key).
4. **Create** calls `CreateFolder` (POST variant — PUT `Create` is deprecated per ADO docs).
5. **Read** calls `GetFolders` filtered by path; a 404-equivalent (empty result) clears the resource ID to signal drift.
6. **Update** calls `UpdateFolder` with the new description.
7. **Delete** calls `DeleteFolder`.
8. Tests use the project's existing `MockReleaseClient` gomock double so no live credentials are needed.
