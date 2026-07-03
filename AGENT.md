# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (WI-2, iter 0)

**Gate failure root cause:** `TestAccServiceEndpointGeneric_basic` was creating a new ADO project (via `betterado_project.project` resource + `HclProjectResource`), but the org is at the 1000-project cap — any project create fails immediately. Also was using `Providers: testutils.GetProviders()` (SDKv2-only) which would panic under the mux after migrating generic off SDKv2.

**What was done:**

1. Created `azuredevops/internal/service/serviceendpoint/resource_serviceendpoint_generic_framework.go` — full framework resource implementing `resource.Resource` + `resource.ResourceWithConfigure`. Key details:
   - TypeName: `req.ProviderTypeName + "_serviceendpoint_generic"` (avoids mis-naming bug)
   - `Configure()`: reads `req.ProviderData.(*client.AggregatedClient)` — never SDKv2 meta
   - Inline plan modifier + defaults helpers (no `stringplanmodifier`/`stringdefault` sub-packages — they don't exist in vendor)
   - Password field: Sensitive=true, preserved from state on Read (API returns empty/masked)
   - authorization: computed Map with `{"scheme": "UsernamePassword"}`

2. Registered `serviceendpoint.NewServiceEndpointGenericResource` in `framework_provider.go` Resources() slice.

3. Removed `"betterado_serviceendpoint_generic": serviceendpoint.ResourceServiceEndpointGeneric()` from `provider.go` ResourcesMap (to avoid duplicate-resource-type panic at apply).

4. Removed `"betterado_serviceendpoint_generic"` from `provider_test.go` expectedResources list.

5. Rewrote `TestAccServiceEndpointGeneric_basic` to:
   - Use `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()` (not `Providers`)
   - Use `SharedFixtureProjectName` (betterado-standing-demo) via `data "betterado_project" "test"` data source — NO new project creation
   - Add idempotency step: `PlanOnly: true, ExpectNonEmptyPlan: false`
   - Call `captureServiceEndpointGenericEvidence` in Check step → `testutils.CaptureLiveEvidence("acceptance-resource-generic", endpointURL, ep)`
   - `checkServiceEndpointGenericDestroyed` uses `getDirectClient()` (not `testutils.GetProvider().Meta()` which is nil under ProtoV6ProviderFactories)
   - Added build tag: `//go:build (all || resource_serviceendpoint_generic) && !exclude_resource_serviceendpoint_generic`

**Status after iter 1:** Code compiles, provider_test.go passes offline, acceptance tests compile. Awaiting live gate run by forge.

## What worked

- Pattern for framework resources: use inline type-prefixed helpers (e.g. `seGenericRequiresReplace()`, `seGenericDefaultString()`) because the `stringplanmodifier`/`stringdefault` sub-packages are NOT in vendor.
- `defaults` package IS in vendor at `vendor/github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults/`
- `SharedFixtureProjectName = "betterado-standing-demo"` (defined in `azuredevops/internal/acceptancetests/shared_fixtures.go`)
- `getDirectClient()` is defined in `resource_task_group_test.go` and accessible within the `acceptancetests` package
- `testutils.GetMuxedProviderFactories()` is defined in `testutils/mux_provider.go` (there's also `GetMuxProviderFactories()` in `testutils/commons.go` — both are equivalent)

## What didn't work

- `stringdefault.StaticString()` and `stringplanmodifier.UseStateForUnknown()` — these convenience sub-packages don't exist in vendor; must implement inline helpers within the package

## Open questions

- Will the idempotency check pass? The framework Read() doesn't return `password` (API masks it), so state preserved from plan. This should be fine since we preserve state in Read.
- generic_v2, generic_git, and generic_v2 data source are still SDKv2 — the gate only tests the `generic` resource so those can follow in later iterations.

## Notes for reflection

- The org-at-1000-project-cap failure is now a hard constraint: every acceptance test for a new resource MUST use `SharedFixtureProjectName` via data source lookup, never create a new project.
- The `getDirectClient()` pattern (direct env-var client build) is the correct approach for CheckDestroy in mux/ProtoV6 tests.
- The pre-existing `serviceendpoint` package unit test build failure (assignment mismatch in many `_test.go` files) is NOT caused by our changes — it was present before iteration 1.
