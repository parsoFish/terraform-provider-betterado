# Fix Plan

> Checklist for WI-9. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN serviceendpoint_npm and serviceendpoint_sonarcloud data sources exist in the SDKv2 DataSourcesMap WHEN both migrated to terraform-plugin-framework THEN framework data source files exist for both; both deregistered from provider.go DataSourcesMap; registered in framework_provider.go DataSources(); no Duplicate resource type panic; CI-equivalent gate passes
- [x] AC2: GIVEN acceptance tests for migrated npm and sonarcloud data sources WHEN test helper provider factory is used THEN tests use ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(); provider_test.go expectedDataSources counts updated
- [x] AC3: GIVEN framework data source files compile correctly WHEN go build -mod=vendor . is run THEN provider binary builds without errors; TypeName for each uses req.ProviderTypeName + suffix pattern

## Completed in iteration 0

All three ACs completed:
- `data_serviceendpoint_npm_framework.go` created with `TypeName = req.ProviderTypeName + "_serviceendpoint_npm"` and `url` field
- `data_serviceendpoint_sonarcloud_framework.go` created with `TypeName = req.ProviderTypeName + "_serviceendpoint_sonarcloud"`
- Both registered in `framework_provider.go` DataSources()
- Both deregistered from `provider.go` DataSourcesMap
- Both removed from `provider_test.go` expectedDataSources
- Acceptance test files updated to use `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`
- `go build -mod=vendor .` passes
- `TestProvider_HasChildDataSources` gate passes
- `make test` all green
- `golangci-lint --new-from-rev=main` 0 issues
