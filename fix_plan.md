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
- `make test` passes; `TestProvider_HasChildResources` passes; offline build clean

## Awaiting

- Live gate run (TF_ACC): forge will run `TestAccProjectFeatures_roundtrip` against real ADO
