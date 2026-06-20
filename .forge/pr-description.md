## Why

`betterado_task_group` was implemented via Terraform Plugin SDK v2. The SDKv2 `schema.Schema` type uses block syntax for nested objects, which conflicted with the mux/framework world we introduced in `[0.3.0]` and produced a perpetual-diff bug: any optional task-step field (`enabled`, `timeout_in_minutes`, `always_run`, `retry_count_on_task_failure`, `inputs`) not explicitly set in config would round-trip as `null` vs. the ADO API's zero-value default, causing Terraform to propose a change on every plan. Migrating to `terraform-plugin-framework` with `ListNestedAttribute` + typed `Default` values on every optional field eliminates the noise and aligns the resource with protocol-6 semantics the mux already serves.

## What

- **New framework resource** (`resource_task_group_framework.go`, 1015 lines): `NewTaskGroupResource() resource.Resource` — a full protocol-6 implementation of `betterado_task_group` with:
  - `ListNestedAttribute` for `task`, `input`, and `version` (array-of-objects HCL: `task = [{ ... }]`).
  - Typed `Default` on every optional field in `task` (`enabled = true`, `always_run = false`, `condition = "succeeded()"`, `timeout_in_minutes = 0`, `retry_count_on_task_failure = 0`, `inputs = {}`, `environment = {}`).
  - Framework-typed `expandTaskGroupFramework` / `flattenTaskGroupFramework` helpers (do not reuse SDKv2 `*schema.ResourceData` helpers).
  - CRUD via `clients.TaskAgentClient`: `AddTaskGroup` → `GetTaskGroups` → `UpdateTaskGroup` (stale-revision retry) → `DeleteTaskGroup`.
- **Provider wiring** (`framework_provider.go`): `taskagent.NewTaskGroupResource` added to `Resources()` slice. SDKv2 `ResourcesMap` entry removed from `provider.go`; `expectedResources` list updated in `provider_test.go`.
- **Acceptance tests updated** (`resource_task_group_test.go`, `data_task_group_test.go`): `hclTaskGroupBasic`, `hclTaskGroupWithGapFields`, `hclTaskGroupDataSourceBasic` all converted from block syntax to array syntax. `captureTaskGroupEvidence` writes live REST evidence to `.forge/live-evidence/`.
- **Mux provider helper** (`testutils/mux_provider.go`): factory for acceptance tests that routes through the mux.
- **Docs regenerated** (`docs/resources/task_group.md`): reflects `ListNestedAttribute` schema; `examples/resources/betterado_task_group/resource.tf` updated to array syntax.
- **Changelog + version**: `CHANGELOG.md` `## Unreleased` entry added; `PROVIDER_VERSION.txt` bumped to `0.4.0`.

## How

All changes are exercised by two layers of evidence:

1. **Unit / offline** — `TestTaskGroupFramework_Schema` (verifies schema attributes and type name) and `TestFrameworkProvider_HasTaskGroupResource` (verifies framework provider registration) run without credentials. These pass in the quality gate: `go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...` → all `ok`.

2. **Live TF_ACC** (run against real Azure DevOps org `davidgparsonson`):
   - `TestAccTaskGroup_basic`: `terraform apply` with array syntax → ADO REST read-back (`GET /distributedtask/taskgroups/<id>`) → idempotency re-plan (no changes) → `destroy`. Evidence in `.forge/live-evidence/acceptance-resource.json` (task group `3b399cdc-04fe-4ca4-8652-91f93c7a33e4`, created `2026-06-20T01:10:28Z`).
   - `TestAccTaskGroup_withGapFields`: proves omitted optional fields produce no perpetual diff.
   - `TestAccTaskGroupDataSource_basic`: data source reads back resource attributes (`name`, `description`, `category`) via `TestCheckResourceAttrPair`. Evidence in `.forge/live-evidence/task-group-datasource-acceptance.json` (task group `7c1199e7-16f8-4af9-bd56-e31a15b66d55`, created `2026-06-20T01:23:06Z`).

Changed files anchored to `git diff --name-only main...HEAD`:
- `azuredevops/internal/service/taskagent/resource_task_group_framework.go` (new)
- `azuredevops/internal/service/taskagent/resource_task_group_framework_test.go` (new)
- `azuredevops/internal/provider/framework_provider.go` (updated)
- `azuredevops/internal/provider/framework_provider_test.go` (new)
- `azuredevops/provider.go` (SDKv2 entry removed)
- `azuredevops/provider_test.go` (expectedResources updated)
- `azuredevops/internal/acceptancetests/resource_task_group_test.go` (array HCL syntax)
- `azuredevops/internal/acceptancetests/data_task_group_test.go` (array HCL syntax)
- `azuredevops/internal/acceptancetests/testutils/mux_provider.go` (new)
- `examples/resources/betterado_task_group/resource.tf` (array syntax)
- `docs/resources/task_group.md` (regenerated)
- `CHANGELOG.md` (DRAFT Unreleased entry)
- `PROVIDER_VERSION.txt` (0.4.0)
