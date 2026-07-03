# Agent Memory — WI-7

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0 (complete)
- Created `resource_list_framework.go` — framework implementation of `betterado_workitemtrackingprocess_list`
- Created `resource_field_framework.go` — framework implementation of `betterado_workitemtrackingprocess_field`
- Registered `NewListResource`, `NewFieldResource` in `framework_provider.go`
- Removed `ResourceList()` and `ResourceField()` from `provider.go` SDKv2 ResourcesMap
- Deleted `resource_list.go`, `resource_field.go`, `order.go`
- Updated `provider_test.go` to remove both resources from expected SDKv2 list
- Updated list acceptance tests to use `ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories()`
- Fixed `checkListDestroyed` to build client from env vars (mux-compatible)
- Added `captureListEvidence()` for AC4
- All unit tests pass (`go test ./azuredevops/...`)

### Iteration 1 (complete)
- **Root cause of gate failure**: `checkFieldDestroyed` in `resource_workitemtracking_field_test.go` (NOT the process field test) used `testutils.GetProvider().Meta().(*client.AggregatedClient)` which panics when `ProtoV6ProviderFactories` is used (muxed provider sets SDKv2 Meta() to nil).
- **Fix**: Updated `checkFieldDestroyed` to build client directly from env vars (`AZDO_ORG_SERVICE_URL` + `AZDO_PERSONAL_ACCESS_TOKEN`) — same pattern as `checkListDestroyed`. Added `os` and `azuredevops` imports.
- **Compile**: Clean (`go build -tags all ./azuredevops/internal/acceptancetests/`)
- **Note**: Pre-existing failures exist in `TestAccTaskGroupStateUpgradeSmoke`, `TestBuildDefinition_Create/Update_DoesNotSwallowError`, etc — those are on main branch too, not caused by our change.

## What worked

- **Inline plan modifiers and defaults**: Vendor lacks `resource/schema/stringplanmodifier`, `resource/schema/booldefault` etc. — must use inline struct implementations. Available sub-packages: `resource/schema/planmodifier`, `resource/schema/defaults`.
- **env-var-based client in CheckDestroy**: `os.Getenv("AZDO_ORG_SERVICE_URL")` + `azuredevops.NewAuthProviderPAT(pat)` — works with both SDKv2 and mux tests.
- **API types**: `GetWorkItemTypeField`, `AddFieldToWorkItemType`, `UpdateWorkItemTypeField` return `*ProcessWorkItemTypeField`. `PickList.Items` is `*[]string`.

## What didn't work

- Attempted to use `resource/schema/stringplanmodifier.RequiresReplace()` and `resource/schema/booldefault.StaticBool()` — those sub-packages are NOT vendored. Got "cannot find module" error.

### Iteration 2 (complete)
- **Root cause of gate failure**: `TestAccWorkitemtrackingprocessField_Identity` — `allow_groups` is write-only (not returned by the Azure DevOps API on read). The `identityField()` config sets `allow_groups = true`, so ImportStateVerify fails with "allow_groups" missing after import.
- **Fix**: Added `ImportStateVerifyIgnore: []string{"allow_groups"}` to the import step in `TestAccWorkitemtrackingprocessField_Identity` only (the other tests — Basic, Integer, Update — don't set `allow_groups` in their configs, so they don't need it).
- **Compile**: Clean, all unit tests pass.

### Iteration 3 (complete)
- **Root cause of gate failure**: `TestAccWorkitemtrackingprocessList_Update` step 5/6 — after reverting from `updatedList` (is_suggested=true, items+Yellow) back to `basicList` (is_suggested=false, items without Yellow), the state still showed the old values (`is_suggested=true`, Yellow in items). This caused perpetual drift.
- **Root cause**: The `UpdateList` Azure DevOps API can return stale data in its response (e.g. returning the OLD `is_suggested` value before the update propagates). The previous polling logic used `listPickListsEqual(updated, readList)` — where `updated` was the UpdateList response — so if UpdateList returned stale data, polling converged on the stale values and wrote them to state. On the next refresh (Read), the API would return the correct values, revealing drift.
- **Fix**:
  1. Changed polling target from `updated` (UpdateList response) to `desired` (a PickList built from plan values). Now we poll until GetList matches what we WANTED to set.
  2. Changed `listPickListsEqual` to treat nil fields in arg `a` as wildcards (skip comparison). This allows the desired struct to omit the `type` field (type requires-replace, not updated) without causing perpetual mismatch.
  3. Made type comparison case-insensitive (`strings.EqualFold`) to handle "String" vs "string" from the API.
- **Compile**: Clean; all unit tests pass.

## Open questions

- All known failures fixed. Awaiting live gate run to confirm all 7 acceptance tests pass.

## Notes for reflection

- Pattern: when migrating SDKv2 resources, `checkXxxDestroyed` functions that use `testutils.GetProvider().Meta()` must be rewritten to build clients from env vars when tests switch to `ProtoV6ProviderFactories`. THIS IS CRITICAL: the check function may live in a DIFFERENT file (e.g., `resource_workitemtracking_field_test.go`) than the test being run (e.g., `resource_workitemtrackingprocess_field_test.go`). Search for all callers.
- The vendor directory determines which framework sub-packages are available; always check with `ls vendor/github.com/hashicorp/terraform-plugin-framework/resource/schema/` before using sub-packages.
