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

## Open questions

_(none)_

## Notes for reflection

_(observations the reflector should capture into the brain; the agent doesn't write them itself, but flags here)_
