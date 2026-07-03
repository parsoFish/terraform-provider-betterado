# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0

**Gate failure**: `TestAccWorkitemtrackingprocessProcess_CreateDisabled` import verify: `is_enabled` showed `true` after import, expected `false`.

**Root cause found and fixed**: The Azure DevOps REST API's `GetProcessByItsId` with `Expand=None` omits the `isEnabled` JSON field for **disabled** processes (the field is absent, so the Go SDK decodes it as `nil`). For **enabled** processes, it returns `"isEnabled": true` explicitly (`IsEnabled = &true`).

**Framework bug**: The old `flattenProcess` code left `model.IsEnabled` as `types.Bool{}` (null) when nil. During import, saving null for an `Optional+Computed+Default(true)` attribute caused the framework to apply the default `true` on the next read.

**SDKv2 comparison**: SDKv2 `d.Set("is_enabled", process.IsEnabled)` where IsEnabled is `(*bool)(nil)` → SDKv2 sets it to the zero value of `TypeBool = false`.

**Fix applied (commit 196e5771)**: In `flattenProcess`, when `process.IsEnabled == nil`, explicitly set `model.IsEnabled = types.BoolValue(false)`.

### Iteration 1

**Gate failure from last-gate-failure.md**: `TestAccWorkitemtrackingprocessProcess_CreateDisabled` **Step 1/2** (not import step): "After applying this test step, the refresh plan was not empty. `~ is_enabled = true -> false`". State has `true`, config wants `false`.

**Root cause**: The ADO `EditProcess` API **returns `IsEnabled = &true`** in its response body even when the request sends `IsEnabled = false`. The Create handler called `flattenProcess(&model, updated)` where `updated` is the EditProcess response → saves `true` to state → post-apply refresh plan shows drift.

**Key ADO API behavior**:
- `GetProcessByItsId(Expand=None)` for **disabled** process → `IsEnabled = nil` (omitted, omitempty behavior)
- `GetProcessByItsId(Expand=None)` for **enabled** process → `IsEnabled = &true` (explicit)
- `EditProcess(IsEnabled=false)` → returns `IsEnabled = &true` in response (API bug/quirk — does NOT reflect the applied change in the response body)

**Fix applied (commit 55d0e323)**: Read-after-write pattern in both `Create` and `Update` handlers. After calling `EditProcess`, call `GetProcessByItsId` to get ground-truth state before saving to state. This bypasses the EditProcess response's unreliable `IsEnabled` field.

## What worked

- Nil-to-false mapping in `flattenProcess` for `GetProcessByItsId` response (iteration 0)
- Read-after-write pattern in Create/Update (iteration 1): trust `GetProcessByItsId`, not `EditProcess` response

## What didn't work

- Trusting the `EditProcess` API response for `IsEnabled` (ADO returns `&true` in response even when set to false)

## Open questions

- Will the live gate confirm the iteration 1 fix? The read-after-write should produce accurate state.
- Are there similar issues in the data source framework implementations? (data_process_framework.go, data_processes_framework.go are read-only so they're fine — they use GetProcessByItsId or GetListOfProcesses)

## Notes for reflection

- ADO API quirk: `EditProcess` response does NOT reflect the updated `isEnabled` value. Always do a GET after PATCH for this resource.
- ADO API convention: disabled process → `GetProcessByItsId` returns `isEnabled=nil`; enabled process → `isEnabled=&true`.
- Framework vs SDKv2: framework must be careful about trusting PATCH responses; SDKv2 had the same PATCH-then-read-from-PATCH bug but the SDKv2 test must have passed due to the SDKv2 nil pointer → zero bool mapping hiding the issue, OR the import step masked it.
- The non-empty plan check happens after EVERY apply step, not just after imports. So Create must save accurate state.

## Current file inventory

**Framework implementations (present):**
- `azuredevops/internal/service/workitemtrackingprocess/resource_process_framework.go` — `NewProcessResource()`
- `azuredevops/internal/service/workitemtrackingprocess/data_process_framework.go` — `NewProcessDataSource()`
- `azuredevops/internal/service/workitemtrackingprocess/data_processes_framework.go` — `NewProcessesDataSource()`

**SDKv2 files (DELETED):**
- `resource_process.go`, `resource_process_test.go`
- `data_process.go`, `data_process_test.go`
- `data_processes.go`, `data_processes_test.go`

**AC3 status**: SDKv2 files deleted; deregistered from provider.go; registered in framework_provider.go; provider_test.go counts updated. ✓
**AC4 status**: captureProcessEvidence + CaptureLiveEvidence call present in test; .forge/live-evidence/acceptance-resource-workitemtrackingprocess-process.json exists. ✓
