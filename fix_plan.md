# Fix Plan — unifier sub-phase

> Initiative-level acceptance criteria. Tick each as you prove it against branch tip. Iteration 1 is initial prep; iterations 2+ react to either gate failures or send-back feedback.

- [x] AC1 (WI-1): GIVEN the file azuredevops/internal/service/release/resource_release_folder.go exists WHEN go vet -mod=vendor ./azuredevops/internal/service/release/ runs THEN the package compiles without errors and ResourceReleaseFolder() returns a non-nil *schema.Resource
  - ✅ iter 1: `go vet -mod=vendor ./azuredevops/internal/service/release/` → VET OK; `ResourceReleaseFolder()` defined at line 17.
- [x] AC2 (WI-1): GIVEN provider.go registers betterado_release_folder via ResourceReleaseFolder() WHEN go build -mod=vendor . runs THEN the binary compiles cleanly with the new resource present in the resource map
  - ✅ iter 1: `go build -tags all -mod=vendor .` → BUILD OK. Note: `-tags all` required (resource gated behind `//go:build (all || resource_release_folder)`).
- [x] AC3 (WI-1): GIVEN the schema defines path (Required, string) and description (Optional, string) and project_id (Required, string) WHEN expandReleaseFolder and flattenReleaseFolder are called with a matching sdk Folder struct THEN a round-trip expand → flatten preserves path, description, and project_id without data loss
  - ✅ iter 1: TestReleaseFolder_ExpandFlatten_Roundtrip passes; schema at resource_release_folder.go lines 24–46.
- [x] AC4 (WI-2): GIVEN a mock release client configured to return a Folder on CreateFolder WHEN resourceReleaseFolderCreate is called with a valid ResourceData containing project_id, path, and description THEN CreateFolder is called exactly once with the correct FolderPath and the resource ID is set to path
  - ✅ iter 1: TestReleaseFolder_Create_Success passes.
- [x] AC5 (WI-2): GIVEN a mock release client configured to return a 404-equivalent error on GetFolders WHEN resourceReleaseFolderRead is called THEN the resource ID is cleared (d.SetId('')) and no error is returned
  - ✅ iter 1: TestReleaseFolder_Read_NotFound passes.
- [x] AC6 (WI-2): GIVEN a mock release client configured to return an error on DeleteFolder WHEN resourceReleaseFolderDelete is called THEN the error is propagated back to Terraform
  - ✅ iter 1: TestReleaseFolder_Delete_Error passes.
- [x] AC7 (WI-2): GIVEN expandReleaseFolder and flattenReleaseFolder are implemented WHEN TestReleaseFolder_ExpandFlatten_Roundtrip runs with a Folder containing path and description THEN the round-trip preserves all fields without data loss
  - ✅ iter 1: same as AC3.
- [x] AC8 (WI-2): GIVEN all five unit tests are present and the package compiles WHEN go test -tags all -run TestReleaseFolder ./azuredevops/internal/service/release/ runs THEN all five tests pass (PASS lines for each TestReleaseFolder_* function)
  - ✅ iter 1: `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` → all ok.
- [ ] AC9 (WI-3): GIVEN TF_ACC=1 and AZDO_ORG_SERVICE_URL + AZDO_PERSONAL_ACCESS_TOKEN are set WHEN go test -tags all -run TestAccReleaseFolder ./azuredevops/internal/acceptancetests/ runs THEN the test creates a real ADO release folder, reads it back, updates its description, and destroys it without leaving orphaned resources
  - ⚠️ iter 1: azuredevops/internal/acceptancetests/resource_release_folder_test.go missing from branch — WI-3 per-WI commit not found in git log.
- [ ] AC10 (WI-3): GIVEN the acceptance test file azuredevops/internal/acceptancetests/resource_release_folder_test.go exists WHEN go build -mod=vendor . runs THEN the acceptance test package compiles without errors
  - ⚠️ iter 1: file missing — same root cause as AC9.
- [ ] AC11 (WI-3): GIVEN website/docs/r/release_folder.html.markdown exists WHEN the file is read THEN it documents the resource attributes (project_id, path, description), includes an HCL example, and lists import instructions
  - ⚠️ iter 1: website/docs/r/release_folder.html.markdown missing from branch — WI-3 per-WI commit not found in git log.
