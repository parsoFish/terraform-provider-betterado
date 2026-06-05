# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 1 (complete)

1. Added `parallel_execution` TypeList/MaxItems:1 schema block inside `deployment_input` with fields: `type` (string, default "none"), `max_number_of_agents` (int), `multipliers` (TypeList string), `continue_on_error` (bool).

2. Discovered type mismatch: `AgentDeploymentInput.ParallelExecution *ExecutionInput` only has `ParallelExecutionType` — `maxNumberOfAgents` / `multipliers` / `continueOnError` are on `MultiConfigInput`/`MultiMachineInput` subtypes and cannot be assigned via the typed struct.

3. **Fix**: Changed `expandDeploymentInput` to return `map[string]interface{}` (raw ADO camelCase map) instead of `*releaseapi.AgentDeploymentInput`. Updated `expandDeployPhases` to build phases as `map[string]interface{}` so `deploymentInput` value is set from the raw map (not the typed struct field). This lets the full `parallelExecution` payload flow through to the API.

4. Added `expandParallelExecution(input []interface{}) map[string]interface{}` — builds ADO camelCase keys from TF snake_case.

5. Added `flattenParallelExecution(pe map[string]interface{}) []interface{}` — reads ADO camelCase from the JSON-decoded API response and produces TF snake_case.

6. Wired both into `expandDeploymentInput` / `flattenDeploymentInput`.

7. Added `TestReleaseDefinition_ParallelExecution_ExpandFlatten` with three sub-tests (AC1, AC2, AC3). All pass.

8. Full package test suite still passes (`go test -tags all ./azuredevops/internal/service/release/`).

## What worked

- Changing `expandDeploymentInput` to `map[string]interface{}` return type avoids the polymorphic subtype problem in the ADO Go SDK.
- Building `expandDeployPhases` phases as raw maps (not typed `AgentBasedDeployPhase` structs) keeps the deploy-phase JSON round-trip working via existing `flattenDeployPhases` (which already JSON-marshals/unmarshals each phase).
- AC3 test: when `parallel_execution` is an empty slice (`[]interface{}{}`), the `len(pe) > 0` guard ensures the key is absent from the result — verified in test.

## What didn't work

- Trying to set `di.ParallelExecution = expandParallelExecution(pe)` where `ParallelExecution *releaseapi.ExecutionInput` — type mismatch since `ExecutionInput` only has `ParallelExecutionType`, loses `maxNumberOfAgents`.

## Open questions

_(none)_

## Notes for reflection

- Pattern: when the ADO Go SDK uses polymorphic structs with limited base fields, use raw `map[string]interface{}` for the expand path so all JSON keys flow through. The flatten path already uses `map[string]interface{}` from JSON decode.
