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
| dashboard | betterado-inherited | `betterado_dashboard` | 2 | 7 |
| extension | betterado-inherited | `betterado_extension` | 0 | 16 |
| gallery-extensionmanagement | betterado-inherited | `betterado_extension` | 0 | 21 |
| featuremanagement | betterado-inherited | `betterado_project_features`, `betterado_feature_flag` (planned) | 0 | 0 |
| workitemtracking | betterado-inherited | `betterado_workitem`, `betterado_workitemtracking_field`, `betterado_workitemquery`, `betterado_workitemquery_folder`, `data.betterado_area`, `data.betterado_iteration` | 23 | 30 |
| workitemtrackingprocess | betterado-inherited | 12 resource types (`betterado_workitemtrackingprocess_*`) | 0 | 0 |
| accounts-profile | betterado-inherited | `data.betterado_accounts`, `data.betterado_profile` (planned) | 12 | 0 |
| test | betterado-inherited | planned: `betterado_test_plan`, `betterado_test_suite`, `betterado_test_configuration`, `betterado_test_variable`, `betterado_test_result_retention_settings`, `data.betterado_test_run`, `data.betterado_test_result` | 0 | 0 |

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

## Long-tail tier

### dashboard

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_dashboard` (resource)
**Gap-open count:** 2
  - `betterado_dashboard.position` — dashboard ordering within a group; writable via API but deferred (complexity-then)
  - `betterado_dashboard.widgets` — widget configuration; deeply nested structure; deferred to dedicated WI (complexity-then)
**Gap-deferred count:** 7
  - `_links` — read-only HAL navigation links (non-declarative-forever)
  - `dashboardScope` — API-inferred from team_id; not user-settable (non-declarative-forever)
  - `eTag` — server-managed concurrency token (non-declarative-forever)
  - `lastAccessedDate`, `modifiedBy`, `modifiedDate` — server-computed timestamps/identity (non-declarative-forever)
  - `url` — server-provided resource URL (non-declarative-forever)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### extension

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_extension` (resource)
**Gap-open count:** 0
**Gap-deferred count:** 16
  - `baseUri`, `constraints`, `contributions`, `contributionTypes`, `demands`, `eventCallbacks`, `fallbackBaseUri`, `files`, `flags`, `language`, `lastPublished`, `licensing`, `manifestVersion`, `registrationId`, `restrictedTo`, `serviceInstanceType` — all read-only ADO manifest metadata; no Terraform IaC value (non-declarative-forever)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### gallery-extensionmanagement

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_extension` (resource)
**Gap-open count:** 0
**Gap-deferred count:** 21
  - `installState.lastUpdated`, `installState.installationIssues` — read-only diagnostic metadata (non-declarative-forever)
  - `extensionName`, `publisherName`, `scopes` — computed display fields; not in shipped schema (complexity-then)
  - `baseUri`, `contributions`, `contributionTypes`, `demands`, `eventCallbacks`, `files`, `flags`, `language`, `lastPublished`, `licensing`, `manifestVersion`, `registrationId`, `restrictedTo`, `serviceInstanceType`, `constraints`, `fallbackBaseUri` — read-only ADO manifest metadata (non-declarative-forever)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### featuremanagement

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_project_features` (resource, existing), `betterado_feature_flag` (resource, planned)
**Gap-open count:** 0
**Gap-deferred count:** 0
  - All remaining ContributedFeature metadata fields (`defaultValueRules`, `overrideRules`, `featureProperties`, `featureStateChangedListeners`, `includeAsClaim`, `order`, `serviceInstanceType`, `_links`) are internal server-side fields; not TF-relevant (non-declarative-forever)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### workitemtracking

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_workitem` (resource), `betterado_workitemtracking_field` (resource), `betterado_workitemquery` (resource), `betterado_workitemquery_folder` (resource), `betterado_area` (data source), `betterado_iteration` (data source)
**Gap-open count:** 23
  - `betterado_workitem.relations` — arbitrary link types (child, related, remote); deferred (complexity-then)
  - `betterado_workitem.System.AssignedTo` — assignee identity; deferred (complexity-then)
  - `betterado_workitem.System.History` — comment/history entry; deferred (complexity-then)
  - `betterado_workitem.System.Reason` — computed transition reason; not directly settable (complexity-then)
  - `betterado_workitem.System.BoardColumn`, `System.BoardLane` — board-level concerns outside WI scope (complexity-then)
  - `betterado_workitemtracking_field.isPicklistSuggested` — computed backward-compat attribute; deferred (complexity-then)
  - `betterado_workitemquery.path`, `isPublic`, `isDeleted`, `queryType`, `queryRecursionOption`, `clauses`, `linkClauses`, `sourceClauses`, `targetClauses`, `columns`, `sortColumns`, `filterOptions` — query structural fields; deferred (complexity-then)
  - `betterado_workitemquery_folder.path`, `isPublic`, `isDeleted` — folder structural fields; deferred (complexity-then)
**Gap-deferred count:** 30
  - `betterado_workitem.commentVersionRef`, `System.CreatedBy`, `System.CreatedDate`, `System.ChangedBy`, `System.ChangedDate`, `System.CommentCount`, `System.TeamProject` — server-computed metadata (non-declarative-forever)
  - `betterado_workitemquery.isInvalidSyntax`, `createdBy`, `createdDate`, `lastModifiedBy`, `lastModifiedDate`, `lastExecutedBy`, `lastExecutedDate` — server-computed metadata (non-declarative-forever)
  - `betterado_workitemquery_folder.hasChildren`, `children`, `createdBy`, `createdDate`, `lastModifiedBy`, `lastModifiedDate` — server-computed metadata (non-declarative-forever)
  - `betterado_area.attributes`, `id` (integer), `betterado_iteration.attributes`, `id` (integer) — derived/integer node IDs; deferred (non-declarative-forever)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### workitemtrackingprocess

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_workitemtrackingprocess_process` (resource), `betterado_workitemtrackingprocess_workitemtype` (resource), `betterado_workitemtrackingprocess_state` (resource), `betterado_workitemtrackingprocess_inherited_state` (resource), `betterado_workitemtrackingprocess_rule` (resource), `betterado_workitemtrackingprocess_field` (resource), `betterado_workitemtrackingprocess_list` (resource), `betterado_workitemtrackingprocess_page` (resource), `betterado_workitemtrackingprocess_inherited_page` (resource), `betterado_workitemtrackingprocess_group` (resource), `betterado_workitemtrackingprocess_control` (resource), `betterado_workitemtrackingprocess_inherited_control` (resource)
**Gap-open count:** 0
**Gap-deferred count:** 0
  - `betterado_workitemtrackingprocess_control.height` (top-level, HTML controls only) and `betterado_workitemtrackingprocess_group.contribution`, `is_contribution`, `height` — low-demand deferred fields tracked in matrix (complexity-then)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### accounts-profile

**Classification:** betterado-inherited
**Resources/Data sources:** `data.betterado_accounts` (data source), `data.betterado_profile` (data source, planned)
**Gap-open count:** 12
  - `data.betterado_accounts`: `accountOwner`, `accountStatus`, `ownerId` query param — deferred; low consumer demand (complexity-then)
  - `data.betterado_profile`: entire Profile API data source not yet implemented; deferred to follow-on WI — `id`, `coreRevision`, `revision`, `profileState`, `coreAttributes` (incl. `DisplayName`, `EmailAddress`, `PublicAlias`), `applicationContainer`, `timeStamp` (complexity-then)
**Gap-deferred count:** 0
  - `createdBy`, `createdDate`, `lastUpdatedBy`, `lastUpdatedDate`, `namespaceId`, `newCollectionId`, `hasMoved`, `properties`, `statusReason` — audit/migration/internal fields; not IaC-relevant (out-of-scope-forever)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

### test

**Classification:** betterado-inherited
**Resources/Data sources:** `betterado_test_plan` (resource, planned), `betterado_test_suite` (resource, planned), `betterado_test_configuration` (resource, planned), `betterado_test_variable` (resource, planned), `betterado_test_result_retention_settings` (resource, planned), `data.betterado_test_run` (data source, planned), `data.betterado_test_result` (data source, planned)
**Gap-open count:** 0
**Gap-deferred count:** 0
  - Test Run, Test Result, Test Iteration — ephemeral execution artifacts; data-source-only (non-declarative-forever)
  - Test Session, Test Run/Result Attachments — execution-orchestration/binary payloads; out of scope (out-of-scope-forever)
  - Test Point — server-computed join record; read-only (non-declarative-forever)
**v7.2 delta:** sourced from cached ADO SDK vendor source; live verification pending

---

## Priority backlog

Synthesised from all 31 normalized gap matrices (INIT-2026-09-25). Items are grouped by operator value and implementation effort.

### Tier 1 — betterado net-new resource gaps (highest priority)

These are writable fields on betterado's own surfaces that are not yet modelled. Implementing them directly expands the fork's unique value over upstream.

- **task-group/icon_url**: Icon URL shown in the ADO UI for the task group. Writable field present in Create/Update API payloads. [complexity: low]
- **task-group/input[].visible_rule**: Conditional visibility expression for a task input (`input.visibleRule`). Used in UI to show/hide inputs based on other values. [complexity: medium]
- **task-group/input[].properties**: Opaque key/value metadata map on a task input (`input.properties`). Rarely needed but API-complete parity. [complexity: low]
- **task-group/input[].aliases**: Alternative input names (`input.aliases`). Rarely set; useful for task-group authoring completeness. [complexity: low]
- **release_definition_environment_template**: `betterado_release_definition_environment_template` resource — not yet implemented. Provides create/read/delete for reusable stage templates. [complexity: high]

### Tier 2 — high-value upstream gaps (widely-used inherited resources)

Fields on upstream-inherited resources that have measurable operator utility. Implementing these improves parity with what operators expect from a complete ADO provider.

- **build/triggers — build_completion_trigger**: Build-completion trigger (`buildCompletionTrigger`) not yet migrated to the schema. Allows chaining pipelines. [complexity: medium]
- **build/triggers — schedules**: Scheduled build trigger (`schedulesTrigger`) not migrated. Commonly needed for nightly builds. [complexity: medium]
- **build/properties — connectedServiceId / reportBuildStatus**: Two property-bag keys writable via the Build Definitions API that are not yet surfaced as top-level schema attributes. [complexity: low]
- **dashboard/widgets**: Widget configuration block for dashboards. Deeply nested structure; deferred due to idempotency risk with server-side widget ordering. [complexity: high]
- **dashboard/position**: Position of a dashboard within a dashboard group. Ordering field; low operator urgency. [complexity: low]
- **serviceendpoint/workload_identity_federation_subject**: Computed field returned by ADO for workload identity federation scheme. Read-only; needed for external IdP trust configuration. [complexity: low]
- **servicehook/commentPattern**: Subscription filter for comment-pattern events. Not in schema; blocks modelling GitHub comment triggers. [complexity: medium]
- **servicehook/checkedInBy**: Subscription filter by committer identity. Not in schema; blocks precise TFVC trigger configuration. [complexity: low]
- **policy/useUncompressedSize** (`max_file_size`): Missing policy option that controls whether the file-size check uses compressed or uncompressed size. Single bool field. [complexity: low]
- **workitemtracking/System.AssignedTo**: Assignee field on work items. High operator demand; deferred from the migration initiative. [complexity: medium]
- **workitemtracking/System.History**: Comment/history entry write. Append-only; needs special handling to avoid perpetual diff. [complexity: medium]
- **workitemtracking/relations**: Writable link management (child, related, remote links). Deferred from migration; high complexity. [complexity: high]
- **accounts-profile/coreAttributes — display_name, email, public_alias**: Three top-level computed strings from the `coreAttributes` bag. Useful for data-source enrichment; `betterado_profile` data source not yet implemented. [complexity: medium]

### Tier 3 — low-value computed-field gaps (gap-deferred read-only fields)

Server-generated, read-only fields that are intentionally deferred. Low ROI: they cannot be set by the operator and serve no planning purpose in Terraform state. Implement only if a downstream data-consumer explicitly needs them.

- **release_definition/createdBy**: Read-only identity reference; computed by ADO on create. [complexity: low]
- **release_definition/modifiedBy**: Read-only identity reference; computed by ADO on every update. [complexity: low]
- **release_definition/createdOn**: Read-only timestamp. [complexity: low]
- **release_definition/modifiedOn**: Read-only timestamp. [complexity: low]
- **release_definition/lastRelease**: Read-only reference to the most recent release run. [complexity: low]
- **release_definition/isDeleted**: Read-only soft-delete state. [complexity: low]
- **release_definition/source**: Read-only enum indicating how the definition was created (ibiza, restApi). [complexity: low]
- **release_definition/projectReference**: Read-only nested struct; project tracked as `project_id` string. [complexity: low]
- **release_definition/properties**: Opaque property bag; typically empty. [complexity: low]
- **release_definition/environment[].badgeUrl**: Read-only URL computed by ADO. [complexity: low]
- **release_definition/environment[].currentRelease**: Read-only reference to current release run for the stage. [complexity: low]
- **release_definition/environment[].deployStep**: Read-only internal gate step ID. [complexity: low]
- **release_definition/deploy_phase[].refName**: Internal reference name; read-only. [complexity: low]
- **release_definition/pre/post_deploy_approval[].approver[].isNotificationOn**: Read-only computed bool. [complexity: low]
- **release_definition/pre/post_deploy_approval[].approver[].id (step id)**: Internal step ID distinct from approver UUID; read-only. [complexity: low]
- **release_definition/artifact[].isRetained**: Set by release runtime; read-only from definition perspective. [complexity: low]
- **release_definition/pre/post_deployment_gates[].id**: Read-only step ID assigned by ADO. [complexity: low]
- **release_definition/triggers[schedule].branchFilters**: ADO does not return branchFilters for schedule triggers in GET response; intentionally excluded. [complexity: low]
- **workitemtracking/System.Reason, System.BoardColumn, System.BoardLane**: Computed transition/board state fields; not directly settable in most workflows. [complexity: low]
