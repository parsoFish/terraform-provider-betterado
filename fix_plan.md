# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN ResourceReleaseDefinitionPermissions is implemented in WI-2 WHEN provider.go ResourcesMap is inspected THEN it contains the key 'betterado_release_definition_permissions' pointing to permissions.ResourceReleaseDefinitionPermissions()
- [x] AC2: GIVEN betterado_release_definition_permissions is added to provider.go WHEN TestProvider_HasChildResources runs (go test -run TestProvider_HasChildResources ./azuredevops/...) THEN the test passes — the expected resource list includes 'betterado_release_definition_permissions' and the count matches
- [x] AC3: GIVEN the resource is registered WHEN the example file examples/resources/betterado_release_definition_permissions/main.tf is present THEN it contains a valid HCL example demonstrating the resource with project_id, release_definition_id, principal, and permissions attributes
