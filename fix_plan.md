# Fix Plan

> Checklist for WI-6. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a pre/post_deployment_gates block that, beyond gates_options, declares one or more `gate {}` blocks each carrying a `task {}` (workflow task) WHEN expandDeploymentGates processes the resource data THEN ReleaseDefinitionGatesStep.Gates is a non-empty []ReleaseDefinitionGate, each gate carrying its workflow task(s) — the gate CHECKS, not just timing
- [x] AC2: GIVEN a ReleaseDefinitionGatesStep from ADO whose Gates array is populated WHEN flattenDeploymentGates processes it THEN the Terraform state round-trips the `gate {}` blocks + tasks (no drops)
- [x] AC3: GIVEN new unit tests TestReleaseDefinition_GatesTasks_ExpandFlatten WHEN go test -tags all -count=1 -run TestReleaseDefinition_GatesTasks ./azuredevops/internal/service/release/ runs THEN the tests pass and exit 0
