# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (complete)

**Goal**: Migrate 4 graph data sources to terraform-plugin-framework.

**Approach**:
1. Created four new `datasource.DataSource` implementations in `azuredevops/internal/service/graph/`:
   - `datasource_descriptor_framework.go` — calls `GraphClient.GetDescriptor()`
   - `datasource_storage_key_framework.go` — calls `GraphClient.GetStorageKey()`
   - `datasource_group_framework.go` — reuses package-level helpers `getProjectDescriptor`, `getGroupsForDescriptor`, `selectGroup` from `data_group.go`; calls `GetStorageKey()` for group_id
   - `datasource_group_membership_framework.go` — reuses `getGroupMemberships()` from `resource_group_membership.go`
2. Removed all four keys from `provider.go` DataSourcesMap
3. Added all four constructors to `framework_provider.go` DataSources()
4. Updated `provider_test.go` TestProvider_HasChildDataSources to remove the 4 migrated entries
5. Created `data_graph_simple_framework_test.go` with `TestAccGraphSimpleDataSources_Framework` (top-level test with 4 subtests, each with idempotency re-plan step)

**Pattern followed**: Same as release framework datasources — struct with `*client.AggregatedClient`, Metadata/Schema/Configure/Read methods, model struct with tfsdk tags.

**Key gate insight**: The previous gate failure was `[no tests to run]` because `TestAccGraphSimpleDataSources_Framework` function didn't exist at all. Creating the function resolves this — the subtests now show as SKIP (not missing).

## What worked

- Reusing package-level helpers directly (getProjectDescriptor, getGroupsForDescriptor, selectGroup, getGroupMemberships) — no need to copy logic
- `types.ListValueFrom(ctx, types.StringType, memberDescriptors)` for converting []string to types.List in group_membership datasource
- Top-level wrapper test `TestAccGraphSimpleDataSources_Framework` calling `t.Run()` subtests — allows gate to target one function name while running all 4 data sources

## What didn't work

### Iteration 1 live gate
All 4 acceptance tests failed in live ADO run with:
`Error: creating project: Failed to add a project as this organization already has 1000 projects`

The test HCL used `resource "betterado_project" "test"` to create a fresh project for each sub-test.
The fix (iteration 2): replace project creates with `data "betterado_project" "shared"` lookup of
`SharedFixtureProjectName = "betterado-standing-demo"` — same pattern as TestAccGroupResource_Framework
and TestAccGroupMembership_Framework.

## Open questions

_(none)_

## Notes for reflection

- The `betterado_group` data source now exists in BOTH the framework and SDKv2 (SDKv2 `data_group.go` remains for its helper functions and the SDKv2 registration was removed). The helpers (`getProjectDescriptor`, etc.) stay in the SDKv2 file — they are package-level and accessible intra-package.
- The serviceendpoint test package has a pre-existing build failure (unrelated to WI-4) — `expandServiceEndpoint*` functions changed signatures but tests weren't updated. Out of scope for WI-4.
