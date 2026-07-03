# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 0)

No forge-brain query (per ADR 010 — planner encoded context into WI spec).

## What I've tried

### Iteration 0 — full implementation (2025-07-03)

Built the complete framework migration from scratch. All offline gates pass.

**Files created/modified:**
- `azuredevops/internal/service/feed/resource_feed_framework.go` — `feedFrameworkResource` struct implementing `resource.Resource` + `resource.ResourceWithImportState`. CRUD mirrors the SDKv2 logic. Uses `feedapi.CreateFeed`, `GetFeed`, `UpdateFeed`, `DeleteFeed`, `PermanentDeleteFeed`, `RestoreDeletedFeed`. Schema uses `schema.ListNestedBlock` for `features`.
- `azuredevops/internal/service/feed/framework_defaults.go` — inline `useStateForUnknown()` and `requiresReplace()` plan modifier constructors.
- `azuredevops/internal/provider/framework_provider.go` — added `feed.NewFeedResource` to `Resources()` slice.
- `azuredevops/provider.go` — removed `"betterado_feed": feed.ResourceFeed()` from SDKv2 `ResourcesMap`; replaced with comment.
- `azuredevops/provider_test.go` — removed `"betterado_feed"` from `expectedResources` list; replaced with comment.
- `azuredevops/internal/acceptancetests/resource_feed_framework_test.go` — new file with build tag `//go:build (all || resource_feed) && !exclude_feed`. Contains `TestAccFeedFramework_basic` (org-scoped) and `TestAccFeedFramework_withProject` (project-scoped), `checkFeedFrameworkDestroyed`, `captureFeedFrameworkEvidence`.

## What worked

- Copy the `framework_defaults.go` pattern from `release/framework_defaults.go` for plan modifiers — the `stringplanmodifier` sub-package is NOT vendored in this project (only base `planmodifier` package exists).
- Use `GetMuxedProviderFactories()` (from `testutils/mux_provider.go`) for `ProtoV6ProviderFactories`.
- Use a separate `*_framework_test.go` file with the build tag to avoid changing existing `resource_feed_test.go`.
- Build tag `//go:build (all || resource_feed) && !exclude_feed` — with `-tags all`, the test is included.
- `getFeedDirectClient()` in the new test file (named differently from task group's `getDirectClient()` to avoid duplicate symbol in same package).
- For org-scoped feeds, ADO returns `feedDetail.Project == nil` — set `model.ProjectID = types.StringNull()` in Read (NOT `types.StringValue("")` — null ≠ empty-string in the framework).

### Iteration 1 — live gate fixes (2025-07-03)

Fixed two live-gate failures reported in `.forge/last-gate-failure.md`:

**Fix 1 — Idempotency diff (TestAccFeedFramework_basic, exit code 1):**
Iteration 0 wrote `model.ProjectID = types.StringValue("")` for org-scoped feeds in Read(). The HCL config omits `project_id` entirely → framework represents that as `types.StringNull()`. Empty string ≠ null → plan sees diff → `requiresReplace` fires → destroy+recreate.
Fix: `model.ProjectID = types.StringNull()` when `feedDetail.Project == nil`.

**Fix 2 — Project create fails (TestAccFeedFramework_withProject):**
Test provisioned `resource "betterado_project"`. Org is at 1000-project cap.
Fix: `data "betterado_project"` referencing `SharedFixtureProjectName` ("betterado-standing-demo"). Pattern from `resource_task_group_test.go` lines 68-99.

**Additional:** `nilIfEmpty(s string) *string` helper added to the resource, `nilIfEmptyStr` added to the test file — so all ADO API calls receive `nil` (not `""`) for org-scoped feeds.

## What didn't work

- Initial import of `github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier` — package doesn't exist in vendor; fixed by defining local helpers in `framework_defaults.go`.
- `types.StringValue("")` for null optionals in Read() — causes perpetual diff when `requiresReplace` is on the attribute. Must use `types.StringNull()`.
- Creating `betterado_project` resources in acceptance tests — org at 1000-project cap, all creates fail.

### Iteration 2 — soft-delete destroy fix (2025-07-03)

**Root cause:** ADO `DeleteFeed` is a **soft delete**. After calling it, `GetFeed` by GUID still returns the feed — but with `DeletedDate` set on the `Feed` struct. The iteration-1 `checkFeedFrameworkDestroyed` treated any non-error response from `GetFeed` as "feed still exists" → test reported dangling resource error.

**Fix 1 — `checkFeedFrameworkDestroyed`:** After `GetFeed` returns without error, check `feedDetail.DeletedDate != nil`. If set, the feed is in the recycle bin — `continue` (treat as destroyed). Only return an error if the feed is returned AND has no `DeletedDate`.

**Fix 2 — `Read()` in `resource_feed_framework.go`:** Same logic: if `feedDetail.DeletedDate != nil`, call `resp.State.RemoveResource(ctx)` and return. This prevents the "restore" path from being confused by a stale recycle-bin entry on next apply.

**Key ADO fact:** `Feed.DeletedDate` (from `models.go`) is non-nil when `DeleteFeed` has been called. `PermanentDeletedDate` is non-nil after `PermanentDeleteFeed`. There is no `IsDeleted` bool on `Feed` (that field is on `FeedCore`/`FeedSettings`).

## Open questions

_(none blocking)_

## Notes for reflection

- The `stringplanmodifier` sub-package being absent from vendor is a gotcha for future framework migrations in this project.
- Mux-path acceptance tests must use a direct client in CheckDestroy (not `testutils.GetProvider().Meta()`) since the SDKv2 meta singleton is not wired when using `ProtoV6ProviderFactories`.
- **null vs empty-string is critical**: optional attributes absent from HCL config arrive as `types.StringNull()`. Read() must write `types.StringNull()` back (not `""`), or any `requiresReplace` on that attribute will trigger perpetual destroy+recreate.
- **Project-create in acceptance tests is forbidden**: org at 1000-project cap. Always use `data "betterado_project"` with `SharedFixtureProjectName`.
