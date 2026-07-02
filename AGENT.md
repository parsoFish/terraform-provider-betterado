# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 2 (current)
- **Root cause identified**: `TestAccProject_importByName` failed with "resource with ID <uuid> not found" at `testing_new_import_state.go:326`. This error originates in `testImportCommand` which compares `importState` (post-import) against `state` (pre-import/"old" state). Since this is a pure import test (no prior terraform apply), `state` is empty. `oldResources` is empty, so the imported resource's ID can't be found in `oldResources` → `oldR == nil` → test fails.
- **Fix applied**: Removed `ImportStateVerify: true` and `ImportStateVerifyIgnore: []string{"description"}` from the Step 1 in `TestAccProject_importByName`. The `checkProjectImportByName` function already verifies all required attributes; Step 2's `PlanOnly: true, ExpectNonEmptyPlan: false` verifies idempotency. Both satisfy AC1 fully.
- Verified: `make test` passes clean (all offline tests). `go build -tags all ./azuredevops/internal/acceptancetests/` succeeds.

### Iteration 1 (prior, autocommit)
- Implemented framework resource `resource_project_framework.go` with full CRUD + ImportState by name/UUID
- Implemented `data_project_framework.go` (lookup by ID or name)
- Implemented `data_projects_framework.go` (list all projects)
- Registered all three in `framework_provider.go`
- Removed from SDKv2 `provider.go` ResourcesMap/DataSourcesMap
- Updated `provider_test.go` counts
- Updated acceptance tests to use `GetMuxedProviderFactories()`

## What worked

- Framework resource implementing `resource.Resource` + `resource.ResourceWithImportState` for import-by-name support: works fine.
- Removing `ImportStateVerify: true` from a test that does import-only (no prior apply) fixes the "resource with ID not found" test framework error. The error is NOT from ADO API; it's from the test framework's state-comparison logic which requires pre-import state to exist.
- `checkProjectImportByName` as `ImportStateCheck` correctly verifies attributes after import.

## What didn't work

- `ImportStateVerify: true` when there's no prior apply step (pre-import state is empty). The test framework's `testImportCommand` compares `newResources` (from import) against `oldResources` (from pre-import state). When pre-import state is empty, `oldResources` is empty and the lookup by ID fails with "resource with ID ... not found" (testing_new_import_state.go:326).

## Key insight for future iterations

- The `ImportStateVerify` mechanism in terraform-plugin-testing requires a prior terraform apply step to work. For import-only tests (where you import an existing resource without first creating it via TF), you must omit `ImportStateVerify` and rely on `ImportStateCheck` for attribute verification.
- The mux provider setup (`GetMuxedProviderFactories()`) works correctly for framework resources alongside SDKv2 resources.
- The `visibility` field from ADO API returns `"private"` (lowercase) for private projects.

## Open questions

_(things that aren't blocking but would be useful to clarify; reflector picks these up)_

## Notes for reflection

- Pattern: import-only acceptance tests (where creating the resource is not possible) MUST NOT use `ImportStateVerify: true` — document in testing conventions/brain.
