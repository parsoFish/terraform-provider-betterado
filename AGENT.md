# Agent Memory — WI-9

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 8 (current)

**Gate failure**: All 5 `TestAccVariableGroup*` tests failing with "Unexpectedly found a variable group that should be deleted" at 143-167 s. The CheckDestroy timeout was 120 s, but ADO's distributed backend keeps returning the VG as alive for 2+ minutes after deletion is confirmed.

**Root cause**: ADO's eventual consistency propagates VG deletions slowly. The provider's `Delete` was using `ContinuousTargetOccurence: 2` with `MinTimeout: 3s` — only requiring 2 consecutive 404s (~6s) before claiming delete done. CheckDestroy then had only 120s. Tests ran for ~143s (30s apply + 90s provider wait + 120s CheckDestroy = 240s budget, but the 404-flicker returned 200 after provider completed).

**Fix applied**:
1. `resource_variable_group_framework.go`: Increased `ContinuousTargetOccurence` from 2 to 4, `MinTimeout` from 3s to 5s, `Timeout` from 60s to 90s. Now requires ~20s of stable 404s before confirming deletion.
2. `resource_variable_group_test.go` `checkVariableGroupDestroyedMux`: Increased `timeout` from 120s to 300s (5 minutes). ADO needs this long to fully propagate VG deletions.

**Total test time estimate**: ~30s steps + 90s provider wait (max) + 300s CheckDestroy = 420s = 7 min. Parallel tests → 7 min wall clock, within 10-min gate budget.

### Iterations 0-7

- Migrated betterado_variable_group resource and data source to terraform-plugin-framework
- Deregistered SDKv2 files from provider.go and framework_provider.go
- Fixed HCL fixtures from block to attribute syntax
- Fixed inconsistent-result errors in read-back
- Fixed post-destroy race conditions and CheckDestroy polling
- Made permissions tests parallel
- Various timeout adjustments

## What worked

- `resource.ParallelTest` for all VG tests ensures the 300s CheckDestroy window doesn't multiply across tests
- `ContinuousTargetOccurence: 4` requires stable deletion signal before provider returns
- Fixture-project approach (reuse standing project) avoids 1000-project limit

## What didn't work

- `ContinuousTargetOccurence: 2` with 120s CheckDestroy — ADO's eventual consistency takes 2+ minutes to propagate deletions fully
- Shorter timeouts (60s, 45s, 120s CheckDestroy) all too short for ADO backend

## Open questions

- None currently blocking

## Notes for reflection

- ADO variable group deletion is far more eventually consistent than other resources (2+ minutes propagation)
- The 300s CheckDestroy timeout is needed; shorter values reliably fail in live env
