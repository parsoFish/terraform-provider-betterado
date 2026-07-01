# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (this iteration)

**Problem (from last-gate-failure.md):** "Invalid Provider Server Combination: Duplicate resource type: betterado_release_folder" — mux fails to load provider schema because `betterado_release_folder` is registered in BOTH the SDKv2 provider (`azuredevops/provider.go`) and the framework provider (`azuredevops/internal/provider/framework_provider.go`).

**Fix applied:**
1. Removed `"betterado_release_folder": release.ResourceReleaseFolder()` from `azuredevops/provider.go` line 91 (SDKv2 resources map). Left a comment explaining why.
2. Updated `azuredevops/provider_test.go` `TestProvider_HasChildResources` to not expect `betterado_release_folder` in the SDKv2 resource map (it's now framework-only).
3. Ran `make fmt` to fix gofmt formatting issues (comment caused whitespace change).
4. All offline gates now pass: `make test` ✓, `golangci-lint run ./azuredevops/...` (0 issues) ✓, `make terrafmt-check` ✓.

**Commit:** `fix: remove betterado_release_folder from SDKv2 provider to eliminate mux duplicate`

## What worked

- `make fmt` is required after any edit to provider.go — gofmt enforced by `make test` (fmtcheck step).
- The `release` import in `provider.go` stays — it's still needed for data sources (`DataReleaseFolder()`, etc.).
- The data source `betterado_release_folder` (line 186 of provider.go) stays in SDKv2 — only the **resource** was migrated.
- `TestProvider_HasChildResources` in `azuredevops/provider_test.go` checks the SDKv2 resource map exhaustively — must be updated when a resource is migrated out.

## What didn't work

_(nothing to record yet — issue was clear from the gate failure message)_

### Iteration 2 (this iteration)

**Problem (from last-gate-failure.md):** "Failed to add a project as this organization already has 1000 projects." — the test HCL contained `resource "betterado_project" "fw_test"` which tried to create a new ADO project. The org is at its 1000-project cap.

**Fix applied:**
1. Rewrote `TestAccReleaseFolderFramework` in `resource_release_folder_framework_test.go` to call `SharedReleaseFixture(t)` to obtain `fixture.ProjectID` from the persistent `betterado-standing-demo` project.
2. Changed `hclReleaseFolderFramework(name string)` → `hclReleaseFolderFramework(name, projectID string)` — HCL no longer creates a project; passes `projectID` directly as a quoted string literal.
3. Removed `resource "betterado_project" "fw_test"` block from the HCL template entirely.
4. All offline gates pass: `make test` ✓, `golangci-lint run ./azuredevops/...` (0 issues) ✓, `make terrafmt-check` ✓.

**Commit:** `test: use SharedReleaseFixture in TestAccReleaseFolderFramework to avoid project create`

## What worked

- `make fmt` is required after any edit to provider.go — gofmt enforced by `make test` (fmtcheck step).
- The `release` import in `provider.go` stays — it's still needed for data sources (`DataReleaseFolder()`, etc.).
- The data source `betterado_release_folder` (line 186 of provider.go) stays in SDKv2 — only the **resource** was migrated.
- `TestProvider_HasChildResources` in `azuredevops/provider_test.go` checks the SDKv2 resource map exhaustively — must be updated when a resource is migrated out.
- **SharedReleaseFixture pattern**: call `SharedReleaseFixture(t)` to get a persistent project; use `fixture.ProjectID` in HCL string formatting. The fixture skips automatically if `TF_ACC` is not set. The project is NEVER deleted.

## What didn't work

- Creating projects via HCL in acceptance tests — the org is at the 1000-project cap and all creates fail.

### Iteration 3 (this iteration)

**Problem (from last-gate-failure.md):** Two panics in `TestAccReleaseFolderFramework`:
1. `checkReleaseFolderFrameworkDestroyed` panicked at line 73 — nil pointer dereference
2. `captureReleaseFolderFrameworkEvidence` panicked at line 104 — nil pointer dereference

Both called `testutils.GetProvider().Meta().(*client.AggregatedClient)`. Root cause: when `ProtoV6ProviderFactories` is used, the mux server builds its own SDKv2 provider instance — the package-level `var provider = azuredevops.Provider()` singleton (what `GetProvider()` returns) is never `Configure()`d by Terraform, so `.Meta()` returns `nil`. The type assertion `.(type)` on nil panics.

**Fix applied:**
1. Replaced both `testutils.GetProvider().Meta().(*client.AggregatedClient)` calls with `getDirectClient()`.
   - `getDirectClient()` is already defined in `resource_task_group_test.go` (same `acceptancetests` package) — builds `AggregatedClient` from `AZDO_ORG_SERVICE_URL` + `AZDO_PERSONAL_ACCESS_TOKEN` env vars directly.
   - In `checkReleaseFolderFrameworkDestroyed`: error from `getDirectClient()` propagates (hard error — test must be able to verify destroy).
   - In `captureReleaseFolderFrameworkEvidence`: error from `getDirectClient()` returns `nil` (best-effort, never fail test on evidence capture).
2. Removed unused `"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"` import.
3. All offline gates pass: `make test` ✓, `golangci-lint run ./azuredevops/...` (0 issues) ✓, `make terrafmt-check` ✓.

**Commit:** `fix: use getDirectClient in framework test destroy/evidence helpers`

## What worked

- `make fmt` is required after any edit to provider.go — gofmt enforced by `make test` (fmtcheck step).
- The `release` import in `provider.go` stays — it's still needed for data sources (`DataReleaseFolder()`, etc.).
- The data source `betterado_release_folder` (line 186 of provider.go) stays in SDKv2 — only the **resource** was migrated.
- `TestProvider_HasChildResources` in `azuredevops/provider_test.go` checks the SDKv2 resource map exhaustively — must be updated when a resource is migrated out.
- **SharedReleaseFixture pattern**: call `SharedReleaseFixture(t)` to get a persistent project; use `fixture.ProjectID` in HCL string formatting. The fixture skips automatically if `TF_ACC` is not set. The project is NEVER deleted.
- **getDirectClient() pattern**: when an acceptance test uses `ProtoV6ProviderFactories` (mux), never call `testutils.GetProvider().Meta()` — the SDKv2 singleton is never configured. Use `getDirectClient()` from `resource_task_group_test.go` instead.

## What didn't work

- Creating projects via HCL in acceptance tests — the org is at the 1000-project cap and all creates fail.
- `testutils.GetProvider().Meta()` returns nil under ProtoV6ProviderFactories → panic on type assertion.

## Open questions

- Will the live acceptance test (`TestAccReleaseFolderFramework` with TF_ACC=1) now pass? The project-create error is fixed, the nil-Meta panic is fixed. The resource CRUD logic should work. The next gate run will verify.

## Notes for reflection

- Pattern: when migrating a resource from SDKv2 to framework via mux, always remove it from both `provider.go` resource map AND `provider_test.go` `TestProvider_HasChildResources`. Otherwise the duplicate causes "Invalid Provider Server Combination" at plan time.
- Pattern: ALL release acceptance tests that need a project must use `SharedReleaseFixture(t)` — never create a project via Terraform HCL because the org is at its 1000-project cap.
