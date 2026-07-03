# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## What I've tried

### Iteration 0 (this iteration)

**Problem:** Gate returned `[no tests to run]` because `resource_feature_flag_test.go` did not exist yet.

**Action:** Created `azuredevops/internal/acceptancetests/resource_feature_flag_test.go` with:
- Build tag: `//go:build (all || resource_feature_flag) && !exclude_resource_feature_flag`
- `TestAccFeatureFlag_basic` — 3 steps: enable, disable, idempotency re-plan
- `hclFeatureFlagBasic(state string)` — HCL using `data "betterado_project" "demo"` with `SharedFixtureProjectName` + `betterado_feature_flag.test` resource
- `checkFeatureFlagDestroyed` — reads back state via `FeatureManagementClient.GetFeatureStateForScope`; fails if still "enabled"
- `captureFeatureFlagEvidence` — calls `testutils.CaptureLiveEvidence("acceptance-resource", url, state)` with the real REST URL

**Key decisions:**
- Used `getDirectClient()` (already defined in `resource_task_group_test.go` in same package) — no need to redefine
- `UserScope` for `GetFeatureStateForScope` must be the scope name (e.g. "project"), not "host" or "me"
- Evidence URL format: `{orgURL}/_apis/FeatureManagement/FeatureStates/host/{scopeName}/{scopeValue}/{featureId}?api-version=7.1-preview.1`
- `CaptureLiveEvidence` already exists in `testutils/live_evidence.go` with signature `(label, url string, response interface{}) error`

## What worked

- `go build -tags all ./azuredevops/internal/acceptancetests/` — clean
- `go vet -tags all ./azuredevops/internal/acceptancetests/` — clean
- `go test -tags all -run TestAccFeatureFlag -list '.*' ...` — lists `TestAccFeatureFlag_basic` (test is discoverable)
- `make test` — 12 packages pass, no failures

## What didn't work

N/A — first iteration found the test file missing and created it correctly.

## Notes for reflection

- **CORRECTED (Iter 2):** `UserScope` must be `"host"` (NOT `"project"`). The URL path is `FeatureStates/{userScope}/{scopeName}/{scopeValue}/{featureId}` where userScope="host", scopeName="project", scopeValue={projectId}.
- Evidence URL format: `/_apis/FeatureManagement/FeatureStates/host/project/{projectId}/{featureId}?api-version=7.1-preview.1`

---

### Iteration 2

**Problem:** Gate failed with:
```
SetFeatureStateForScope failed for feature "ms.vss-work.agile" scope "project"/"6ddb680c-...": userId
```

**Root cause:** The ADO FeatureManagement SDK distinguishes two separate "scope" concepts:
- `UserScope` — who the setting applies to: `"me"` (current user) or `"host"` (all users / organisation-wide). This is a **route parameter** that determines which feature store to use. For org-wide project-scoped features it must be `"host"`.
- `ScopeName` / `ScopeValue` — the named scope level (e.g., `"project"`) and its ID.

We were incorrectly passing `scopeName` (e.g., `"project"`) as `UserScope`, which made the ADO REST URL:
`/_apis/FeatureManagement/FeatureStates/project/project/{projectId}/...`
…which ADO rejects because `"project"` is not a valid userId segment.

The correct call requires `UserScope: "host"`, `ScopeName: "project"`, `ScopeValue: {projectId}`.

**Confirmed by:** WI spec line 107 evidence URL: `...FeatureStates/host/project/{projectId}/...` = `userScope/scopeName/scopeValue`.

**Fix:** Changed `UserScope` from `converter.String(scopeName)` to `converter.String("host")` in:
- `setFeatureFlag()` in `resource_feature_flag_framework.go`
- `readFeatureFlag()` in `resource_feature_flag_framework.go`
- `deleteFeatureFlag()` in `resource_feature_flag_framework.go`
- `checkFeatureFlagDestroyed()` in `resource_feature_flag_test.go`
- `captureFeatureFlagEvidence()` in `resource_feature_flag_test.go`
- Updated unit test assertions from `"project"` to `"host"` for UserScope.

**Verified:**
- `go build -tags all ./...` — clean
- `go test -tags all -count=1 ./azuredevops/internal/service/featuremanagement/...` — all 6 tests pass
- `make test` — all offline tests pass
- `TestAccFeatureFlag_basic` is discoverable in the acceptance test list

---

### Iteration 1

**Problem:** Gate failed with "The provider hashicorp/betterado does not support resource type `betterado_feature_flag`". The resource implementation existed in `azuredevops/internal/service/featuremanagement/resource_feature_flag_framework.go` but was never registered in the framework provider.

**Root cause:** `featuremanagement.NewFeatureFlagResource` was absent from `azuredevops/internal/provider/framework_provider.go`'s `Resources()` slice.

**Fix:** Added `featuremanagement` import and `featuremanagement.NewFeatureFlagResource` to `Resources()` in `framework_provider.go`. One 2-line change.

**Verified:**
- `go build -tags all ./...` — clean
- `go vet -tags all ./azuredevops/internal/provider/... ./azuredevops/internal/acceptancetests/...` — clean
- `go test -tags all ./azuredevops/internal/provider/... ./azuredevops/internal/service/featuremanagement/...` — all pass
- `go test -tags all -run TestAccFeatureFlag -list '.*' ...` — lists `TestAccFeatureFlag_basic` (discoverable)
- Committed: `feat(provider): register betterado_feature_flag in framework provider`
