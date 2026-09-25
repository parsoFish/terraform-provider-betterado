# Gap Registry

Tracks implementation coverage for the `betterado` Terraform provider against the ADO API surface.

## Vocabulary

Canonical status tokens used throughout this registry:

- `covered` — implemented and acceptance-tested
- `gap-open` — missing; should be implemented
- `gap-deferred` — intentionally skipped; reason documented
- `out-of-scope` — non-declarative / imperative-only; will not be implemented

## Classification

Three classification values for resources in this provider:

- `betterado-net-new` — resources the fork ships that upstream never released
- `betterado-extended` — upstream resources the fork has extended with additional fields/capabilities
- `betterado-inherited` — resources inherited unchanged from upstream

## Area index

Stub table — WI-2 through WI-4b fill the data cells.

| Area | Classification | Resources / Data Sources | Gap-Open | Gap-Deferred |
|------|----------------|--------------------------|----------|--------------|
| release-definition | | | | |
| release-folder | | | | |
| release-definition-permissions | | | | |
| task-group | | | | |
| taskagent | | | | |
| approvalsandchecks | | | | |
| pipelinesapproval | | | | |
| pipelines-v2 | | | | |
| serviceendpoint | | | | |
| core | | | | |
| build | | | | |
| policy | | | | |
| git | | | | |
| feed | | | | |
| wiki | | | | |
| identity | | | | |
| graph | | | | |
| security | | | | |
| permissions | | | | |
| securityroles | | | | |
| memberentitlementmanagement | | | | |
| notification | | | | |
| servicehook | | | | |
| dashboard | | | | |
| extension | | | | |
| gallery-extensionmanagement | | | | |
| featuremanagement | | | | |
| workitemtracking | | | | |
| workitemtrackingprocess | | | | |
| accounts-profile | | | | |
| test | | | | |

## Release + Pipeline tier

### release-definition

**Classification:** betterado-net-new
**Resources/Data sources:** `betterado_release_definition` (resource), `betterado_release_definition` (data source), `betterado_release_definitions` (data source)
**Gap-open count:** 0
**Gap-deferred count:** 25
  - `createdBy`, `createdOn`, `modifiedBy`, `modifiedOn` (server-computed identity/timestamp metadata) (complexity-then)
  - `_links`, `url` (read-only REST navigation metadata) (non-declarative-forever)
  - `comment` (per-save comment string; rarely used in IaC) (complexity-then)
  - `workflow_task[].overrideInputs` (runtime task override map; deferred) (complexity-then)
  - `deploy_phase[].deploymentInput.artifactsDownloadInput` (fine-grained per-artifact download control; deferred) (complexity-then)
  - remaining deprecated/read-only fields (`retentionPolicy` deprecated, `queueId` legacy, `runOptions` deprecated, `sourceId` deprecated) (non-declarative-forever)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### release-folder

**Classification:** betterado-net-new
**Resources/Data sources:** `betterado_release_folder` (resource), `betterado_release_folder` (data source)
**Gap-open count:** 0
**Gap-deferred count:** 4
  - `CreatedBy`, `CreatedOn`, `LastChangedBy`, `LastChangedDate` (server-computed identity/timestamp metadata; no consumer use case for exposing as Computed) (non-declarative-forever)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### release-definition-permissions

**Classification:** betterado-net-new
**Resources/Data sources:** `betterado_release_definition_permissions` (resource)
**Gap-open count:** 0
**Gap-deferred count:** 0
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### task-group

**Classification:** betterado-net-new
**Resources/Data sources:** `betterado_task_group` (resource), `betterado_task_group` (data source)
**Gap-open count:** 4
  - `icon_url` (historical tracking — was gap-open before WI-1; now covered) (7.2-api-improvement-feasible)
  - `input[].visible_rule` (historical tracking — was gap-open before WI-1; now covered) (7.2-api-improvement-feasible)
  - `input[].properties` (historical tracking — was gap-open before WI-1; now covered) (7.2-api-improvement-feasible)
  - `input[].aliases` (historical tracking — was gap-open before WI-1; now covered) (7.2-api-improvement-feasible)
**Gap-deferred count:** 12
  - `createdBy`, `createdOn`, `modifiedBy`, `modifiedOn` (server-computed metadata; not needed for TF management) (non-declarative-forever)
  - `demands`, `groups`, `satisfies`, `sourceDefinitions`, `dataSourceBindings`, `execution`, `preJobExecution`, `postJobExecution` (all server-computed from task definitions) (non-declarative-forever)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### taskagent

**Classification:** betterado-extended
**Resources/Data sources:** `betterado_agent_pool`, `betterado_agent_queue`, `betterado_deployment_group`, `betterado_elastic_pool`, `betterado_environment`, `betterado_environment_resource_kubernetes`, `betterado_variable_group`, `betterado_variable_group_variable` (resources); `betterado_agent_pool`, `betterado_agent_pools`, `betterado_agent_queue`, `betterado_environment`, `betterado_variable_group`, `betterado_task_group` (data sources)
**Gap-open count:** 0
**Gap-deferred count:** 29
  - `agentCloudId` (agent_pool) — no `agent_cloud` resource exists yet; pre-requisite WI needed (complexity-then)
  - `authorizePipelines` (agent_queue) — hard-coded `false` at create; enhancement WI to expose (complexity-then)
  - `tags` (deployment_group) — list of string labels; deferred enhancement (complexity-then)
  - `osType`, `maxSavedNodeCount` (elastic_pool) — deferred enhancement WIs (complexity-then)
  - `type` as explicit attribute (variable_group) — inferred from `key_vault` presence; cosmetic enhancement deferred (complexity-then)
  - server-assigned timestamps and identity refs across all resources (non-declarative-forever)
  - 12 `ValidateFunc` entries not ported from SDKv2 to framework validators (complexity-then)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### approvalsandchecks

**Classification:** betterado-inherited
**Resources/Data sources:** `azuredevops_check_approval`, `azuredevops_check_branch_control`, `azuredevops_check_business_hours`, `azuredevops_check_exclusive_lock`, `azuredevops_check_required_template`, `azuredevops_check_rest_api` (resources)
**Gap-open count:** 0
**Gap-deferred count:** 0
  - `settings.approvalType` (Approval) — ADO field for group vs individual approval; not exposed (complexity-then)
  - `settings.allowApproversToApproveOwnPipeline` (Approval) — niche ADO field; not exposed (complexity-then)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### pipelinesapproval

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_pipeline_approval` (resource), `betterado_pipeline_approvals` (data source)
**Gap-open count:** 0
**Gap-deferred count:** 0
  - `approvedBy` (read-only identity set by ADO on approval action) (non-declarative-forever)
  - `createdOn`, `lastModifiedOn` (read-only timestamps) (non-declarative-forever)
  - Import via `terraform import` (approval IDs are ephemeral/pipeline-run-bound) (non-declarative-forever)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### pipelines-v2

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_build_definition` (resource), `betterado_pipeline` (resource)
**Gap-open count:** 0
**Gap-deferred count:** 0
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

## Priority backlog

Populated by WI-5.
