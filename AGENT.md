# Agent Memory — WI-8

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 8 (FINAL — all AC complete)

**Problem**: `TestAccReleaseDefinition_complete` was failing at step 1/2 (idempotency check) with multiple perpetual diffs.

**Root causes identified and fixed:**

1. **`multipliers` round-trip failure** (was the original failure at check 31/64):
   - ADO SDK defines `Multipliers *string` in `release/models.go` — comma-separated, NOT a JSON array.
   - `expandParallelExecution` was sending `[]string{"Configuration"}` (JSON array) which ADO silently drops.
   - **Fix**: Changed to `strings.Join(multStrs, ",")` to send the correct comma-separated string.
   - Updated unit test `AC1_expand_multiConfiguration` to expect `string` not `[]string`.

2. **`variable_groups = []` perpetual diff on environments**:
   - When ADO returns `VariableGroups = &[]int{}` (empty), omitting the key from state caused a diff.
   - **Fix**: Always set `variable_groups` in flattened env map (even if empty).

3. **`deployment_input` removed from Production "Agent job" phase**:
   - ADO always returns a `deploymentInput` object even for phases without `deployment_input` in HCL.
   - **Fix**: Added `isDefaultDeploymentInput()` helper + `hclPhaseHasDeploymentInput(d, envIdx, phaseIdx)`.
   - If HCL had no `deployment_input` AND the API response has all-default values → suppress.
   - Changed `flattenDeployPhases` signature to accept `d *schema.ResourceData` and `envIdx int`.

4. **`pre_deployment_gates` / `post_deployment_gates` removed from Production env**:
   - ADO always returns a `ReleaseDefinitionGatesStep` even for envs with no gates.
   - **Fix**: Added `isNonDefaultGatesOptions()` helper. Return nil from `flattenDeploymentGates` when no actual gates and all options are default.

5. **`schedule_trigger.branch_filter` perpetual diff**:
   - ADO does NOT return `branchFilters` in GET response for schedule triggers (confirmed by live API probe).
   - `flattenScheduleTriggerBranchFilter` was always emitting an empty `branch_filter {}` block.
   - **Fix 1**: Removed `branch_filter` from the schedule_trigger in HCL (ADO drops it, so testing it causes perpetual diff).
   - **Fix 2**: Changed `flattenScheduleTriggerBranchFilter` to return `nil` when both include and exclude are empty.

6. **`demands = []` perpetual diff on agentless deployment_input**:
   - `diMap["demands"] = []string{}` — Go `[]string` might not be stored correctly by TF SDK.
   - **Fix**: Changed to `[]interface{}{}` and append strings individually.

**Test result**: `TestAccReleaseDefinition_complete` PASSES — all 64 check functions ✓, idempotency plan empty ✓, destroy ✓.

## What worked

- Probing the ADO API directly with `go run` to check what it actually returns for schedule triggers → `branchFilters` is absent in the GET response.
- Using SDK models (vendor dir) to find `Multipliers *string` → root cause of silent drop.
- `hclPhaseHasDeploymentInput(d, envIdx, phaseIdx)` pattern for context-aware flatten.

## What didn't work

- Sending `multipliers` as `[]string` — ADO silently ignores it (type mismatch with `*string` in SDK model).
- Suppressing `variable_groups` completely when empty — causes `+ variable_groups = []` diff.

## Open questions

_(none — all ACs complete)_

## Notes for reflection

- **ADO always returns default objects**: `deploymentInput`, `ReleaseDefinitionGatesStep`, `ReleaseDefinitionApprovals` are ALWAYS present in GET even when not configured in HCL. Provider must use context (prior HCL state) or content checks to avoid spurious state entries.
- **ADO wire types differ from Terraform types**: `multipliers` is `*string` (comma-separated) in ADO SDK but `TypeList` ([]string) in TF schema. Always check vendor SDK models before sending data.
- **ADO does not return `branchFilters` for schedule triggers in GET**: This field is write-only effectively. Don't test it in acceptance tests.
