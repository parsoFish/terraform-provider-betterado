# Agent Memory — WI-6

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (COMPLETED)

Created all three framework data source implementations and wired them up:

1. **datasource_identity_group_framework.go** — `identityGroupDataSource` struct, Metadata/Schema/Configure/Read. Reuses package-internal helpers `getIdentityGroupsWithProjectID` and `selectIdentityGroup` from data_identity_group.go. Calls `IdentityClient.ReadIdentity` to get `SubjectDescriptor`.

2. **datasource_identity_groups_framework.go** — `identityGroupsDataSource` struct. Uses `SetNestedAttribute` with `identityGroupItemAttrTypes`. Calls `getIdentityGroupsWithProjectID` then `IdentityClient.ReadIdentities` with comma-joined IDs. Builds `types.Set` via `types.SetValueFrom`.

3. **datasource_identity_user_framework.go** — `identityUserDataSource` struct. `search_filter` is both Optional and Computed (defaults to "General" in Read logic). Reuses `getIdentityUsersWithFilterValue` and `validateIdentityUser` from data_identity_user.go.

4. **framework_provider.go** — Added `identity` import, added `identity.NewIdentityGroupDataSource`, `identity.NewIdentityGroupsDataSource`, `identity.NewIdentityUserDataSource` to `DataSources()`.

5. **provider.go** — Removed three identity entries from `DataSourcesMap`, removed unused `identity` import.

6. **provider_test.go** — Removed three identity entries from `expectedDataSources` list, added comment.

7. **data_identity_group_framework_test.go** — `TestAccIdentityDataSources_Framework` dispatching 3 sub-tests via `t.Run`. Uses `testutils.GetMuxedProviderFactories()` and `SharedFixtureProjectName` (persistent project). Group name built in Go before embedding in HCL (terrafmt requires no `%[n]s` with brackets in HCL strings).

8. **docs/** — Ran `make docs` → regenerated identity_group.md, identity_groups.md, identity_user.md.

9. **CHANGELOG.md** — Added three entries under `[Unreleased] / FEATURES`.

## What worked

- Reusing intra-package helpers from SDKv2 files — same package, no export needed.
- `types.SetValueFrom(ctx, types.ObjectType{AttrTypes: ...}, items)` pattern for nested set (same as graph/datasource_groups_framework.go).
- `search_filter` needs both `Optional: true, Computed: true` because it has a default that the server/Read computes (otherwise plan detects "value was set in config" vs "value must be computed").
- For terrafmt: avoid Go format directives with brackets (like `[%[1]s]`) directly inside HCL string literals. Pre-compute the string in Go and use `%[2]q` to embed the result as a quoted literal.
- Identity group name format in ADO: `[ProjectName]\GroupName` — e.g. `[betterado-standing-demo]\Build Administrators`.
- **"Project Collection Build Service ({OrgName})"** is the CORRECT stable user identity for identity_user tests. See below.

## What didn't work

- `fmt.Sprintf` with `%[1]s` inside HCL template strings where `[%[1]s]` looks like HCL interpolation to terrafmt — causes parse error. Fix: pre-compute.
- Including unused import `identity` in provider.go after removing all three DataSourcesMap entries — causes compile error.
- **"{ProjectName} Build Service ({OrgName})"** — project-scoped build service account only exists after a pipeline has been run in that project. The `betterado-standing-demo` project has NOT had pipelines run, so this identity doesn't exist. DO NOT use this for tests with the standing demo project.

## Iteration 2 fix (FAILED in live gate)

**Problem:** Live gate failed with `Could not find user with name: Project Collection Build Service, with filter: DisplayName`.
Tried hard-coding "Project Collection Build Service" — that's not the full name; it needs the org name appended.

## Iteration 3 fix

**Problem:** Live gate failed with `Could not find user with name: betterado-standing-demo Build Service (davidgparsonson), with filter: DisplayName`.

**Root cause:** The per-project build service `"{ProjectName} Build Service ({OrgName})"` only exists after a pipeline has been run in that project. The `betterado-standing-demo` project has NOT had pipelines run, so this account doesn't exist.

**Fix:** Use `"Project Collection Build Service ({OrgName})"` instead. This is the **collection-level** build service that is a permanent system identity in every ADO org, regardless of pipeline history.

Updated `hclIdentityUserFrameworkRead()` in `data_identity_group_framework_test.go`:
```go
return `
data "betterado_client_config" "current" {}

data "betterado_identity_user" "test" {
  name          = "Project Collection Build Service (${compact(split("/", data.betterado_client_config.current.organization_url))[2]})"
  search_filter = "DisplayName"
}
`
```

The org name (`davidgparsonson`) is extracted dynamically at apply time. The "Project Collection Build Service" prefix (without a project name) is the collection-level identity.

## Key facts about ADO build service identities

| Identity | Format | When exists |
|----------|--------|-------------|
| Collection-level | `Project Collection Build Service ({OrgName})` | ALWAYS — permanent system identity |
| Project-level | `{ProjectName} Build Service ({OrgName})` | Only after a pipeline run in that project |

**For tests with the standing demo project: always use the collection-level identity.**

## Open questions

_(none)_

## Notes for reflection

- The identity package helpers (`getIdentityGroupsWithProjectID`, `selectIdentityGroup`, `getIdentityUsersWithFilterValue`, `validateIdentityUser`) are all package-internal, so new framework files in the same package can call them directly — no export needed. Good pattern for future identity-package migrations.
- `ReadIdentities` with a comma-joined ID string is how the SDKv2 did it; the framework version follows the same pattern.
