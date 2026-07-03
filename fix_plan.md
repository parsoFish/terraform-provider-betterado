# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC3: GIVEN SDKv2 resource_process.go, data_process.go, data_processes.go and their unit test files deleted and deregistered from provider.go WHEN provider.go DataSourcesMap and ResourcesMap inspected THEN betterado_workitemtrackingprocess_process resource and both process data sources are registered ONLY in framework_provider.go (azuredevops/internal/provider/framework_provider.go); provider_test.go resource/datasource counts updated; no orphaned SDKv2 files remain ← DONE in iteration 0 (commit eee2ce73)

- [x] AC4: captureProcessEvidence() call wired into TestAccWorkitemtrackingprocessProcess_Basic; writes .forge/live-evidence/acceptance-resource-workitemtrackingprocess-process.json ← Code written, needs live TF_ACC run to produce file

- [x] Root cause of gate failure fixed (iteration 1, commit 86dec7be): TestAccWorkitemtrackingprocessProcessPermissions_SetPermissions_InheritedProcess was failing because it used `ProviderFactories: testutils.GetProviderFactories()` (SDKv2-only) while its HCL references `betterado_workitemtrackingprocess_process` which is now framework-only. Fixed by updating to ProtoV6ProviderFactories with GetMuxedProviderFactories(). Also proactively fixed 12 other workitemtrackingprocess test files with the same issue (state, inherited_control, inherited_state, system_control, control, field, group, inherited_page, page, rule, data workitemtype, data workitemtypes).

- [ ] AC1: Live TF_ACC must pass TestAccWorkitemtrackingprocessProcess_Basic, TestAccWorkitemtrackingprocessProcess_CreateDisabled, TestAccWorkitemtrackingprocessProcess_CreateAndUpdate
  - Tests use ProtoV6ProviderFactories (GetMuxedProviderFactories) ✓
  - Framework resource registered in framework_provider.go ✓
  - SDKv2 registration removed from provider.go ✓
  - Permissions test (which was blocking the gate run) now also fixed ✓

- [ ] AC2: Live TF_ACC must pass TestAccWorkitemtrackingprocessProcess_DataSource_Get and TestAccWorkitemtrackingprocessProcesses_DataSource_AllProcesses
  - Tests use ProtoV6ProviderFactories (GetMuxedProviderFactories) ✓
  - Data sources registered in framework_provider.go ✓

## What remains for live gate to pass

1. The blocking permissions test is now fixed — gate can now reach AC1/AC2 tests
2. The framework resource Create correctly calls EditProcess after creation when is_enabled=false — should fix CreateDisabled
3. Framework and data sources are correctly registered in the mux provider
