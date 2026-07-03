# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN no betterado_test_plan resource or data source exists WHEN WI-2 is complete THEN azuredevops/internal/service/testplan/resource_test_plan_framework.go defines a framework resource.Resource for betterado_test_plan with CRUD methods; azuredevops/internal/service/testplan/data_test_plan_framework.go defines a datasource.DataSource; both are registered in azuredevops/internal/provider/framework_provider.go under Resources() and DataSources() respectively; provider.go (SDKv2) has zero new test_plan registrations
- [x] AC2: GIVEN the resource schema is defined WHEN unit tests run (no TF_ACC) THEN go test -tags all -run TestUnitTestPlan ./azuredevops/internal/service/testplan/ passes; tests cover expand/flatten roundtrip for project_id, name, area_path, iteration_path, start_date, end_date
- [x] AC3: GIVEN the framework provider resource list is updated WHEN go test -run TestProvider_HasChildResources ./azuredevops/ runs (no TF_ACC) THEN the test still passes (betterado_test_plan must NOT appear in the SDKv2 ResourcesMap; its absence keeps the count valid)
