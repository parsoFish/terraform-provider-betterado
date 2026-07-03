# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

### Iteration 0

- Read the gate failure: gate found no tests (no test file yet for the framework resource).
- Examined the SDKv2 resource (`resource_servicehook_storage_queue_pipelines.go`) for schema/CRUD shape.
- Examined the framework pattern reference (`resource_release_folder_framework.go`, `framework_defaults.go`).
- Wrote `resource_servicehook_storage_queue_pipelines_framework.go` implementing `resource.Resource` + `resource.ResourceWithImportState`.
  - Inline plan modifier / default helpers (sqpUseStateForUnknown, sqpRequiresReplace, sqpStaticInt64) in the servicehook package.
  - Two ListNestedBlock blocks for stage_state_changed_event / run_state_changed_event.
  - Read() uses `utils.ResponseWasNotFound` → `resp.State.RemoveResource(ctx)` on 404.
  - account_key is Sensitive and preserved from state on Read/Update (API never returns it).
- Wrote test file `resource_servicehook_storage_queue_pipelines_framework_test.go` with build tag `all || resource_servicehook_storage_queue`.
  - First attempt: called `r.Configure()` on the `resource.Resource` interface — compile error (method not in interface).
  - Fix: cast to `resource.ResourceWithConfigure` first; `r.(resource.ResourceWithConfigure).Configure(...)`.
- Updated `framework_provider.go` Resources() to include `servicehook.NewServicehookStorageQueuePipelinesResource`.
- Removed `betterado_servicehook_storage_queue_pipelines` from `provider.go` ResourcesMap.
- Removed it from the expected list in `provider_test.go`; added comment explaining the migration.
- Gate test passes: 3 subtests in TestServicehookStorageQueuePipelinesFramework_Configure all PASS.
- `go build -mod=vendor .` passes clean.
- `TestProvider_HasChildResources` passes.

## What worked

- Casting the `resource.Resource` interface to `resource.ResourceWithConfigure` to call Configure() in tests.
- Inlining plan modifiers and defaults locally (same pattern as `release` package's `framework_defaults.go`).
- Using `schema.ListNestedBlock` (not `schema.ListNestedAttribute`) for the event blocks — they have sub-attributes.
- Using `path.Root("id")` for ImportStatePassthroughID.

## What didn't work

- Calling `r.Configure()` directly on `resource.Resource` — that interface doesn't expose Configure; must use `resource.ResourceWithConfigure`.

## Open questions

_(things that aren't blocking but would be useful to clarify; reflector picks these up)_

## Notes for reflection

- WI-2 completed in a single iteration (0). All 3 ACs implemented and gate test passes.
- The framework pattern for ListNestedBlock event blocks (stage_state_changed_event / run_state_changed_event) is straightforward: declare as `schema.ListNestedBlock`, model as `[]EventModel` with `tfsdk` tags.
- The `resource.ResourceWithConfigure` interface cast is the correct way to test Configure() since `resource.Resource` only exposes Metadata/Schema/Create/Read/Update/Delete.
