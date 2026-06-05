# Fix Plan

> Checklist for WI-5. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a `deploy_phase` block specifies `phase_type: runOnServer` and a `deployment_input` block with an `agentless_input` sub-block (timeoutInMinutes, cancelTimeoutInMinutes) WHEN `expandDeployPhase` processes the resource data THEN the resulting deploy-phase map carries `phaseType: 'runOnServer'` and a `deploymentInput` map containing the timeout fields (without `queueId`)
- [x] AC2: GIVEN a deploy-phase from ADO carries `phaseType: runOnServer` with `deploymentInput.timeoutInMinutes: 120` WHEN `flattenDeployPhase` processes it THEN the Terraform state sets `phase_type = 'runOnServer'` and `deployment_input.0.timeout_in_minutes = 120` (or `agentless_input` sub-block if separated)
- [x] AC3: GIVEN new unit tests `TestReleaseDefinition_AgentlessPhase_ExpandFlatten` exist in the release package WHEN `go test -tags all -count=1 -run TestReleaseDefinition_AgentlessPhase ./azuredevops/internal/service/release/` is executed THEN the tests pass and exit 0
- [x] AC4: GIVEN all prior WIs' unit tests plus the new agentless tests are present WHEN `go test -tags all -count=1 -run TestReleaseDefinition ./azuredevops/internal/service/release/` is executed (the full release CI gate) THEN all tests pass and the command exits 0
