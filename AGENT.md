# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (complete)

- Created `data_release_folder.go`:
  - `DataReleaseFolder()` returns `*schema.Resource` with non-Context `Read` field (matching `data_release_definition.go` pattern), 5-minute read timeout, schema with `project_id` (Required, IsUUID), `path` (Required, StringIsNotWhiteSpace), `description` (Computed).
  - `dataReleaseFolderRead` calls `clients.ReleaseClient.GetFolders` with the project/path args, returns error if empty/nil result, otherwise calls existing `flattenReleaseFolder` helper.
- Created `data_release_folder_test.go`:
  - Build tag: `//go:build (all || data_release_folder) && !exclude_data_release_folder`
  - `TestDataReleaseFolder_Read_Populates` — mocks GetFolders returning one folder, asserts no error, Id == path, description populated.
  - `TestDataReleaseFolder_Read_NotFound` — mocks GetFolders returning empty slice, asserts error contains "not found" and path.
  - Reuses package-level fixtures `testReleaseFolderProjectID`, `testReleaseFolderPath`, `testReleaseFolderDescription` from `resource_release_folder_test.go`.
- Quality gate (`go test -mod=vendor -tags all -count=1 -run TestDataReleaseFolder ./azuredevops/internal/service/release/`) → `ok` in 0.004s.
- `gofmt -l` → clean (no output).

## What worked

- Mirroring `data_release_definition.go` pattern exactly (non-Context Read, Timeouts struct).
- Reusing `flattenReleaseFolder` from `resource_release_folder.go` — it already sets Id, path, description, project_id.
- Reusing package-level fixtures from `resource_release_folder_test.go` — no redeclaration needed.
- Using `gomock.Any()` for the ctx arg in the not-found test (avoids strict arg matching when we don't need it).

## What didn't work

_(nothing — completed in first iteration)_

## Open questions

_(none)_

## Notes for reflection

_(nothing flagged)_
