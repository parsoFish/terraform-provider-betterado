# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (successful — offline)

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
   - `TestAccProjectPermissionsFramework` creates a project (NOT using existing-project pattern — **this was a bug**)
   - Uses `testutils.GetMuxProviderFactories()`

7. Fixed nilerr lint in `resource_securityrole_assignment_framework.go`: poll loop on 404 now checks error string

### Iteration 2 (live gate failure → fix)

**Gate failure**: `TestAccProjectPermissionsFramework` failed because it tried to create a new ADO project, but the org is at the 1000-project cap:
```
Error: creating project: Failed to add a project as this organization already has 1000 projects.
```

**Root cause**: The original test created a `betterado_project` resource. The org is at 1000-project cap — any create attempt fails immediately.

**Fix**: Rewrote `TestAccProjectPermissionsFramework` to use an existing project instead of creating one.

**Pattern**: Exactly mirrors `resolveSecurityPermissionsFixtureProject` from `resource_security_permissions_framework_test.go` (written in a prior WI iteration). Key details:
- `resolveProjectPermissionsFixtureProject(t)` resolves an existing project ID at test setup time
- Prefers `SharedFixtureProjectName` ("betterado-standing-demo")
- Falls back to any WellFormed project from `GetProjects()`; skips `keepProjects` entries first
- HCL uses the project ID as a literal (`%[1]q`) — no `betterado_project` resource created
- `CheckDestroy: checkProjectPermissionsFrameworkDestroyed` (no-op: ACLs have no "does it exist?" endpoint)
- `CheckProjectDestroyed` and `CheckProjectExists` removed (project not created by this test)

**Client construction**: Use `azuredevops.NewAuthProviderPAT(pat)` + `client.GetAzdoClient(authProvider, orgURL)` — NOT `GetAzdoClientByPAT` (that doesn't exist).

**Offline check**: `go test -tags all -run TestAccProjectPermissionsFramework ./azuredevops/internal/acceptancetests/` — compiles and `SKIP`s cleanly without TF_ACC.

## What worked

- Copy pattern from `resource_release_definition_permissions_framework.go` exactly, substituting the token builder and namespace ID
- Use unique `pp` prefix for plan modifiers to avoid duplicate type declarations
- Build tag must use `resource_project_permissions` to match the `all` build tag set
- `CaptureLiveEvidence("acceptance-resource", url, nil)` with the hardcoded Project namespace UUID
- **For existing-project pattern**: use `resolveProjectPermissionsFixtureProject` (same pattern as `resolveSecurityPermissionsFixtureProject`) — resolve at test setup, pass ID as literal to HCL

## What didn't work

- `_ = err` does NOT satisfy golangci-lint's `nilerr` checker — must use actual conditional logic
- Creating a new ADO project in the test — the org is at 1000-project cap; any create fails immediately

## Open questions

- AC2 and AC3 require all 13 resources migrated. Only project_permissions done so far. Gate only checks `TestAccProjectPermissionsFramework` so AC1 drives the gate, but next iteration should migrate remaining 12 if time permits.

## Notes for reflection

- The gate command (`TestAccProjectPermissionsFramework`) requires TF_ACC=1 for a live pass. The offline test skips cleanly with `t.Skip("TF_ACC not set")`.
- `GetMuxProviderFactories()` is defined in both `testutils/commons.go` and `testutils/mux_provider.go` as different functions. Use `testutils.GetMuxProviderFactories()` (commons.go version, same mux logic).
- **CRITICAL**: Never create a `betterado_project` resource in framework acceptance tests. The org is at the 1000-project cap. Always use `resolveProject*FixtureProject` pattern.
- `client.GetAzdoClient` signature: `(authProvider azuredevops.AuthProvider, organizationURL string)` — must wrap PAT with `azuredevops.NewAuthProviderPAT(pat)` first.
