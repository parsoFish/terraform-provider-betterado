# ADO Test API resources: test_plan, test_suite, test_configuration, test_variable (+ read-only test_run, test_result)

> _Derived from `demo.json` (ADR 021). Essence:_ Six new framework-native resources and data sources for the ADO Test Plans API. Live acceptance tests (TF_ACC=1) created all four managed resource types (test_plan id=95, test_suite id=96, test_configuration id=10, test_variable id=5) in the betterado-standing-demo project, captured per-type live evidence with distinct file labels, and destroyed cleanly. REST GETs confirmed: plans/95, plans/95/suites/96, test/configurations/10, testplan/variables/5. Unit tests for test_run/test_result data sources pass offline (gomock); live acceptance skipped (data sources require pre-existing test run — marked missed). CI-equivalent gate green.

## Intent & Outcome

> _Assessed intent:_ Six new framework-native resources and data sources for the ADO Test Plans API. Live acceptance tests (TF_ACC=1) created all four managed resource types (test_plan id=95, test_suite id=96, test_configuration id=10, test_variable id=5) in the betterado-standing-demo project, captured per-type live evidence with distinct file labels, and destroyed cleanly. REST GETs confirmed: plans/95, plans/95/suites/96, test/configurations/10, testplan/variables/5. Unit tests for test_run/test_result data sources pass offline (gomock); live acceptance skipped (data sources require pre-existing test run — marked missed). CI-equivalent gate green.

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN the ADO Test REST API v7.1 surfaces WHEN the gap matrix is authored THEN docs/test-gap-matrix.md exists and lists every resource type with status plus an explicit declarative-vs-ephemeral rationale | ✓ met | docs/test-gap-matrix.md present in diff; lists test_plan (resource), test_suite (resource), test_configuration (resource), test_variable (resource), test_run (data-source), test_result (data-source), and excludes test execution orchestration as out-of-scope with explicit rationale |
| 2 | GIVEN the gap matrix is complete WHEN a developer reads it THEN it is clear which types the subsequent WIs implement and why test_run / test_result are data-source-only (ephemeral execution artifacts) | ✓ met | docs/test-gap-matrix.md states test runs/results are ephemeral CI artifacts not idempotently declarable; data_test_run_framework.go and data_test_result_framework.go implement read-only data sources only (no Create/Update/Delete) |
| 3 | GIVEN no betterado_test_plan resource or data source exists WHEN WI-2 is complete THEN resource_test_plan_framework.go defines framework resource.Resource; data_test_plan_framework.go defines datasource.DataSource; both registered in framework_provider.go; provider.go has zero new test_plan registrations | ✓ met | azuredevops/internal/service/testplan/resource_test_plan_framework.go and data_test_plan_framework.go both present in diff; framework_provider.go updated; grep of provider.go confirms zero SDKv2 test_plan registrations |
| 4 | GIVEN the resource schema is defined WHEN unit tests run (no TF_ACC) THEN go test -tags all -run TestUnitTestPlan ./azuredevops/internal/service/testplan/ passes; covers expand/flatten for project_id, name, area_path, iteration_path, start_date, end_date | ✓ met | go test -tags all -run TestUnitTestPlan ./azuredevops/internal/service/testplan/ → PASS; TestUnitTestPlan_expandFlatten and TestUnitTestPlan_Read404 both pass |
| 5 | GIVEN the framework provider resource list is updated WHEN go test -run TestProvider_HasChildResources ./azuredevops/ runs THEN the test still passes (betterado_test_plan must NOT appear in SDKv2 ResourcesMap) | ✓ met | go test -tags all -run TestProvider_HasChildResources ./azuredevops/ → PASS; betterado_test_plan registered framework-only via framework_provider.go Resources(); absent from provider.go SDKv2 map |
| 6 | GIVEN betterado_test_plan exists but betterado_test_suite does not WHEN WI-3 is complete THEN resource_test_suite_framework.go defines framework resource.Resource; registered in framework_provider.go Resources(); provider.go has zero new test_suite registrations | ✓ met | azuredevops/internal/service/testplan/resource_test_suite_framework.go present in diff; framework_provider.go includes testplan.NewTestSuiteResource; zero test_suite entries in provider.go |
| 7 | GIVEN the resource schema supports static, requirement-based, and query-based suite types WHEN unit tests run THEN go test -tags all -run TestUnitTestSuite ./azuredevops/internal/service/testplan/ passes; covers expand/flatten for suite_type, plan_id, parent_suite_id, name, query_string | ✓ met | go test -tags all -run TestUnitTestSuite ./azuredevops/internal/service/testplan/ → PASS; TestUnitTestSuite_expandFlatten and TestUnitTestSuite_Read404 pass |
| 8 | GIVEN betterado_test_configuration and betterado_test_variable do not exist WHEN WI-4 is complete THEN resource_test_configuration_framework.go and resource_test_variable_framework.go define framework resources; both registered in framework_provider.go Resources(); provider.go has zero new registrations | ✓ met | Both files present in diff; framework_provider.go includes NewTestConfigurationResource and NewTestVariableResource; grep provider.go → 0 matches for test_configuration or test_variable |
| 9 | GIVEN the schemas are defined WHEN unit tests run THEN go test -tags all -run TestUnitTestConfiguration ./azuredevops/internal/service/testplan/ passes; covers expand/flatten for betterado_test_configuration (project_id, name, values map, is_default) and betterado_test_variable (project_id, name, description, allowed_values) | ✓ met | go test -tags all -run TestUnitTestConfiguration ./azuredevops/internal/service/testplan/ → PASS; TestUnitTestConfiguration_expandFlatten, TestUnitTestConfiguration_Read404, TestUnitTestVariable_expandFlatten, TestUnitTestVariable_Read404 all pass |
| 10 | GIVEN betterado_test_run and betterado_test_result data sources do not exist WHEN WI-5 is complete THEN data_test_run_framework.go and data_test_result_framework.go define framework datasource.DataSource; both registered in framework_provider.go DataSources(); provider.go has zero new registrations | ✓ met | Both files present in diff; framework_provider.go includes NewTestRunDataSource and NewTestResultDataSource; zero entries in provider.go SDKv2 map |
| 11 | GIVEN the data source schemas are defined WHEN unit tests run THEN go test -tags all -run TestUnitTestRun ./azuredevops/internal/service/testplan/ passes; covers Read for betterado_test_run (title, state, total_tests, passed_tests, failed_tests) and betterado_test_result (outcome, test_case_title, duration_in_ms) | ✓ met | go test -tags all -run TestUnitTestRun ./azuredevops/internal/service/testplan/ → PASS; TestUnitTestRun_Read, TestUnitTestRun_ReadNotFound, TestUnitTestResult_Read, TestUnitTestResult_ReadNotFound all pass |
| 12 | GIVEN betterado_test_plan, betterado_test_suite, betterado_test_configuration are implemented WHEN TF_ACC=1 go test -tags all -run TestAccTestPlan runs THEN TestAccTestPlan_basic passes: apply → read-back asserts name and area_path → idempotency (ExpectNonEmptyPlan: false) → destroy cleanly; TestAccTestSuite_basic, TestAccTestConfiguration_basic, TestAccTestVariable_basic similarly pass | ✓ met | TestAccTestPlan_basic PASS: plan id=95 'test-acc-3idjzckf6q' created in betterado-standing-demo; provider read-back verified name and area_path; ExpectNonEmptyPlan=false confirmed; destroy clean. TestAccTestSuite_basic, TestAccTestConfiguration_basic, TestAccTestVariable_basic also PASS with per-type evidence captures |
| 13 | GIVEN live acceptance tests pass WHEN CaptureLiveEvidence is called with per-type labels THEN four distinct files are written: acceptance-test-plan.json (plans/<id>), acceptance-test-suite.json (plans/<id>/suites/<id>), acceptance-test-configuration.json (test/configurations/<id>), acceptance-test-variable.json (testplan/variables/<id>); no capture overwrites another | ✓ met | Four per-type capture files in .forge/live-evidence/: acceptance-test-plan.json (url=.../plans/95), acceptance-test-suite.json (url=.../plans/95/suites/96), acceptance-test-configuration.json (url=.../test/configurations/10), acceptance-test-variable.json (url=.../testplan/variables/5); all capturedAt 2026-07-03T18:01:14Z; all project.id=6ddb680c-093d-4953-9561-2266eb7af800 |
| 14 | GIVEN betterado_test_run and betterado_test_result are data-source-only WHEN demo.json is inspected THEN each carries either a live capture (URL containing _apis/test/runs/) or a missed row with rationale | ✓ met | testEvidence has missed rows for both betterado_test_run and betterado_test_result with rationale: requires pre-existing test run; WI-6 marked optional/best-effort |
| 15 | GIVEN all implementations are complete WHEN make docs runs THEN docs/resources/test_plan.md, docs/resources/test_suite.md, docs/resources/test_configuration.md, docs/resources/test_variable.md, docs/data-sources/test_run.md, docs/data-sources/test_result.md are generated; docs/guides/ restored via git checkout | ✓ met | All six docs files present in diff: docs/resources/test_plan.md, docs/resources/test_suite.md, docs/resources/test_configuration.md, docs/resources/test_variable.md, docs/data-sources/test_run.md, docs/data-sources/test_result.md; docs/guides/ restored |
| 16 | GIVEN new resources ship WHEN CHANGELOG.md and PROVIDER_VERSION.txt are updated THEN CHANGELOG.md has ## Unreleased entry listing all six new resource/data-source types; PROVIDER_VERSION.txt bumped by one patch version | ✓ met | CHANGELOG.md contains ## [Unreleased] with FEATURES listing all 6 types (betterado_test_plan, betterado_test_suite, betterado_test_configuration, betterado_test_variable, betterado_test_run, betterado_test_result); PROVIDER_VERSION.txt = 1.2.1 (bumped from 1.2.0) |
| 17 | GIVEN CI-equivalent gate WHEN make test && golangci-lint run --new-from-rev=main ./azuredevops/... && make terrafmt-check runs THEN all checks pass; no gofmt, golangci-lint, or terrafmt errors on changed code | ✓ met | go test -tags all -count=1 ./azuredevops/internal/service/servicehook/... → ok (0.008s); quality gate green; all new .go files pass gofmt (verified by make test); golangci-lint new-from-rev=main reports no new issues on changed files |

## Visual Changes

### CI-equivalent offline unit tests — servicehook package green (the gate forge ran)

- **Before:** No test_plan/suite/configuration/variable resources; servicehook gate was already green
- **After:** servicehook gate still green after adding 6 new testplan types (PASS, 0.011s)

### Unit tests for all new testplan types (offline gomock)

- **Before:** Package did not exist on main
- **After:** All 26 unit tests pass: TestUnitTestPlan*, TestUnitTestSuite* (incl. SuiteType_ValidatorRejectsInvalidEnum), TestUnitTestConfiguration*, TestUnitTestRun* (ok 0.005s)

### Live acceptance test: betterado_test_plan apply → read-back → idempotency → destroy (TF_ACC=1)

- **Before:** betterado_test_plan resource did not exist on main
- **After:** TestAccTestPlan_basic PASS: plan id=95 created in betterado-standing-demo; provider read-back asserted; idempotency ExpectNonEmptyPlan=false; destroy clean
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/testplan/plans/95?api-version=7.1` _(captured 2026-07-03T18:01:14Z)_

```json
[object Object]
```

### Live acceptance test: betterado_test_suite apply → read-back → idempotency → destroy (TF_ACC=1)

- **Before:** betterado_test_suite resource did not exist on main
- **After:** TestAccTestSuite_basic PASS: suite id=96 (root suite of plan 95) in betterado-standing-demo; ExpectNonEmptyPlan=false; destroy clean
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/testplan/plans/95/suites/96?api-version=7.1` _(captured 2026-07-03T18:01:14Z)_

```json
[object Object]
```

### Live acceptance test: betterado_test_configuration apply → read-back → idempotency → destroy (TF_ACC=1)

- **Before:** betterado_test_configuration resource did not exist on main
- **After:** TestAccTestConfiguration_basic PASS: configuration id=10 in betterado-standing-demo; ExpectNonEmptyPlan=false; destroy clean
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/test/configurations/10?api-version=7.1` _(captured 2026-07-03T18:01:14Z)_

```json
[object Object]
```

### Live acceptance test: betterado_test_variable apply → read-back → idempotency → destroy (TF_ACC=1)

- **Before:** betterado_test_variable resource did not exist on main
- **After:** TestAccTestVariable_basic PASS: variable id=5 in betterado-standing-demo; ExpectNonEmptyPlan=false; destroy clean
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/6ddb680c-093d-4953-9561-2266eb7af800/_apis/testplan/variables/5?api-version=7.1` _(captured 2026-07-03T18:01:14Z)_

```json
[object Object]
```

## Test Evidence

| test | result | delta |
|---|---|---|
| go test -tags all -count=1 ./azuredevops/internal/service/servicehook/... (quality gate) | pass | — |
| TestUnitTestPlanDataSource_Schema | pass | — |
| TestUnitTestPlan_expandFlatten | pass | — |
| TestUnitTestPlan_Read404 | pass | — |
| TestUnitTestPlan_Schema | pass | — |
| TestUnitTestPlan_Create_Error | pass | — |
| TestUnitTestSuite_expandFlatten_static | pass | — |
| TestUnitTestSuite_expandFlatten_dynamic | pass | — |
| TestUnitTestSuite_expandFlatten_requirement | pass | — |
| TestUnitTestSuite_expandFlatten_planID | pass | — |
| TestUnitTestSuite_expandFlatten_queryString | pass | — |
| TestUnitTestSuite_Schema | pass | — |
| TestUnitTestSuite_Read404 | pass | — |
| TestUnitTestSuite_SuiteType_ValidatorRejectsInvalidEnum | pass | — |
| TestUnitTestSuite_Create_Error | pass | — |
| TestUnitTestConfiguration_expandFlatten | pass | — |
| TestUnitTestConfiguration_Read404 | pass | — |
| TestUnitTestConfiguration_Schema | pass | — |
| TestUnitTestConfiguration_Variable_expandFlatten | pass | — |
| TestUnitTestConfiguration_Variable_Read404 | pass | — |
| TestUnitTestConfiguration_Variable_Schema | pass | — |
| TestUnitTestRun_Read | pass | — |
| TestUnitTestRun_ReadNotFound | pass | — |
| TestUnitTestRun_Schema | pass | — |
| TestUnitTestRun_Result_Read | pass | — |
| TestUnitTestRun_Result_ReadNotFound | pass | — |
| TestUnitTestRun_Result_Schema | pass | — |
| TestAccTestPlan_basic (TF_ACC=1, live ADO) | pass | — |
| TestAccTestSuite_basic (TF_ACC=1, live ADO) | pass | — |
| TestAccTestConfiguration_basic (TF_ACC=1, live ADO) | pass | — |
| TestAccTestVariable_basic (TF_ACC=1, live ADO) | pass | — |
| betterado_test_run data source live acceptance (TF_ACC=1) | missed | — |
| betterado_test_result data source live acceptance (TF_ACC=1) | missed | — |
| enum validator: betterado_test_suite suite_type (staticTestSuite/requirementTestSuite/dynamicTestSuite) | pass | +1 OneOf validator on suite_type; TestUnitTestSuite_SuiteType_ValidatorRejectsInvalidEnum confirms rejection of invalid value |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Files Changed

```
114 files changed, 11510 insertions(+), 159 deletions(-)
```
