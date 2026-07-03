# Fix Plan

> Checklist for WI-4. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a betterado_feed_retention_policy resource configured with feed_id, project_id, count_limit=25, and days_to_keep_recently_downloaded_packages=45 (non-default values) WHEN terraform apply runs via the mux ProtoV6 provider path (GetMuxProviderFactories) THEN the retention policy is created in ADO, the provider reads it back asserting count_limit=25 and days=45, an idempotency re-plan produces no diff (ExpectNonEmptyPlan: false), and destroy succeeds
  - Implemented: `TestAccFeedRetentionPolicyFramework_projectBasic` with idempotency step
- [x] AC2: GIVEN a betterado_feed_retention_policy configured with count_limit=25 then updated to count_limit=30 WHEN a second apply step runs THEN the provider updates the policy in-place; the read-back shows count_limit=30; idempotency holds
  - Implemented: `TestAccFeedRetentionPolicyFramework_update` with 3 steps
- [x] AC3: GIVEN the framework betterado_feed_retention_policy resource is live and readable via the Feed Retention API WHEN CaptureLiveEvidence is called inside the acceptance test check step with label acceptance-resource THEN .forge/live-evidence/acceptance-resource.json is written with a real ADO GET URL (retention policies endpoint)
  - Implemented: `captureFeedRetentionPolicyFrameworkEvidence` called from AC1 test check step

## Implementation summary (iteration 0)

1. Created `azuredevops/internal/service/feed/resource_feed_retention_policy_framework.go` — full CRUD + ImportState
2. Registered `feed.NewFeedRetentionPolicyResource` in `framework_provider.go`
3. Deregistered `betterado_feed_retention_policy` from `provider.go` + `provider_test.go`
4. Added `TestAccFeedRetentionPolicyFramework_projectBasic` and `TestAccFeedRetentionPolicyFramework_update` to retention policy test file
5. Added `//go:build (all || resource_feed_retention_policy)` to retention policy test file
6. Fixed pre-existing offline build failure: added `//go:build (all || resource_feed_permission)` to permission test file
7. Added `examples/resources/betterado_feed_retention_policy/resource.tf`; ran `make docs`
8. Updated CHANGELOG.md under `## [Unreleased]`

## Awaiting live gate

The quality gate (`go test -tags all -run TestAccFeedRetentionPolicyFramework ./azuredevops/internal/acceptancetests/`) runs with TF_ACC=1 in the forge serve environment and will validate the full create→read→idempotency→update→destroy lifecycle against real ADO.
