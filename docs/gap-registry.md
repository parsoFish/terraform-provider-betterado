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

### Service Endpoint

- **betterado resources/data sources:** `betterado_serviceendpoint_aws`, `betterado_serviceendpoint_azurecr`, `betterado_serviceendpoint_azurerm`, `betterado_serviceendpoint_kubernetes`, `betterado_serviceendpoint_github`, `betterado_serviceendpoint_generic_v2`, and 36 additional typed endpoint resources; `data.betterado_serviceendpoint_azurecr`, `data.betterado_serviceendpoint_azurerm`, `data.betterado_serviceendpoint_bitbucket`, `data.betterado_serviceendpoint_dockerregistry`, `data.betterado_serviceendpoint_generic_v2`, `data.betterado_serviceendpoint_github`, `data.betterado_serviceendpoint_npm`, `data.betterado_serviceendpoint_sonarcloud`
- **Classification:** `betterado-inherited` (all 42 endpoint type resources inherit the upstream ADO service endpoint envelope; betterado adds no net-new endpoint types, but `betterado_serviceendpoint_generic_v2` is `betterado-extended` for cross-project `shared_project_ids`)
- **gap-open count:** 2
  - `betterado_serviceendpoint_kubernetes` — `data.namespace` for `service_account` auth type: writable ADO field absent from TF schema; ADO defaults to Kubernetes `default` namespace
  - `betterado_serviceendpoint_octopus_deploy.api_key` — schema quality gap: field lacks `Sensitive: true`; credential leakage risk in plan output
- **gap-deferred count:** 55 (credential-rotation-only fields across all endpoint resources) + 4 (read-only computed: `workload_identity_federation_subject` ×2, `is_ready`, `is_shared`)
  - credential-rotation-only fields — re-evaluation: `non-declarative-forever` (ADO never returns secret values in GET responses; state-only tracking is the correct pattern)
  - `workload_identity_federation_subject` (`azurerm` + `azurecr`) — re-evaluation: `7.2-api-improvement-feasible` (read-only computed; useful for WIF trust config in external IdP)
  - `is_ready` — re-evaluation: `non-declarative-forever` (transient readiness state; not a TF lifecycle concern)
  - `shared_project_ids` on typed resources — re-evaluation: `complexity-then` (only `generic_v2` exposes cross-project sharing today)
- **v7.1→v7.2 delta:** no known changes to `ServiceEndpoint` struct fields in v7.2; `serviceEndpointProjectReferences` and `authorization.parameters` models are stable (sourced from learn.microsoft.com; live verification pending)

---

### Core

- **betterado resources/data sources:** `betterado_project`, `betterado_project_features`, `betterado_project_pipeline_settings`, `betterado_project_tags`, `betterado_team`, `betterado_team_administrators`, `betterado_team_members`
- **Classification:** `betterado-inherited` (project, team, and feature management resources are all inherited from the upstream ADO provider; betterado adds no net-new core resources)
- **gap-open count:** 5
  - `betterado_project.abbreviation` — writable short abbreviation string on `TeamProject`; absent from TF schema; low IaC priority
  - `betterado_project_pipeline_settings.disableClassicBuildPipelineCreation` — policy enforcement field added in later ADO versions
  - `betterado_project_pipeline_settings.disableClassicReleasePipelineCreation` — same category as above
  - `betterado_project_pipeline_settings.enforceNoAccessToSecretsFromForks` — security-hardening field for fork PR pipelines
  - `betterado_project_pipeline_settings.isCommentRequiredForPullRequest` — PR policy setting at project scope
- **gap-deferred count:** 0
  - All 5 gap-open items are tracked open (not deferred); no explicitly deferred writable fields exist in core resources
- **v7.1→v7.2 delta:** `disableClassicBuildPipelineCreation` and `disableClassicReleasePipelineCreation` fields were added to `PipelineGeneralSettings` in v7.1+; v7.2 adds no further changes to `TeamProject`, `WebApiTeam`, or `ContributedFeatureState` schemas known at time of writing

---

### Build

- **betterado resources/data sources:** `betterado_build_definition`, `betterado_build_folder`, `betterado_pipeline_authorization`, `betterado_resource_authorization`, `data.betterado_build_definition`
- **Classification:** `betterado-inherited` (build definitions and folders inherit from upstream ADO build API) + `betterado-extended` (`betterado_build_definition` is extended with `skip_first_run`, `job_authorization_scope`, and framework-native CI/PR trigger blocks that exceed upstream schema coverage)
- **gap-open count:** 8
  - `betterado_build_definition.badge_enabled` — writable bool; absent from both SDKv2 and framework schemas
  - `betterado_build_definition.build_number_format` — writable string; absent from both schemas
  - `betterado_build_definition.description` — writable string; absent from both schemas
  - `betterado_build_definition.job_cancel_timeout_in_minutes` — writable int; absent from both schemas
  - `betterado_build_definition.job_timeout_in_minutes` — writable int; absent from both schemas
  - `betterado_build_definition.tags` — writable string list; absent from both schemas
  - `betterado_build_definition.repository.clean` — writable bool; absent from both schemas
  - `betterado_build_definition.repository.checkout_submodules` — writable bool; absent from both schemas
- **gap-deferred count:** 4
  - `variable_groups` on `betterado_build_definition` — re-evaluation: `complexity-then` (complex int-set type with no direct framework parallel in current vendored version)
  - `build_completion_trigger` — re-evaluation: `complexity-then` (complex nested structure; low usage)
  - `schedules` — re-evaluation: `complexity-then` (complex nested structure with timezone enumeration)
  - `jobs` (OtherGit only) — re-evaluation: `complexity-then` (large nested block; OtherGit pipeline type only)
- **v7.1→v7.2 delta:** no new writable fields added to `BuildDefinition` or `BuildRepository` in v7.2; `AgentPoolQueue` and trigger structs are stable (sourced from learn.microsoft.com; live verification pending)

---

### Policy

- **betterado resources/data sources:** `azuredevops_branch_policy_build_validation`, `azuredevops_branch_policy_min_reviewers`, `azuredevops_branch_policy_auto_reviewers`, `azuredevops_branch_policy_comment_resolution`, `azuredevops_branch_policy_merge_strategy`, `azuredevops_branch_policy_status_check`, `azuredevops_branch_policy_work_item_linking`, `azuredevops_repository_policy_author_email_patterns`, `azuredevops_repository_policy_file_path_patterns`, `azuredevops_repository_policy_case_enforcement`, `azuredevops_repository_policy_reserved_names`, `azuredevops_repository_policy_max_file_size`, `azuredevops_repository_policy_max_path_length`, `azuredevops_repository_policy_check_credentials`
- **Classification:** `betterado-inherited` (all branch and repository policy resources wrap the upstream ADO Policy API `PolicyConfiguration` envelope; betterado adds no net-new policy types)
- **gap-open count:** 3
  - `azuredevops_branch_policy_min_reviewers.enforceTeamMemberCount` — niche ADO settings field; not surfaced in TF schema
  - `azuredevops_branch_policy_min_reviewers.allowCompletionWithRejectsOrWaitsFromNonRequiredReviewers` — niche edge-case field; safely defaulted by ADO
  - `azuredevops_repository_policy_max_file_size.useUncompressedSize` — ADO defaults to `false`; no user demand identified
- **gap-deferred count:** 0
  - All gap-open items above are tracked open; no explicitly deferred writable fields
  - `azuredevops_repository_policy_check_credentials` is deprecated by ADO (feature withdrawn); resource retained for state compatibility only (`non-declarative-forever`)
- **v7.1→v7.2 delta:** no known changes to branch or repository `PolicyConfiguration` schemas in v7.2; policy type UUIDs and `settings` structures are stable across v7.x (sourced from learn.microsoft.com; live verification pending)

---

### Git

- **betterado resources/data sources:** `betterado_git_repository`, `betterado_git_repository_branch`, `betterado_git_repository_file`, `data.betterado_git_repository`, `data.betterado_git_repositories`, `data.betterado_git_repository_file`
- **Classification:** `betterado-inherited` (git repository, branch, and file resources are inherited from the upstream ADO Git API; betterado adds no net-new git resources)
- **gap-open count:** 1
  - `betterado_git_repository_branch.isLocked` — writable branch lock state; branch locking is an admin operation rarely managed via IaC; ADO branch policies are preferred
- **gap-deferred count:** 6
  - `betterado_git_repository.isInMaintenance` — re-evaluation: `non-declarative-forever` (maintenance mode is set by ADO internally, not via TF config)
  - `data.betterado_git_repositories.isFork` — re-evaluation: `complexity-then` (not yet surfaced in list element schema; low demand for fork detection in bulk list context)
  - `data.betterado_git_repository.parentRepository` — re-evaluation: `complexity-then` (parent repo UUID useful for fork detection workflows; low priority for data source)
  - `data.betterado_git_repository_file.committer.*` / `author.*` — re-evaluation: `complexity-then` (committer identity useful for auditing; low complexity to add in a future WI)
  - `data.betterado_git_repository_file.commitId` — re-evaluation: `complexity-then` (exposing SHA allows downstream pinning; useful for reproducibility)
- **v7.1→v7.2 delta:** no known changes to `GitRepository`, `GitRef`, or `GitItem` declarative surfaces in v7.2; `WriteOnly: true` on `initialization.password` remains correct (sourced from learn.microsoft.com; live verification pending)

---

### Feed

- **betterado resources/data sources:** `betterado_feed`, `betterado_feed_permission`, `betterado_feed_retention_policy`, `data.betterado_feed`
- **Classification:** `betterado-inherited` (Artifacts Feed resources are inherited from the upstream ADO Artifacts API) + `betterado-extended` (`betterado_feed_retention_policy` and `betterado_feed_permission` extend the base feed with dedicated permission and retention management that the upstream provider lacks)
- **gap-open count:** 5
  - `betterado_feed.upstream_enabled` — writable bool controlling upstream package proxy; single-field addition; high value
  - `betterado_feed.upstream_sources` — writable `UpstreamSource` list; key ADO feed feature; complex nested type warrants dedicated WI
  - `betterado_feed.description` — writable string (≤255 chars); straightforward
  - `betterado_feed.hide_deleted_package_versions` — writable bool; single-field addition
  - `betterado_feed.badges_enabled` — writable bool enabling package badge generation
- **gap-deferred count:** 1
  - `betterado_feed_retention_policy.ageLimitInDays` — re-evaluation: `non-declarative-forever` (deprecated by ADO; SDK marks field as not honoured by retention)
- **v7.1→v7.2 delta:** no known changes to `Feed`, `FeedPermission`, or `FeedRetentionPolicy` structs in v7.2; `FeedUpdate` writable fields are stable (sourced from learn.microsoft.com; live verification pending)

---

### Wiki

- **betterado resources/data sources:** `betterado_wiki`, `betterado_wiki_page`
- **Classification:** `betterado-inherited` (Wiki resources are inherited from the upstream ADO Wiki REST API; betterado adds no net-new wiki resources)
- **gap-open count:** 2
  - `betterado_wiki.properties` — writable `*map[string]string` of arbitrary key/value metadata on `WikiV2`; freeform map; low IaC adoption; schema complexity outweighs benefit
  - `betterado_wiki_page.order` — writable `*int` controlling sibling page sort position; causes perpetual drift on every plan due to ADO reorder-on-read behaviour; limited declarative IaC value
- **gap-deferred count:** 0
  - Both gap-open items are tracked open; `betterado_wiki.type` lacks `ForceNew` (low priority deferral, not a writable field gap)
- **v7.1→v7.2 delta:** no known changes to `WikiV2` or `WikiPage` declarative surfaces in v7.2; `WikiPageCreateOrUpdateParameters` and `GitVersionDescriptor` remain stable (sourced from learn.microsoft.com; live verification pending)

---

## Priority backlog

<!-- Populated by WI-5 (Synthesis) -->
