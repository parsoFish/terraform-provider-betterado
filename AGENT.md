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
- Identity user lookup: `"Project Collection Build Service"` with `search_filter = "DisplayName"` is a stable system identity in every ADO project.

## What didn't work

- `fmt.Sprintf` with `%[1]s` inside HCL template strings where `[%[1]s]` looks like HCL interpolation to terrafmt — causes parse error. Fix: pre-compute.
- Including unused import `identity` in provider.go after removing all three DataSourcesMap entries — causes compile error.

## Iteration 2 fix

**Problem:** Live gate failed with `Could not find user with name: Project Collection Build Service, with filter: DisplayName`.
"Project Collection Build Service" is not a valid display name in this ADO org.

**Root cause:** ADO build service accounts follow the naming convention `"{ProjectName} Build Service ({OrgName})"`.
The org name varies per ADO tenant — it is NOT "Project Collection Build Service" universally.

**Fix:** Updated `hclIdentityUserFrameworkRead()` in `data_identity_group_framework_test.go` to construct
the user name dynamically:
```go
return fmt.Sprintf(`
data "betterado_client_config" "current" {}

data "betterado_identity_user" "test" {
  name          = "%[1]s Build Service (${compact(split("/", data.betterado_client_config.current.organization_url))[2]})"
  search_filter = "DisplayName"
}
`, SharedFixtureProjectName)
```
This is identical to the pattern in `resource_git_permissions_test.go` (line 203), which is known to work.

## Open questions

_(none)_

## Notes for reflection

- The identity package helpers (`getIdentityGroupsWithProjectID`, `selectIdentityGroup`, `getIdentityUsersWithFilterValue`, `validateIdentityUser`) are all package-internal, so new framework files in the same package can call them directly — no export needed. Good pattern for future identity-package migrations.
- `ReadIdentities` with a comma-joined ID string is how the SDKv2 did it; the framework version follows the same pattern.
