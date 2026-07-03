# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## What I've tried

### Iteration 0 (current)

**Root cause of gate failure:**
- `TestAccWorkitemtrackingprocessWorkItemType_Basic`: Used `GetProviderFactories()` (SDKv2-only), so when `betterado_workitemtrackingprocess_process` (now in framework provider) was referenced in the config, TF said "does not support resource type betterado_workitemtrackingprocess_process" - because the SDKv2-only provider factory doesn't know about framework resources.
- `TestAccWorkitemtrackingprocessWorkItemType_DataSource_Get`: Used `GetMuxedProviderFactories()` which is correct, but `CheckDestroy: testutils.CheckProcessDestroyed` panicked because `testutils.CheckProcessDestroyed` calls `GetProvider().Meta().(*client.AggregatedClient)` which returns nil when using the mux provider (the SDKv2 singleton Meta is never set).

**What was done:**
1. Created `resource_work_item_type_framework.go` — full framework resource with:
   - `witStaticStringDefault` / `witStaticBoolDefault` inline (sub-packages not vendored)
   - `witUseStateForUnknown` / `witRequiresReplace` inline planmodifiers
   - `witHexColorValidator` for #RRGGBB format
   - Nested pages/sections/groups/controls via `types.ListNestedAttribute`
   - `refreshModel()` using `GetWorkItemTypeExpandValues.Layout`
   - Import format: `<processId>/<referenceName>`

2. Created `data_work_item_type_framework.go` — single work item type lookup by `process_id` + `reference_name`

3. Created `data_work_item_types_framework.go` — list all work item types for a process using `SetNestedAttribute` (so `TestCheckTypeSetElemNestedAttrs` works)

4. Registered all three in `framework_provider.go` Resources()/DataSources()

5. Deregistered from `provider.go` ResourcesMap/DataSourcesMap

6. Deleted 6 SDKv2 files (resource + data + their tests) — `color.go` kept (used by `resource_state.go`)

7. Updated `provider_test.go` counts (removed 3 entries)

8. Rewrote resource acceptance test:
   - No build tag (was no tag originally — sibling tests use `basicWorkItemType()` helper)
   - Uses `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`
   - `checkWorkItemTypeDestroyed` uses `getWorkItemTypeDirectClient()` (direct from env vars)
   - `captureWorkItemTypeEvidence()` calls `CaptureLiveEvidence("acceptance-resource-workitemtrackingprocess-workitemtype", url, wit)`

9. Rewrote data source tests with proper build tags and `checkWorkItemTypeDestroyed` (via shared resource test helper)

## What worked

- Pattern from `resource_process_framework.go` for inline defaults/planmodifiers
- Pattern from `resource_task_group_test.go` for `getDirectClient()` + `captureEvidence()`
- Using `SetNestedAttribute` for the list-data-source to match `TestCheckTypeSetElemNestedAttrs`

## What didn't work

- `stringplanmodifier`, `booldefault`, `stringdefault` sub-packages are NOT vendored — must use inline implementations
- Adding a build tag to `resource_workitemtrackingprocess_workitemtype_test.go` breaks sibling tests that use `basicWorkItemType()` helper

## Open questions

- Will the `parent_work_item_reference_name` default `""` cause import idempotency issues? If yes, may need `UseStateForUnknown` handling.
- The `description` default `""` and the API returning `nil` description — the test uses `TestCheckResourceAttrPair` for description between resource and data source; if resource defaults to `""` and data source returns `""` for nil, this should match.

## Notes for reflection

- `CheckProcessDestroyed` in testutils always uses `GetProvider().Meta()` which panics with muxed provider — each resource using muxed provider must define its own CheckDestroy using direct client
