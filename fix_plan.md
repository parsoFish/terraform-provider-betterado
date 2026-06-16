# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN TF_ACC=1, AZDO_ORG_SERVICE_URL and AZDO_PERSONAL_ACCESS_TOKEN are set WHEN TestAccTaskGroup_basic runs against a live ADO org THEN a task group is created with a UUID-prefixed name, non-default description and category, at least one input parameter, and at least one task step
- [x] AC2: GIVEN the task group is created WHEN the acceptance test read-back step runs THEN resource.TestCheckResourceAttr assertions on name, description, category, input.0.name, task.0.display_name all pass (not TestCheckResourceAttrSet)
- [x] AC3: GIVEN the task group state is applied WHEN a PlanOnly step with ExpectNonEmptyPlan: false runs THEN no perpetual diff is detected (idempotency confirmed)
- [x] AC4: GIVEN the test completes WHEN CheckDestroy runs THEN GetTaskGroups returns 404 confirming the task group is gone from ADO

## Status (iteration 0 complete)

All ACs implemented in `azuredevops/internal/acceptancetests/resource_task_group_test.go` (commit a01f0915).

- `go test -tags all -list TestAccTaskGroup_basic` → found ✓
- `gofmt -l` → no issues ✓
- `golangci-lint run` → no issues ✓
- `go vet` → clean ✓

Awaiting orchestrator to run quality gate with `TF_ACC=1` against live ADO.
