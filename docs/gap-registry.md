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

## Infrastructure tier

### serviceendpoint

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_serviceendpoint_aws`, `betterado_serviceendpoint_azurecr`, `betterado_serviceendpoint_azurerm`, `betterado_serviceendpoint_azuredevops`, `betterado_serviceendpoint_azure_service_bus`, `betterado_serviceendpoint_bitbucket`, `betterado_serviceendpoint_black_duck`, `betterado_serviceendpoint_checkmarx_one`, `betterado_serviceendpoint_checkmarx_sast`, `betterado_serviceendpoint_checkmarx_sca`, `betterado_serviceendpoint_dockerregistry`, `betterado_serviceendpoint_dynamic_lifecycle_services`, `betterado_serviceendpoint_externaltfs`, `betterado_serviceendpoint_gcp_terraform`, `betterado_serviceendpoint_generic`, `betterado_serviceendpoint_generic_git`, `betterado_serviceendpoint_generic_v2`, `betterado_serviceendpoint_github`, `betterado_serviceendpoint_github_enterprise`, `betterado_serviceendpoint_gitlab`, `betterado_serviceendpoint_incomingwebhook`, `betterado_serviceendpoint_jenkins`, `betterado_serviceendpoint_jfrog_artifactory_v2`, `betterado_serviceendpoint_jfrog_distribution_v2`, `betterado_serviceendpoint_jfrog_platform_v2`, `betterado_serviceendpoint_jfrog_xray_v2`, `betterado_serviceendpoint_kubernetes`, `betterado_serviceendpoint_maven`, `betterado_serviceendpoint_nexus`, `betterado_serviceendpoint_npm`, `betterado_serviceendpoint_nuget`, `betterado_serviceendpoint_octopus`, `betterado_serviceendpoint_openshift`, `betterado_serviceendpoint_argocd`, `betterado_serviceendpoint_artifactory`, `betterado_serviceendpoint_runpipeline`, `betterado_serviceendpoint_servicefabric`, `betterado_serviceendpoint_snyk`, `betterado_serviceendpoint_sonarcloud`, `betterado_serviceendpoint_sonarqube`, `betterado_serviceendpoint_ssh`, `betterado_serviceendpoint_visualstudiomarketplace` (resources); data sources per-type
**Gap-open count:** 2
**Gap-deferred count:** 5
  - `workload_identity_federation_subject` on `azurerm` and `azurecr` — ADO returns this read-only for WIF scheme; Computed attribute absent (non-declarative-forever)
  - `isShared` flag and `shared_project_ids` on non-generic resources — cross-project sharing not exposed on typed resources (complexity-then)
  - `is_ready` — endpoint readiness not surfaced as Computed attribute (complexity-then)
  - `api_key` `Sensitive: true` absent on Octopus Deploy endpoint — credential leakage risk; schema fix needed (complexity-then)
  - Kubernetes ServiceAccount `namespace` — absent from `service_account` auth block (complexity-then)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### core

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_project`, `betterado_project_features`, `betterado_project_pipeline_settings`, `betterado_project_tags`, `betterado_team`, `betterado_team_administrators`, `betterado_team_members` (resources)
**Gap-open count:** 0
**Gap-deferred count:** 5
  - `betterado_project.abbreviation` — short project abbreviation; low IaC demand (complexity-then)
  - `betterado_project_pipeline_settings.disableClassicBuildPipelineCreation` — policy enforcement field; follow-on WI (complexity-then)
  - `betterado_project_pipeline_settings.disableClassicReleasePipelineCreation` — same category (complexity-then)
  - `betterado_project_pipeline_settings.enforceNoAccessToSecretsFromForks` — security-hardening field; follow-on WI (complexity-then)
  - `betterado_project_pipeline_settings.isCommentRequiredForPullRequest` — lower-priority policy field (complexity-then)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### build

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_build_definition` (resource), `betterado_build_folder` (resource), `betterado_pipeline_authorization` (resource), `betterado_resource_authorization` (resource, deprecated), `data.betterado_build_definition` (data source)
**Gap-open count:** 3
**Gap-deferred count:** 5
  - `variable_groups` — complex int-set type; not migrated to framework schema this iteration (complexity-then)
  - `build_completion_trigger` — complex nested trigger; not migrated (complexity-then)
  - `schedules` — timezone-list trigger block; not migrated (complexity-then)
  - `jobs` (OtherGit only) — large nested block; not migrated (complexity-then)
  - `features` list wrapper — replaced by `skip_first_run` top-level attribute; wrapper omitted (complexity-then)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### policy

**Classification:** betterado-inherited
**Resources/Data sources:** `azuredevops_branch_policy_build_validation`, `azuredevops_branch_policy_min_reviewers`, `azuredevops_branch_policy_auto_reviewers`, `azuredevops_branch_policy_comment_resolution`, `azuredevops_branch_policy_merge_strategy`, `azuredevops_branch_policy_status_check`, `azuredevops_branch_policy_work_item_linking`, `azuredevops_repository_policy_author_email_patterns`, `azuredevops_repository_policy_file_path_patterns`, `azuredevops_repository_policy_case_enforcement`, `azuredevops_repository_policy_reserved_names`, `azuredevops_repository_policy_max_file_size`, `azuredevops_repository_policy_max_path_length`, `azuredevops_repository_policy_check_credentials` (resources)
**Gap-open count:** 2
**Gap-deferred count:** 3
  - `min_reviewers.enforceTeamMemberCount` — niche field; not exposed (complexity-then)
  - `min_reviewers.allowCompletionWithRejectsOrWaitsFromNonRequiredReviewers` — niche edge case; safely defaulted by ADO (complexity-then)
  - `max_file_size.useUncompressedSize` — ADO defaults to false; no user demand identified (complexity-then)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### git

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_git_repository` (resource), `betterado_git_repository_branch` (resource), `betterado_git_repository_file` (resource), `betterado_git_repositories` (data source), `betterado_git_repository` (data source)
**Gap-open count:** 0
**Gap-deferred count:** 4
  - `betterado_git_repository.isInMaintenance` — ADO-internal read-only flag; not user-configurable (non-declarative-forever)
  - `betterado_git_repository_branch.isLocked` — admin-only branch lock; rarely managed declaratively (complexity-then)
  - `data.betterado_git_repositories.isFork` — fork detection in bulk list; low demand (complexity-then)
  - `data.betterado_git_repository.parentRepository` — parent repo for fork detection; low priority for data source (complexity-then)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### feed

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_feed` (resource), `betterado_feed_permission` (resource), `betterado_feed_retention_policy` (resource), `data.betterado_feed` (data source)
**Gap-open count:** 0
**Gap-deferred count:** 11
  - `betterado_feed.upstream_enabled` — single bool writable via FeedUpdate; high value (complexity-then)
  - `betterado_feed.upstream_sources` — complex nested block; dedicated WI warranted (complexity-then)
  - `betterado_feed.description` — string; low effort follow-on (complexity-then)
  - `betterado_feed.hide_deleted_package_versions` — single bool; low risk (complexity-then)
  - `betterado_feed.badges_enabled` — single bool; low risk (complexity-then)
  - `betterado_feed.default_view_id` — UUID; needs view lookup support (complexity-then)
  - `data.betterado_feed`: 5 computed-attribute gaps mirroring resource writable gaps above (upstream_enabled, upstream_sources, badges_enabled, default_view_id, description) (complexity-then)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### wiki

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_wiki` (resource), `betterado_wiki_page` (resource)
**Gap-open count:** 0
**Gap-deferred count:** 2
  - `betterado_wiki.properties` — freeform key/value map; low IaC adoption; open tracking issue (complexity-then)
  - `betterado_wiki_page.order` — sibling page ordering; causes plan drift on every read; limited IaC value (complexity-then)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

## Identity / Security tier

### identity

**Classification:** betterado-inherited
**Resources/Data sources:** `data.betterado_identity_group`, `data.betterado_identity_groups`, `data.betterado_identity_user`
**Gap-open count:** 0
**Gap-deferred count:** 0
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### graph

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_group` (resource + data source), `betterado_descriptor`, `betterado_group_membership`, `data.betterado_groups`, `betterado_service_principal`, `betterado_storage_key`, `betterado_user`, `data.betterado_users`
**Gap-open count:** 0
**Gap-deferred count:** 0
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### security

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_security_permissions` (resource), `data.betterado_security_namespace`, `data.betterado_security_namespace_token`, `data.betterado_security_namespaces`
**Gap-open count:** 0
**Gap-deferred count:** 14
  - `AccessControlList.IncludeExtendedInfo` — read-only query flag; not a schema field (non-declarative-forever)
  - `DataspaceCategory` — read-only namespace metadata; no TF consumer value (non-declarative-forever)
  - `ElementLength` — read-only internal separator config (non-declarative-forever)
  - `ExtensionType` — read-only plugin extension type string (non-declarative-forever)
  - `IsRemotable` — read-only boolean; low IaC value (non-declarative-forever)
  - `ReadPermission`, `WritePermission` — read-only bitmasks (non-declarative-forever)
  - `SeparatorValue`, `StructureValue`, `SystemBitMask`, `UseTokenTranslator` — read-only internal config (non-declarative-forever)
  - `Actions[].NamespaceId` — read-only backlink; redundant (non-declarative-forever)
  - 3 additional read-only internal bookkeeping fields (non-declarative-forever)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### permissions

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_area_permissions`, `betterado_build_definition_permissions`, `betterado_build_folder_permissions`, `betterado_git_permissions`, `betterado_iteration_permissions`, `betterado_library_permissions`, `betterado_project_permissions`, `betterado_serviceendpoint_permissions`, `betterado_servicehook_permissions`, `betterado_tagging_permissions`, `betterado_variable_group_permissions`, `betterado_workitemquery_permissions`, `betterado_workitemtrackingprocess_process_permissions`
**Gap-open count:** 0
**Gap-deferred count:** 0
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### securityroles

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_securityrole_assignment` (resource), `data.betterado_securityrole_definitions`
**Gap-open count:** 0
**Gap-deferred count:** 11
  - `identity.displayName`, `identity.uniqueName` — read-only identity display fields (non-declarative-forever)
  - `role.displayName`, `role.identifier`, `role.description` — read-only role metadata (non-declarative-forever)
  - `role.allowPermissions`, `role.denyPermissions` — read-only bitmasks (non-declarative-forever)
  - `assignment.access`, `assignment.accessDisplayName` — read-only assignment metadata; negligible IaC value (non-declarative-forever)
  - 2 summary/resolved rows
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### memberentitlementmanagement

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_user_entitlement` (resource), `betterado_group_entitlement` (resource), `betterado_service_principal_entitlement` (resource)
**Gap-open count:** 3
  - `betterado_user_entitlement.user` — nested GraphUser sub-object not yet exposed (complexity-then)
  - `betterado_group_entitlement.group` — nested GraphGroup sub-object not yet exposed (complexity-then)
  - `betterado_service_principal_entitlement.servicePrincipal` — nested GraphServicePrincipal sub-object not yet exposed (complexity-then)
**Gap-deferred count:** 58
  - `dateCreated`, `lastAccessedDate` — read-only timestamps (non-declarative-forever)
  - `groupAssignments` — read-only group membership aggregate (non-declarative-forever)
  - `projectEntitlements` — complex nested project-level entitlements; deferred for dedicated WI (complexity-then)
  - `extensions` — deprecated extension licenses (non-declarative-forever)
  - `accessLevel.assignmentSource`, `accessLevel.licenseDisplayName`, `accessLevel.status`, `accessLevel.statusMessage` — read-only access level metadata (non-declarative-forever)
  - `accessLevel.msdnLicenseType` — legacy MSDN license; deferred (complexity-then)
  - 52 additional read-only GraphUser/GraphGroup/GraphServicePrincipal sub-fields (displayName, url, legacyDescriptor, subjectKind, domain, mailAddress, directoryAlias, isDeletedInOrigin, metaType, _links, applicationId, etc.) across all three entitlement types (non-declarative-forever)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### notification

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_notification_subscription` (planned; WI-2 of initiative INIT-2026-07-01-new-api-notification)
**Gap-open count:** 0
**Gap-deferred count:** 0
  - 15 fields marked `**implement**` in matrix pending WI-2 of parent initiative (not yet gap-open; tracked in matrix)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### servicehook

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_servicehook_storage_queue_pipelines` (resource), `betterado_servicehook_webhook_tfs` (resource)
**Gap-open count:** 2
  - `betterado_servicehook_webhook_tfs.commentPattern` — TFVC comment filter; not in schema (complexity-then)
  - `betterado_servicehook_storage_queue_pipelines.checkedInBy` — TFVC check-in identity filter; not in schema (complexity-then)
**Gap-deferred count:** 1
  - `betterado_servicehook_storage_queue_pipelines.sasToken` — SAS token auth; not in schema; write-only secret (complexity-then)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

## Priority backlog

Populated by WI-5.
