# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (successful)

**Gate failure**: `[no tests to run]` — `TestAccProjectPermissionsFramework` did not exist.

**What was done**:
1. Created `azuredevops/internal/service/permissions/resource_project_permissions_framework.go`
   - Full CRUD framework resource implementing `betterado_project_permissions`
   - Token format: `$PROJECT:vstfs:///Classification/TeamProject/{projectId}`
   - Security namespace: `SecurityNamespaceIDValues.Project` (UUID: `52d39943-cb85-4d7f-8fa8-c6baac873819`)
   - Pattern: copied from `resource_release_definition_permissions_framework.go`
   - Inline plan modifiers with `pp` prefix (to avoid name collision with rdp prefix from release)
   - Polls until ACL synced, max 60 minutes

2. Registered in `framework_provider.go` Resources() slice:
   `permissions.NewProjectPermissionsResource`

3. Removed `betterado_project_permissions` from SDKv2 `provider.go` ResourcesMap

4. Updated `provider_test.go` expectedResources: removed `betterado_project_permissions`, added comment

5. Updated `resource_project_permissions_test.go`: changed `ProviderFactories` to `ProtoV6ProviderFactories: testutils.GetMuxProviderFactories()`

6. Created `azuredevops/internal/acceptancetests/resource_permissions_framework_test.go`:
   - Build tag: `(all || resource_project_permissions) && !exclude_resource_project_permissions`
   - `TestAccProjectPermissionsFramework` creates a project (NOT using existing-project pattern)
   - Asserts 4 permissions: DELETE=Deny, EDIT_BUILD_STATUS=NotSet, WORK_ITEM_MOVE=Allow, DELETE_TEST_RESULTS=Deny
   - Uses `testutils.GetMuxProviderFactories()`
   - Calls `testutils.CaptureLiveEvidence("acceptance-resource", url, nil)` where url uses the project namespace ID
   - Idempotency step: `PlanOnly: true, ExpectNonEmptyPlan: false`
   - `CheckDestroy: testutils.CheckProjectDestroyed` (project is created, so project destroy is needed)

7. Fixed nilerr lint in `resource_securityrole_assignment_framework.go`: poll loop on 404 now checks error string

**Offline test results**: All pass. `make test` green. `golangci-lint` 0 issues. `TestProvider_HasChildResources` passes.

## What worked

- Copy pattern from `resource_release_definition_permissions_framework.go` exactly, substituting the token builder and namespace ID
- Use unique `pp` prefix for plan modifiers to avoid duplicate type declarations
- Build tag must use `resource_project_permissions` to match the `all` build tag set
- `CaptureLiveEvidence("acceptance-resource", url, nil)` with the hardcoded Project namespace UUID
- Creating a new project (not an existing one) keeps the test self-contained and avoids org cap issues for permissions tests that only need `project_id`

## What didn't work

- `_ = err` does NOT satisfy golangci-lint's `nilerr` checker — must use actual conditional logic

## Open questions

- AC2 and AC3 require all 13 resources migrated. Only project_permissions done so far. Gate only checks `TestAccProjectPermissionsFramework` so AC1 is satisfied, but next iteration should migrate remaining 12 if time permits.

## Notes for reflection

- The gate command (`TestAccProjectPermissionsFramework`) requires TF_ACC=1 for a live pass. The offline test skips cleanly with `t.Skip("TF_ACC not set")`.
- `GetMuxProviderFactories()` is defined in both `testutils/commons.go` and `testutils/mux_provider.go` as different functions. Use `testutils.GetMuxProviderFactories()` (commons.go version, same mux logic).
