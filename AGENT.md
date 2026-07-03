# Agent Memory — WI-1

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 2 (current)

**Gate failure:** `TestAccWorkitemtrackingprocessProcess_CreateDisabled` — "Provider produced inconsistent result after apply: .is_enabled: was cty.False, but now cty.True"

**Root cause identified:**
- After `CreateNewProcess` (which always creates with is_enabled=true), we called `EditProcess` to set is_enabled=false, then re-read with `GetProcessByItsId` to get ground-truth state.
- `GetProcessByItsId` suffers from **eventual consistency lag** — it returns the stale `isEnabled=true` immediately after the PATCH, causing Terraform's post-apply consistency check to see plan=false vs state=true.
- The live evidence confirms `GetProcessByItsId` DOES return `"isEnabled": true` explicitly for enabled processes (not nil/omitted) — so the API is consistent, just slow to reflect PATCH changes.

**Fix applied (commit 089fb845):**
- Changed both `Create` and `Update` to use the `EditProcess` PATCH response directly (mirrors old SDK resource behaviour which also used the EditProcess response).
- Added nil-fallback: if `edited.IsEnabled == nil`, fall back to plan value.
- Removed now-unused `getProcessByID` helper (was a lint error too).

### Iteration 1

- Attempted read-after-write fix: after EditProcess, re-read via GetProcessByItsId.
- This was insufficient because GetProcessByItsId has eventual consistency lag.

### Iteration 0

- Gap matrix created at docs/workitemtrackingprocess-gap-matrix.md.
- Framework migration of resource_process + data sources.
- Fixed nil IsEnabled default in flattenProcess (nil → false).

## What worked

- Using `EditProcess` PATCH response directly for is_enabled/is_default (same as old SDK code did).
- Adding nil-fallback from plan model when PATCH response omits is_enabled/is_default.
- Removing the post-create/post-update `GetProcessByItsId` re-read (it's unreliable immediately after PATCH).

## What didn't work

- Re-reading via `GetProcessByItsId` immediately after `EditProcess`: returns stale is_enabled value due to eventual consistency.

## Open questions

_(things that aren't blocking but would be useful to clarify; reflector picks these up)_

## Notes for reflection

- ADO `GetProcessByItsId` has eventual consistency after PATCH — don't re-read immediately; use the PATCH response directly.
- The old SDK resource trusted EditProcess response directly; the framework migration should do the same.
