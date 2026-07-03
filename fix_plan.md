# Fix Plan

> Checklist for WI-5. Tick items as you complete them; add items as you discover sub-problems.

- [ ] AC1: GIVEN betterado_environment resource and betterado_environment data source are migrated to terraform-plugin-framework WHEN TestAccEnvironment acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values, ExpectNonEmptyPlan: false, destroy is clean
  - Framework resource/datasource files exist and compile: resource_environment_framework.go + data_environment_framework.go
  - Registered in framework_provider.go (Resources + DataSources)
  - TestAccEnvironmentKubernetes_createUpdate FIXED: now uses ProtoV6ProviderFactories (mux) instead of SDKv2-only Providers; testutils.GetDirectClient() used instead of GetProvider().Meta()
  - Awaiting live TF_ACC gate run to confirm full AC1
- [x] AC2: GIVEN SDKv2 environment files are deregistered and deleted WHEN provider.go ResourcesMap and DataSourcesMap are inspected THEN betterado_environment is absent from SDKv2 maps; source files deleted; framework_provider.go includes new factories; provider_test.go counts updated
  - betterado_environment absent from provider.go ResourcesMap and DataSourcesMap (removed in prior iteration)
  - framework_provider.go already includes NewEnvironmentResource + NewEnvironmentDataSource
  - SDKv2 files deleted (resource_environment.go, data_environment.go, resource_environment_test.go)
- [ ] AC3: GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-environment", url, apiResponse) writes .forge/live-evidence/acceptance-resource-environment.json
  - captureEnvironmentEvidence() implemented in resource_environment_test.go calling testutils.CaptureLiveEvidence("acceptance-resource-environment", ...)
  - Awaiting live gate run to write the file
