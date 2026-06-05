# Agent Memory — WI-5

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (COMPLETE — all ACs done in one pass)

**Approach taken:**
1. Changed `queue_id` schema from `Required: true, ValidateFunc: IntAtLeast(1)` to `Optional: true, Default: 0` — agentless phases have no queue.
2. Modified `expandDeploymentInput` to accept a variadic `phaseType ...string` parameter. When `phaseType[0] == "runOnServer"` OR `queueID == 0`, the `queueId` key is omitted from the output map.
3. Modified `expandDeployPhases` to pass `phaseMap["phase_type"].(string)` as the second argument to `expandDeploymentInput`.
4. `flattenDeploymentInput` already handled absent `queueId` safely (guarded by `if queueID, ok := di["queueId"].(float64); ok`), defaulting to 0.
5. Added `TestReleaseDefinition_AgentlessPhase_ExpandFlatten` with three sub-tests:
   - `AC1_expand_runOnServer_no_queueId` — verifies queueId absent, timeouts present
   - `AC2_flatten_runOnServer_no_queueId_no_panic` — verifies flatten doesn't panic, queue_id=0
   - `AC3_roundtrip_agent_and_agentless_phases` — full roundtrip through expandDeployPhases + JSON marshal/unmarshal + flattenDeployPhases

**Test results:**
- `go test -tags all -count=1 -run TestReleaseDefinition_AgentlessPhase ./internal/service/release/` → PASS
- `go test -tags all -count=1 -run TestReleaseDefinition ./internal/service/release/` → PASS (full CI gate)

## What worked

- Variadic `phaseType ...string` parameter on `expandDeploymentInput` — backward-compatible signature change; existing callers without the arg get default empty-string behavior (queueId still included when non-zero).
- The condition `pt != "runOnServer" && queueID != 0` covers both explicit agentless (phase_type=runOnServer) and implicit (queue_id=0 with any phase type).

## What didn't work

_(none — completed in one iteration)_

## Open questions

_(none)_

## Notes for reflection

- The schema change (Required→Optional for `queue_id`) is additive and backward-compatible; existing configs with queue_id set continue to work.
- The `flattenDeploymentInput` already had safe handling for absent `queueId` from prior WI refactoring — no change needed there.
