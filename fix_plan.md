# Fix Plan

> Checklist for WI-9. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN betterado_variable_group resource and betterado_variable_group data source are migrated to terraform-plugin-framework WHEN TestAccVariableGroup acceptance tests run live (TF_ACC=1) THEN apply succeeds, provider read-back verifies all attributes with non-default values (including secret variables), ExpectNonEmptyPlan: false, destroy is clean
  - resource_variable_group_framework.go: CRUD + Import + secret_value recovery + allow_access + key_vault
  - data_variable_group_framework.go: Read data source
  - Acceptance tests: ProtoV6ProviderFactories (mux) + fixture project + ExpectNonEmptyPlan: false + GetDirectClient for CheckDestroy
- [x] AC2: GIVEN SDKv2 variable_group files are deregistered and deleted WHEN provider.go ResourcesMap and DataSourcesMap are inspected THEN betterado_variable_group is absent from SDKv2 maps; source files deleted; framework_provider.go includes new factories; provider_test.go counts updated
  - data_variable_group.go deleted; data_variable_group_test.go deleted
  - resource_variable_group.go KEPT (shared helpers needed by resource_variable_group_variable.go for WI-10)
  - ResourcesMap and DataSourcesMap entries removed; provider_test.go counts updated
- [x] AC3: GIVEN live acceptance test runs WHEN CaptureLiveEvidence is called THEN CaptureLiveEvidence("acceptance-resource-variable-group", url, apiResponse) writes .forge/live-evidence/acceptance-resource-variable-group.json
  - captureVariableGroupEvidence() wired in TestAccVariableGroup_basic Check step (best-effort, non-fatal)
  - Requires TF_ACC=1 live run to produce the evidence file

## Status: ALL ACs structurally complete, awaiting live gate TF_ACC=1 run
