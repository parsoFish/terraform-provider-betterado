# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the file azuredevops/internal/service/release/resource_release_folder.go exists WHEN go vet -mod=vendor ./azuredevops/internal/service/release/ runs THEN the package compiles without errors and ResourceReleaseFolder() returns a non-nil *schema.Resource
- [x] AC2: GIVEN provider.go registers betterado_release_folder via ResourceReleaseFolder() WHEN go build -mod=vendor . runs THEN the binary compiles cleanly with the new resource present in the resource map
- [x] AC3: GIVEN the schema defines path (Required, string) and description (Optional, string) and project_id (Required, string) WHEN expandReleaseFolder and flattenReleaseFolder are called with a matching sdk Folder struct THEN a round-trip expand → flatten preserves path, description, and project_id without data loss
