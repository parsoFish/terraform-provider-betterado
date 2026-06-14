# Changelog

All notable changes to the `betterado` provider are documented here. This
changelog starts at the first public release of the fork. The inherited history
from the upstream `microsoft/azuredevops` provider is preserved in
[`CHANGELOG-upstream.md`](./CHANGELOG-upstream.md).

## Unreleased

## 0.1.2 (2026-06-14)

DOCS:

- Added `docs/guides/` with all five authentication guides updated to use `parsoFish/betterado`
  source and `betterado` provider/resource names (replaces `website/docs/guides/` which still
  referenced `microsoft/azuredevops`).
- Added `## Example Usage` sections to all five betterado-specific data source docs
  (`betterado_release_definition`, `betterado_release_definitions`, `betterado_release_folder`,
  `betterado_release_definition_history`, `betterado_release_definition_revision`).

## 0.1.1 (2026-06-14)

DOCS:

- Registry documentation fixes: removed upstream provider links, added
  betterado-specific resource/data-source tables to provider index, added
  `Release Pipelines` subcategory and descriptions to `betterado_release_definition`,
  `betterado_task_group`, `betterado_release_folder`, and
  `betterado_release_definition_permissions` docs.
- Updated examples in `examples/azdo-based-cicd/` and
  `examples/github-based-cicd-simple/` to use `betterado_*` resources and
  `parsoFish/betterado` provider source.

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

- All inherited `azuredevops_*` resources are available under the `betterado_*`
  prefix.
