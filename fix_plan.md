# Fix Plan

> Checklist for WI-9. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC2: GIVEN SDKv2 variable_group files are deregistered and deleted WHEN provider.go ResourcesMap and DataSourcesMap are inspected THEN betterado_variable_group is absent from SDKv2 maps; source files deleted; framework_provider.go includes new factories; provider_test.go counts updated — **DONE** (committed in prior iteration)
- [x] AC3: GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-variable-group", url, apiResponse) writes .forge/live-evidence/acceptance-resource-variable-group.json — **DONE** (committed in prior iteration)
- [ ] AC1: GIVEN betterado_variable_group resource and betterado_variable_group data source are migrated to terraform-plugin-framework WHEN TestAccVariableGroup acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values (including secret variables), ExpectNonEmptyPlan: false, destroy is clean
  - [x] Root cause fixed iter-1: All HCL test fixtures using block syntax (`variable { ... }`) converted to framework attribute syntax (`variable = [{ ... }]`) — **DONE** (commit 63044dd2)
  - [x] Root cause fixed iter-2: `.key_vault: was null, but now cty.ListValEmpty(...)` — return `types.ListNull(...)` for non-keyvault VGs — **DONE** (commit 1f556edf)
  - [x] Root cause fixed iter-2: `.variable: inconsistent values for sensitive attribute` — changed `variable` from `SetNestedAttribute` to `ListNestedAttribute`; added `mergeVariableListWithPlan` to preserve plan order + sensitive values in Apply result — **DONE** (commit 1f556edf)
  - [ ] Pending: Live gate must confirm TestAccVariableGroup tests pass end-to-end
