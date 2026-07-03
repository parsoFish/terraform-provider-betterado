# Agent Memory — WI-9

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (2025-07)

Full migration of betterado_variable_group resource + data source to terraform-plugin-framework.

**Created:**
- `azuredevops/internal/service/taskagent/resource_variable_group_framework.go`
  - VariableGroupResource: Create/Read/Update/Delete/ImportState
  - Uses package-level `defaultString()`, `defaultBool()`, `defaultInt64()`, `requiresReplace()`, `useStateForUnknown()` helpers from resource_task_group_framework.go (same package)
  - `types.SetNestedAttribute` for variable set; `types.ListNestedAttribute` for key_vault
  - `allow_access` via `build.AuthorizeProjectResources` / `build.GetProjectResources`
  - Secret values recovered from prior state (API returns empty for secrets)
  - `searchAzureKVSecrets` called from resource_variable_group.go (same package — kept for WI-10 dependency)
  - `splitImportID` for import ID parsing ("projectID/vgIntID")
  - Retry loop for Create to wait for variable count to stabilize
- `azuredevops/internal/service/taskagent/data_variable_group_framework.go`
  - VariableGroupDataSource: Read only, uses GetVariableGroups by name

**Modified:**
- `azuredevops/internal/provider/framework_provider.go`: added NewVariableGroupResource + NewVariableGroupDataSource
- `azuredevops/provider.go`: removed betterado_variable_group from ResourcesMap and DataSourcesMap (kept comment explaining removal)
- `azuredevops/provider_test.go`: removed betterado_variable_group from both SDKv2 count lists
- `azuredevops/internal/acceptancetests/resource_variable_group_test.go`: rewrote to use ProtoV6ProviderFactories, fixture project (`betterado-standing-demo`), GetDirectClient for CheckDestroy, CaptureLiveEvidence in TestAccVariableGroup_basic
- `azuredevops/internal/acceptancetests/data_variable_group_test.go`: rewrote to use inline HCL with fixture project (avoids HclVariableGroupResource which references `betterado_project.project` resource)

**Deleted:**
- `azuredevops/internal/service/taskagent/data_variable_group.go` (SDKv2, replaced by framework)
- `azuredevops/internal/service/taskagent/data_variable_group_test.go` (all-commented-out, deleted)

**Kept:**
- `azuredevops/internal/service/taskagent/resource_variable_group.go` — NOT deleted because resource_variable_group_variable.go calls `updateVariableGroup` from it, and resource_variable_group_framework.go calls `searchAzureKVSecrets` and `isKeyVaultVariableGroupType`. WI-10 will handle resource_variable_group_variable.

## What worked

- Import sub-packages like `booldefault`, `stringdefault`, `stringplanmodifier` are NOT vendored. Must use the custom package-level helpers defined in `resource_task_group_framework.go` (same package): `defaultString()`, `defaultBool()`, `defaultInt64()`, `requiresReplace()`, `useStateForUnknown()`.
- Fixture project pattern: `data "betterado_project" "fixture" { name = "betterado-standing-demo" }` avoids 1000-project limit.
- `GetDirectClient()` from testutils is required for CheckDestroy when using ProtoV6ProviderFactories (SDKv2 provider singleton's Meta() is nil in mux mode).
- `variableAttrTypes` and `keyVaultAttrTypes` must be declared at package level as `map[string]attr.Type` for use with `types.ObjectValueFrom` and `types.SetValue`.
- Watch out for variable shadowing: `d *VariableGroupDataSource` receiver shadows `:=` assignments. Use distinct variable names (e.g., `dSet` instead of `d`) for diagnostics from nested calls.
- The data source test should not use `HclVariableGroupResource` helper because it references `betterado_project.project` (resource), not a data source lookup. Write inline HCL instead.

## What didn't work

- Using `booldefault`, `stringdefault`, `stringplanmodifier` from typed sub-packages — not in vendor directory, causes build failure.
- Using `:=` with `d` as both a diag result AND the method receiver `d *VariableGroupDataSource` — causes compile error due to shadowing.

## Open questions

- None at this time.

## Notes for reflection

- Pattern: custom helper functions for defaults and plan modifiers are reused across all framework resources in the taskagent package. New framework resources must use these, not the vendored typed sub-packages.
- Pattern: fixture project lookup for acceptance tests that avoid the 1000-project limit.
- The data source helper `HclVariableGroupResource` (testutils/hcl.go) still references `betterado_project.project` (a resource). Since this WI migrated the VG to framework using a data source lookup instead, there may be a future cleanup opportunity to update this helper too.
