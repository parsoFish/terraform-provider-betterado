# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (commit f13104ee) — COMPLETE

Created all required files:

1. **`resource_securityrole_assignment_framework.go`** — terraform-plugin-framework resource
   - Pattern: inline plan modifiers (`sraRequiresReplace`, `sraUseStateForUnknown`) — the vendor does NOT have `stringplanmodifier` package; use inline structs implementing `planmodifier.String` interface instead (same pattern as `resource_security_permissions_framework.go`).
   - `waitForAssignment` polling loop (5s intervals, 10m timeout) — mirrors SDKv2 StateChangeConf.
   - Read: uses `GetSecurityRoleAssignment`; nil-or-empty assignment → `RemoveResource`.
   - Create/Update: `SetSecurityRoleAssignment` → poll until confirmed.
   - Delete: `DeleteSecurityRoleAssignment`.

2. **`data_securityrole_definitions_framework.go`** — terraform-plugin-framework data source
   - Computes a `types.Set` of objects with fields: name, display_name, allow_permissions, deny_permissions, identifier, description, scope.

3. **`framework_provider.go`** — added `securityroles` import + registered both in Resources()/DataSources() slices.

4. **`provider.go`** — removed `betterado_securityrole_assignment` from ResourcesMap, `betterado_securityrole_definitions` from DataSourcesMap, and the now-unused `securityroles` import.

5. **`provider_test.go`** — removed both entries from expected resource/datasource lists (counts now match).

6. **`resource_securityrole_assignment_framework_test.go`** — `TestAccSecurityRoleAssignmentFramework`:
   - Resolves `betterado-standing-demo` project (avoids creating new project — org at 1000-cap).
   - Creates an environment via `clients.TaskAgentClient.AddEnvironment` with field `EnvironmentCreateParameter` (NOT `Environment` — that field doesn't exist).
   - Gets identity via `resolveApproverIdentity` (reuses function from shared_fixtures.go).
   - HCL uses literal IDs — no dependent Terraform resources.
   - `CheckDestroy`: `resource.TestCheckFunc` type (not `resource.TestCheckDestroyFunc` which doesn't exist).
   - Registered under build tag `resource_securityrole_assignment`.

7. **`examples/resources/betterado_securityrole_assignment/resource.tf`** — example.

### Iteration 2 (commit 41dc6aad) — FIX: destroy check failure

**Gate failure**: `checkSecurityRoleAssignmentFrameworkDestroyed` reported "assignment still exists after destroy".

**Root cause**: The ADO security roles PATCH delete (`DeleteSecurityRoleAssignment`) does NOT fully remove the assignment — it reverts the identity from an **explicit** assignment to an **inherited** one. For Project Administrators on any environment, they always have an inherited Administrator access at the project level. After our PATCH delete, `ListSecurityRoleAssignments` still returns them with `Access="inherited"`.

**Fixes applied**:

1. **`checkSecurityRoleAssignmentFrameworkDestroyed`**: Added check for `Access="inherited"` — treats inherited-access assignments as "gone" (not a dangling resource). Only explicit `Access="assigned"` entries count as still existing.

2. **`waitForDeletion` (new function)**: Added post-delete polling that mirrors `waitForAssignment`. Polls until the assignment is gone OR has `Access="inherited"`. Called after `DeleteSecurityRoleAssignment` in the Delete function. Handles ADO's eventual-consistency window.

3. **`Read` scope bug fix**: Stopped overwriting `state.Scope` from `assignment.Role.Scope`. The role-definition's own `Scope` field is the role's internal scope attribute, which may differ from the query scope (`distributedtask.environmentreferencerole`). Overwriting state.Scope would corrupt the identity key used in subsequent Delete calls.

## What worked

- The existing `resource_security_permissions_framework.go` was the ideal template for both the inline plan modifiers and the overall structure.
- The `resolveApproverIdentity` function in `shared_fixtures.go` is the right way to get a real identity UUID for test assertions.
- `GetMuxProviderFactories()` in `testutils/commons.go` (NOT the `GetMuxedProviderFactories` in `mux_provider.go`) is used by all other framework tests — use it consistently.
- `AddEnvironmentArgs.EnvironmentCreateParameter` (not `Environment`) is the correct field name.

## What didn't work

- `stringplanmodifier` package doesn't exist in vendor — use inline structs.
- `resource.TestCheckDestroyFunc` type doesn't exist — use `resource.TestCheckFunc`.
- `GetProjects` doesn't accept nil — pass `core.GetProjectsArgs{}`.
- The original `checkSecurityRoleAssignmentFrameworkDestroyed` didn't account for ADO's inherited-access behaviour — Project Administrators always retain inherited Administrator access even after explicit assignment is deleted.
- Setting `state.Scope` from `assignment.Role.Scope` in Read corrupts the scope for subsequent operations.

## Open questions

None.

## Notes for reflection

- The ADO project cap (1000 projects) forces acceptance tests to use pre-existing projects + create sub-resources (environments, groups) via direct ADO client calls rather than Terraform HCL. This is the established pattern for this repo.
- ADO security roles use `Access="assigned"` for explicit and `Access="inherited"` for inherited. The PATCH delete reverts from explicit to inherited, not full removal. Any destroy check must handle this distinction.
- `GetSecurityRoleAssignment` always returns a non-nil `*SecurityRoleAssignment` (even when not found — zero struct). Check `Identity == nil && Role == nil` for "not found".
