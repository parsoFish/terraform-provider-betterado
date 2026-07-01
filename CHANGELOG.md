# Changelog

All notable changes to the `betterado` provider are documented here. This
changelog starts at the first public release of the fork. The inherited history
from the upstream `microsoft/azuredevops` provider is preserved in
[`CHANGELOG-upstream.md`](./CHANGELOG-upstream.md).

## [Unreleased]

## [1.2.0] - 2026-07-01

### FEATURES

- **`betterado_release_folder` migrated to terraform-plugin-framework.** The
  resource now uses the terraform-plugin-framework implementation served through
  the mux provider, alongside the existing SDKv2 path. CRUD operations continue
  to target the Release Management API at `vsrm.dev.azure.com`; the schema is
  unchanged (`project_id`, `path`, `description`). Verified by live acceptance
  test `TestAccReleaseFolderFramework`.

- **`betterado_release_definition_permissions` migrated to terraform-plugin-framework.**
  The resource now uses the framework implementation (ReleaseManagement2 security
  namespace, `c788c23e-1b46-4162-8f5e-d7585343b5de`) served through the mux provider.
  Schema is unchanged (`project_id`, `principal`, `release_definition_id`, `permissions`,
  `replace`). Supports all writable ACL bits with idempotent apply. Verified by
  live acceptance test `TestAccReleaseDefinitionPermissionsFramework`.

- **`betterado_release_definition` data source migrated to terraform-plugin-framework.**
  Reads a single release definition by `project_id` + `name` (or `release_definition_id`);
  returns `id`, `description`, `path`, and `release_name_format`.
  Served through the mux provider. Verified by live acceptance test
  `TestAccDataReleaseDefinition_Basic`.

- **`betterado_release_definition_history` data source migrated to terraform-plugin-framework.**
  Reads the revision history for a release definition; exposes a `revisions` list with
  `revision`, `changed_by`, `changed_date`, `change_type`, and `comment` per entry.
  Served through the mux provider. Verified by live acceptance test
  `TestAccDataReleaseDefinitionHistory_Basic`.

- **`betterado_release_definition_revision` data source migrated to terraform-plugin-framework.**
  Fetches the serialised JSON snapshot of a specific release definition revision via
  `project_id`, `definition_id`, and `revision`; exposes `json_content`.
  Served through the mux provider. Verified by live acceptance test
  `TestAccDataReleaseDefinitionRevision_Basic`.

- **`betterado_release_definitions` data source migrated to terraform-plugin-framework.**
  Lists all release definitions in a project (filtered by optional `path`); exposes a
  `definitions` list with `id`, `name`, and `path` per entry.
  Served through the mux provider. Verified by live acceptance test
  `TestAccDataReleaseDefinitions_Basic`.

- **`betterado_release_folder` data source migrated to terraform-plugin-framework.**
  Reads release folder metadata by `project_id` + `path`; exposes `id`,
  `path`, and `description`. Served through the mux provider. Verified by live
  acceptance test `TestAccDataReleaseFolder_Basic`.

## [1.0.5] - 2026-06-21

### ENHANCEMENTS

- **Quieter plans for ADO-assigned computed fields.** `revision` and each stage's
  `id` now use a `UseStateForUnknown` plan modifier, so they no longer render as
  `(known after apply)` churn on plans that change unrelated attributes. They are
  pinned to their prior-state value when the plan leaves them unknown; if ADO
  actually changed them (e.g. a revision bump or a stage-id reassignment on a
  structural change) the value reconciles on the next refresh. Safe alongside the
  v1.0.4 plan-faithful create/update path (`mergePlanComputed` keeps plan-known
  values consistent at apply). Verified live across idempotency, update, and
  structural add/remove-stage scenarios.

## [1.0.4] - 2026-06-21

### BUG FIXES

- **Plan-faithful create/update — the definitive fix for `inconsistent values for
  sensitive attribute`.** Create and Update previously rebuilt the entire resource
  state from ADO's create/update response (`flattenReleaseDefinitionFramework`).
  Because ADO normalises some sent fields and the variable `value` attribute is
  `Sensitive`, any field ADO returned differently from the plan produced a hard
  `Provider produced inconsistent result after apply` error — a whack-a-mole that
  earlier point fixes (value preservation in 1.0.2, stage ordering in 1.0.3) could
  not fully close while the result was re-derived from the API. Create/Update now
  build the result from the **plan**: every configured (plan-known) attribute —
  stage order, conditions, variables and their sensitive values, variable_groups,
  tasks, gates, options — is kept exactly as planned, and only server-assigned
  **computed** values that the plan left unknown (resource id, revision, stage
  ids, schedule job_ids, owner, …) are overlaid from the API response, matched by
  stage name. This makes the apply result consistent with the plan for every
  configured attribute *by construction*. Read still reconciles from the API (with
  secret-value preservation), so genuine out-of-band drift continues to surface on
  the next plan — but create/update can no longer fail with an inconsistent-result
  error for a configured value. Verified by a from-scratch 6-stage acceptance test
  (chained environmentState conditions, per-stage variable_groups, gates,
  environment_options, execution_policy, an agentless WAIT stage, per-stage secret
  variables) plus the full release acceptance suite.

## [1.0.3] - 2026-06-21

### BUG FIXES

- **Stage order plan fidelity (framework apply consistency).**
  `betterado_release_definition` failed on apply with `Provider produced
  inconsistent result after apply: .stages: inconsistent values for sensitive
  attribute` for releases with multiple stages — most reliably when the
  configured `stages` list order differed from rank order, but also on real
  multi-stage (e.g. 6-stage) releases, because ADO's create/update response can
  return environments in a different order than the configured list. The variable
  `value` attribute is `Sensitive`, and the framework compares list elements
  positionally, so any stage-order mismatch makes the per-index sensitive
  comparison fail. v1.0.2 preserved variable *values* by name but not stage
  *order*. Create/Update/Read now reorder the resulting `stages` list to match the
  plan/prior order (by stage name), keeping every per-stage configured attribute —
  including sensitive values — index-aligned with the plan. Stages not present in
  the plan order are preserved (never dropped). Combined with the existing
  value-preservation, the apply result is plan-faithful for every configured
  attribute.

## [1.0.2] - 2026-06-21

### BUG FIXES

- **Variable values on create/update (sensitive write-only consistency).**
  `betterado_release_definition` failed on apply with `Provider produced
  inconsistent result after apply: .stages: inconsistent values for sensitive
  attribute` for any release that declared definition-level or stage-level
  `variables`. The variable `value` attribute is unconditionally `Sensitive`
  (write-only), and the framework requires a sensitive attribute's post-apply
  value to match the plan exactly — but Create/Update overwrote the planned value
  with the ADO API response, which does not faithfully echo variable values
  (empty for secrets, normalised for others). Create/Update/Read now preserve the
  configured/planned value for every variable (definition-level and per-stage),
  keyed by name and stage, falling back to the API value only when there is no
  prior value (e.g. import).

### TESTS

- `betterado_release_definition_permissions` acceptance tests
  (`Set`/`Update`/`AllWritable`) were red since the v1.0.0 Plugin Framework
  migration (SDK v2-only provider factory + block-syntax HCL + self-created
  project). Moved them onto the mux provider factories, attribute-syntax HCL, and
  the shared fixture project.

## [1.0.1] - 2026-06-21

### BUG FIXES

- **Framework provider client wiring (completes WI-5).** `betterado_release_definition`
  and `betterado_task_group` (migrated to the Plugin Framework in v1.0.0) failed
  on apply with `Client not configured` when the provider was configured via an
  HCL `provider "betterado" { ... }` block (for example a `personal_access_token`
  sourced from a data source) instead of environment variables. The framework
  provider's `Configure` now reads `org_service_url` and `personal_access_token`
  from the provider configuration, falling back to the `AZDO_ORG_SERVICE_URL` /
  `AZDO_PERSONAL_ACCESS_TOKEN` environment variables — matching the SDK v2
  provider's behaviour. Proven by a live acceptance test that clears the PAT env
  var and supplies the credential only via the HCL block.

### CI

- `tag-on-changelog` now dispatches `release.yml` (via `workflow_dispatch`) after
  pushing the version tag. A tag pushed by the Actions `GITHUB_TOKEN` does not
  trigger the `push` event (GitHub's anti-recursion rule), so auto-publish never
  fired for v0.3.0–v1.1.0; `workflow_dispatch` is the documented exception that
  `GITHUB_TOKEN` may trigger, so releases now publish automatically on merge.

## [1.0.0] - 2026-06-20

First major release. Migrates the two betterado resources to the
terraform-plugin-framework and restores the full classic-release surface on the
framework resource, with live-proven idempotency. (Consolidates the internal
0.3.0–0.5.0 development tags, which were never published to the Terraform
Registry — 0.2.0 was the last public release.)

### BREAKING CHANGES

- **HCL syntax (framework migration).** `betterado_release_definition` and
  `betterado_task_group` are now terraform-plugin-framework resources and use
  attribute (array-of-objects) HCL syntax instead of SDK v2 blocks. Update
  configurations: `stages = [{ … }]`, `deploy_phase = [{ … }]`,
  `cd_artifact_trigger = [{ … }]`, `variables = { NAME = { … } }`,
  `task = [{ … }]`, `input = [{ … }]`, `version = [{ … }]`, etc.
- `betterado_release_definition`: the `environment` block was renamed to `stages`
  (no alias). Existing v0.x Terraform state is automatically upgraded from schema
  version 0 to 1 on `terraform init` (state upgraders for both resources).

### FEATURES

- **Plugin Framework migration.** `main.go` serves the provider via
  `tf6muxserver`, multiplexing the SDK v2 provider (upgraded to protocol 6 via
  `tf5to6server`) with a new terraform-plugin-framework provider.
  `betterado_task_group` and `betterado_release_definition` are now framework
  resources; all other resources/data sources are unaffected.
- **New: `pull_request_trigger`** on `betterado_release_definition` triggers —
  declare pull-request CD triggers (`artifact_alias`, `target_branches`, `tags`,
  `use_artifact_reference`). This was the last writable gap versus the deployed
  ADO trigger surface.

### ENHANCEMENTS — full SDK-v2 parity restored on the framework resource

The framework migration had dropped a large part of the `release_definition`
surface; this release restores all of it (mirroring the ADO Release API):

- **Triggers:** restored `source_repo_trigger`, `container_image_trigger`, and
  the `cd_artifact_trigger` filters `tag_filter`, `use_build_definition_branch`,
  and `create_release_on_build_tagging`.
- **Stages:** restored `execution_policy` (`concurrency_count` — `0` means
  unlimited — and `queue_depth_count`), `environment_trigger`, stage-level
  `schedule`, `process_parameters`, `properties`, `owner`, and the computed
  stage `id`.
- **Deploy phases:** restored `deployment_input.override_inputs` and
  `deployment_input.job_cancel_timeout_in_minutes`.
- **Environment options:** restored `pull_request_deployment_enabled` plus the
  (API-deprecated) `email_recipients`, `skip_artifacts_download`,
  `timeout_in_minutes`, and `enable_access_token`.
- **Workflow tasks / gates:** restored `workflow_task.definition_type`; gave
  `pre_deployment_gates` / `post_deployment_gates` full gate-task fidelity
  (display name, task id, version, condition, inputs, …) in place of the prior
  name/task-id stub.
- **Definition-level `tags`** restored.
- Terraform Registry docs (`docs/resources/`, `docs/data-sources/`, `examples/`)
  regenerated from the framework schema (attribute syntax).

### BUG FIXES

- **`revision` idempotency.** `betterado_release_definition` no longer shows a
  perpetual `~ revision = N -> (known after apply)` diff. A resource-level
  `ModifyPlan` detects a true no-op (the plan differs from prior state only in
  framework-injected unknowns) and snaps the plan back to prior state, while
  server-(re)assigned computed values (revision, stage ids, ADO-assigned
  environment ids) correctly go "known after apply" on real changes — avoiding
  "inconsistent result after apply" when ADO bumps them.
- **Secret variables.** `is_secret` variable values round-trip without drift —
  ADO never returns secret values on read, so the planned/prior value is
  preserved and the `value` attribute is marked sensitive.

### NOTES

- Removed the orphaned SDK v2 `release_definition` / `task_group` resource
  implementations left dangling by the migration, and added a
  `framework-migration-guard` skill (`forge/skills/framework-migration-guard/`)
  whose `check.sh` fails when any resource constructor is registered in neither
  provider — guarding future migration steps against dead code.
- The restored deprecated `environment_options` fields are superseded by their
  `deployment_input` equivalents; prefer the latter.
- Every `betterado_release_definition` change in this release is proven by a live
  `TF_ACC` acceptance test against a real Azure DevOps org (apply → API
  read-back → idempotency re-plan → destroy).

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
