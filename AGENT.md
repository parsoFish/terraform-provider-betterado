# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (completed, commit 8a143226)

Implemented full framework migration from scratch in one iteration:

1. **resource_security_permissions_framework.go** — framework `resource.Resource`:
   - Mirrors SDKv2 `resource_security_permissions.go` ACE-bit logic exactly
   - Model: `id`, `namespace_id`, `token`, `principal`, `permissions` (MapAttribute), `replace` (BoolAttribute, default=true)
   - Inline plan modifiers: `spRequiresReplaceString()`, `spUseStateForUnknownString()`, `spStaticBool()`
   - Vendor has NO typed sub-packages like `booldefault`/`stringplanmodifier` — must use inline implementations (same as `resource_release_definition_permissions_framework.go`)

2. **data_security_namespace_framework.go** — framework `datasource.DataSource`

3. **data_security_namespace_token_framework.go** — framework `datasource.DataSource`; reuses `namespaceTokenTemplates` from same package

4. **data_security_namespaces_framework.go** — framework `datasource.DataSource`

5. **framework_provider.go** — added 1 resource + 3 data sources

6. **provider.go** — removed security import + commented out 4 registrations

7. **provider_test.go** — commented out 4 types from SDKv2 count lists

8. **resource_security_permissions_framework_test.go** — `TestAccSecurityPermissionsFramework` (mux, idempotency, live evidence)

9. **data_security_namespace_framework_test.go** — `TestAccDataSecurityNamespaceFramework`

10. **examples/resources/betterado_security_permissions/resource.tf** — created

## What worked

- Copying inline plan modifier pattern from `resource_release_definition_permissions_framework.go`
- Reusing `namespaceTokenTemplates` and `resolveIdentityDescriptor` from same `security` package
- Sharing `actionAttrTypes` var between namespace and namespaces data source files (same package)

## What didn't work

- Importing `booldefault` / `stringplanmodifier` typed sub-packages — NOT vendored

## Key patterns / gotchas

- `go test -run TestAccFoo ./pkg/` with no TF_ACC returns `ok ... [no tests to run]` which the forge gate REJECTS
- golangci-lint gocritic catches if-else chains that should be switch
- `types.ListValueMust` signature: `[]attr.Value`, NOT `[]interface{}`
- Unused struct types get flagged by `unused` linter — removed `actionModel` struct since it wasn't needed

## Iteration 2 (fix, commit 930af1f8)

Root cause of live gate failure (from `.forge/last-gate-failure.md`):
1. **1000-project cap**: `betterado_project.project` apply fails because the org is at its 1000-project cap
2. **nil-Meta panic**: After apply failure, destroy step runs `testutils.CheckProjectDestroyed` which calls `GetProvider().Meta().(*client.AggregatedClient)` — but `GetProvider()` returns the module-level singleton which was never configured when using `ProtoV6ProviderFactories` (mux path), so `Meta()` is nil → panic

### Fix applied:
- Switched to `SharedReleaseFixture(t)` for a pre-existing `betterado-standing-demo` project — no new project created
- Replaced `CheckProjectDestroyed` with local `checkSecurityPermissionsFrameworkDestroyed` that uses `getDirectClient()` (same pattern as `resource_task_group_test.go` and `resource_mux_sdkv2_passthrough_test.go`)
- Replaced `betterado_identity_group` (needs `project.name` for group lookup) with `betterado_group` (just `project_id` + short name `"Readers"`)
- Updated HCL: no `betterado_project` resource; project_id is a literal string from fixture

### Key pattern confirmed:
- **All tests using `ProtoV6ProviderFactories` (mux)** MUST NOT call `GetProvider().Meta()` in `CheckDestroy` or check functions — use `getDirectClient()` instead (defined in `resource_task_group_test.go`)
- `betterado_group` data source uses `descriptor` attr (= the resource ID); `betterado_identity_group` uses `subject_descriptor`
- `SharedReleaseFixture` skips automatically when `TF_ACC` is unset (line 79-81 in `shared_fixtures.go`)

## Iteration 3 (fix, commit 4772ff30)

Root cause of live gate failure (iteration 2):
- `SharedReleaseFixture` calls `resolveOrCreateFixtureProject` which tries GetProject("betterado-standing-demo").
- If that project doesn't exist in this org, it falls through to `QueueCreateProject` → fails at 1000-project cap.

### Fix applied:
- Replaced `SharedReleaseFixture(t)` entirely with `resolveSecurityPermissionsFixtureProject(t)`.
- New helper: tries `GetProject("betterado-standing-demo")` first (fast path); if that fails, falls back to `GetProjects(WellFormed)` and picks the first project.
- NEVER calls `QueueCreateProject` — zero project-creation path.
- Only needs a `ProjectID` string; the full release fixture (repo, build def, vg, release def) was overkill.

### Key pattern for this iteration:
- `SharedReleaseFixture` is the WRONG fixture for tests that only need a project ID. It provisions an entire release pipeline and tries to create the project if it doesn't exist.
- Use `resolveSecurityPermissionsFixtureProject` (or a similar minimal helper) for tests that just need an existing project.

## Open questions

_(none)_

## Notes for reflection

- The security namespace `sync.Once` cache in `data_security_namespaces.go` is NOT carried into framework implementation — each call queries fresh
