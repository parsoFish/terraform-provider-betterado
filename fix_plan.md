# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1 (partial): TestAccWorkitemtrackingprocessProcess_CreateDisabled import verify fix committed
      - Root cause: ADO API omits `isEnabled` (nil) for disabled processes on GetProcessByItsId(Expand=None).
        Framework left model.IsEnabled as types.Bool{} (null) → schema Default(true) applied on import.
        Fix: explicitly set model.IsEnabled = false when process.IsEnabled is nil.
- [ ] AC1 (live gate): TestAccWorkitemtrackingprocessProcess_Basic, TestAccWorkitemtrackingprocessProcess_CreateDisabled, TestAccWorkitemtrackingprocessProcess_CreateAndUpdate must all pass with live TF_ACC
- [ ] AC2: TestAccWorkitemtrackingprocessProcess_DataSource_Get and TestAccWorkitemtrackingprocessProcesses_DataSource_AllProcesses must pass live
- [ ] AC3: SDKv2 files deleted, only registered in framework_provider.go; provider_test.go counts updated
- [ ] AC4: testutils.CaptureLiveEvidence called in acceptance test; .forge/live-evidence file written
