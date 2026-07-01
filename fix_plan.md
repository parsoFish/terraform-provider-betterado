# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] Iteration 0: Migrated 5 release data sources from SDKv2 to framework DataSources()
  - Created datasource_release_definition_framework.go
  - Created datasource_release_definition_history_framework.go
  - Created datasource_release_definition_revision_framework.go
  - Created datasource_release_definitions_framework.go
  - Created datasource_release_folder_framework.go
  - Registered all 5 in framework_provider.go DataSources()
  - Removed 5 data sources from SDKv2 provider.go DataSourcesMap (and unused import)
  - Updated TestProvider_HasChildDataSources to remove migrated data sources
  - Updated TestAccDataReleaseFolder_Basic to use ProtoV6ProviderFactories + GetMuxProviderFactories()
  - All make test, golangci-lint, terrafmt-check pass

- [ ] AC1: GIVEN all five release data sources (`betterado_release_definition`, `betterado_release_definition_history`, `betterado_release_definition_revision`, `betterado_release_definitions`, `betterado_release_folder`) migrated to `datasource.DataSource` and registered in `DataSources()` in `framework_provider.go` WHEN each is read live against the standing project (TF_ACC=1) THEN each returns the same fields as its SDKv2 predecessor; `TestAccDataReleaseDefinition_ById`, `TestAccDataReleaseDefinition_ByName`, `TestAccDataReleaseDefinitions_List`, `TestAccDataReleaseDefinitionRevision_Basic`, `TestAccDataReleaseDefinitionHistory_Basic`, `TestAccDataReleaseFolder_Basic` all pass using `ProtoV6ProviderFactories`
  - Framework datasources implemented ✓
  - Tests use ProtoV6ProviderFactories ✓
  - Needs live TF_ACC gate run to confirm

- [ ] AC2: GIVEN the acceptance tests for these data sources are updated to use `testutils.GetMuxProviderFactories()` WHEN any of the above test functions run under the mux provider THEN the tests compile and pass; `ExpectNonEmptyPlan: false` on the idempotency re-plan step for each
  - data_release_folder_test.go updated ✓
  - data_release_definition_test.go already uses GetMuxProviderFactories() ✓
  - data_release_definition_revision_history_test.go already uses GetMuxProviderFactories() ✓
  - Needs live TF_ACC gate run to confirm pass

- [x] AC3: GIVEN the CI-equivalent gate runs (no TF_ACC) WHEN `make test && golangci-lint run ./azuredevops/... && make terrafmt-check` executes THEN all checks pass with zero new lint findings on changed files
  - All pass locally ✓
