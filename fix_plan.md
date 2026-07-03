# Fix Plan

> Checklist for WI-9. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC2: GIVEN SDKv2 variable_group files are deregistered and deleted WHEN provider.go ResourcesMap and DataSourcesMap are inspected THEN betterado_variable_group is absent from SDKv2 maps; source files deleted; framework_provider.go includes new factories; provider_test.go counts updated
- [x] AC3: GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-variable-group", url, apiResponse) writes .forge/live-evidence/acceptance-resource-variable-group.json
- [ ] AC1: GIVEN betterado_variable_group resource and betterado_variable_group data source are migrated to terraform-plugin-framework WHEN TestAccVariableGroup acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values (including secret variables), ExpectNonEmptyPlan: false, destroy is clean
  - [x] Framework resource + data source implemented
  - [x] SDKv2 files deregistered
  - [x] HCL fixtures fixed
  - [x] Inconsistent-result errors fixed
  - [x] Tests use ParallelTest + fixture project
  - [x] Provider Delete waits for 4 consecutive 404s (ContinuousTargetOccurence: 4, Timeout: 90s)
  - [x] CheckDestroy timeout increased to 300s (5 minutes)
  - [ ] Live gate must pass — test failing: "Unexpectedly found a variable group that should be deleted" after 120s (now extended to 300s for current iteration)
