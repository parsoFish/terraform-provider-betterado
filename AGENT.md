# AGENT.md — WI-1 Institutional Memory

## Work Item
Migrate `betterado_release_folder` to terraform-plugin-framework, wiring it through the mux provider.

## Status: COMPLETE (all ACs satisfied)

## What Was Built

### Core Resource (AC1)
`azuredevops/internal/service/release/resource_release_folder_framework.go`:
- Full framework resource: Metadata, Schema, Configure, Create, Read, Update, Delete, ImportState
- Model: `releaseFolderModel` with id, project_id, path, description
- Schema: id (Computed+useStateForUnknown), project_id (Required+requiresReplace), path (Required+requiresReplace), description (Optional+Computed, default "")
- Create: calls `ReleaseClient.CreateFolder`, sets ID=path
- Read: calls `ReleaseClient.GetFolders`, removes from state on 404
- Update: only description is updatable (path is ForceNew); calls `ReleaseClient.UpdateFolder`
- Delete: calls `ReleaseClient.DeleteFolder`
- ImportState: parses "<project_id>/<path>" format

### Acceptance Test (AC2)
`azuredevops/internal/acceptancetests/resource_release_folder_framework_test.go`:
- `TestAccReleaseFolderFramework`: uses `SharedReleaseFixture` to reuse existing project (avoids 1000-project cap hit)
- Step 1: apply + assert attributes + `captureReleaseFolderFrameworkEvidence`
- Step 2: idempotency check with `ExpectNonEmptyPlan: false`
- `checkReleaseFolderFrameworkDestroyed`: uses `getDirectClient()` (not SDKv2 singleton Meta) to verify destroy
- `captureReleaseFolderFrameworkEvidence`: calls `CaptureLiveEvidence("acceptance-resource", vsrmURL, folder)` — writes `.forge/live-evidence/acceptance-resource.json`
- Uses `ProtoV6ProviderFactories: testutils.GetMuxProviderFactories()` for mux path
- Build tag: `//go:build (all || resource_release_folder) && !exclude_resource_release_folder`

### Provider Registration (AC1/AC2)
`azuredevops/internal/provider/framework_provider.go`:
- `release.NewReleaseFolderResource` added to Resources() slice

### Live Evidence (AC2)
`.forge/live-evidence/acceptance-resource.json` was written during a live acceptance test run.
The evidence URL is: `https://vsrm.dev.azure.com/davidgparsonson/<projectID>/_apis/release/folders%5CAccTestFW-<name>?api-version=7.1`

## CI Gate Results (AC3)
- `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` → EXIT:0
- `golangci-lint run ./azuredevops/...` → 0 issues
- `make terrafmt-check` → EXIT:0

## Key Patterns / Anti-Patterns

### DO
- Use `SharedReleaseFixture(t)` to get a pre-existing project — the org is at 1000-project cap
- Use `getDirectClient()` in destroy/evidence helpers (not `testutils.GetProvider().Meta()`) — ProtoV6ProviderFactories doesn't configure the SDKv2 singleton
- Register helpers in `framework_defaults.go` (useStateForUnknown, requiresReplace, staticString already there)
- Wire resources in `framework_provider.go` Resources() slice only — no main.go changes needed

### DON'T
- Don't add `betterado_release_folder` to the SDKv2 provider (provider.go) — it was removed to avoid mux duplicate
- Don't call `testutils.GetProvider().Meta()` in ProtoV6 test helpers — Meta is nil there

## Other Resources Built This Branch (context)
- `datasource_release_folder_framework.go` — betterado_release_folder data source
- `datasource_release_definition_framework.go` — betterado_release_definition data source
- `datasource_release_definition_history_framework.go`
- `datasource_release_definition_revision_framework.go`
- `datasource_release_definitions_framework.go`
- `resource_release_definition_permissions_framework.go` — permissions resource
- All registered in framework_provider.go

## Iteration History
- Iteration 0 (prior branch): built resource_release_folder_framework.go skeleton + test + provider registration
- Iteration 1 (prior): migrated data sources, added SharedReleaseFixture
- Iteration 2 (prior): fixed getDirectClient usage, added CaptureLiveEvidence
- Iteration 3 (prior): fixed nil-Meta issue in framework test helpers
- Current iteration: verified all gates pass (no code changes needed); updated fix_plan + AGENT
