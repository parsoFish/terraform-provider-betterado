# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (current)

**Gate failure**: `TestAccWorkitemtrackingprocessProcess_CreateDisabled` import verify: `is_enabled` showed `true` after import, expected `false`.

**Root cause found and fixed**: The Azure DevOps REST API's `GetProcessByItsId` with `Expand=None` omits the `isEnabled` JSON field for **disabled** processes (the field is absent, so the Go SDK decodes it as `nil`). For **enabled** processes, it returns `"isEnabled": true` explicitly (`IsEnabled = &true`).

**Framework bug**: The old `flattenProcess` code:
```go
if process.IsEnabled != nil {
    model.IsEnabled = types.BoolValue(*process.IsEnabled)
}
// else: leaves model.IsEnabled as types.Bool{} (null)
```
When nil, `model.IsEnabled` remained as `types.Bool{}` (null). During import, saving null for an `Optional+Computed+Default(true)` attribute caused the framework to apply the default `true` on the next read, making a disabled process appear enabled.

**SDKv2 comparison**: SDKv2 `d.Set("is_enabled", process.IsEnabled)` where IsEnabled is `(*bool)(nil)` → SDKv2 sets it to the zero value of `TypeBool = false`. This matched the expected `false` for disabled processes. This is why SDKv2 tests passed but framework tests fail.

**Fix applied**: In `flattenProcess`:
```go
if process.IsEnabled != nil {
    model.IsEnabled = types.BoolValue(*process.IsEnabled)
} else {
    model.IsEnabled = types.BoolValue(false)  // nil = disabled (field omitted by ADO API)
}
```
Also defaulted `IsDefault` to `false` when nil for consistency.

**Files changed**: `azuredevops/internal/service/workitemtrackingprocess/resource_process_framework.go`

**Commit**: `196e5771` - fix: default nil IsEnabled to false in flattenProcess for disabled process import

## What worked

- Defaulting nil IsEnabled to false in flattenProcess mirrors SDKv2 behavior and correctly handles the ADO API's omitempty behavior for disabled processes.

## What didn't work

_(none yet)_

## Open questions

- Will the live gate confirm the fix? The IsEnabled logic matches SDKv2 exactly, so it should pass.
- Are there similar nil-defaulting issues in the data source framework implementations? (data_process_framework.go, data_processes_framework.go also have the same pattern but those are read-only and their tests may not check disabled processes explicitly.)

## Notes for reflection

- ADO API quirk: `GetProcessByItsId(Expand=None)` omits `isEnabled` for disabled processes; field absent = nil = false. This is a documented API behavior difference from `EditProcess` which always returns the field.
- Framework vs SDKv2: SDKv2 nil pointer sets zero value; framework nil leaves attribute null which triggers schema defaults. Always explicitly set all fields in flattenProcess.
