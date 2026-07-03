# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1 (partial iter 0): TestAccWorkitemtrackingprocessProcess_CreateDisabled import verify fix committed
      - Root cause: ADO API omits `isEnabled` (nil) for disabled processes on GetProcessByItsId(Expand=None).
        Framework left model.IsEnabled as types.Bool{} (null) → schema Default(true) applied on import.
        Fix: explicitly set model.IsEnabled = false when process.IsEnabled is nil.
- [x] AC1 (partial iter 1): Read-after-write fix for Create/Update to avoid trusting EditProcess response
      - Root cause: EditProcess API returns IsEnabled=&true in response even when IsEnabled=false was sent.
        This caused state to save true, and the post-apply refresh plan showed drift (true → false).
        Fix: after EditProcess, do GetProcessByItsId to get ground-truth state.
- [ ] AC1 (live gate): TestAccWorkitemtrackingprocessProcess_Basic, TestAccWorkitemtrackingprocessProcess_CreateDisabled, TestAccWorkitemtrackingprocessProcess_CreateAndUpdate must all pass with live TF_ACC
- [ ] AC2 (live gate): TestAccWorkitemtrackingprocessProcess_DataSource_Get and TestAccWorkitemtrackingprocessProcesses_DataSource_AllProcesses must pass live
- [x] AC3: SDKv2 files deleted, deregistered from provider.go, registered in framework_provider.go; provider_test.go counts updated
- [x] AC4: testutils.CaptureLiveEvidence called in acceptance test; .forge/live-evidence/acceptance-resource-workitemtrackingprocess-process.json exists
