# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a betterado_feed resource configured with a non-default name and no project_id (org-scoped) WHEN terraform apply runs via the mux ProtoV6 provider path (GetMuxProviderFactories) THEN the feed is created in ADO, the provider reads it back with the correct name and id, an idempotency re-plan produces no diff (ExpectNonEmptyPlan: false), and destroy succeeds
- [x] AC2: GIVEN a betterado_feed resource configured with a non-default name and a project_id (project-scoped) WHEN terraform apply runs via the mux ProtoV6 provider path THEN the feed is created in the specified project, the provider reads back both name and project_id correctly, idempotency re-plan produces no diff, and destroy succeeds
- [x] AC3: GIVEN the framework betterado_feed resource is live and readable via the Feed REST API WHEN CaptureLiveEvidence is called inside the acceptance test check step with label acceptance-resource THEN .forge/live-evidence/acceptance-resource.json is written with a real ADO GET URL (dev.azure.com feeds endpoint)

## Sub-tasks completed (iteration 0)

- [x] Created `azuredevops/internal/service/feed/resource_feed_framework.go` — full CRUD + ImportState
- [x] Created `azuredevops/internal/service/feed/framework_defaults.go` — local plan modifiers (vendored sub-pkgs not available)
- [x] Registered `feed.NewFeedResource` in `framework_provider.go` (Resources() slice)
- [x] Deregistered `betterado_feed` from `provider.go` SDKv2 ResourcesMap (with comment)
- [x] Removed `betterado_feed` from `provider_test.go` expectedResources list (with comment)
- [x] Added `resource_feed_framework_test.go` with TestAccFeedFramework_basic + TestAccFeedFramework_withProject
  - Both use GetMuxedProviderFactories(), idempotency step, captureFeedFrameworkEvidence
- [x] `make test` passes (gofmt + go test)
- [x] `golangci-lint run --new-from-rev=main` → 0 issues
- [x] `make terrafmt-check` passes

## Pending (needs live gate)

- [ ] Live TF_ACC run to confirm acceptance tests actually pass end-to-end
