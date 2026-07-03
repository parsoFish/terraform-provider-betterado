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

## Open questions

- Live acceptance test results: will the `ImportStateVerify` for `betterado_workitemtrackingprocess_list` pass? The SDKv2 version has eventual-consistency polling on Update — the framework version replicates that. The `type` field is normalized to lowercase in flattenListFramework but the API might return it uppercase. Watch for diffs.
- For `betterado_workitemtrackingprocess_field`: `allow_groups` is write-only (never read back). ImportStateVerify will set it to null — if the test config sets it, ImportStateVerify may fail. The field tests that set `allow_groups` may need `ImportStateVerifyIgnore: []string{"allow_groups"}`.

## Notes for reflection

- Pattern: when migrating SDKv2 resources, `checkXxxDestroyed` functions that use `testutils.GetProvider().Meta()` must be rewritten to build clients from env vars when tests switch to `ProtoV6ProviderFactories`. THIS IS CRITICAL: the check function may live in a DIFFERENT file (e.g., `resource_workitemtracking_field_test.go`) than the test being run (e.g., `resource_workitemtrackingprocess_field_test.go`). Search for all callers.
- The vendor directory determines which framework sub-packages are available; always check with `ls vendor/github.com/hashicorp/terraform-plugin-framework/resource/schema/` before using sub-packages.
