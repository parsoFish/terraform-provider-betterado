# Fix Plan

> Checklist for WI-9. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN serviceendpoint_npm and serviceendpoint_sonarcloud data sources exist in the SDKv2 DataSourcesMap WHEN both migrated to terraform-plugin-framework THEN framework data source files exist for both; both deregistered from provider.go DataSourcesMap; registered in framework_provider.go DataSources(); no Duplicate resource type panic; CI-equivalent gate passes
- [x] AC2: GIVEN acceptance tests for migrated npm and sonarcloud data sources WHEN test helper provider factory is used THEN tests use ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(); provider_test.go expectedDataSources counts updated
- [x] AC3: GIVEN framework data source files compile correctly WHEN go build -mod=vendor . is run THEN provider binary builds without errors; TypeName for each uses req.ProviderTypeName + suffix pattern
