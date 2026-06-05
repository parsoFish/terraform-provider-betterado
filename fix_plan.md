# Fix Plan

> Checklist for WI-4. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a `deploy_phase` block includes a `deployment_input` block with a `parallel_execution` sub-block specifying `type: multiConfiguration` and `maxNumberOfAgents: 3` WHEN `expandDeploymentInput` processes the resource data THEN the resulting deployment-input map carries `parallelExecution: {parallelExecutionType: 'multiConfiguration', maxNumberOfAgents: 3}`
- [x] AC2: GIVEN a deploy-phase deployment-input from ADO carries `parallelExecution` with type `multiMachine` and `maxNumberOfAgents: 2` WHEN `flattenDeploymentInput` processes it THEN the Terraform state contains `deployment_input.0.parallel_execution.0.type = 'multiMachine'` and `deployment_input.0.parallel_execution.0.max_number_of_agents = 2`
- [x] AC3: GIVEN a deploy-phase deployment-input omits `parallel_execution` (or specifies `type: none`) WHEN `expandDeploymentInput` processes it THEN the resulting map carries `parallelExecution: {parallelExecutionType: 'none'}` or omits the field without panicking
- [x] AC4: GIVEN new unit tests `TestReleaseDefinition_ParallelExecution_ExpandFlatten` exist in the release package WHEN `go test -tags all -count=1 -run TestReleaseDefinition_ParallelExecution ./azuredevops/internal/service/release/` is executed THEN the tests pass and exit 0
