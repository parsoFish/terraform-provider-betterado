## Why

The servicehook resources (`betterado_servicehook_storage_queue_pipelines` and `betterado_servicehook_webhook_tfs`) were implemented in the legacy Terraform SDKv2. Migrating them to terraform-plugin-framework aligns them with the provider's strategic direction (all new resources use framework), eliminates the dual-path maintenance burden, and enables consistent diagnostics and plan-modifier support. Schema and user-visible behaviour are unchanged — this is a pure internal migration.

## What

- **`docs/servicehook-gap-matrix.md`** — new document listing every consumer input, publisher input, and event configuration field for both servicehook resources against the ADO ServiceHooks REST API v7.1; writable gaps flagged as `open` or `deferred`.
- **`azuredevops/internal/service/servicehook/resource_servicehook_storage_queue_pipelines_framework.go`** — new terraform-plugin-framework `resource.Resource` implementation for `betterado_servicehook_storage_queue_pipelines`; `account_key` marked `Sensitive`; `stage_state_changed_event` and `run_state_changed_event` as `ListNestedAttribute`; 404 in Read triggers `RemoveResource`.
- **`azuredevops/internal/service/servicehook/resource_servicehook_storage_queue_pipelines_framework_test.go`** — unit tests for `Configure()` (nil-guard + valid AggregatedClient).
- **`azuredevops/internal/service/servicehook/resource_servicehook_webhook_tfs_framework.go`** — new framework implementation for `betterado_servicehook_webhook_tfs`; all 19 TFS event type blocks as `ListNestedAttribute`; `basic_auth_password` marked `Sensitive`; `http_headers` as `MapAttribute`; HTTP headers encoding/decoding preserved.
- **`azuredevops/internal/service/servicehook/resource_servicehook_webhook_tfs_framework_test.go`** — unit tests for `Configure()`.
- **`azuredevops/internal/provider/framework_provider.go`** — both framework constructors registered in `Resources()`.
- **`azuredevops/provider.go`** — both SDKv2 entries removed from `ResourcesMap` (no duplicate resource type at apply).
- **`azuredevops/provider_test.go`** — resource count updated to reflect SDKv2 removals.
- **`azuredevops/internal/acceptancetests/resource_servicehook_storage_queue_pipelines_framework_test.go`** — live acceptance test `TestAccServicehookStorageQueuePipelinesFramework_basic` using `ProtoV6ProviderFactories`, idempotency step, `CaptureLiveEvidence`.
- **`azuredevops/internal/acceptancetests/resource_servicehook_webhook_tfs_framework_test.go`** — live acceptance test `TestAccServicehookWebhookTfsFramework_basic` using `ProtoV6ProviderFactories`, idempotency step, `CaptureLiveEvidence`.
- **`docs/resources/servicehook_storage_queue_pipelines.md`** and **`docs/resources/servicehook_webhook_tfs.md`** — regenerated Terraform registry docs via `make docs`.
- **`examples/resources/betterado_servicehook_storage_queue_pipelines/resource.tf`** and **`examples/resources/betterado_servicehook_webhook_tfs/resource.tf`** — example HCL blocks embedded in registry docs.
- **`CHANGELOG.md`** — draft changelog entry under `## [Unreleased]` for both migrations.
- **`PROVIDER_VERSION.txt`** — bumped to `1.2.1`.

## How

1. **Gap matrix (WI-1):** Inspected the SDKv2 schemas in `resource_servicehook_storage_queue_pipelines.go` and `resource_servicehook_webhook_tfs.go` alongside the ADO REST API v7.1 subscription model; produced `docs/servicehook-gap-matrix.md` with field-by-field comparison and gap resolution column.

2. **Framework implementation — storage queue (WI-2):** Created `resource_servicehook_storage_queue_pipelines_framework.go` mirroring the `resource_release_folder_framework.go` pattern. Wrote unit `Configure` test; removed SDKv2 provider registration in the same commit.

3. **Framework implementation — webhook TFS (WI-3):** Created `resource_servicehook_webhook_tfs_framework.go` with all 19 TFS event type blocks as `ListNestedAttribute`. Preserved the HTTP-header newline-delimited encoding in `expand`/`flatten` helpers. Wrote unit `Configure` test; removed SDKv2 provider registration in the same commit.

4. **Live acceptance tests (WI-4, WI-5):** Both acceptance tests use `ProtoV6ProviderFactories` (mux path required for framework resources), include an idempotency plan step (`ExpectNonEmptyPlan: false`), call `CaptureLiveEvidence("acceptance-resource", url, subscription)` before destroy, and verify `checkDestroy`.

5. **Docs, examples, changelog, version (WI-6):** Ran `make docs` (tfplugindocs), restored `docs/guides/` with `git checkout`, verified `make terrafmt-check` passes, updated CHANGELOG.md and PROVIDER_VERSION.txt.

6. **Quality gate:** `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` — **green** on branch HEAD.
