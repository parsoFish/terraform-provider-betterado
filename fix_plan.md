# Fix Plan — unifier sub-phase

> Initiative-level acceptance criteria. Tick each as you prove it against branch tip. Iteration 1 is initial prep; iterations 2+ react to either gate failures or send-back feedback.

- [x] AC1 (WI-1): GIVEN no betterado_release_folder resource exists in the worktree WHEN the resource is implemented with schema fields project_id (Required, UUID), path (Required, string), description (Optional, string), CRUD functions wired to CreateFolder/GetFolders/UpdateFolder/DeleteFolder, registered as betterado_release_folder in provider.go, and an HCL example added under examples/ THEN 5 canonical gomock unit tests pass — expand/flatten roundtrip, create-error, read-404-clears-id, update-args, delete-error — scoped with -run TestReleaseFolder; the test file compiles and all 5 tests are green; CI gate is green
  — VERIFIED iter 1: all 5 TestReleaseFolder tests PASS; gate go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/... exits 0
