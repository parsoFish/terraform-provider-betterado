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

- The `UserScope` parameter to `GetFeatureStateForScope` doubles as the scope name (pass "project" for project-scoped features), matching the pattern in `resource_feature_flag_framework.go`
- Evidence URL format for FeatureManagement is the "host" path: `/_apis/FeatureManagement/FeatureStates/host/{scopeName}/{scopeValue}/{featureId}`
