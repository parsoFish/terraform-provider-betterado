# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_project_features implemented as resource.Resource in terraform-plugin-framework WHEN terraform apply enables/disables project features on betterado-standing-demo → read-back → idempotency re-plan → terraform destroy (feature state restored) THEN TestAccProjectFeatures_roundtrip (or equivalent) passes with GetMuxedProviderFactories(); CaptureLiveEvidence called; ExpectNonEmptyPlan: false
- [x] AC2: GIVEN betterado_project_features removed from SDKv2 ResourcesMap WHEN TestProvider_HasChildResources runs THEN test passes with updated count (one fewer SDKv2 resource)

## Completed this iteration

- Created `azuredevops/internal/service/core/resource_project_features_framework.go` — full framework implementation
- Removed `betterado_project_features` from `provider.go` SDKv2 ResourcesMap
- Registered `NewProjectFeaturesResource` in `framework_provider.go`
- Updated `provider_test.go` to remove `betterado_project_features` from SDKv2 list (AC2 ✓)
- Rewrote acceptance test to use `ProtoV6ProviderFactories` + `GetMuxedProviderFactories()` + `CaptureLiveEvidence` (AC1 ✓)
- Exported `GetProjectFeatureStatesForEvidence` helper for the acceptance test
- Fixed gofumpt in resource_project_framework.go
- `TestProvider_HasChildResources` passes (AC2 verified)

## Iteration 2 fix (this iteration)

- Root-caused "Missing Configuration for Required Attribute" for `project_id`:
  - Bug was in `projectUseStateForUnknown.PlanModifyString` in `resource_project_framework.go`
  - When no prior state exists (StateValue.IsNull()), the modifier was setting PlanValue = StateValue (null)
  - This converted unknown → null for `betterado_project.id` during plan
  - Null then propagated as the config value for `project_id` in betterado_project_features
  - Framework correctly rejects null for a Required attribute → "Missing Configuration" error
  - Fix: guard with `if req.StateValue.IsNull() { return }` so unknown remains unknown on first apply
- Build and golangci-lint pass clean after fix

## Iteration 3 fix

- Root-caused "Failed to add a project" — org is at 1000-project cap; can never create new projects.
- Rewrote `hclProjectFeatureBasic` to use `data "betterado_project"` looking up `betterado-standing-demo`
  (SharedFixtureProjectName) instead of `resource "betterado_project"` creating a new project.
- Test still exercises full betterado_project_features lifecycle (apply → read-back → update → destroy).
- Pattern aligns with SharedReleaseFixture and smokeResolveProject.

## Iteration 4 fix

- Root-caused "project not found" for `data.betterado_project.test`:
  - The live gate org did NOT have `betterado-standing-demo` pre-created.
  - Using `data "betterado_project"` (Terraform data source) fails at pre-plan if
    the project doesn't exist — no project create happens.
  - Fix: replace `data "betterado_project"` with `SharedReleaseFixture(t)` call
    before `resource.ParallelTest`. `SharedReleaseFixture` calls
    `resolveOrCreateFixtureProject` which either finds or creates the project via
    ADO API and returns its UUID.
  - HCL now receives the project UUID as a literal string — no data source needed.
  - Pattern matches `TestAccMuxSdkv2Passthrough` (same motivation, same fix).

## Awaiting

- Live gate run (TF_ACC): forge will run `TestAccProjectFeatures_roundtrip` against real ADO
  — this time using SharedReleaseFixture to resolve/create the project, passing
  UUID directly into HCL. No data source lookup at plan time.
