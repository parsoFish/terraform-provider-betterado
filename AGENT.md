# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## What I've tried

### Iteration 0 (initial build)

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

### Iteration 1 (fix display_name unknown after apply)

**Gate failure:** `provider still indicated an unknown value for betterado_feed_permission.test.display_name. All values must be known after apply`

**Root cause:** `display_name` is `Optional + Computed` but had:
1. No `PlanModifiers` — so on every plan where user doesn't set it, it remains "unknown"
2. No guaranteed resolution to a known value in Create — only set if `findPermission` returned perm with non-nil DisplayName

**Fix applied (commit e92be2db):**
1. Added `useStateForUnknown()` PlanModifier to `display_name` schema — on re-plan, carries forward state value instead of "unknown"
2. In Create: after polling, if `display_name` is unknown/null, call `findPermission` to get ADO's display name; fall back to `""` (empty string, always known) if that fails
3. In Update: guard `plan.DisplayName.IsUnknown()` and fall back to state or empty string before writing state

**Key insight:** In terraform-plugin-framework, `Optional + Computed` attributes that are not set in config are "unknown" in the plan. After Create/Update, ALL attributes MUST be known in state, or Terraform raises the "unknown value after apply" error. `useStateForUnknown()` handles re-plans but the INITIAL apply must resolve the value.

## What worked

- Using the custom `useStateForUnknown()` and `requiresReplace()` from `framework_defaults.go` instead of `stringplanmodifier.*` (not vendored)
- Using `SharedFixtureProjectName` data source instead of creating a new project (org is at cap)
- Using `getFeedDirectClient()` (already defined in `resource_feed_framework_test.go`) for CheckDestroy
- Using `nilIfEmptyStr()` (defined in `resource_feed_framework_test.go`) for nil-safe project ID
- `findPermission` called after `pollPermissions` succeeds is reliable (poll already confirmed perm exists)

## What didn't work

- `github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier` — NOT in vendor; use `framework_defaults.go` helpers instead
- Initial `display_name` schema without PlanModifiers → "unknown value after apply" error

## Framework rules learned

- In terraform-plugin-framework, `Optional + Computed` attributes not set in config → "unknown" in plan
- After Create/Update, MUST write known value for all Computed attributes to state
- `useStateForUnknown()` only helps for re-plans (carry forward state), NOT first apply
- For first apply: must explicitly resolve Computed-only/Optional-Computed values via API re-read or fallback

## Notes for reflection

- The `betterado_feed_permission` deregistration from SDKv2 means old `TestAccFeedPermission_*` tests using `GetProviderFactories()` (pure SDKv2) will fail at live acceptance time since SDKv2 no longer has the resource. This is expected — they need to be migrated to the mux path too, but that's out of scope for WI-3.
