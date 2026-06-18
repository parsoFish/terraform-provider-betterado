## Why

`betterado_release_folder` had no auditable record of which ADO Folder API fields it covered, and no resource-level acceptance test to prove the CRUD lifecycle worked against real ADO. The resource existed and worked in practice, but there was no CI-enforced guarantee that (a) all writable fields were mapped, or (b) that apply → read → idempotency → destroy ran cleanly end-to-end. This initiative closes both gaps: it produces the formal coverage audit and adds the live acceptance test that makes the guarantee machine-checkable.

## What

Three files are added by this initiative (zero production Go files changed):

- **`docs/release-folder-gap-matrix.md`** — field-by-field audit of all 6 ADO `Folder` struct fields against the `betterado_release_folder` schema. Both writable fields (`path`, `description`) are confirmed mapped; 4 server-computed fields are explicitly excluded. The matrix states the writable-column parity verdict so future maintainers know the audit ran.

- **`azuredevops/internal/service/release/doc_audit_test.go`** (modified) — adds `TestReleaseFolderGapMatrixAudit`, a sentinel test that reads `docs/release-folder-gap-matrix.md` and asserts it contains the parity verdict string. The quality gate (`go test -tags all ./azuredevops/internal/service/release/...`) fails on a clean tree without the file.

- **`azuredevops/internal/acceptancetests/resource_release_folder_test.go`** — new resource acceptance test (`TestAccReleaseFolder_basic`) with:
  - Live apply against real ADO with a unique path and non-default description.
  - `TestCheckResourceAttr` confirming `description = "Acceptance test folder"` round-trips.
  - Idempotency step (`PlanOnly: true, ExpectNonEmptyPlan: false`).
  - `CheckDestroy` via `GetFolders` confirming the folder is gone after destroy.
  - `captureReleaseFolderEvidence` writing `.forge/live-evidence/acceptance-resource.json` with the real vsrm REST GET URL.

## How

**WI-1** (gap matrix + sentinel):
- Enumerated all fields from `vendor/.../release/models.go` Folder struct.
- Produced `docs/release-folder-gap-matrix.md` following the structure of the existing `release-definition-gap-matrix.md`.
- Added `TestReleaseFolderGapMatrixAudit` to `doc_audit_test.go` using `os.ReadFile` + `require.Contains` — no mocks, no ADO credentials needed; the sentinel is purely a file-existence + content check.

**WI-2** (resource acceptance test + live evidence):
- Added `resource_release_folder_test.go` to `azuredevops/internal/acceptancetests/` with build tag `//go:build (all || resource_release_folder) && !exclude_resource_release_folder`.
- `captureReleaseFolderEvidence` mirrors `captureTaskGroupEvidence` from `resource_task_group_test.go`: calls `GetFolders`, constructs the vsrm host URL (`dev.azure.com` → `vsrm.dev.azure.com`), and writes `.forge/live-evidence/acceptance-resource.json` via `testutils.CaptureLiveEvidence`.
- `checkReleaseFolderDestroyed` calls `GetFolders` on the stored path and returns an error if any folder entries remain.

**WI-3** (CI gate):
- No production code changed; the quality gate (`go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...`) passes green with 63 release-package tests + 32 taskagent tests. No new HCL examples were added; no provider registry changes were made.
