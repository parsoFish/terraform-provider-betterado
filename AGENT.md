# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1

The gate failure on iteration 0 was:
```
--- FAIL: TestAccProjectFeatures_EnableUpdateFeature (0.13s)
    resource_project_features_test.go:15: Step 1/2 error: Error running pre-apply plan: exit status 1
    Error: Invalid resource type
      on terraform_plugin_test.tf line 12, in resource "betterado_project" "test":
      12: resource "betterado_project" "test" {
    The provider hashicorp/betterado does not support resource type "betterado_project".
```

Root cause: the old test used `ProviderFactories: testutils.GetProviderFactories()` (SDKv2-only), but
`betterado_project` had already been migrated to the framework provider in WI-2. Also, `betterado_project_features` was not yet a framework resource.

**Actions taken:**
1. Created `resource_project_features_framework.go` — full framework implementation.
2. Removed `betterado_project_features` from `provider.go` SDKv2 `ResourcesMap`.
3. Registered `NewProjectFeaturesResource` in `framework_provider.go Resources()`.
4. Updated `provider_test.go`: removed `betterado_project_features` from the SDKv2 list.
5. Rewrote acceptance test: `TestAccProjectFeatures_roundtrip`, `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`, `ExpectNonEmptyPlan: false`, `captureProjectFeaturesEvidence()`.
6. Fixed pre-existing `gofumpt` issue in `resource_project_framework.go`.

### Iteration 2 (this iteration)

Gate failure after iteration 1:
```
Error: Missing Configuration for Required Attribute
  with betterado_project_features.test, line 21: project_id = betterado_project.test.id
  Must set a configuration value for the project_id attribute as the provider has marked it as required.
```

**Root cause identified:**
- `projectUseStateForUnknown.PlanModifyString` in `resource_project_framework.go` had a bug:
  ```go
  if !req.PlanValue.IsUnknown() { return }
  resp.PlanValue = req.StateValue  // BUG on first apply: StateValue is null
  ```
- On first apply (no prior state), `req.StateValue` is `null`.
- The modifier was setting `resp.PlanValue = null`, converting `unknown → null`.
- The `betterado_project.id` plan value thus became `null` instead of `unknown`.
- Downstream `betterado_project_features.project_id = betterado_project.test.id` resolved to `null` config.
- Framework validation: `Required + configHasNullValue` → "Missing Configuration for Required Attribute".

**Fix:**
Added `if req.StateValue.IsNull() { return }` guard before assigning `resp.PlanValue = req.StateValue`.
This keeps the plan value as **unknown** when there is no prior state, allowing proper unknown propagation
to downstream resource references.

**Location:** `azuredevops/internal/service/core/resource_project_framework.go`, `projectUseStateForUnknown.PlanModifyString`

## What worked

- Reusing `projectStateForUnknown()` / `projectForceReplace()` from the same package avoids the
  `stringplanmodifier` package that isn't in vendor.
- `featuresFromMap(ctx, types.Map) (map[string]string, diag.Diagnostics)` via `m.ElementsAs()` is
  the correct framework way to extract string map elements.
- For `CaptureLiveEvidence` best-effort: wrap in `if err == nil { ... }` nesting rather than
  `if err != nil { return nil }` to avoid the `nilerr` golangci-lint finding.

## What didn't work

- `stringplanmodifier.UseStateForUnknown()` / `stringplanmodifier.RequiresReplace()` — this helper
  package is NOT in the vendor dir. Use the inline plan-modifier structs from `resource_project_framework.go`
  instead (they're in the same `core` package and already exported as `projectStateForUnknown()` etc.).
- **`projectUseStateForUnknown` without null-state guard** — DO NOT use `resp.PlanValue = req.StateValue`
  unconditionally. Always guard with `if req.StateValue.IsNull() { return }` first.

## Key lessons for next iteration

- **Framework `UseStateForUnknown` pattern requires null guard:** Always check `req.StateValue.IsNull()`
  before assigning state to plan. Without the guard, unknown → null conversion breaks cross-resource refs.
- **"Missing Configuration for Required Attribute"** — this means the framework sees a null config value
  for a Required attribute. Trace it to the plan modifier that is producing null instead of unknown.
- `graph` and `serviceendpoint` packages have pre-existing test build failures (unrelated to WI-3).
- The quality gate command is in `.forge/quality_gate_cmd`; it tests release + taskagent packages, not acceptance tests.

### Iteration 3 (this iteration)

Gate failure after iteration 2:
```
--- FAIL: TestAccProjectFeatures_roundtrip (0.98s)
    resource_project_features_test.go:27: Step 1/2 error: Error running apply: exit status 1
    Error: creating project
      with betterado_project.test, line 12
      Failed to add a project as this organization already has 1000 projects.
```

**Root cause:**
- The test used `resource "betterado_project" "test"` with a dynamically generated name.
- ADO org is at the 1000-project cap — any project-create attempt fails.
- This is org-level infrastructure constraint, not a code bug in the resource itself.
- `shared_fixtures.go` already documents this in `SharedFixtureProjectName` const comment:
  "ADO soft-deletes projects (28-day retention) and soft-deleted projects count toward
  the 1000-project cap, so a create+delete per run fills the recycle bin."

**Fix:**
- Rewrote `hclProjectFeatureBasic` to use `data "betterado_project" "test"` looking up
  `betterado-standing-demo` (= `SharedFixtureProjectName`) instead of `resource "betterado_project"`.
- `betterado_project` data source is registered in `framework_provider.go` via `core.NewProjectDataSource`.
- `betterado_project_features` manage features on this existing project — no new project created.
- Pattern matches `resource_state_upgrade_smoke_test.go`'s `smokeResolveProject` + data source approach.
- `resource_mux_sdkv2_passthrough_test.go` does the same via `SharedReleaseFixture`.

**Verification:**
- `go build -tags all ./azuredevops/internal/acceptancetests/` — clean.
- `golangci-lint run --new-from-rev=main ./azuredevops/...` — 0 issues.
- `make test` — all pass.
- `go test -tags all -run TestAccProjectFeatures ./azuredevops/internal/acceptancetests/ -v` — skips
  without TF_ACC (as expected); will run live with TF_ACC set.

### Iteration 4 (this iteration)

Gate failure after iteration 3:
```
--- FAIL: TestAccProjectFeatures_roundtrip (0.58s)
    resource_project_features_test.go:35: Step 1/2 error: Error running pre-apply plan: exit status 1
    Error: project not found
      with data.betterado_project.test, line 12
    Project with name or ID "betterado-standing-demo" does not exist
```

**Root cause:**
- The test used `data "betterado_project" "test"` to look up `betterado-standing-demo` in the
  live gate ADO org. However, `betterado-standing-demo` does NOT yet exist in that org.
- A Terraform data source read at plan time goes directly to ADO API — if the project isn't there,
  `GetProject` returns 404 → `ResponseWasNotFound` → "project not found" error → plan fails.
- `resolveOrCreateFixtureProject` (in `SharedReleaseFixture`) handles the same scenario correctly:
  it tries `GetProject`, and if the project doesn't exist, it creates it.

**Fix:**
- Replaced `data "betterado_project"` lookup with a call to `SharedReleaseFixture(t)` BEFORE
  the `resource.ParallelTest` call.
- `SharedReleaseFixture` returns `SharedFixtureResult.ProjectID` (a UUID string).
- `hclProjectFeatureBasic` now takes `projectID` (UUID) instead of `projectName` and injects
  the UUID as a string literal — no Terraform data source needed.
- Pattern is identical to `TestAccMuxSdkv2Passthrough`.

**Key lesson:**
- `data "betterado_project"` in test HCL assumes the project exists at plan time.
  If the project may not exist, ALWAYS use `SharedReleaseFixture(t)` to get the project ID
  via `resolveOrCreateFixtureProject`, then pass the UUID directly into HCL.
- `SharedReleaseFixture` calls `t.Skip` if `TF_ACC` is not set — so the test still correctly
  skips in the offline unit-test environment.

## Open questions

_(none blocking)_
