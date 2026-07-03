# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Iteration 0 summary

### What I did

Full implementation committed in one iteration (commit `092411a7`):

1. **`resource_feed_retention_policy_framework.go`** — new terraform-plugin-framework resource implementing:
   - `feedRetentionPolicyModel` with `id`, `feed_id`, `project_id`, `count_limit` (Int64), `days_to_keep_recently_downloaded_packages` (Int64)
   - `Schema`: `id` = computed with useStateForUnknown, `feed_id` + `project_id` with requiresReplace, two int64 required attrs
   - `Configure`: asserts `*client.AggregatedClient`
   - `Create`: calls `SetFeedRetentionPolicies`, sets `id = feedID`, then re-reads via `readPolicy()`
   - `Read`: calls `readPolicy()`, removes state on 404
   - `Update`: calls `SetFeedRetentionPolicies` with plan values, then re-reads
   - `Delete`: calls `DeleteFeedRetentionPolicies`
   - `ImportState`: handles `feedId` or `projectId/feedId` format
   - Helper `readPolicy()` returns `diag.Diagnostics` (not an interface) — avoids the interface type problem
   - Uses `nilIfEmpty(projectID)` (already defined in `resource_feed_permission_framework.go`)

2. **`framework_provider.go`** — added `feed.NewFeedRetentionPolicyResource` to `Resources()` slice

3. **`provider.go`** — removed `betterado_feed_retention_policy: feed.ResourceFeedRetentionPolicy()`, replaced with comment

4. **`provider_test.go`** — removed `betterado_feed_retention_policy` from SDKv2 resource list, replaced with comment

5. **Acceptance tests** in `resource_feed_retention_policy_test.go`:
   - Added `//go:build (all || resource_feed_retention_policy) && !exclude_feed_retention_policy` to fix offline build
   - Added `os` and `strings` imports
   - `TestAccFeedRetentionPolicyFramework_projectBasic` — AC1+AC3
   - `TestAccFeedRetentionPolicyFramework_update` — AC2
   - `checkFeedRetentionPolicyFrameworkDestroyed` — destroy check using `getFeedDirectClient`
   - `captureFeedRetentionPolicyFrameworkEvidence` — calls CaptureLiveEvidence with retention policies URL
   - HCL uses `SharedFixtureProjectName` (betterado-standing-demo) to avoid project-create

6. **Pre-existing bug fix**: `resource_feed_permission_test.go` was missing `//go:build` tag, causing offline build failure when `getFeedDirectClient` + `nilIfEmptyStr` (defined in the build-tagged `resource_feed_framework_test.go`) were not visible. Added `//go:build (all || resource_feed_permission) && !exclude_feed_permission`.

7. **Docs + example**: `make docs` run; `examples/resources/betterado_feed_retention_policy/resource.tf` created

8. **CHANGELOG**: entry added under `## [Unreleased]`

### Key patterns

- `getFeedDirectClient()` and `nilIfEmptyStr()` are in `resource_feed_framework_test.go` (build-tagged `all || resource_feed`). Any test file that calls them must ALSO be build-tagged.
- `nilIfEmpty(s string) *string` is defined in `resource_feed_permission_framework.go` in the `feed` package — reuse it, don't redeclare.
- `diag.Diagnostics` is a concrete slice type, not an interface — pass it directly, not via an interface.
- The retention policy ID is the feed ID itself (same as SDKv2 behavior: `d.SetId(feedId)`).
- HCL for framework tests uses `data "betterado_project"` lookup on `SharedFixtureProjectName` (betterado-standing-demo) to avoid project creation (org is at 1000-project cap).

### Quality gates

- `go build ./...` — clean
- `make test` (gofmt + `go test -v ./...`) — 0 failures
- `go test -tags all -c ./azuredevops/internal/acceptancetests/` — compiles
- `golangci-lint run --new-from-rev=main ./azuredevops/...` — 0 issues
- `make terrafmt-check` — clean

### Live gate status

The forge live gate (`go test -tags all -run TestAccFeedRetentionPolicyFramework ./azuredevops/internal/acceptancetests/` with TF_ACC=1) is the next step — awaiting orchestrator execution.

## Open questions

_(none)_

## Notes for reflection

- The WI-3 offline-build failure (permission test missing build tag) was a latent issue that would have caused CI problems; fixed proactively here since it's the same root cause as what would block retention policy tests.
