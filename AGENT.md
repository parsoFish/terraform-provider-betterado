# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 2 (current)

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
- Plan modifier null-guard (check `req.StateValue.IsNull()`) prevents converting Unknown→null for Computed-only fields on create.
- `staticCheckInt64(v)` helper (patterned after existing `staticCheckBool`, `staticCheckString`) works for Int64 defaults.

## What didn't work

- Original `checkUseStateForUnknownInt64Val()` on Optional+Computed `timeout` / `retry_interval` without a Default: on create, plan was null, API returned non-null → "inconsistent result".

## Open questions

- Why did the ADO API return `Version` and `Timeout` for InvokeRESTAPI checks but apparently not for BranchControl/BusinessHours/etc. checks? This might mean those other resources have a latent bug if the API behavior changes, but it's not blocking the current gate.

## Notes for reflection

- Pattern: For Optional+Computed fields where the API always returns a non-null default, use `Default` (not UseStateForUnknown) to make the plan concrete from the start.
- Pattern: For Computed-only fields (set only by provider), `UseStateForUnknown` plan modifier MUST guard against null state (create path). Standard `stringplanmodifier.UseStateForUnknown()` from the framework handles this correctly; custom implementations must replicate that guard.
