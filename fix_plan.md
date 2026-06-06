# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN no betterado_release_folder resource exists in the worktree WHEN the resource is implemented with schema fields project_id (Required, UUID), path (Required, string), description (Optional, string), CRUD functions wired to CreateFolder/GetFolders/UpdateFolder/DeleteFolder, registered as betterado_release_folder in provider.go, and an HCL example added under examples/ THEN 5 canonical gomock unit tests pass — expand/flatten roundtrip, create-error, read-404-clears-id, update-args, delete-error — scoped with -run TestReleaseFolder; the test file compiles and all 5 tests are green; CI gate is green
