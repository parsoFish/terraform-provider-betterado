# AGENT.md — WI-3 Institutional Memory

## What this WI does

Migrates 5 release data sources from SDKv2 to terraform-plugin-framework and registers
them in the mux framework provider, while updating acceptance tests to use mux provider
factories.

## Critical facts (do NOT re-research)

### The gate failure (iteration 0)

- `TestAccDataReleaseFolder_Basic` failed: "The provider does not support resource type betterado_release_folder"
- Root cause: the test used `Providers: testutils.GetProviders()` (SDKv2 only)
- The resource `betterado_release_folder` was already migrated to framework provider in prior WI
- SDKv2 provider no longer has it → test can't find it

### What was done in iteration 0

1. Created 5 framework datasource files in `azuredevops/internal/service/release/`:
   - `datasource_release_definition_framework.go` → `NewReleaseDefinitionDataSource()`
   - `datasource_release_definition_history_framework.go` → `NewReleaseDefinitionHistoryDataSource()`
   - `datasource_release_definition_revision_framework.go` → `NewReleaseDefinitionRevisionDataSource()`
   - `datasource_release_definitions_framework.go` → `NewReleaseDefinitionsDataSource()`
   - `datasource_release_folder_framework.go` → `NewReleaseFolderDataSource()`

2. Registered all 5 in `framework_provider.go` `DataSources()` (was returning empty slice)

3. Removed all 5 data sources from `provider.go` `DataSourcesMap` (mux duplicate prevention)
   - Also removed the now-unused `release` package import from `provider.go`

4. Updated `azuredevops/provider_test.go` `TestProvider_HasChildDataSources` to not expect
   the 5 migrated data sources (they're framework now, not SDKv2)

5. Updated `data_release_folder_test.go`:
   - Changed `Providers: testutils.GetProviders()` → `ProtoV6ProviderFactories: testutils.GetMuxProviderFactories()`
   - Removed `Providers` field, added `ProtoV6ProviderFactories`

6. Build + `make test` + `golangci-lint` + `terrafmt-check` all pass locally

### File layout for framework datasources

Pattern: each datasource file has:
- `var _ datasource.DataSource = &<name>DataSource{}` (compile-time interface check)
- `New<Name>DataSource() datasource.DataSource` constructor
- `<name>DataModel` struct with `tfsdk` tags
- `Metadata()`, `Schema()`, `Configure()`, `Read()` methods

### Data source schemas (match SDKv2 predecessor fields exactly)

- `betterado_release_definition`: project_id (R), release_definition_id (O+C), name (O+C), path (C), description (C), release_name_format (C), id (C)
- `betterado_release_definition_history`: project_id (R), release_definition_id (R), revisions (list: revision/changed_by/changed_date/change_type/comment), id (C)
- `betterado_release_definition_revision`: project_id (R), release_definition_id (R), revision (R), json_content (C), id (C)
- `betterado_release_definitions`: project_id (R), path (O+C), name (O+C), release_definitions (list: id/name/path), id (C)
- `betterado_release_folder`: project_id (R), path (R), description (C), id (C)

### Tests that use which provider factory (IMPORTANT)

- `data_release_definition_test.go` → `ProtoV6ProviderFactories: testutils.GetMuxProviderFactories()` ✓
- `data_release_definition_revision_history_test.go` → `ProtoV6ProviderFactories: testutils.GetMuxProviderFactories()` ✓
- `data_release_folder_test.go` → `ProtoV6ProviderFactories: testutils.GetMuxProviderFactories()` ✓ (fixed in iter 0)

### Gate failure (iteration 1) and fix

**Failure:** `TestAccDataReleaseFolder_Basic` failed:
```
Error: creating project: Failed to add a project as this organization already has 1000 projects.
```

**Root cause:** `hclDataReleaseFolderBasic` created a `betterado_project` resource in the HCL config. Org is at 1000-project cap.

**Fix applied:** Rewrote `data_release_folder_test.go` to:
1. Call `fixture := SharedReleaseFixture(t)` in the test function
2. Pass `fixture` to `hclDataReleaseFolderBasic(name, fixture)`
3. Change `hclDataReleaseFolderBasic` signature to `(name string, fixture SharedFixtureResult)` and use `fixture.ProjectID` directly (no `betterado_project` resource created)

This matches the exact pattern used in `resource_release_folder_framework_test.go` (commit `1a941e31`).

### What remains

- Live TF_ACC gate run to confirm all 6 acceptance tests pass against real Azure DevOps
- The orchestrator decides done vs failed based on the gate result

### Known patterns in this codebase

- `flattenReleaseFolder(d, &folder, projectID)` is SDKv2 helper in resource_release_folder.go
- For framework datasources, replicate logic inline in Read()
- `d.client.ReleaseClient.*` — use `d.client.Ctx` (not `context.Background()`)
- `GetMuxProviderFactories()` in testutils creates the mux with both SDKv2+framework
