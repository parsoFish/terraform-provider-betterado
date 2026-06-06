# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (complete)

- Read WI-1.md spec and existing `resource_release_definition.go` for patterns.
- Confirmed `azdosdkmocks.MockReleaseClient` has `CreateFolder`, `GetFolders`, `UpdateFolder`, `DeleteFolder` mocked.
- Confirmed `releaseapi.Folder` struct has `Path *string` and `Description *string` (no `Id` field — path is the stable identifier).
- Confirmed `AggregatedClient.ReleaseClient` already exists — no new client field needed.
- Wrote `resource_release_folder.go` with full CRUD + `expandReleaseFolder` / `flattenReleaseFolder` helpers.
- Wrote `resource_release_folder_test.go` with 5 tests (all function names start with `TestReleaseFolder`).
- Added `"betterado_release_folder": release.ResourceReleaseFolder()` to `provider.go` after `betterado_release_definition`.
- Created `examples/resources/betterado_release_folder/main.tf`.
- Quality gate `go test -mod=vendor -tags all -count=1 -run TestReleaseFolder ./azuredevops/internal/service/release/` → **PASS** (all 5 tests green).
- Committed: `adfa7576`.

## What worked

- Mirroring `resource_release_definition.go` structure closely (build tag, package, import paths, converter usage).
- Using `path` as the Terraform resource ID (ADO addresses folders by project + path; `Folder` struct has no integer Id).
- Test pattern: `schema.TestResourceDataRaw` + `resourceData.SetId(...)` for tests that need a pre-set ID.
- Checking `diags[0].Summary` for error message substring (not exact match) — robust against formatting changes.
- For `TestReleaseFolderUpdate_Args`: using `gomock.Eq(releaseapi.UpdateFolderArgs{...})` with pointer fields matched via `converter.String` to verify exact args.

## What didn't work

_(none — iteration 0 succeeded on first attempt)_

## Open questions

_(none)_

## Notes for reflection

- `Folder` struct in the ADO SDK has no integer `Id` field — `path` is the natural key. WI spec confirms this.
- `path` is `ForceNew: true` because the ADO API's `UpdateFolder` takes the *current* path as a separate URL arg; renaming would need a special path-rename call. Simpler to treat path changes as destroy+create.
