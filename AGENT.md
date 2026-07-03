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

### Iteration 1 — Fix gate-blocking permissions test (committed: 86dec7be)

**Gate failure:** `TestAccWorkitemtrackingprocessProcessPermissions_SetPermissions_InheritedProcess`
was matched by the gate regex `TestAccWorkitemtrackingprocessProcess` and failing with:
  `The provider hashicorp/betterado does not support resource type "betterado_workitemtrackingprocess_process"`

**Root cause:** This test used `ProviderFactories: testutils.GetProviderFactories()` (SDKv2-only),
but its HCL configuration creates a `betterado_workitemtrackingprocess_process` resource which is
now ONLY in the framework provider.

**Fix:** Changed `ProviderFactories: testutils.GetProviderFactories()` to
`ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()` in ALL workitemtrackingprocess
test files that reference the process resource in their HCL configurations (13 files total):
resource_workitemtrackingprocess_process_permissions_test.go (InheritedProcess only),
resource_workitemtrackingprocess_{state,inherited_control,inherited_state,system_control,control,field,group,inherited_page,page,rule}_test.go,
data_workitemtrackingprocess_workitemtype{,s}_test.go.

Files NOT updated (no process resource in HCL):
- resource_workitemtrackingprocess_list_test.go (uses process client API, not resource HCL)
- resource_workitemtrackingprocess_process_permissions_test.go SystemProcess function

**KEY LESSON:** When migrating a resource from SDKv2 to framework, search ALL acceptance tests
that reference the resource type in HCL (not just the resource's own test file). The gate regex
`TestAccWorkitemtrackingprocessProcess` matches ALL tests with that prefix.

## Open questions

- If live gate still fails on CreateDisabled: check whether ADO API actually honours `IsEnabled: false` in EditProcess or whether there's a race/cache. The framework Read calls GetProcessByItsId which should return the current value.

## Notes for reflection

- The mux provider (GetMuxedProviderFactories) is required for all framework resources — tests using ProviderFactories (SDKv2 only) will silently skip the framework resource.
- Framework data sources for nested lists (projects inside processes) require explicit `types.ObjectValue` construction — the SDKv2 `map[string]any` pattern doesn't apply.
- **When migrating a resource from SDKv2 to framework, search for ALL acceptance tests in the repo that reference the resource type name in HCL, not just the resource's own test file.** Other resource tests may use it as a dependency and will break if they use the old SDKv2-only provider factories.
