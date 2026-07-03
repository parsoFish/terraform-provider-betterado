# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC3: GIVEN SDKv2 resource_process.go, data_process.go, data_processes.go and their unit test files deleted and deregistered from provider.go WHEN provider.go DataSourcesMap and ResourcesMap inspected THEN betterado_workitemtrackingprocess_process resource and both process data sources are registered ONLY in framework_provider.go (azuredevops/internal/provider/framework_provider.go); provider_test.go resource/datasource counts updated; no orphaned SDKv2 files remain ← DONE in this iteration

- [x] AC4: captureProcessEvidence() call wired into TestAccWorkitemtrackingprocessProcess_Basic; writes .forge/live-evidence/acceptance-resource-workitemtrackingprocess-process.json ← Code written, needs live TF_ACC run to produce file

- [ ] AC1: Live TF_ACC must pass TestAccWorkitemtrackingprocessProcess_Basic, TestAccWorkitemtrackingprocessProcess_CreateDisabled, TestAccWorkitemtrackingprocessProcess_CreateAndUpdate
  - Root cause of prior failure: SDKv2 CreateDisabled not setting is_enabled=false properly
  - Framework fix: on Create, if API returns is_enabled=true but plan says false, immediately call EditProcess to set is_enabled=false
  - Tests now use ProtoV6ProviderFactories (GetMuxedProviderFactories)

- [ ] AC2: Live TF_ACC must pass TestAccWorkitemtrackingprocessProcess_DataSource_Get and TestAccWorkitemtrackingprocessProcesses_DataSource_AllProcesses
  - Tests now use ProtoV6ProviderFactories (GetMuxedProviderFactories)

## What remains for live gate to pass

1. The framework resource Create correctly calls EditProcess after creation when is_enabled=false — this should fix the CreateDisabled perpetual-diff
2. Data sources now registered in framework provider — should work with mux
3. Acceptance tests now use GetMuxedProviderFactories() — required for framework resources
