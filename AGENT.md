# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 5 (current)

**Problems**: Gate failure (iter 4) showed two classes of failures:

1. `TestAccCheckRestAPI_update` (step 2): `.version: was cty.NumberIntVal(1), but now cty.NumberIntVal(2)` — the plan had version=1 (preserved by UseStateForUnknown), but after update the ADO API returned version=2 (it always increments on every write).

2. `TestAccCheckRequiredTemplate_update` and `TestAccCheckBranchControl_complete`: "Service connection with name serviceendpoint already exists" — all 5 check test files used hardcoded `"serviceendpoint"` as the service endpoint name, causing name collisions when tests run in parallel.

**Fix 1: version plan modifier** (commit `c9d4bb01`):
- Added `checkVersionPlanModifierFn()` / `checkVersionPlanModifier` to `framework_helpers.go`
- Uses `tftypes.Value` comparison of the full plan vs state (excluding the version attr itself)
- On no-op re-plan (all other attrs unchanged): preserves state version → idempotency maintained
- On actual update (any other attr changed): keeps version as Unknown → ADO API's incremented value accepted without conflict
- `tftypes.Value` uses `IsKnown()` (not `IsUnknown()`) and `.As(&map[string]tftypes.Value{})` for object traversal
- Applied to all 6 check framework resources (approval, branch_control, business_hours, exclusive_lock, required_template, rest_api)

**Fix 2: generated service endpoint names** (same commit):
- All 5 test files (approval, branch_control, business_hours, exclusive_lock, required_template) updated to call `testutils.GenerateResourceName()` for service endpoint names
- Added `serviceEndpointName` param to all HCL helper functions; threaded through each `betterado_serviceendpoint_generic` resource
- For multi-step tests (update): same `serviceEndpointName` used in both steps (endpoint stays the same; only the check resource changes)

**Why branches NOT in gate failure (iter 4)**: The branch_control_complete and required_template_update tests fail at step 1 with serviceendpoint collision — they never reach step 2 where version inconsistency would manifest. Once collision is fixed, version inconsistency would have become visible for these too.

### Iteration 4 (prior)

**Problem**: Gate failure (iter 3) showed TestAccCheckRestAPI_complete and TestAccCheckRestAPI_update still failing with:
`Error: Provider returned invalid result object after apply` — unknown values for: body, headers, success_criteria, url_suffix, variable_group_name.

**Root cause**: These Optional+Computed fields used `checkUseStateForUnknown()` plan modifier. On create (no prior state), this leaves the plan value as `Unknown`. After apply:
- In the basic test (fields not in config): API doesn't return them → flattenRestAPIFW didn't set them → model kept Unknown value → state had Unknown → error.
- In the complete test (fields in config): plan had concrete values from config, but if API echoes them inconsistently, they could remain Unknown (though likely the fix for basic also resolves complete since the API probably returns them for complete test).

**Fix applied** (commit `ae77dfd4`):
- Changed `body`, `headers`, `success_criteria`, `url_suffix`, `variable_group_name` from `checkUseStateForUnknown()` plan modifier to `Default: staticCheckString("")`.
- Updated `flattenRestAPIFW` to always set these fields to `""` (empty string) when absent from API response (instead of leaving model value unchanged).
- When settings map is nil, all 5 optional fields are set to `""`.
- `variable_group_name` has explicit else branch: set to `""` when `linkedVariableGroup` not in settings.
- Before reading inputs, pre-initialize headers/body/successCriteria/urlSuffix to `""`.

**Why Default is better than UseStateForUnknown for these fields**:
- These fields are Optional — when user doesn't configure them, API returns empty (absent from inputs map).
- The API round-trip is: user omits → API doesn't store → API doesn't return → state = "".
- Default="" ensures plan is always concrete (""), state is always known. No Unknown-after-apply.
- SDKv2 handled this by retaining plan values in optional fields; Framework requires explicit always-known values.

### Iteration 3

**Problem**: Gate failure (iter 2) showed `TestAccCheckEnvironment` panicking with:
`interface conversion: interface {} is nil, not *client.AggregatedClient`

**Root cause**: `pipelinechecks.go:getCheckFromState()` called `GetProvider().Meta().(*client.AggregatedClient)`. When tests use `GetMuxedProviderFactories()` (proto v6 mux), the SDKv2 singleton `provider` variable in `commons.go` never has its `Meta()` set by the Terraform test lifecycle (only `ProviderFactories`/`Providers` wire `Meta`). Result: `Meta()` returns `nil`, type assertion panics.

**Additional bug fixed**: `CheckPipelineCheckDestroyed` was calling `getSvcEndpointFromState` (wrong — that's for service endpoints, not checks). Fixed to call `getCheckFromState`.

**Fix applied** (commit `852c4283`):
- Replaced `GetProvider().Meta()` with `getADOClientsFromEnv()` helper that builds client from env vars directly.
- `CheckPipelineCheckDestroyed` now uses `getCheckFromState` (semantically correct: check API).
- Uses `context.Background()` instead of `clients.Ctx`.

### Iteration 2

**Problem**: Gate failure (iter 1) showed `TestAccCheckRestAPI_complete` and `TestAccCheckRestAPI_update` failing with "Provider produced inconsistent result after apply":
- `.timeout: was null, but now cty.NumberIntVal(1440)` (basic test config, timeout not set)
- `.version: was null, but now cty.NumberIntVal(1)` (both tests)
- `.id: was null, but now cty.StringVal("17")` (both tests)
- `.retry_interval: was null, but now cty.NumberIntVal(0)` (basic test, retry_interval not set)

**Root causes identified**:

1. **timeout + retry_interval**: Optional+Computed fields with no Default. When not in config, plan has `null`. After apply, API returns 1440 / 0 respectively. Terraform core sees null→1440 as "inconsistent".

2. **version + id**: Computed-only fields with `checkUseStateForUnknownInt64Val()` / `checkUseStateForUnknown()` plan modifiers. On create (no prior state), plan should be `Unknown`. But the plan modifier was converting `Unknown → null` (since state is null on create, `resp.PlanValue = req.StateValue` set plan to null). After apply, API returns version=1, id="17" → `null → 1` is "inconsistent" (but `Unknown → 1` is OK).

**Fix applied** (commit `bdfd0e7f`):
1. Added `Default: staticCheckInt64(1440)` for `timeout` and `Default: staticCheckInt64(0)` for `retry_interval` in `resource_check_rest_api_framework.go`. Added `staticCheckInt64` helper to `framework_helpers.go`.
2. Added null/unknown guard to `checkUseStateForUnknownString` and `checkUseStateForUnknownInt64` plan modifiers: when `req.StateValue.IsNull() || req.StateValue.IsUnknown()`, return early to keep plan value as Unknown (don't convert to null).

**Why only rest_api failed** (not branch_control etc.):
- The ADO API for `InvokeRESTAPI` task checks returns `Timeout` and `retryInterval` in its response, making them non-null after apply.
- Other check types (BranchControl, BusinessHours, etc.) may not return `Timeout` from ADO API (returns nil), so `flattenFW` doesn't change the model value, plan stays consistent.
- For `version` and `id`: the plan modifier null-guard fix should also help other resources, but they might have been passing because ADO API doesn't return `Version` for those check types.

### Iteration 1

**Context**: All 6 framework files created in one autocommit (7707fe77). Gate ran all TestAccCheck* tests. Only TestAccCheckRestAPI_* failed. All other check tests passed.

**Files created**:
- `resource_check_approval_framework.go`
- `resource_check_branch_control_framework.go`
- `resource_check_business_hours_framework.go`
- `resource_check_exclusive_lock_framework.go`
- `resource_check_required_template_framework.go`
- `resource_check_rest_api_framework.go`
- `framework_helpers.go`

## What worked

- `Default` on Optional+Computed fields prevents "inconsistent result after apply" when the API always returns a concrete value for those fields.
- `Default: staticCheckString("")` for Optional+Computed string fields that the API may not return: plan is always concrete, state is always known. Better than UseStateForUnknown for fields that don't round-trip through the API.
- Always set Optional+Computed string fields to a concrete value (not null, not unknown) in flatten — either the API value or ""/0 when absent.
- Plan modifier null-guard (check `req.StateValue.IsNull()`) prevents converting Unknown→null for Computed-only fields on create.
- `staticCheckInt64(v)` helper (patterned after existing `staticCheckBool`, `staticCheckString`) works for Int64 defaults.
- Build ADO client from env vars directly in testutils verify functions when tests use `GetMuxedProviderFactories()` — avoids nil `GetProvider().Meta()` panic (same pattern used in `shared_fixtures.go`).
- `checkVersionPlanModifierFn()`: uses `tftypes.Value` comparison of full plan vs state (excluding version itself) to detect no-op re-plan vs actual update. On no-op: preserves state version for idempotency. On update: keeps Unknown so ADO's post-write incremented version is accepted.
- `tftypes.Value` API: `IsKnown()` (not `IsUnknown()`), `IsNull()`, `.As(&map[string]tftypes.Value{})` for object traversal, `.Equal()` for comparison.
- Generate unique service endpoint names in acceptance tests to prevent parallel-test name collisions in shared ADO environments.

## What didn't work

- Original `checkUseStateForUnknownInt64Val()` on Optional+Computed `timeout` / `retry_interval` without a Default: on create, plan was null, API returned non-null → "inconsistent result".
- `checkUseStateForUnknown()` on Optional+Computed string fields (body/headers/url_suffix/success_criteria/variable_group_name) that the ADO API may not return: on create, plan was Unknown; after apply, if API doesn't return them, state had Unknown → "Provider returned invalid result object after apply".

## Open questions

- Why did the ADO API return `Version` and `Timeout` for InvokeRESTAPI checks but apparently not for BranchControl/BusinessHours/etc. checks? This might mean those other resources have a latent bug if the API behavior changes, but it's not blocking the current gate.

## Notes for reflection

- Pattern: For Optional+Computed fields where the API always returns a non-null default, use `Default` (not UseStateForUnknown) to make the plan concrete from the start.
- Pattern: For Computed-only fields (set only by provider), `UseStateForUnknown` plan modifier MUST guard against null state (create path). Standard `stringplanmodifier.UseStateForUnknown()` from the framework handles this correctly; custom implementations must replicate that guard.
