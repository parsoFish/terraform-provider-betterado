# betterado API Parity Gap Registry

> Single canonical source of API coverage status for terraform-provider-betterado.
> Generated from 31 per-area gap matrices. Last updated: <date>.

## Vocabulary

| Token | Meaning |
|---|---|
| `covered` | Field/resource is implemented and acceptance-tested |
| `gap-open` | Field/resource is missing and should be implemented |
| `gap-deferred` | Intentionally skipped; reason documented |
| `out-of-scope` | Non-declarative (imperative/runtime-only); will not be implemented |

**Forbidden tokens** (must not appear in any matrix or this registry):
`mapped`, `supported`, `implemented`, `partial`, `missing`, `present`, `gap-resolved`, `✅`, `⚠️`, `🚫`

## Classification legend

| Classification | Meaning |
|---|---|
| `betterado-net-new` | Resource/data source introduced by betterado (not in upstream azuredevops provider) |
| `betterado-extended` | Upstream resource extended with additional fields or functionality |
| `betterado-inherited` | Purely inherited from upstream; betterado adds nothing |

## Registry

<!-- Populated by WI-2 (Release/Pipeline), WI-3 (Infrastructure), WI-4a/4b (Identity/Security/Collaboration/Long-tail) -->

### Release Definition

- **betterado resources/data sources:** `betterado_release_definition`, `data.betterado_release_definition`, `data.betterado_release_definitions`
- **Classification:** `betterado-net-new`
- **gap-open count:** 1 (matches docs/release-definition-gap-matrix.md)
  - `tags` — ADO silently ignores tags on write; field is Computed to avoid perpetual diff; gap-open as write parity is unachievable without API fix
- **gap-deferred count:** 3
  - `comment` — re-evaluation: `complexity-then` (per-save comment rarely used via TF; low ROI)
  - `triggers[pullRequestTrigger]` — re-evaluation: `non-declarative-forever` (no SDK type exists; ADO does not expose PR deployment trigger in v7.1)
  - `workflow_task[].overrideInputs` — re-evaluation: `complexity-then` (runtime override inputs; requires task-level injection)
- **v7.1→v7.2 delta:** v7.2 REST API adds no new fields to `ReleaseDefinition` or `ReleaseDefinitionEnvironment`; no known changes to declarative surface (sourced from learn.microsoft.com; live verification pending)

---

### Release Folder

- **betterado resources/data sources:** `betterado_release_folder`, `data.betterado_release_folder`
- **Classification:** `betterado-net-new`
- **gap-open count:** 0
- **gap-deferred count:** 0
  - All 4 non-covered fields (`CreatedBy`, `CreatedOn`, `LastChangedBy`, `LastChangedDate`) are server-computed read-only metadata — classified `out-of-scope`
- **v7.1→v7.2 delta:** no known changes to declarative surface; `Folder` struct is stable across v7.x

---

### Release Definition Permissions

- **betterado resources/data sources:** `betterado_release_definition_permissions`
- **Classification:** `betterado-net-new`
- **gap-open count:** 0
- **gap-deferred count:** 0
  - All 13 writable permission bits (`ViewReleaseDefinition` … `ManageReleaseSettings`) are fully covered and acceptance-tested; `ManageTaskHubExtension` is `out-of-scope` (extension/org-level only)
- **v7.1→v7.2 delta:** no known changes to the `ReleaseManagement` security namespace permission bits in v7.2; namespace bit table is stable

---

### Task Group

- **betterado resources/data sources:** `betterado_task_group`, `data.betterado_task_group`
- **Classification:** `betterado-net-new`
- **gap-open count:** 0
- **gap-deferred count:** 0
  - All previously-open writable fields (`icon_url`, `input[].visible_rule`, `input[].properties`, `input[].aliases`) were covered by INIT-2026-06-17 WI-1; all server-computed fields (`createdBy`, `modifiedBy`, `demands`, `groups`, etc.) are `out-of-scope`
- **v7.1→v7.2 delta:** no known changes to `TaskGroup` or `TaskGroupCreateParameter` declarative surface in v7.2

---

### Task Agent (agent pools, queues, deployment groups, elastic pools, environments, variable groups)

- **betterado resources/data sources:** `betterado_agent_pool`, `betterado_agent_queue`, `betterado_deployment_group`, `betterado_elastic_pool`, `betterado_environment`, `betterado_environment_resource_kubernetes`, `betterado_variable_group`, `betterado_variable_group_variable`
- **Classification:** `betterado-inherited` (agent pool/queue infrastructure inherited from upstream) + `betterado-extended` (elastic pool, environments, variable groups extended beyond upstream)
- **gap-open count:** 0 (3 gap-open rows in matrix are informational coverage notes on read-back fields, not actionable write gaps)
- **gap-deferred count:** 5
  - `agent_pool.agentCloudId` — re-evaluation: `complexity-then` (no `agent_cloud` resource pre-requisite yet)
  - `agent_queue.authorizePipelines` — re-evaluation: `7.2-api-improvement-feasible` (hard-coded `false`; enhancement WI when demand confirmed)
  - `deployment_group.tags` — re-evaluation: `complexity-then` (string label list; straightforward to add in a follow-up WI)
  - `elastic_pool.osType` — re-evaluation: `complexity-then` (OS type enum; deferred enhancement)
  - `elastic_pool.maxSavedNodeCount` — re-evaluation: `complexity-then` (warm-spare count; deferred enhancement)
- **v7.1→v7.2 delta:** v7.2 adds `osType` to `ElasticPool` API response (already tracked above as gap-deferred `complexity-then`); no other declarative-surface changes known

---

### Approvals and Checks

- **betterado resources/data sources:** `betterado_check_approval`, `betterado_check_branch_control`, `betterado_check_business_hours`, `betterado_check_exclusive_lock`, `betterado_check_required_template`, `betterado_check_rest_api`
- **Classification:** `betterado-inherited`
- **gap-open count:** 0
- **gap-deferred count:** 2
  - `check_approval.approvalType` — re-evaluation: `complexity-then` (ADO group-vs-individual distinction; not needed for standard approval workflows)
  - `check_approval.allowApproversToApproveOwnPipeline` — re-evaluation: `complexity-then` (niche permission flag; safely defaulted by ADO)
- **v7.1→v7.2 delta:** no known changes to the Pipelines Checks API check-type UUIDs or settings schemas in v7.2; `check_required_template` timeout field remains non-configurable

---

### Pipelines Approval

- **betterado resources/data sources:** `betterado_pipeline_approval`, `data.betterado_pipeline_approvals`
- **Classification:** `betterado-inherited`
- **gap-open count:** 0
- **gap-deferred count:** 3
  - `Approval.approvedBy` (IdentityRef) — re-evaluation: `non-declarative-forever` (read-only identity set by ADO upon action; reflects service principal, not human approver)
  - `Approval.createdOn` — re-evaluation: `non-declarative-forever` (read-only server timestamp; no actionable TF use)
  - `Approval.lastModifiedOn` — re-evaluation: `non-declarative-forever` (read-only server timestamp; no actionable TF use)
- **v7.1→v7.2 delta:** no known changes to `Approval` struct fields or `UpdateApprovals` semantics in v7.2; ephemeral-ID model unchanged

---

### Pipelines v2

- **betterado resources/data sources:** `betterado_pipeline`
- **Classification:** `betterado-extended`
- **gap-open count:** 0
- **gap-deferred count:** 0
  - All fields of `Pipeline` and `CreatePipelineParameters` are fully covered; `_links` (HAL hypermedia bag) is `out-of-scope`
- **v7.1→v7.2 delta:** no known changes to declarative surface; `Pipeline.Folder` and `Pipeline.Revision` remain stable; no new writable fields added in v7.2 Pipelines v2 API

---

## Priority backlog

<!-- Populated by WI-5 (Synthesis) -->
