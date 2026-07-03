# Agent Memory — UWI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (2026-07-04) — COMPLETE

**AC2 — Rewrote TestDataSource_404NotFound:**
- Old test only called `mockClient.GetSubscription(nil, ...)` directly, never the datasource `Read()` method — `RemoveResource` was never exercised.
- New test constructs `tfsdk.Config` and `tfsdk.State` using the datasource schema (obtained via `ds.Schema()`), builds `tftypes.Object` raw values, then calls `ds.Read(ctx, readReq, &readResp)` directly.
- After 404: asserts `readResp.State.Raw.IsNull()` — proves `RemoveResource` was called.
- Also added `TestFlattenNotificationSubscriptionData_NilSubscription` and `TestFlattenNotificationSubscriptionData_PartialFields` (already existed) helpers.

**AC1 — Fixed demo.json and DEMO.md:**
- Subscription ID 886543 → 886548 (with capturedAt 2026-07-03T08:02:23Z) to match `.forge/live-evidence/acceptance-resource.json`.
- Added test names `TestDataSource_404NotFound`, `TestFlattenNotificationSubscriptionData*` to testEvidence.
- Data-source AC evidence rewritten to cite unit tests by name and the `stringvalidator.OneOf` channel_type fix (UWI-2 AC1).

**Gate result:** All 3 review gates pass (gate 1: test names in demo.json; gate 2: subscription ID match; gate 3: `.Read(` in test + suite green).

## What worked

- Calling `ds.Schema()` on the unexported datasource struct to get the framework schema directly in unit tests — no provider server needed.
- `tftypes.NewValue(tftypes.String, nil)` for null string values in the config object.
- `rawVal.Copy()` for pre-populating `readResp.State.Raw` before calling `Read()`.
- `tftypes.Object{AttributeTypes: map[string]tftypes.Type{...}}` with all schema fields explicitly typed.

## What didn't work

- Old `TestDataSource_404NotFound` pattern: calling mock directly and only checking `utils.ResponseWasNotFound` without driving `Read()` — does not exercise `RemoveResource`.
- `demo.json` liveEvidence with old subscription ID 886543 — gate 2 rejects it.

## Open questions

_(things that aren't blocking but would be useful to clarify; reflector picks these up)_

## Notes for reflection

_(observations the reflector should capture into the brain; the agent doesn't write them itself, but flags here)_
