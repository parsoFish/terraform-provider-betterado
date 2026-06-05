# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a release environment block in Terraform HCL includes `pre_deployment_gates` and `post_deployment_gates` nested blocks with `gatesOptions` (isEnabled, timeout, samplingInterval, stabilizationTime, minimumSuccessDuration) WHEN `expandReleaseDefinitionEnvironment` processes the resource data THEN the resulting `ReleaseDefinitionEnvironment` carries `PreDeploymentGates` and `PostDeploymentGates` with all `GatesOptions` fields correctly populated
- [x] AC2: GIVEN a `ReleaseDefinitionEnvironment` from ADO with `PreDeploymentGates` and `PostDeploymentGates` set WHEN `flattenReleaseDefinitionEnvironment` processes it THEN the resulting Terraform state contains the correct `pre_deployment_gates` and `post_deployment_gates` blocks with all `gates_options` sub-fields matching
- [x] AC3: GIVEN new unit tests `TestReleaseDefinition_Gates_ExpandFlatten` exist in the release package WHEN `go test -tags all -count=1 -run TestReleaseDefinition_Gates ./azuredevops/internal/service/release/` is executed THEN the tests pass and exit 0
