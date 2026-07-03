# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## What I've tried

### Iteration 0 (this iteration)

**State before this iteration:** Gate reported "no tests to run" for TestAccFeedPermissionFramework — the test function didn't exist yet.

**What was built:**
1. `azuredevops/internal/service/feed/resource_feed_permission_framework.go` — full framework resource implementing CRUD + ImportState for betterado_feed_permission. Key decisions:
   - Uses `useStateForUnknown()` and `requiresReplace()` from `framework_defaults.go` (NOT `stringplanmodifier.*` — those are not vendored)
   - Reuses `syncing`, `succeed`, `failed` constants from `resource_feed_permission.go`
   - Uses `nilIfEmpty()` from `resource_feed_framework.go`
   - Polls ADO with `retry.StateChangeConf` (same as SDKv2 version)
   - Conflict check on Create to mirror SDKv2 behavior

2. `azuredevops/internal/provider/framework_provider.go` — added `feed.NewFeedPermissionResource` to Resources()

3. `azuredevops/provider.go` — removed `"betterado_feed_permission": feed.ResourceFeedPermission()` from SDKv2 ResourcesMap (mux requires no duplicate registrations across SDKv2 + framework providers)

4. `azuredevops/provider_test.go` — removed `"betterado_feed_permission"` from expected SDKv2 resources list

5. `azuredevops/internal/acceptancetests/resource_feed_permission_test.go` — rewrote file to add:
   - `TestAccFeedPermissionFramework_basic` using `GetMuxedProviderFactories()` + `ProtoV6ProviderFactories`
   - Uses `SharedFixtureProjectName` (betterado-standing-demo) data source — avoids creating new project (org is at 1000-project cap)
   - `captureFeedPermissionFrameworkEvidence()` calls `CaptureLiveEvidence("acceptance-resource", url, permissions)` where url = `<org>/<projectId>/_apis/packaging/feeds/<feedId>/permissions?api-version=7.1`
   - Added `os` and `strings` imports

## What worked

- Using the custom `useStateForUnknown()` and `requiresReplace()` from `framework_defaults.go` instead of `stringplanmodifier.*` (not vendored)
- Using `SharedFixtureProjectName` data source instead of creating a new project (org is at cap)
- Using `getFeedDirectClient()` (already defined in `resource_feed_framework_test.go`) for CheckDestroy
- Using `nilIfEmptyStr()` (defined in `resource_feed_framework_test.go`) for nil-safe project ID

## What didn't work

- `github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier` — NOT in vendor; use `framework_defaults.go` helpers instead

## Open questions

- Will the ADO group's `descriptor` attribute be available immediately when `betterado_group` is created in the mux path? (Should be, since group is still SDKv2)
- idempotency: `display_name` is Computed+Optional. On re-plan, ADO returns the display name — if it differs from what terraform has in state (or null vs set), this might cause a diff. May need to handle carefully.

## Notes for reflection

- The `betterado_feed_permission` deregistration from SDKv2 means old `TestAccFeedPermission_*` tests using `GetProviderFactories()` (pure SDKv2) will fail at live acceptance time since SDKv2 no longer has the resource. This is expected — they need to be migrated to the mux path too, but that's out of scope for WI-3.
- The `display_name` field on `feedPermissionModel` is `Optional + Computed`. During read, ADO returns the group's display name. If config doesn't set it, it will be null in config but set in state after apply → potential idempotency diff. To avoid this, the Read sets it from ADO, and Create re-reads after polling to populate it. If this causes issues, consider making it Computed-only or using `UseStateForUnknown`.
