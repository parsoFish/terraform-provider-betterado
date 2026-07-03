# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 0 (final)

WI-1 was already completed by a prior iteration (pre-existing commit `bc358c12`). Both required files exist:
- `docs/policy-gap-matrix.md` (370 lines) — committed, in diff vs main
- `docs/approvalsandchecks-gap-matrix.md` (240 lines) — committed, in diff vs main

WI-1 quality gate `TestProvider_HasChildResources` passes: `ok ... 0.005s`

The `last-gate-failure.md` present in `.forge/` is about WI-4/5's live acceptance tests (`TestAccCheck*`) — NOT WI-1's gate. WI-1 is documentation-only; its gate is `TestProvider_HasChildResources`, not live acc tests.

## What worked

- Creating comprehensive gap matrix docs that enumerate every ADO Policy API type and Checks/Approvals type with coverage columns.
- WI-1 quality gate (`TestProvider_HasChildResources`) is an offline test that passes without TF_ACC.

## What didn't work

_(nothing failed for WI-1 — it was completed successfully)_

## Open questions

_(none)_

## Notes for reflection

- WI-1 is a documentation-only WI (behaviour_preserving: true). Its quality gate is an offline structural test, not a live acceptance test.
- The `.forge/last-gate-failure.md` from WI-4/5 should not block WI-1 completion; they are separate work items with separate gates.
