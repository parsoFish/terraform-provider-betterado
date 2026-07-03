# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a betterado_feed_permission resource configured with feed_id, project_id, role=reader, and identity_descriptor pointing to a real ADO group WHEN terraform apply runs via the mux ProtoV6 provider path (GetMuxProviderFactories) THEN the feed permission is created in ADO, the provider reads it back asserting role and identity_descriptor, an idempotency re-plan produces no diff (ExpectNonEmptyPlan: false), and destroy succeeds (permission removed)
- [x] AC2: GIVEN the framework betterado_feed_permission resource is live and readable via the Feed Permissions REST API WHEN CaptureLiveEvidence is called inside the acceptance test check step with label acceptance-resource THEN .forge/live-evidence/acceptance-resource.json is written with a real ADO GET URL (dev.azure.com feed permissions endpoint)

## Status

Both ACs are COMPLETE:

- `resource_feed_permission_framework.go` — full framework resource (Create/Read/Update/Delete/ImportState + identity resolution + polling)
- `framework_provider.go` — registers `feed.NewFeedPermissionResource` in framework Resources()
- `provider.go` — `betterado_feed_permission` deregistered from SDKv2 ResourcesMap (comment added)
- `resource_feed_permission_test.go` — `TestAccFeedPermissionFramework_basic` with:
  - `GetMuxedProviderFactories()` for ProtoV6 mux path
  - role=reader + identity_descriptor assertions
  - Idempotency step (PlanOnly:true, ExpectNonEmptyPlan:false)
  - `checkFeedPermissionFrameworkDestroyed` CheckDestroy
  - `captureFeedPermissionFrameworkEvidence` → writes `.forge/live-evidence/acceptance-resource.json`
- `.forge/live-evidence/acceptance-resource.json` — written by live test run (real dev.azure.com URL)

All offline gates pass: `make test` green, golangci-lint 0 issues, terrafmt clean.
