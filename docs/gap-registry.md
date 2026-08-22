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

### Identity

- **betterado resources/data sources:** `data.betterado_identity_group`, `data.betterado_identity_groups`, `data.betterado_identity_user`
- **Classification:** `betterado-inherited` (identity lookup is upstream; betterado adds no net-new identity resources)
- **gap-open count:** 7
  - `betterado_identity_group.provider_display_name` — display name used for lookup but not returned as a computed output attribute
  - `betterado_identity_group.is_active` — active/inactive flag returned by `ReadIdentity`; useful for health-check automations
  - `betterado_identity_group.is_container` — always `true` for groups; useful as type assertion
  - `betterado_identity_groups[].is_active` — active flag per group item; consistency with single-group data source
  - `betterado_identity_groups[].is_container` — container flag per group item; consistency with single-group data source
  - `betterado_identity_user.provider_display_name` — display name available from API but not returned as output attribute
  - `betterado_identity_user.is_active` — active/inactive flag; useful for detecting disabled accounts
- **gap-deferred count:** 10
  - `betterado_identity_group.member_ids` — re-evaluation: `non-declarative-forever` (consumers should use dedicated membership data sources)
  - `betterado_identity_group.member_of` — re-evaluation: `complexity-then` (complex nested type; low priority)
  - `betterado_identity_group.members` — re-evaluation: `non-declarative-forever` (duplicate of memberIds as descriptors; use group_membership)
  - `betterado_identity_group.custom_display_name` — re-evaluation: `complexity-then` (custom override; rarely set; low value)
  - `betterado_identity_group.master_id` — re-evaluation: `non-declarative-forever` (internal identity master ID; not useful in TF)
  - `betterado_identity_group.meta_type_id` — re-evaluation: `non-declarative-forever` (internal enum; not in public API docs)
  - `betterado_identity_group.properties` — re-evaluation: `complexity-then` (opaque property bag; schema varies per identity)
  - `betterado_identity_group.resource_version` — re-evaluation: `non-declarative-forever` (internal version counter)
  - `betterado_identity_group.social_descriptor` — re-evaluation: `complexity-then` (MSA social descriptor; low relevance for AAD-backed orgs)
  - `betterado_identity_group.unique_user_id` — re-evaluation: `non-declarative-forever` (internal user ID; not useful in TF)
- **v7.1→v7.2 delta:** no known changes to `Identity` struct or `ReadIdentity`/`ReadIdentities`/`ListGroups` API surface in v7.2; identity descriptor model is stable (sourced from learn.microsoft.com; live verification pending)

---

### Graph

- **betterado resources/data sources:** `betterado_group`, `data.betterado_group`, `data.betterado_groups`, `betterado_group_membership`, `data.betterado_group_membership`, `betterado_descriptor`, `betterado_storage_key`, `data.betterado_user`, `data.betterado_users`, `data.betterado_service_principal`
- **Classification:** `betterado-inherited` (group/user/service-principal graph is upstream; betterado adds no net-new graph resources)
- **gap-open count:** 14
  - `data.betterado_group.description` — not exposed in data source read-back; already in flattenGroup
  - `data.betterado_group.mail` (mailAddress) — not exposed in data source read-back; needed for AAD group identification
  - `data.betterado_group.domain` — not exposed in data source read-back; useful for project scope identification
  - `data.betterado_group.principal_name` — needed for cross-reference with other resources
  - `data.betterado_group.subject_kind` — useful for disambiguating group type
  - `data.betterado_group.url` — REST URL for cross-referencing
  - `data.betterado_groups[].subject_kind` — not exposed per-item; consistency with other data sources
  - `data.betterado_service_principal.application_id` — AAD application ID; required for identity federation workflows
  - `data.betterado_service_principal.subject_kind` — consistency; always `"servicePrincipal"`
  - `data.betterado_service_principal.domain` — tenant domain for the service principal
  - `data.betterado_service_principal.principal_name` — UPN-style name useful for role assignments
  - `data.betterado_service_principal.mail_address` — email of service principal
  - `data.betterado_users[].subject_kind` — not exposed per-item; consistency with single-user data source
  - `data.betterado_users[].domain` — domain info already available in GraphUser struct
- **gap-deferred count:** 29
  - `_links` on all types — re-evaluation: `non-declarative-forever` (REST navigation links; no consumer value in TF state)
  - `legacyDescriptor` on all types — re-evaluation: `non-declarative-forever` (internal use only per SDK doc comment)
  - `url` on service_principal/users — re-evaluation: `complexity-then` (low-value read-only REST URL)
  - `directoryAlias` on user/service_principal/users — re-evaluation: `complexity-then` (rarely needed in TF)
  - `isDeletedInOrigin` on user/service_principal — re-evaluation: `non-declarative-forever` (soft-delete state; provider handles implicitly)
  - `metaType` on user/service_principal — re-evaluation: `non-declarative-forever` (internal meta type)
  - `isDeleted` on group resource — re-evaluation: `non-declarative-forever` (used internally to trigger d.SetId(""); not needed in state)
- **v7.1→v7.2 delta:** no known changes to `GraphGroup`, `GraphUser`, `GraphServicePrincipal`, or `GraphMembership` declarative surfaces in v7.2; subject descriptor model and storage key API are stable (sourced from learn.microsoft.com; live verification pending)

---

### Security

- **betterado resources/data sources:** `betterado_security_permissions`, `data.betterado_security_namespace`, `data.betterado_security_namespaces`, `data.betterado_security_namespace_token`
- **Classification:** `betterado-inherited` (security namespace, ACL, and token resources are upstream; betterado adds no net-new security resources)
- **gap-open count:** 0
- **gap-deferred count:** 6
  - `betterado_security_permissions.inherit_permissions` (`AccessControlList.InheritPermissions`) — re-evaluation: `complexity-then` (useful for controlling ACL token inheritance; medium complexity; requires separate ACL write path)
  - `data.betterado_security_namespace_token` — `ReleaseManagement` token template — re-evaluation: `complexity-then` (useful for release pipeline permissions)
  - `data.betterado_security_namespace_token` — `DistributedTask` token template — re-evaluation: `complexity-then` (complex scope/environment tokens)
  - `data.betterado_security_namespace_token` — `Library` token template — re-evaluation: `complexity-then` (covered by dedicated permissions resources)
  - `data.betterado_security_namespace_token` — `MetaTask` token template — re-evaluation: `complexity-then` (task group permissions)
  - `data.betterado_security_namespace` `ReadPermission`/`WritePermission` — re-evaluation: `complexity-then` (low priority; useful as documentation for consumers)
- **v7.1→v7.2 delta:** no known changes to `AccessControlList`, `AccessControlEntry`, or `SecurityNamespaceDescription` schemas in v7.2; security namespace bit tables and ACL API are stable (sourced from learn.microsoft.com; live verification pending)

---

### Permissions

- **betterado resources/data sources:** `betterado_area_permissions`, `betterado_build_definition_permissions`, `betterado_build_folder_permissions`, `betterado_git_permissions`, `betterado_iteration_permissions`, `betterado_library_permissions`, `betterado_project_permissions`, `betterado_serviceendpoint_permissions`, `betterado_servicehook_permissions`, `betterado_tagging_permissions`, `betterado_variable_group_permissions`, `betterado_workitemquery_permissions`, `betterado_workitemtrackingprocess_process_permissions`
- **Classification:** `betterado-inherited` (all 13 permission resources wrap the upstream ADO ACL API; betterado adds no net-new permission namespaces)
- **gap-open count:** 0
- **gap-deferred count:** 13
  - `InheritPermissions` on all 13 resources — re-evaluation: `complexity-then` (the `InheritPermissions` flag controls whether an ACL token inherits from parent tokens; medium complexity; requires `inherit_permissions` attribute in base schema and separate ACL write; deferred to framework migration follow-up)
- **v7.1→v7.2 delta:** no known changes to security namespace permission bit tables or ACL API in v7.2; token format templates for all 13 namespaces are stable (sourced from learn.microsoft.com; live verification pending)

---

### Security Roles

- **betterado resources/data sources:** `betterado_securityrole_assignment`, `data.betterado_securityrole_definitions`
- **Classification:** `betterado-inherited` (security role assignment and definition resources are upstream; betterado adds no net-new role scopes)
- **gap-open count:** 0
- **gap-deferred count:** 1
  - `betterado_securityrole_assignment.access` (assigned vs inherited) — re-evaluation: `complexity-then` (useful to expose as computed field to distinguish explicit vs inherited assignments; medium priority future enhancement)
- **v7.1→v7.2 delta:** the SecurityRoles API remains at `7.1-preview.1`; no promotion to stable in v7.2 known at time of writing; `SecurityRoleAssignment` and `SecurityRoleDefinition` struct fields are stable (sourced from learn.microsoft.com; live verification pending)

---

### Member Entitlement

- **betterado resources/data sources:** `betterado_user_entitlement`, `betterado_group_entitlement`, `betterado_service_principal_entitlement`
- **Classification:** `betterado-inherited` (user, group, and service principal entitlement resources are upstream; betterado adds no net-new entitlement types)
- **gap-open count:** 0
- **gap-deferred count:** 9
  - `betterado_user_entitlement.project_entitlements` — re-evaluation: `complexity-then` (complex nested type; per-project license assignment)
  - `betterado_user_entitlement.extensions` (deprecated) — re-evaluation: `non-declarative-forever` (deprecated by ADO; no action)
  - `betterado_user_entitlement.access_level.msdn_license_type` — re-evaluation: `complexity-then` (MSDN license type enum; low practitioner demand)
  - `betterado_group_entitlement.members` — re-evaluation: `complexity-then` (create-only hint in SDK; complex nested UserEntitlement list)
  - `betterado_group_entitlement.project_entitlements` — re-evaluation: `complexity-then` (complex nested type; per-project license rules)
  - `betterado_group_entitlement.extension_rules` (deprecated) — re-evaluation: `non-declarative-forever` (deprecated by ADO)
  - `betterado_group_entitlement.license_rule.msdn_license_type` — re-evaluation: `complexity-then` (MSDN license type; low demand)
  - `betterado_group_entitlement.group.description` — re-evaluation: `complexity-then` (writable group description; low priority)
  - `betterado_service_principal_entitlement.project_entitlements` — re-evaluation: `complexity-then` (complex nested type; per-project license assignment)
- **v7.1→v7.2 delta:** no known changes to `UserEntitlement`, `GroupEntitlement`, or `ServicePrincipalEntitlement` structs in v7.2; `AccessLevel` and `LicensingSource` enums are stable (sourced from learn.microsoft.com; live verification pending)

---

### Notification

- **betterado resources/data sources:** `betterado_notification_subscription` (planned)
- **Classification:** `betterado-inherited` (notification subscription resource wraps the upstream ADO Notifications API; betterado adds no net-new notification channels)
- **gap-open count:** 6
  - `betterado_notification_subscription.subscriber_id` — required field: UUID of user/group receiving notifications; not yet covered (resource in-flight for INIT-2026-07-01-new-api-notification WI-2/WI-3)
  - `betterado_notification_subscription.channel` — delivery channel block (EmailHtml, EmailPlaintext, User, Group types); not yet covered
  - `betterado_notification_subscription.filter` — ExpressionFilter block with clauses; not yet covered
  - `betterado_notification_subscription.description` — optional subscription description; not yet covered
  - `betterado_notification_subscription.scope_id` — optional project UUID scoping; not yet covered
  - `betterado_notification_subscription.status` — enable/disable the subscription; not yet covered
- **gap-deferred count:** 7
  - `RoleBasedFilter` channel support — re-evaluation: `complexity-then` (requires identity/role resolution; deferred post-MVP)
  - `ActorFilter` channel support — re-evaluation: `complexity-then` (identity-specific subscriptions; deferred post-MVP)
  - `ArtifactFilter` channel support — re-evaluation: `complexity-then` (follow subscriptions on specific artifacts; niche use case)
  - `betterado_notification_subscription_template` data source — re-evaluation: `complexity-then` (read-only API; useful but not required for subscription management)
  - `betterado_notification_subscriber` resource — re-evaluation: `non-declarative-forever` (org-level subscriber delivery preferences; separate concern)
  - `betterado_notification_admin_settings` resource — re-evaluation: `non-declarative-forever` (org-level admin settings; single resource with minimal value)
  - `SubscriptionUserSettings` per-user opt-in/out — re-evaluation: `non-declarative-forever` (user-centric setting; not infrastructure state)
- **v7.1→v7.2 delta:** the Notifications subscription API remains at `7.1-preview.1`; no promotion to stable in v7.2 known; `NotificationSubscription` create/update parameters are stable (sourced from learn.microsoft.com; live verification pending)

---

### Service Hook

- **betterado resources/data sources:** `betterado_servicehook_storage_queue_pipelines`, `betterado_servicehook_webhook_tfs`
- **Classification:** `betterado-inherited` (service hook resources wrap the upstream ADO ServiceHooks API; betterado adds no net-new service hook consumer types)
- **gap-open count:** 2
  - `betterado_servicehook_webhook_tfs` — `git_pull_request_commented.comment_pattern` (`publisherInputs.commentPattern`): API supports filtering by comment body substring; absent from TF schema
  - `betterado_servicehook_webhook_tfs` — `tfvc_checkin.checked_in_by` (`publisherInputs.checkedInBy`): API supports filtering by committer identity; absent from TF schema
- **gap-deferred count:** 1
  - `sas_token` consumer input on all service hook resources — re-evaluation: `complexity-then` (SAS token as alternative auth mechanism for storage queue; requires auth-type discriminator; separate work item)
- **v7.1→v7.2 delta:** no known changes to ServiceHooks `consumerInputs` or `publisherInputs` schemas for `pipelines` or `tfs` publishers in v7.2; service hook subscription envelope is stable (sourced from learn.microsoft.com; live verification pending)

---

### Dashboard

- **betterado resources/data sources:** `betterado_dashboard`
- **Classification:** `betterado-inherited` (dashboard resource wraps the upstream ADO Dashboard API; betterado adds no net-new dashboard features)
- **gap-open count:** 0
- **gap-deferred count:** 2
  - `betterado_dashboard.position` — re-evaluation: `complexity-then` (dashboard ordering in a group; rarely used via IaC; the ADO UI is the natural owner; low demand)
  - `betterado_dashboard.widgets` — re-evaluation: `complexity-then` (complex nested widget schema with server-side ordering; dedicated WI required to avoid idempotency issues; high effort, deferred pending demand)
- **v7.1→v7.2 delta:** no known changes to the `Dashboard` struct or `DashboardScope` enum in v7.2; `Widget` API surface is stable (sourced from learn.microsoft.com; live verification pending)

---

### Extension

- **betterado resources/data sources:** `betterado_extension`, `betterado_extension_install`
- **Classification:** `betterado-inherited` (extension resources wrap the upstream ADO ExtensionManagement API; betterado adds no net-new extension features)
- **gap-open count:** 3
  - `betterado_extension_install.extension_name` — Computed display name; not in framework resource schema (only in legacy `betterado_extension`)
  - `betterado_extension_install.publisher_name` — Computed display name; not in framework resource schema
  - `betterado_extension_install.scope` — Computed OAuth scope list; not in framework resource schema
- **gap-deferred count:** 3
  - `betterado_extension.installState.lastUpdated` — re-evaluation: `complexity-then` (server-computed timestamp; low IaC value; deferred pending demand)
  - `betterado_extension.installState.installationIssues` — re-evaluation: `non-declarative-forever` (read-only diagnostic set by ADO platform; not user-controllable)
  - other `InstallState` flags (VersionCheckError, Warning) — re-evaluation: `non-declarative-forever` (platform error/warning flags; not user-writable)
- **v7.1→v7.2 delta:** no known changes to `InstalledExtension` or `ExtensionStateFlags` in v7.2; ExtensionManagement API remains at `7.1-preview.1` (sourced from learn.microsoft.com; live verification pending)

---

### Gallery / ExtensionManagement

- **betterado resources/data sources:** `betterado_extension`, `betterado_extension_install`, `betterado_marketplace_extension` (gap-deferred data source)
- **Classification:** `betterado-inherited` (Gallery API wraps the Visual Studio Marketplace; betterado adds no net-new gallery features)
- **gap-open count:** 2
  - `betterado_marketplace_extension` data source — not yet in schema; deferred pending operator demand for version lookup
  - `betterado_extension_settings` resource — not yet in schema; ExtensionManagement SDK lacks client methods for extension data collection; deferred
- **gap-deferred count:** 4
  - `betterado_marketplace_extension` — re-evaluation: `complexity-then` (Gallery SDK client not yet wired into provider's AggregatedClient; requires `client.go` + provider registration updates)
  - `betterado_extension_settings` — re-evaluation: `complexity-then` (HTTP client methods absent from SDK; requires raw HTTP or custom client; narrow use case)
  - `PublishedExtension.statistics` — re-evaluation: `complexity-then` (complex nested stat object; low IaC value)
  - `GetInstalledExtensions` (list endpoint) — re-evaluation: `complexity-then` (batch list; useful for import; separate WI)
- **v7.1→v7.2 delta:** no known changes to Gallery `PublishedExtension` or ExtensionManagement `installedextensions` schemas in v7.2; both APIs remain at `7.1-preview.1` (sourced from learn.microsoft.com; live verification pending)

---

### Feature Management

- **betterado resources/data sources:** `betterado_feature_flag` (gap-open — not yet in schema), `betterado_project_features` (existing; covers 5 hardcoded project features)
- **Classification:** `betterado-inherited` (feature management resources wrap the upstream ADO Feature Management API; betterado adds no net-new feature flag surfaces)
- **gap-open count:** 1
  - `betterado_feature_flag` resource — not yet in schema; planned for a follow-on WI; exposes `SetFeatureState`/`SetFeatureStateForScope` for arbitrary feature IDs at host or project scope
- **gap-deferred count:** 2
  - user-scoped feature flags (`userScope="me"`) — re-evaluation: `non-declarative-forever` (personal preferences for the authenticated user; non-idempotent across runs; excluded from `betterado_feature_flag` scope)
  - `data.betterado_feature_flag` data source (feature definition metadata) — re-evaluation: `complexity-then` (read-only; lower priority than the state resource; deferred until resource is live)
- **v7.1→v7.2 delta:** Feature Management API remains at `7.1-preview.1`; no promotion to stable in v7.2 known; `ContributedFeatureState` create/update parameters are stable (sourced from learn.microsoft.com; live verification pending)

---

### Work Item Tracking

- **betterado resources/data sources:** `betterado_workitem`, `betterado_workitemtracking_field`, `betterado_workitemquery`, `betterado_workitemquery_folder`, `data.betterado_area`, `data.betterado_iteration`
- **Classification:** `betterado-inherited` (work item tracking resources wrap the upstream ADO Work Item Tracking API; workitem and workitemquery are upstream; betterado adds no net-new WIT surfaces)
- **gap-open count:** 3
  - `betterado_workitem.System.AssignedTo` — assignee identity field; writable via JSON Patch but not exposed in schema; deferred out of scope for migration initiative
  - `betterado_workitem.System.History` — comment/history HTML entry; writable but deferred out of scope
  - `betterado_workitem` arbitrary link types — child links, related links, remote links beyond parent hierarchy; deferred out of scope
- **gap-deferred count:** 6
  - `betterado_workitemquery.isPublic` — re-evaluation: `complexity-then` (public/private query visibility; deferred out of scope for migration initiative)
  - `betterado_workitemquery.columns` — re-evaluation: `complexity-then` (display column list; deferred)
  - `betterado_workitemquery.sortColumns` — re-evaluation: `complexity-then` (sort column list; deferred)
  - `betterado_workitemquery.filterOptions` — re-evaluation: `complexity-then` (link filter mode enum; deferred)
  - `betterado_workitemquery.clauses` / `linkClauses` / `sourceClauses` / `targetClauses` — re-evaluation: `complexity-then` (structured clause trees; WIQL covers the functional need; deferred)
  - `data.betterado_iteration.attributes` (`startDate`, `finishDate`) — re-evaluation: `7.2-api-improvement-feasible` (sprint dates visible in ADO boards; high practitioner value; straightforward to add as Computed strings)
- **v7.1→v7.2 delta:** no known changes to `WorkItem`, `WorkItemField2`, or `QueryHierarchyItem` structs in v7.2; Work Item Tracking REST API is stable at v7.1 (sourced from learn.microsoft.com; live verification pending)

---

### Work Item Tracking Process

- **betterado resources/data sources:** `betterado_workitemtrackingprocess_process`, `betterado_workitemtrackingprocess_workitemtype`, `betterado_workitemtrackingprocess_state`, `betterado_workitemtrackingprocess_inherited_state`, `betterado_workitemtrackingprocess_rule`, `betterado_workitemtrackingprocess_field`, `betterado_workitemtrackingprocess_list`, `betterado_workitemtrackingprocess_page`, `betterado_workitemtrackingprocess_inherited_page`, `betterado_workitemtrackingprocess_group`, `betterado_workitemtrackingprocess_control`, `betterado_workitemtrackingprocess_inherited_control`, `betterado_workitemtrackingprocess_system_control`
- **Classification:** `betterado-inherited` (work item tracking process resources wrap the upstream ADO Work Item Tracking Process API; betterado adds no net-new process customisation surfaces)
- **gap-open count:** 0
- **gap-deferred count:** 8
  - `betterado_workitemtrackingprocess_workitemtype.behaviors` — re-evaluation: `complexity-then` (separate sub-resource endpoint; dedicated WI required)
  - `betterado_workitemtrackingprocess_field.allowed_values` — re-evaluation: `complexity-then` (list/picklist integration handled via `_process_list`; duplicating here conflicts; deferred)
  - `betterado_workitemtrackingprocess_page.contribution` / `is_contribution` — re-evaluation: `complexity-then` (extension contribution pages; deferred pending demand)
  - `betterado_workitemtrackingprocess_group.contribution` / `is_contribution` / `height` — re-evaluation: `complexity-then` (extension contribution groups; deferred pending demand)
  - `betterado_workitemtrackingprocess_inherited_page.visible` / `order` — re-evaluation: `complexity-then` (uncommon inherited-page override; deferred)
  - `betterado_workitemtrackingprocess_inherited_control.read_only` — re-evaluation: `complexity-then` (uncommon override for inherited controls; deferred)
  - `betterado_workitemtrackingprocess_system_control.metadata` / `watermark` — re-evaluation: `complexity-then` (uncommon system control overrides; deferred)
  - `customization_type` Computed fields (various resources) — re-evaluation: `complexity-then` (low-priority server-assigned enum; could be added as Computed; deferred)
- **v7.1→v7.2 delta:** no known changes to Work Item Tracking Process API structs (`ProcessInfo`, `ProcessWorkItemType`, `WorkItemStateResultModel`, `ProcessRule`, `Page`, `Group`, `Control`) in v7.2; API remains stable (sourced from learn.microsoft.com; live verification pending)

---

### Accounts / Profile

- **betterado resources/data sources:** `data.betterado_accounts`, `data.betterado_profile` (gap-open — not yet in schema)
- **Classification:** `betterado-inherited` (accounts and profile surfaces wrap the upstream ADO Accounts and Profile APIs; betterado adds no net-new account or profile management)
- **gap-open count:** 2
  - `data.betterado_profile` data source — not yet in schema; Profile API fields (`display_name`, `email_address`, `public_alias` from `coreAttributes`) are the primary gap; follow-on WI required
  - `data.betterado_accounts.ownerId` query parameter — not exposed; deferred as `member_id` covers the primary PAT-based use case
- **gap-deferred count:** 3
  - `data.betterado_accounts.account_owner` — re-evaluation: `complexity-then` (owner UUID; low consumer demand identified in initiative)
  - `data.betterado_accounts.account_status` — re-evaluation: `complexity-then` (account enabled/disabled/deleted enum; useful but no consumer use-case identified)
  - `data.betterado_profile.coreAttributes` (all fields) — re-evaluation: `complexity-then` (dynamic map requiring careful modelling; `display_name`, `email_address`, `public_alias` are the priority fields for a follow-on WI)
- **v7.1→v7.2 delta:** no known changes to Accounts API `Account` struct or Profile API `Profile` struct in v7.2; both APIs use `vssps.visualstudio.com` endpoints which are version-independent (sourced from learn.microsoft.com; live verification pending)

---

### Test

- **betterado resources/data sources:** `betterado_test_plan` (gap-open), `betterado_test_suite` (gap-open), `betterado_test_configuration` (gap-open), `betterado_test_variable` (gap-open), `betterado_test_result_retention_settings` (gap-open), `data.betterado_test_run` (gap-open), `data.betterado_test_result` (gap-open)
- **Classification:** `betterado-inherited` (test resources wrap the upstream ADO Test and TestPlan APIs; betterado adds no net-new test infrastructure)
- **gap-open count:** 7
  - `betterado_test_plan` resource — not yet in schema; requires `_apis/testplan/Plans` client (not in vendored legacy SDK); WI-2 target
  - `betterado_test_suite` resource — not yet in schema; requires `_apis/testplan/Suites` client; WI-3 target
  - `betterado_test_configuration` resource — not yet in schema; requires `_apis/testplan/Configurations` client; WI-4 target
  - `betterado_test_variable` resource — not yet in schema; requires `_apis/testplan/Variables` client; WI-4 target
  - `betterado_test_result_retention_settings` resource — not yet in schema; singleton project-level policy; Get+Update only; WI-5 target
  - `data.betterado_test_run` data source — not yet in schema; query test runs by build/pipeline; WI-6 target
  - `data.betterado_test_result` data source — not yet in schema; query test results for a given run; WI-7 target
- **gap-deferred count:** 3
  - Test Point as managed resource — re-evaluation: `non-declarative-forever` (server-computed join record; no Create or Delete; execution-time only)
  - Test Session — re-evaluation: `non-declarative-forever` (exploratory testing execution container; transient; not declarative IaC)
  - Test Run / Result Attachment — re-evaluation: `non-declarative-forever` (binary payload attachments uploaded by test runners; ephemeral execution artifacts)
- **v7.1→v7.2 delta:** legacy `_apis/test` package remains at v7.1; modern `_apis/testplan` package endpoints are not versioned via v7.x query param; no breaking changes known to `TestPlan`, `TestSuite`, `TestConfiguration`, or `TestVariable` API schemas in v7.2 (sourced from learn.microsoft.com; live verification pending)

---

## Priority backlog

<!-- Populated by WI-5 (Synthesis) -->
