## Why

Terraform configs that reference — but do not manage — a release folder had no
way to look one up. The `betterado_release_folder` **resource** already existed,
but without a matching **data source** any cross-stack or cross-root reference
required hard-coding IDs or duplicating folder management. This is the classic
resource/data-source parity gap: `betterado_release_definition` already shipped
both `data.betterado_release_definition` and `data.betterado_release_definitions`;
release folders deserve the same read-only lookup capability.

## What

- **`data.betterado_release_folder`** — new read-only data source (`data_release_folder.go`)
  that accepts `project_id` + `path` and surfaces `description` (and any other
  folder fields the API returns) by calling the SDK `GetFolders(project, path)`
  — the same call the resource's read path already uses.
- **Provider registration** — one line added to `DataSourcesMap` in `provider.go`;
  `provider_test.go` count assertion updated.
- **Unit tests** — `data_release_folder_test.go` covers the happy path
  (`TestDataReleaseFolder_Read_Populates`) and the folder-not-found path
  (`TestDataReleaseFolder_Read_NotFound`) using `azdosdkmocks` + gomock; no
  credentials required.
- **Acceptance test** — `TestAccDataReleaseFolder_Basic` in `azuredevops/internal/acceptancetests/`
  creates a folder via the resource, reads it back via the new data source, asserts
  `description` matches, verifies idempotency (re-plan → no diff), then destroys cleanly.
- **Example + docs** — `examples/data-sources/betterado_release_folder/main.tf`
  shows minimal HCL usage; `docs/resources/release_folder.md` documents the data
  source's required arguments and computed attributes.

## How

Files changed (7, all additions):

| File | Change |
|------|--------|
| `azuredevops/internal/service/release/data_release_folder.go` | New — `DataReleaseFolder()` + `dataReleaseFolderRead` |
| `azuredevops/internal/service/release/data_release_folder_test.go` | New — unit tests (Read_Populates, Read_NotFound) |
| `azuredevops/provider.go` | +1 line — `"betterado_release_folder": release.DataReleaseFolder()` in DataSourcesMap |
| `azuredevops/provider_test.go` | +1 entry — `"betterado_release_folder"` in expectedDataSources |
| `azuredevops/internal/acceptancetests/data_release_folder_test.go` | New — TestAccDataReleaseFolder_Basic |
| `examples/data-sources/betterado_release_folder/main.tf` | New — usage example |
| `docs/resources/release_folder.md` | New — data source documentation |

Implementation mirrors `data_release_definition.go` (non-Context Read function,
plain `*schema.Resource`, reuses the resource's existing `flattenReleaseFolder`
helper). No new SDK methods introduced — `GetFolders` was already used by the
resource read. Build tag on test files only; production file carries no tag
(CI-equivalent contract satisfied).

Quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` — 3 packages, 63 tests, 0 failures.
