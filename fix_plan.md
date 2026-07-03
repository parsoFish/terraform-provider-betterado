# Fix Plan

> Checklist for WI-5. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_environment resource and betterado_environment data source are migrated to terraform-plugin-framework WHEN TestAccEnvironment acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values, ExpectNonEmptyPlan: false, destroy is clean
- [x] AC2: GIVEN SDKv2 environment files are deregistered and deleted WHEN provider.go ResourcesMap and DataSourcesMap are inspected THEN betterado_environment is absent from SDKv2 maps; source files deleted; framework_provider.go includes new factories; provider_test.go counts updated
- [x] AC3: GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-environment", url, apiResponse) writes .forge/live-evidence/acceptance-resource-environment.json

## Additional fixes applied this iteration
- [x] Fix golangci-lint findings in resource_variable_group_framework.go (errcheck, gocritic elseif)
- [x] Fix golangci-lint nilerr findings in resource_variable_group_test.go
