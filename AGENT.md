# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 — Full framework migration (committed: eee2ce73)

**Gate failure addressed:** `TestAccWorkitemtrackingprocessProcess_CreateDisabled` was failing
with perpetual-diff: `is_enabled = true -> false`. Root cause: old SDKv2 was still registered
and the Create->Update sequence wasn't persisting is_enabled=false.

**What was done:**
1. Created `resource_process_framework.go` — full framework resource. Key fix: after
   CreateNewProcess (which always returns is_enabled=true), if plan says false, immediately
   calls EditProcess.
2. Created `data_process_framework.go` — framework data source for single process by ID.
3. Created `data_processes_framework.go` — framework data source for all processes.
4. Registered all three in `framework_provider.go`.
5. Deregistered from `provider.go` SDKv2 maps.
6. Deleted 6 SDKv2 files (resource_process.go, data_process.go, data_processes.go + tests).
7. Updated `provider_test.go` counts.
8. Updated 3 acceptance test files → `ProtoV6ProviderFactories: GetMuxedProviderFactories()`.
9. Added `captureProcessEvidence()` to resource test for AC4 live evidence.

**Key technical facts:**
- `stringplanmodifier`, `booldefault`, `stringdefault` sub-packages NOT vendored; use inline implementations with `process`-prefixed type names to avoid package-level collisions.
- `processExpandLevelMap` defined in `data_process_framework.go`, reused by `data_processes_framework.go` (same package).
- Offline gates all pass: go build, go test, gofmt.
- `data_processes_framework.go` uses `types.ListNestedAttribute` with explicit `types.ObjectValue` construction.

## What worked

- Framework migration pattern from `release` package (framework_defaults.go + resource_release_folder_framework.go).
- Post-create EditProcess call for is_enabled=false.
- Inline plan modifiers (processUseStateForUnknownMod, processRequiresReplaceMod).
- Using `processStaticString`/`processStaticBool` names to avoid type collision with other packages' inline defaults.

## What didn't work

- SDKv2 resource create: the old createResourceProcess called updateResourceProcess after create to handle is_enabled=false, but this caused a perpetual diff in live tests (the read was returning true). Framework implementation is now authoritative.

## Open questions

- If live gate still fails on CreateDisabled: check whether ADO API actually honours `IsEnabled: false` in EditProcess or whether there's a race/cache. The framework Read calls GetProcessByItsId which should return the current value.

## Notes for reflection

- The mux provider (GetMuxedProviderFactories) is required for all framework resources — tests using ProviderFactories (SDKv2 only) will silently skip the framework resource.
- Framework data sources for nested lists (projects inside processes) require explicit `types.ObjectValue` construction — the SDKv2 `map[string]any` pattern doesn't apply.
