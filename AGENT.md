# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (this iteration)

- Read WI-2.md spec, SDKv2 resource (resource_user_entitlement.go), framework_provider.go, framework_provider_test.go, task_group_framework.go (reference pattern), acceptance tests for task_group (reference pattern for getDirectClient + CaptureLiveEvidence).
- Wrote `resource_user_entitlement_framework.go` — full framework implementation following task_group_framework.go pattern. Key differences from task group:
  - Inline plan modifiers with `ue`-prefixed names (ueRequiresReplace, ueStateForUnknown) to avoid conflicts within same package
  - readIntoModel returns a Go error (not diag.Diagnostics) for simplicity since there's no nested data
  - flattenUserEntitlementFramework normalises licensing_source to lowercase (ADO returns mixed case, SDKv2 had CaseDifference suppress — framework needs explicit normalisation)
  - defaults applied in Create/Update before calling expand (accountLicenseType="express", licensingSource="account")
- Wrote unit test `resource_user_entitlement_framework_test.go` — TestNewUserEntitlementResource_Metadata.
- Registered `memberentitlementmanagement.NewUserEntitlementResource` in `framework_provider.go` Resources() and removed `betterado_user_entitlement` from `provider.go` ResourcesMap **in the same commit** (WI mandates this to avoid "Duplicate resource type" error at apply).
- Added `TestFrameworkProvider_HasUserEntitlementResource` to framework_provider_test.go.
- Removed `"betterado_user_entitlement"` from expectedResources in provider_test.go (SDKv2 count decremented by 1).
- Rewrote `resource_user_entitlement_test.go` acceptance test:
  - `Providers: testutils.GetProviders()` → `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`
  - `testutils.GetProvider().Meta().(*client.AggregatedClient)` → `getDirectClient()` (reuses the helper defined in resource_task_group_test.go in the same package)
  - Added idempotency step: `{PlanOnly: true, ExpectNonEmptyPlan: false}`
  - Added `captureUserEntitlementEvidence(tfNode)` check that writes `.forge/live-evidence/acceptance-resource-user-entitlement.json`
  - Added build tag `//go:build (all || resource_user_entitlement) && !exclude_resource_user_entitlement`

All offline tests pass:
- `go build -mod=vendor .` ✓
- `TestNewUserEntitlementResource_Metadata` ✓
- `TestFrameworkProvider_HasUserEntitlementResource` ✓
- `TestProvider_HasChildResources` ✓
- All `./azuredevops/internal/service/memberentitlementmanagement/...` ✓

Pre-existing failure (not introduced here): serviceendpoint test build (assignment mismatch).

## What worked

- Following the task_group_framework.go pattern exactly (inline plan modifiers, getDirectClient() for check functions).
- **stringplanmodifier package is NOT in vendor** — must use inline plan modifiers.
- Single atomic commit for de-registering from SDKv2 + registering in framework prevents "Duplicate resource type" error.

## What didn't work

- `stringplanmodifier.RequiresReplace()` — the `stringplanmodifier` package is not vendored; must define inline types.

## Open questions

- Does `licensing_source` returned by ADO API exactly match what we store (lowercase normalisation)? Will need live gate to confirm idempotency.
- Does `principal_name` from ADO API match case-sensitively? Live gate will reveal if we need CaseDifference suppression (can't do it in framework schema declaratively — would need PlanModifier).

## Notes for reflection

- The `stringplanmodifier` package gap in vendor is worth noting — other migrated resources will face the same constraint.
- Pattern: `getDirectClient()` is defined once in resource_task_group_test.go in the `acceptancetests` package and reused by all framework acceptance tests in the same package.
