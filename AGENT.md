# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 3 (current)

**Problem**: `TestAccWorkitemtrackingprocessProcess_CreateAndUpdate` fails at step 3/4 with drift:
`is_enabled = true -> false` after updating a process to disabled.

**Root cause**: ADO API eventual consistency. After `EditProcess(is_enabled=false)`:
- The PATCH response can return `is_enabled=true` (stale) — seen in iteration 2
- A single immediate `GetProcessByItsId` can also return `is_enabled=true` (stale) — seen in iteration 1
- Terraform's post-apply consistency check calls Read, which gets stale `true`, causing drift

**Fix applied**: Added `waitForProcessIsEnabled` helper in `resource_process_framework.go` that uses
`retry.StateChangeConf` (from `github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry`) to poll
`GetProcessByItsId` with `ContinuousTargetOccurence=2` and `MinTimeout=2s` until the API returns
the expected `is_enabled` value, up to 2 minutes.

Both Create (post-EditProcess for disabled) and Update now call this helper.

### Iteration 2

Switched from read-after-write (single GetProcessByItsId) to using EditProcess PATCH response directly.
Fixed `TestAccWorkitemtrackingprocessProcess_CreateDisabled` but broke `CreateAndUpdate` step 3/4.

### Iteration 1

Tried read-after-write: GetProcessByItsId right after EditProcess. Fixed some issues but the single
immediate read was still stale for the disabled case.

### Iteration 0

Fixed `flattenProcess` to default nil IsEnabled to false. This handles the case where
`GetProcessByItsId` omits isEnabled for disabled processes (omitempty).

## What worked

- Using `retry.StateChangeConf` with `ContinuousTargetOccurence` pattern for eventual consistency
  (already used in `resource_group.go` and other resources)
- The `sdkretry` alias for `github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry` import compiles
  cleanly alongside the framework resource code

## What didn't work

- Trusting the EditProcess PATCH response directly (iteration 2): PATCH response returns stale is_enabled
- Single GetProcessByItsId read-after-write (iteration 1): single read still returns stale data
- Neither approach addresses the drift in the post-apply consistency check

## Key ADO API behaviour

- `CreateNewProcess` always creates with `is_enabled=true`, `is_default=false`
- `EditProcess` PATCH response can return `is_enabled=true` even when you set it false
- `GetProcessByItsId` omits `isEnabled` when the process is disabled (omitempty JSON) → nil in Go
- nil `IsEnabled` from `GetProcessByItsId` means DISABLED (map to false in `flattenProcess`)
- The API has eventual consistency: GET can lag PATCH by multiple seconds

## Open questions

None currently blocking.

## Notes for reflection

- The pattern `retry.StateChangeConf + ContinuousTargetOccurence=2` with a long Timeout is the
  correct approach for ADO eventual consistency in both SDKv2 and framework resources.
- The `processMinPollInterval` var allows tests to set it to 0 to speed up unit test execution.
