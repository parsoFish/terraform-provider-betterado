# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN a new TestAccReleaseDefinition_environmentConfig acceptance test that creates a release definition with environment_trigger, schedule, and properties blocks configured WHEN TF_ACC=1 is set and the test runs against live ADO THEN the test creates the release definition successfully, a PlanOnly step with ExpectNonEmptyPlan:false confirms idempotency, and destroy completes without error
- [x] AC2: GIVEN the acceptance test file azuredevops/internal/acceptancetests/resource_release_definition_test.go WHEN go test -tags all -run TestAccReleaseDefinition_environmentConfig ./azuredevops/internal/acceptancetests/ is executed with TF_ACC=1 THEN the test function is found and runs (it will skip if TF_ACC is absent, fail the gate guard if TF_ACC/AZDO_ORG_SERVICE_URL/AZDO_PERSONAL_ACCESS_TOKEN are missing per acceptance_gate config)
