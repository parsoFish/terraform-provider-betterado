# Fix Plan

> Checklist for WI-1. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN docs/dashboard-gap-matrix.md does not exist WHEN WI-1 implementation runs THEN docs/dashboard-gap-matrix.md is created and lists every field of the ADO SDK Dashboard struct with mapped/missing/writable status; writable gaps either resolved in the schema or explicitly deferred with rationale
- [x] AC2: GIVEN betterado_dashboard is registered in azuredevops/provider.go SDKv2 ResourcesMap WHEN WI-1 implementation runs THEN betterado_dashboard is removed from provider.go ResourcesMap and its SDKv2 import dropped; provider_test.go TestProvider_HasChildResources updated to remove betterado_dashboard from the expected list with a comment that it is now a framework resource
- [x] AC3: GIVEN azuredevops/internal/provider/framework_provider.go Resources() does not include a DashboardResource WHEN WI-1 implementation runs THEN a new azuredevops/internal/service/dashboard/resource_dashboard_framework.go implements resource.Resource (terraform-plugin-framework); Configure() wires *client.AggregatedClient from ProviderData; framework_provider.go Resources() includes NewDashboardResource
- [ ] AC4: GIVEN the framework Dashboard resource exists WHEN TF_ACC acceptance tests run (TestAccDashboard_project_basic, TestAccDashboard_project_update, TestAccDashboard_team_basic, TestAccDashboard_team_update) THEN all TestAccDashboard_* tests in azuredevops/internal/acceptancetests/resource_dashboard_test.go pass; ProviderFactories is changed to GetMuxedProviderFactories() if not already; idempotency re-plan (ExpectNonEmptyPlan: false) holds
  - [x] ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories() used
  - [x] getDirectDashboardClient() helper for checkDashboardDestroyed/checkDashboardExist
  - [x] Idempotency steps added (ExpectNonEmptyPlan: false, PlanOnly: true)
  - [x] Switched all HCL helpers from `resource "betterado_project" "test"` to `data "betterado_project" "test"` using SharedFixtureProjectName — avoids ADO 1000-project cap
  - [ ] Live gate: awaiting forge re-run to confirm pass
- [x] AC5: GIVEN the framework Dashboard resource runs the live acceptance test WHEN the acceptance test performs a live read-back before destroy THEN testutils.CaptureLiveEvidence("acceptance-resource", <dashboard GET URL>, <apiResponse>) is called so .forge/live-evidence/acceptance-resource.json is written
- [x] AC6: GIVEN docs/resources/dashboard.md exists (from upstream docs/) WHEN WI-1 implementation runs THEN examples/resources/betterado_dashboard/resource.tf is created with non-default values for all writable fields; make docs is run and docs/resources/dashboard.md is updated; git checkout -- docs/guides/ restores hand-written guides
- [x] AC7: GIVEN changed Go files WHEN CI-equivalent gate runs (make test + golangci-lint --new-from-rev=main ./azuredevops/... + make terrafmt-check) THEN gate is green with zero new lint findings on changed files

## Gate failure from iteration 1 — FIXED in iteration 2
The live gate `TestAccDashboard_project_basic` was failing with:
```
Error: creating project: Failed to add a project as this organization already has 1000 projects.
```
Fix (iteration 2): switched all dashboard HCL helpers to use `data "betterado_project" "test"` pointing
at `SharedFixtureProjectName` ("betterado-standing-demo") — the persistent shared project. Pattern
is consistent with resource_task_group_test.go. No new project creation; teams are still created
per-test (don't count toward the project cap).
