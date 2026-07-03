# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (complete — all ACs done)

**Problem:** Gate reported `[no tests to run]` because the gate command is `-run TestFeatureFlagCRUD` but the existing test file only had `TestFeatureFlagSchemaHasRequiredFields` and `TestFeatureFlagMetadata` — no functions matching `TestFeatureFlagCRUD*`.

**Solution:**
1. Replaced the CRUD stubs in `resource_feature_flag_framework.go` with real implementations using internal helper functions (`setFeatureFlag`, `readFeatureFlag`, `deleteFeatureFlag`) that take `featuremanagementapi.Client` directly.
2. Added four `TestFeatureFlagCRUD*` tests in `resource_feature_flag_framework_test.go` that call the helpers directly with the `MockClient` from WI-2.
3. Tests match the gate's `-run TestFeatureFlagCRUD` prefix.

## What worked

- Extracting `setFeatureFlag`, `readFeatureFlag`, `deleteFeatureFlag` as package-level helpers that accept `featuremanagementapi.Client` directly — this makes them trivially testable with gomock without needing tfsdk.State/Plan objects.
- Using `hasNotFoundDiag(diags)` to detect the 404/undefined sentinel instead of propagating error diags directly from Read; Read CRUD calls `resp.State.RemoveResource(ctx)` when `hasNotFoundDiag` returns true.
- The mock (`NewMockClient`) is in the same package (featuremanagement), so tests in `package featuremanagement` can access both production helpers and the mock.
- `t.Run` subtests under `TestFeatureFlagCRUDRead` count as part of `TestFeatureFlagCRUDRead` for the prefix match.

## API mapping (confirmed)

- `scope_name` → `UserScope` AND `ScopeName` (both set to same value, per WI-3 spec and `resource_project_features.go` pattern)
- `scope_value` → `ScopeValue`
- `feature_id` → `FeatureId`
- `state` → `Feature.State` (cast to `ContributedFeatureEnabledValue`)
- Delete: `SetFeatureStateForScope` with `state = "undefined"`
- 404 detection: `utils.ResponseWasNotFound(err)` which checks for `azuredevops.WrappedError{StatusCode: 404}`
- `ContributedFeatureEnabledValueValues.Undefined` = `"undefined"` = the delete sentinel

## What didn't work

- The prior iteration (WI-2) scaffolded the package but left CRUD stubs as `notImplemented()` — that caused `[no tests to run]` because no `TestFeatureFlagCRUD*` tests existed.

## Open questions

_(none — all ACs satisfied)_

## Notes for reflection

- Gate uses prefix match `-run TestFeatureFlagCRUD`, so tests must be named `TestFeatureFlagCRUD<Op>` not `TestFeatureFlag<Op>`.
- Framework CRUD unit tests in this codebase don't construct tfsdk.State/Plan — they test extracted helper functions directly.
