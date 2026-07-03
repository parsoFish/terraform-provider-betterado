# Agent Memory — WI-4

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0

**Gate failure:** `[no tests to run]` — the test `TestAccServicehookStorageQueuePipelinesFramework_basic` did not exist in the acceptancetests package. The gate command targets `./azuredevops/internal/acceptancetests/` but the prior WI-2 commit put a unit test for `TestServicehookStorageQueuePipelinesFramework_Configure` in `./azuredevops/internal/service/servicehook/` — wrong package.

**Action:** Created `azuredevops/internal/acceptancetests/resource_servicehook_storage_queue_pipelines_framework_test.go` with:
- Build tag: `(all || resource_servicehook_storage_queue_pipelines) && !exclude_servicehooks`
- `TestAccServicehookStorageQueuePipelinesFramework_basic` using `ProtoV6ProviderFactories: testutils.GetMuxProviderFactories()` and `SharedReleaseFixture(t)` (reuses betterado-standing-demo)
- Two-step: create+assert (project_id, queue_name, account_key, stage_state_changed_event), then idempotency re-plan
- `checkServicehookStorageQueuePipelinesFrameworkDestroyed` using `getDirectClient()` + `ServiceHooksClient.GetSubscription`
- `captureServicehookStorageQueueEvidence` calling `testutils.CaptureLiveEvidence("acceptance-resource", url, subscription)` where url = `<orgURL>/_apis/hooks/subscriptions/<id>?api-version=7.1`

**Compile check:** `go build -tags all ./azuredevops/internal/acceptancetests/` — clean.
**Test list check:** `go test -tags all -run TestAccServicehookStorageQueuePipelinesFramework_basic -list '.*'` — test appears in listing.
**Offline test suite:** `make test` — 0 failures.

## What worked

- `getDirectClient()` is defined in `resource_task_group_test.go` (same package) and usable from the new file without redeclaration.
- Pattern from `resource_release_folder_framework_test.go` and `resource_task_group_test.go` is the authoritative reference for framework acceptance tests in this project.
- `testutils.GetMuxProviderFactories()` (not `GetMuxedProviderFactories()`) is the correct function in testutils/commons.go for mux path.
- ADO REST URL for servicehook subscriptions is `<orgURL>/_apis/hooks/subscriptions/<id>?api-version=7.1` (no vsrm host prefix — servicehooks live on dev.azure.com directly).

## What didn't work

- Looking for the acceptance test in `azuredevops/internal/service/servicehook/` — wrong package. The gate command is explicitly `./azuredevops/internal/acceptancetests/`.

### Iteration 1

**Gate failure:** `The argument "account_name" is required, but no definition was found.` — the `hclServicehookStorageQueuePipelinesFramework` HCL template was missing the `account_name` attribute. The resource schema defines `account_name` as Required (distinct from `account_key`). The existing SDKv2 tests in `testutils/hcl.go` already use `account_name = "teststorageacc"`.

**Action:** Updated `hclServicehookStorageQueuePipelinesFramework` to:
1. Accept `accountName` as a new parameter (between `projectID` and `accountKey`).
2. Include `account_name = %[2]q` in the HCL block with re-numbered format specifiers.
3. Updated both call sites in `TestAccServicehookStorageQueuePipelinesFramework_basic` (Step 1 + Step 2) to pass `accountName = "teststorageacc"`.

**Compile check:** `go build -tags all ./azuredevops/internal/acceptancetests/` — clean.
**Test list check:** `TestAccServicehookStorageQueuePipelinesFramework_basic` appears in listing.

### Iteration 2

**Gate failure:** `Provider produced inconsistent result after apply` — `.stage_state_changed_event[0].stage_name was null, but now cty.StringVal("")` and `.stage_state_changed_event[0].pipeline_id was null, but now cty.StringVal("")`.

**Root cause:** In `flattenSubscription`, all optional string fields were set via `types.StringValue(pi["..."])`. When the attribute is Optional and absent from config, the plan holds `types.StringNull()`. After apply, the Read call returns `""` from ADO for unset fields, and `types.StringValue("")` ≠ `null` — Terraform reports inconsistency.

**Fix:** Added `sqpOptionalString(v string) types.String` helper in the framework resource file: maps `"" → types.StringNull()`, non-empty → `types.StringValue(v)`. Applied to all optional string fields in both `stage_state_changed_event` and `run_state_changed_event` flatten paths.

**Key insight:** For framework resources with Optional (non-Computed) string attributes, ALWAYS use null-preserving helpers when flattening from API responses. The API never returns null — it returns `""` — but the plan has null for unset Optional attributes. Storing `""` causes the "Provider produced inconsistent result" error.

## Open questions

_(none)_

## Notes for reflection

_(observations the reflector should capture into the brain; the agent doesn't write them itself, but flags here)_
