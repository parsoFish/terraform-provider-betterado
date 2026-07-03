# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 5 (final gate fix)

**Gate failure**: `TestAccFeedFramework_withProject` failed in post-destroy check with:
```
feed (ID: bcc6c6f6-ce36-420b-b26f-8ad8a7adb9ad) should not exist after destroy
```

**Root cause**: `checkFeedFrameworkDestroyed` used `GetFeed(GUID)` and checked `feedDetail.DeletedDate != nil` after `DeleteFeed`. ADO's `GetFeed` by GUID does **not** reliably populate `DeletedDate` after a soft-delete — it returns the feed object but with `DeletedDate` null.

**Fix**: Replaced the primary check with `GetFeedChange(name)` which returns `ChangeType="delete"` after a soft-delete. `GetFeed` by GUID + `DeletedDate` check kept as fallback. Both `feedapi.GetFeedChangeArgs` and `feedapi.ChangeTypeValues.Delete` are in the already-imported `feedapi` package.

**Key insights**:
- `GetFeed` by GUID after `DeleteFeed` → returns feed with `DeletedDate == nil` (unreliable)
- `GetFeedChange` by name after `DeleteFeed` → returns `ChangeType="delete"` (reliable)
- The `Delete()` resource function calls `DeleteFeed` by **name** (not GUID)
- Destroy check in test has feed **name** available via `res.Primary.Attributes["name"]`

## What worked

- Using `GetFeedChange(name)` to detect soft-delete in the destroy check (reliable)
- `isFeedRestorable()` in the resource already used this same pattern
- `feedapi.ChangeTypeValues.Delete` is the correct enum value (`"delete"`)

## What didn't work

- `GetFeed(GUID).DeletedDate != nil` — not reliable for detecting ADO soft-deletes in destroy check

## Open questions

_(things that aren't blocking but would be useful to clarify; reflector picks these up)_

## Notes for reflection

- ADO `GetFeed` by GUID does NOT reliably populate `DeletedDate` after soft-delete. `GetFeedChange` by name is the correct signal for destroy verification.
- Pattern: for ADO feed destroy checks, use `GetFeedChange(name)` (ChangeType="delete") not `GetFeed(GUID)` (DeletedDate) — the latter is unreliable.
