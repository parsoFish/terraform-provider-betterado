# Changelog

All notable changes to the `betterado` provider are documented here. This
changelog starts at the first public release of the fork. The inherited history
from the upstream `microsoft/azuredevops` provider is preserved in
[`CHANGELOG-upstream.md`](./CHANGELOG-upstream.md).

## Unreleased

ENHANCEMENTS:

- `betterado_task_group`: migrated from Terraform Plugin SDK v2 to Terraform Plugin Framework; `task`, `input`, and `version` are now list-of-object attributes (HCL array-of-objects syntax) with typed defaults eliminating null-fill boilerplate.

FEATURES:

- **New Resource (framework):** `betterado_task_group` is now implemented via `terraform-plugin-framework` (`ListNestedAttribute` for `task`, `input`, and `version`). Configurations must use array-of-objects HCL syntax (`task = [{ ... }]`, `input = [{ ... }]`, `version = [{ ... }]`). Optional task-step fields (`enabled`, `timeout_in_minutes`, `retry_count_on_task_failure`, `always_run`, `inputs`) default to typed zero-values and do not produce a perpetual diff when omitted.

## [0.3.0] - 2026-06-20

FEATURES:

- **Mux entrypoint:** `main.go` now serves the provider via `tf6muxserver`, multiplexing the existing SDKv2 provider (upgraded to protocol 6 via `tf5to6server`) with a new `terraform-plugin-framework` provider stub (`azuredevops/internal/provider/framework_provider.go`). All existing SDKv2 resources are unaffected; new framework resources are registered by adding to `Resources()` / `DataSources()` in `framework_provider.go` without touching `main.go`.

NOTES:

- Added direct dependencies: `github.com/hashicorp/terraform-plugin-framework`, `github.com/hashicorp/terraform-plugin-mux`, `github.com/hashicorp/terraform-plugin-go`. Vendored via `go mod vendor`.
- SDKv2 passthrough proven by `TestAccMuxSdkv2Passthrough` (live `TF_ACC` test against real ADO org); live REST evidence at `https://vsrm.dev.azure.com/davidgparsonson/…/_apis/release/folders`.

## 0.2.0

`betterado_release_definition` schema coverage push — full release-pipeline
surface against the ADO Release API, proven by live TF_ACC acceptance tests with
REST evidence capture.

FEATURES:

- **New Resource:** `betterado_release_folder` — organise release definitions
  into ADO release folders, with a live-acceptance test and a gap-matrix audit.

BREAKING CHANGES:

- `betterado_release_definition`: pipeline stages are now declared as repeated
  `stages { ... }` blocks (renamed from `environment { ... }`). This is a rename
  with **no alias** — update existing configurations from `environment` to
  `stages`. (The intermediate array/`ConfigMode: SchemaConfigModeAttr` syntax was
  reverted to plain blocks so optional stage fields need no null-filling.)

ENHANCEMENTS:

- `betterado_release_definition`: full `betterado_release_definition_permissions`
  coverage — all 12 writable `ReleaseManagement2` permission bits, with
  project-scoped and edge-case token handling, documented in a gap matrix.
- `betterado_release_definition`: new `container_image_trigger` — declare
  container-image CD triggers natively (the final writable gap from the trigger
  coverage matrix).
- `betterado_release_definition`: `deployment_gate` support on stages —
  pre/post-deploy gates with `sampling_interval` and stabilization windows.
- Registry docs (`docs/resources/`, `docs/data-sources/`, `examples/`)
  regenerated from the schema for every new/changed attribute.

NOTES:

- Every schema change in this release is exercised by a `TF_ACC` acceptance test
  against a live Azure DevOps org (apply → API GET read-back → idempotency
  re-plan → destroy), with the live REST evidence captured into the cycle demo.

## 0.1.0 (2026-06-14)

First public release of the `betterado` provider — a fork of
`microsoft/azuredevops` that adds classic release pipeline support.

FEATURES:

- **New Resource:** `betterado_release_definition` — classic release pipelines
  with environments, agent/agentless deploy phases, pre/post-deploy approvals,
  approval options, definition- and environment-level variables (including
  secrets) and variable groups, artifacts, workflow tasks (task and task-group
  references), environment options, execution and retention policies, conditions,
  and triggers (artifact, schedule, CD artifact, source repo, environment).
- **New Resource:** `betterado_task_group` — reusable task groups referenced from
  release definitions as `metaTask` workflow tasks.

ENHANCEMENTS:

- `betterado_release_definition`: environment-scoped secret variables now
  round-trip without a perpetual diff.
- `betterado_release_definition`: the ADO `TF400898` failure on a
  `rollbackRedeploy` environment trigger is wrapped with an actionable hint.
- `betterado_release_definition`: added acceptance-test coverage for
  environment-scoped secret variables.
- `betterado_release_definition`: new `workflow_task` fields `timeout_in_minutes`
  and `retry_count_on_task_failure`.
- `betterado_release_definition`: new `deployment_input.override_inputs` map for
  phase-level task input overrides.

NOTES:

- Includes the full set of Azure DevOps resources and data sources under the
  `betterado_*` prefix.
