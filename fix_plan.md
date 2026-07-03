# Fix Plan

> Checklist for WI-4. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a live ADO org with TF_ACC=1, AZDO_ORG_SERVICE_URL, and AZDO_PERSONAL_ACCESS_TOKEN set WHEN go test -tags all -run TestAccFeatureFlag ./azuredevops/internal/acceptancetests/ runs THEN the test applies a betterado_feature_flag targeting a known project-scoped feature (e.g. ms.vss-work.agile on the betterado-standing-demo project), reads back state, asserts ExpectNonEmptyPlan:false (idempotency), then destroys and confirms the feature reverts to undefined/default
- [x] AC2: GIVEN the live read-back step in TestAccFeatureFlag WHEN CaptureLiveEvidence is called with label 'acceptance-resource' THEN .forge/live-evidence/acceptance-resource.json is written with a real REST GET URL from the featuremanagement endpoint

## Status

Both ACs are implemented in `azuredevops/internal/acceptancetests/resource_feature_flag_test.go`:

- TestAccFeatureFlag_basic: 3-step test (enable → disable → idempotency) + checkFeatureFlagDestroyed
- captureFeatureFlagEvidence: calls testutils.CaptureLiveEvidence("acceptance-resource", url, state) where url
  is the real REST GET for FeatureManagement (orgURL/_apis/FeatureManagement/FeatureStates/host/project/{projectId}/{featureId}?api-version=7.1-preview.1)

Gate will show [no tests to run] in offline mode (no TF_ACC) — this is expected.
Live gate (TF_ACC=1 + env creds) will execute the test.

---

## Iteration 1 fix (root cause of gate failure)

The gate failed with "Invalid resource type `betterado_feature_flag`" because the resource
implementation was never registered in the mux provider.

**Fixed in:** `azuredevops/internal/provider/framework_provider.go`
- Added `featuremanagement` import
- Added `featuremanagement.NewFeatureFlagResource` to `Resources()` slice

All unit tests pass, build is clean, test is discoverable. Live gate should now pass.
