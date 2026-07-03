# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [ ] AC1: GIVEN betterado_workitemtrackingprocess_process resource migrated to terraform-plugin-framework WHEN TF_ACC acceptance tests run live THEN TestAccWorkitemtrackingprocessProcess_Basic, TestAccWorkitemtrackingprocessProcess_CreateDisabled, TestAccWorkitemtrackingprocessProcess_CreateAndUpdate all pass; provider compiles; no duplicate resource type error
  - [x] Framework resource `resource_process_framework.go` implemented (iter 0)
  - [x] `flattenProcess` nil IsEnabled → false (iter 0)
  - [x] Post-create EditProcess for disabled processes (iter 0)
  - [x] Eventual consistency fix for CreateDisabled: use PATCH response (iter 2)
  - [x] Eventual consistency fix for CreateAndUpdate: poll with StateChangeConf (iter 3) ← CURRENT
- [ ] AC2: GIVEN betterado_workitemtrackingprocess_process and betterado_workitemtrackingprocess_processes data sources migrated to framework WHEN TF_ACC acceptance tests run live THEN TestAccWorkitemtrackingprocessProcess_DataSource_Get and TestAccWorkitemtrackingprocessProcesses_DataSource_AllProcesses pass
  - [x] Framework data sources `data_process_framework.go` and `data_processes_framework.go` exist
- [ ] AC3: GIVEN SDKv2 resource_process.go, data_process.go, data_processes.go and their unit test files deleted and deregistered from provider.go WHEN provider.go DataSourcesMap and ResourcesMap inspected THEN betterado_workitemtrackingprocess_process resource and both process data sources are registered ONLY in framework_provider.go; provider_test.go resource/datasource counts updated; no orphaned SDKv2 files remain
- [ ] AC4: GIVEN live acceptance test running WHEN process is read back before destroy THEN testutils.CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-process", url, apiResponse) is called; .forge/live-evidence/acceptance-resource-workitemtrackingprocess-process.json written
  - [x] `captureProcessEvidence` implemented in acceptance test (iter 0)

## Notes

- Gate command: `go test -tags all -run TestAccWorkitemtrackingprocessProcess ./azuredevops/internal/acceptancetests/`
- Last gate failure: CreateAndUpdate step 3/4 — is_enabled drift (true → false) after update
- Fix: `waitForProcessIsEnabled` helper using `retry.StateChangeConf` to poll until API consistent
- Prior approaches that failed: PATCH response direct use (iter 2), single read-after-write (iter 1)
