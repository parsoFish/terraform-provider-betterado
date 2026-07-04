# Changelog

All notable changes to the `betterado` provider are documented here. This
changelog starts at the first public release of the fork. The inherited history
from the upstream `microsoft/azuredevops` provider is preserved in
[`CHANGELOG-upstream.md`](./CHANGELOG-upstream.md).

## [Unreleased]

### Added

- New data source `betterado_accounts` — list ADO accounts accessible to the authenticated user.
- New data source `betterado_profile` — look up a user's ADO profile (display name, email, avatar URL).

### FEATURES

- **`betterado_security_permissions` migrated to terraform-plugin-framework.**
  The resource now uses the framework implementation served through the mux
  provider. Schema is unchanged (`namespace_id`, `token`, `principal`,
  `permissions`, `replace`). Supports full ACL management via the Azure DevOps
  Security REST API with idempotent apply. Verified by live acceptance test
  `TestAccSecurityPermissionsFramework`.

- **`betterado_security_namespace` data source migrated to terraform-plugin-framework.**
  Reads a security namespace by `id` or `name`; returns `display_name` and an
  `actions` list with `name`, `display_name`, and `bit` per permission action.
  Served through the mux provider. Verified by live acceptance test
  `TestAccDataSecurityNamespaceFramework`.

- **`betterado_security_namespace_token` data source migrated to terraform-plugin-framework.**
  Generates a scoped security token for a given namespace and set of resource
  identifiers (`project_id`, `repository_id`, etc.); also supports
  `return_identifier_info` mode to discover required and optional identifiers.
  Served through the mux provider. Verified by live acceptance test
  `TestAccDataSecurityNamespaceFramework`.

- **`betterado_security_namespaces` data source migrated to terraform-plugin-framework.**
  Lists all available security namespaces in the Azure DevOps organisation;
  exposes a `namespaces` set with full namespace metadata per entry.
  Served through the mux provider.

- **`betterado_securityrole_assignment` migrated to terraform-plugin-framework.**
  The resource now uses the framework implementation served through the mux
  provider. Schema is unchanged (`scope`, `resource_id`, `identity_id`,
  `role_name`). Handles ADO inherited-access after delete. Verified by live
  acceptance test `TestAccSecurityroleAssignmentFramework`.

- **`betterado_securityrole_definitions` data source migrated to terraform-plugin-framework.**
  Reads all security role definitions for a given `scope`; exposes a
  `definitions` set with `name`, `display_name`, `description`, `identifier`,
  `scope`, `allow_permissions`, and `deny_permissions` per entry.
  Served through the mux provider. Verified by live acceptance test
  `TestAccDataSecurityroleDefinitionsFramework`.

- **`betterado_project_permissions` migrated to terraform-plugin-framework.**
  The resource now uses the framework implementation served through the mux
  provider. Schema is unchanged (`project_id`, `principal`, `permissions`,
  `replace`). Verified by live acceptance test
  `TestAccProjectPermissionsFramework`.

- **`betterado_area_permissions` migrated to terraform-plugin-framework.**
  Manages area (classification node) ACL permissions. Schema is unchanged
  (`project_id`, `token`, `principal`, `permissions`, `replace`). Served through
  the mux provider. Verified by live acceptance test
  `TestAccAreaPermissionsFramework`.

- **`betterado_build_definition_permissions` migrated to terraform-plugin-framework.**
  Manages build definition ACL permissions. Schema is unchanged
  (`project_id`, `build_definition_id`, `principal`, `permissions`, `replace`).
  Served through the mux provider. Verified by live acceptance test
  `TestAccBuildDefinitionPermissionsFramework`.

- **`betterado_build_folder_permissions` migrated to terraform-plugin-framework.**
  Manages build folder ACL permissions. Schema is unchanged
  (`project_id`, `path`, `principal`, `permissions`, `replace`). Served through
  the mux provider. Verified by live acceptance test
  `TestAccBuildFolderPermissionsFramework`.

- **`betterado_git_permissions` migrated to terraform-plugin-framework.**
  Manages Git repository ACL permissions. Schema is unchanged
  (`project_id`, `repository_id`, `branch_name`, `principal`, `permissions`,
  `replace`). Served through the mux provider. Verified by live acceptance test
  `TestAccGitPermissionsFramework`.

- **`betterado_iteration_permissions` migrated to terraform-plugin-framework.**
  Manages iteration (classification node) ACL permissions. Schema is unchanged
  (`project_id`, `token`, `principal`, `permissions`, `replace`). Served through
  the mux provider. Verified by live acceptance test
  `TestAccIterationPermissionsFramework`.

- **`betterado_library_permissions` migrated to terraform-plugin-framework.**
  Manages library ACL permissions. Schema is unchanged
  (`project_id`, `principal`, `permissions`, `replace`). Served through the mux
  provider. Verified by live acceptance test `TestAccLibraryPermissionsFramework`.

- **`betterado_serviceendpoint_permissions` migrated to terraform-plugin-framework.**
  Manages service endpoint ACL permissions. Schema is unchanged
  (`project_id`, `service_endpoint_id`, `principal`, `permissions`, `replace`).
  Served through the mux provider. Verified by live acceptance test
  `TestAccServiceEndpointPermissionsFramework`.

- **`betterado_servicehook_permissions` migrated to terraform-plugin-framework.**
  Manages service hook ACL permissions. Schema is unchanged
  (`project_id`, `principal`, `permissions`, `replace`). Served through the mux
  provider. Verified by live acceptance test `TestAccServiceHookPermissionsFramework`.

- **`betterado_tagging_permissions` migrated to terraform-plugin-framework.**
  Manages tagging ACL permissions. Schema is unchanged
  (`project_id`, `principal`, `permissions`, `replace`). Served through the mux
  provider. Verified by live acceptance test `TestAccTaggingPermissionsFramework`.

- **`betterado_variable_group_permissions` migrated to terraform-plugin-framework.**
  Manages variable group ACL permissions. Schema is unchanged
  (`project_id`, `variable_group_id`, `principal`, `permissions`, `replace`).
  Served through the mux provider. Verified by live acceptance test
  `TestAccVariableGroupPermissionsFramework`.

- **`betterado_workitemquery_permissions` migrated to terraform-plugin-framework.**
  Manages work item query ACL permissions. Schema is unchanged
  (`project_id`, `path`, `principal`, `permissions`, `replace`). Served through
  the mux provider. Verified by live acceptance test
  `TestAccWorkItemQueryPermissionsFramework`.

- **`betterado_workitemtrackingprocess_process_permissions` migrated to terraform-plugin-framework.**
  Manages work item tracking process ACL permissions. Schema is unchanged
  (`principal`, `permissions`, `replace`). Served through the mux provider.
  Verified by live acceptance test
  `TestAccWorkItemTrackingProcessPermissionsFramework`.

### FEATURES

- **7 branch policy resources migrated to terraform-plugin-framework** (`betterado_branch_policy_auto_reviewers`, `betterado_branch_policy_build_validation`, `betterado_branch_policy_comment_resolution`, `betterado_branch_policy_merge_types`, `betterado_branch_policy_min_reviewers`, `betterado_branch_policy_status_check`, `betterado_branch_policy_work_item_linking`). All resources now use the framework implementation served through the mux provider. Schema and CRUD semantics are unchanged; `settings` and `scope` continue to use block syntax. Verified by live acceptance tests `TestAccBranchPolicy*`.

- **7 repository policy resources migrated to terraform-plugin-framework** (`betterado_repository_policy_author_email_pattern`, `betterado_repository_policy_case_enforcement`, `betterado_repository_policy_check_credentials`, `betterado_repository_policy_file_path_pattern`, `betterado_repository_policy_max_file_size`, `betterado_repository_policy_max_path_length`, `betterado_repository_policy_reserved_names`). All resources now use the framework implementation served through the mux provider. Schema is unchanged; `repository_ids` is a flat list attribute. Verified by live acceptance tests `TestAccRepositoryPolicy*`.

- **6 approvalsandchecks resources migrated to terraform-plugin-framework** (`betterado_check_approval`, `betterado_check_branch_control`, `betterado_check_business_hours`, `betterado_check_exclusive_lock`, `betterado_check_required_template`, `betterado_check_rest_api`). All resources now use the framework implementation served through the mux provider. Schema and CRUD semantics are unchanged; resources target pipeline environments and other resource types via `target_resource_id`/`target_resource_type`. Verified by live acceptance tests `TestAccCheck*`.

- Added `docs/policy-gap-matrix.md` and `docs/approvalsandchecks-gap-matrix.md` documenting parity between the betterado provider and the upstream microsoft/azuredevops provider for policy and checks resources.

### FEATURES

- **New resource `betterado_pipeline`** — manages a Pipelines v2 pipeline
  (`_apis/pipelines`) in Azure DevOps via terraform-plugin-framework. Supports
  Create, Read, Update (name/folder via PATCH), and Delete. Schema: `project_id`
  (required, ForceNew), `name` (required), `folder` (optional/computed, default `\`),
  `configuration_type` (optional/computed, ForceNew, default `yaml`, allowed values:
  `yaml`/`designerJson`/`justInTime`, validated at plan time via `stringvalidator.OneOf`),
  `id` (computed), `revision` (computed), `url` (computed). Coexists with
  `betterado_build_definition` — see `docs/pipelines-v2-gap-matrix.md` for the full
  overlap analysis.
- **New data source `betterado_pipeline`** — reads an existing Azure Pipelines v2
  pipeline definition by `pipeline_id`. Returns `name`, `folder`,
  `configuration_type`, `revision`, and `url`.
- **New data source `betterado_pipeline_run`** — reads a pipeline run's status
  and result by `pipeline_id` and `run_id`. Returns `state`, `result`,
  `created_date`, `finished_date`, and `url`.

### FEATURES

- **`betterado_notification_subscription` resource and data source.** Manages
  Azure DevOps notification subscriptions via the ADO Notification API v7.1.
  The resource supports creating, reading, updating, and deleting subscriptions
  with `subscription_type` (event type ID), `subscriber_id` (identity), `channel_type`,
  `channel_address`, `filter_type`, and `filter_criteria` attributes scoped to a
  `project_id`. A companion data source reads an existing subscription by ID
  and exposes all schema attributes as computed outputs. Registered in the
  framework provider only (not SDKv2 `provider.go`). Verified by live acceptance
  test `TestAccNotificationSubscription_basic`.

### FEATURES

- **`betterado_area` and `betterado_iteration` data sources migrated to terraform-plugin-framework.** Both data sources are now served through the mux provider using framework implementations (`data_area_framework.go`, `data_iteration_framework.go`). The SDKv2 implementations (`data_area.go`, `data_iteration.go`) have been removed. Schema is unchanged: `project_id`, `path`, `fetch_children`, `name`, `has_children`, `children`. Verified by live acceptance tests `TestAccAreaDataSource_Read` and `TestAccIterationDataSource_Read`.
- **`betterado_workitemquery` and `betterado_workitemquery_folder` migrated to terraform-plugin-framework.** Both resources are now served through the mux provider using framework implementations (`resource_workitemquery_framework.go`, `resource_workitemquery_folder_framework.go`). The SDKv2 implementations have been removed. All schema attributes are preserved including the `ExactlyOneOf(parent_id, area)` constraint and `ForceNew` fields. Verified by live acceptance tests `TestAccWorkItemQuery_UnderArea` and `TestAccWorkItemQueryFolder_UnderArea`.
- **`betterado_workitem` migrated to terraform-plugin-framework with full validator parity.** The resource
  is now served through the mux provider using the terraform-plugin-framework
  implementation (`resource_workitem_framework.go`). The SDKv2 implementation
  (`resource_workitem.go`) has been removed. CRUD operations target the Work
  Item Tracking REST API; the schema is unchanged. Validators restored to SDKv2
  parity: UUID pattern on `project_id`, non-whitespace on `title`/`type`/`state`/
  `area_path`/`iteration_path`, `ConflictsWith` between `custom_fields` and
  `additional_fields_json`, size floor on `tags`, and `AtLeast(1)` on `parent_id`.
  Orphaned SDKv2 helper (`utils/classification.go`) deleted. All 8 acceptance
  tests converted to use `SharedFixtureProjectName` (no ad-hoc project creation).
  Verified by live acceptance test `TestAccWorkItem_basic`.
- **`betterado_workitemtracking_field` migrated to terraform-plugin-framework.**
  The resource is now served through the mux provider using the framework
  implementation (`resource_field_framework.go`). The SDKv2 implementation
  (`resource_field.go`) has been removed. All schema attributes are preserved
  including computed attributes (`can_sort_by`, `is_queryable`, `is_identity`,
  `is_picklist`, `supported_operations`) and the `restore`/`WriteOnly` field.
  Verified by live acceptance test `TestAccWorkItemTrackingField_Basic`.

### Changed (Framework Migration)

- `betterado_group` resource: migrated to terraform-plugin-framework; schema unchanged (`scope`, `display_name`, `description`, `mail`, `members`, `origin_id` optional; `descriptor`, `domain`, `group_id`, `origin`, `principal_name`, `subject_kind`, `url` computed).
- `betterado_group_membership` resource: migrated to terraform-plugin-framework; schema unchanged (`group` required, `members` required set of strings, `mode` optional).
- `betterado_descriptor` data source: migrated to terraform-plugin-framework; schema unchanged (`storage_key` required; `descriptor`, `id` computed).
- `betterado_storage_key` data source: migrated to terraform-plugin-framework; schema unchanged (`descriptor` required; `storage_key`, `id` computed).
- `betterado_group` data source: migrated to terraform-plugin-framework; schema unchanged (`name` required, `project_id` optional; `descriptor`, `group_id`, `id`, `origin`, `origin_id` computed).
- `betterado_group_membership` data source: migrated to terraform-plugin-framework; schema unchanged (`group_descriptor` required; `members` list of string, `id` computed).
- `betterado_groups` data source: migrated to terraform-plugin-framework; schema unchanged (`project_id` optional; `groups` set with `id`, `descriptor`, `display_name`, `origin`, `origin_id`, `domain`, `mail_address`, `principal_name` computed).
- `betterado_user` data source: migrated to terraform-plugin-framework; schema unchanged (`descriptor` required; `display_name`, `domain`, `mail_address`, `origin`, `origin_id`, `principal_name`, `subject_kind` computed).
- `betterado_users` data source: migrated to terraform-plugin-framework; schema unchanged (`principal_name`, `origin`, `origin_id`, `subject_types` optional; `users` set computed).
- `betterado_service_principal` data source: migrated to terraform-plugin-framework; schema unchanged (`display_name` required; `descriptor`, `origin`, `origin_id` computed).
- `betterado_identity_group` data source: migrated to terraform-plugin-framework; schema unchanged (`name`, `project_id` required; `descriptor`, `id`, `subject_descriptor` computed).
- `betterado_identity_groups` data source: migrated to terraform-plugin-framework; schema unchanged (`project_id` optional; `groups` set with `id`, `name`, `descriptor`, `subject_descriptor` computed).
- `betterado_identity_user` data source: migrated to terraform-plugin-framework; schema unchanged (`name` required, `search_filter` optional; `descriptor`, `id`, `subject_descriptor` computed).

### FEATURES

- **`betterado_identity_group` data source migrated to terraform-plugin-framework.**
  The data source now uses the framework implementation served through the mux provider.
  Schema is unchanged (`name` required, `project_id` required; `descriptor`,
  `subject_descriptor` computed).
  Verified by live acceptance test `TestAccIdentityDataSources_Framework/IdentityGroup`.

- **`betterado_identity_groups` data source migrated to terraform-plugin-framework.**
  The data source now uses the framework implementation served through the mux provider.
  Schema is unchanged (`project_id` optional; `groups` set computed with `id`, `name`,
  `descriptor`, `subject_descriptor`).
  Verified by live acceptance test `TestAccIdentityDataSources_Framework/IdentityGroups`.

- **`betterado_identity_user` data source migrated to terraform-plugin-framework.**
  The data source now uses the framework implementation served through the mux provider.
  Schema is unchanged (`name` required; `search_filter` optional/computed defaulting to
  `General`; `descriptor`, `subject_descriptor` computed).
  Verified by live acceptance test `TestAccIdentityDataSources_Framework/IdentityUser`.

- **`betterado_user` data source migrated to terraform-plugin-framework.** The
  data source now uses the framework implementation served through the mux provider.
  Schema is unchanged (`descriptor` required; `display_name`, `domain`, `mail_address`,
  `origin`, `origin_id`, `principal_name`, `subject_kind` computed).
  Verified by live acceptance test `TestAccGraphComplexDataSources_Framework/User`.

- **`betterado_users` data source migrated to terraform-plugin-framework.** The
  data source now uses the framework implementation served through the mux provider.
  Schema is unchanged (`principal_name`, `origin`, `origin_id`, `subject_types` optional
  filters; `users` set computed with `id`, `descriptor`, `principal_name`, `origin`,
  `origin_id`, `display_name`, `mail_address`).
  Verified by live acceptance test `TestAccGraphComplexDataSources_Framework/Users`.

- **`betterado_groups` data source migrated to terraform-plugin-framework.** The
  data source now uses the framework implementation served through the mux provider.
  Schema is unchanged (`project_id` optional; `groups` set computed with `id`,
  `descriptor`, `display_name`, `origin`, `origin_id`, `domain`, `mail_address`,
  `principal_name`).
  Verified by live acceptance test `TestAccGraphComplexDataSources_Framework/Groups`.

- **`betterado_service_principal` data source migrated to terraform-plugin-framework.**
  The data source now uses the framework implementation served through the mux provider.
  Schema is unchanged (`display_name` required; `descriptor`, `origin_id`, `origin`
  computed).
  Verified by live acceptance test `TestAccGraphComplexDataSources_Framework/ServicePrincipal`.

### Changed

- Migrated `betterado_feed`, `betterado_feed_permission`, `betterado_feed_retention_policy`, and `data.betterado_feed` from terraform-plugin-sdk/v2 to terraform-plugin-framework (mux provider).

### FEATURES

- **`betterado_feed` migrated to terraform-plugin-framework.**
  The resource now uses the terraform-plugin-framework implementation served
  through the mux provider. CRUD operations target the ADO Packaging Feeds API
  (`_apis/packaging/feeds`). Schema is unchanged (`name`, `project_id`, `id`,
  `features`). Soft-deleted feeds are treated as destroyed and re-created on
  next apply. Verified by live acceptance tests
  `TestAccFeedFramework_basic` and
  `TestAccFeedFramework_withProject`.

- **`betterado_feed_permission` migrated to terraform-plugin-framework.**
  The resource now uses the terraform-plugin-framework implementation served
  through the mux provider. Manages a single permission entry per resource
  instance; `display_name`, `identity_id`, `identity_descriptor`, `role`,
  `feed_id`, and `project_id` are all exposed. Verified by live acceptance tests
  `TestAccFeedPermissionFramework_basic`.

- **`data.betterado_feed` migrated to terraform-plugin-framework.**
  The data source now uses the terraform-plugin-framework implementation served
  through the mux provider. Lookup by `name` (org-scoped) or `feed_id` (UUID,
  project-scoped) is supported; `name`, `feed_id`, `project_id`, and `id` are
  all exposed. Verified by live acceptance tests
  `TestAccFeedDataSourceFramework_byName` and
  `TestAccFeedDataSourceFramework_byId`.

- **`betterado_feed_retention_policy` migrated to terraform-plugin-framework.**
  The resource now uses the terraform-plugin-framework implementation served
  through the mux provider alongside the existing SDKv2 path. CRUD operations
  target the ADO Packaging Feeds retention-policy API
  (`_apis/packaging/feeds/{feedId}/retentionpolicies`). Schema is unchanged
  (`feed_id`, `project_id`, `count_limit`,
  `days_to_keep_recently_downloaded_packages`). Verified by live acceptance
  tests `TestAccFeedRetentionPolicyFramework_projectBasic` and
  `TestAccFeedRetentionPolicyFramework_update`.

### ENHANCEMENTS

- **Migrated `betterado_serviceendpoint_*` (24 resources + 8 data sources) from terraform-plugin-sdk/v2 to terraform-plugin-framework** via the mux provider.
  All service endpoint resources and data sources now use the framework implementation; SDKv2 registrations
  have been removed. The following types are included:
  - Resources: `betterado_serviceendpoint_generic`, `betterado_serviceendpoint_generic_v2`,
    `betterado_serviceendpoint_generic_git`, `betterado_serviceendpoint_azurerm`, `betterado_serviceendpoint_aws`,
    `betterado_serviceendpoint_azure_service_bus`, `betterado_serviceendpoint_gcp_terraform`,
    `betterado_serviceendpoint_dockerregistry`, `betterado_serviceendpoint_azurecr`,
    `betterado_serviceendpoint_github`, `betterado_serviceendpoint_github_enterprise`,
    `betterado_serviceendpoint_gitlab`, `betterado_serviceendpoint_bitbucket`,
    `betterado_serviceendpoint_jenkins`, `betterado_serviceendpoint_argocd`,
    `betterado_serviceendpoint_incomingwebhook`, `betterado_serviceendpoint_externaltfs`,
    `betterado_serviceendpoint_azuredevops`, `betterado_serviceendpoint_black_duck`,
    `betterado_serviceendpoint_checkmarx_one`, `betterado_serviceendpoint_checkmarx_sca`,
    `betterado_serviceendpoint_checkmarx_sast`, `betterado_serviceendpoint_artifactory`,
    `betterado_serviceendpoint_dynamics_lifecycle_services`
  - Data sources: `betterado_serviceendpoint_generic_v2`, `betterado_serviceendpoint_azurerm`,
    `betterado_serviceendpoint_dockerregistry`, `betterado_serviceendpoint_azurecr`,
    `betterado_serviceendpoint_github`, `betterado_serviceendpoint_bitbucket`,
    `betterado_serviceendpoint_npm`, `betterado_serviceendpoint_sonarcloud`

### FEATURES

- **`betterado_serviceendpoint_generic` resource migrated to terraform-plugin-framework.**
  Supports `url`, `username`, `password`, and optional `description` attributes.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_generic_v2` resource and data source migrated to terraform-plugin-framework.**
  Supports arbitrary endpoint types, authorization schemes, authorization parameters, and data parameters.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_generic_git` resource migrated to terraform-plugin-framework.**
  Supports `repository_url`, `username`, `password`, and `enable_pipelines_access` attributes.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_azurerm` resource and data source migrated to terraform-plugin-framework.**
  Supports Service Principal (manual and automatic) and Managed Identity authentication modes.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_aws` resource migrated to terraform-plugin-framework.**
  Supports `access_key_id`, `secret_access_key`, and optional `session_token` and `role_to_assume` attributes.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_azure_service_bus` resource migrated to terraform-plugin-framework.**
  Supports `connection_string` for Service Bus endpoint authentication.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_gcp_terraform` resource migrated to terraform-plugin-framework.**
  Supports GCP service account JSON credentials for Terraform operations.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_dockerregistry` resource and data source migrated to terraform-plugin-framework.**
  Supports Docker Hub and custom registry types with username + password authentication.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_azurecr` resource and data source migrated to terraform-plugin-framework.**
  Supports service principal and managed identity authentication for Azure Container Registry.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_github` resource and data source migrated to terraform-plugin-framework.**
  Supports PAT (`personal_access_token`) and OAuth (`oauth_configuration_id`) authentication schemes.
  Deregistered from the SDKv2 provider; served through the mux provider. Verified by live acceptance
  test `TestAccServiceEndpointGitHub_basic` (apply → read-back → idempotency re-plan → destroy).

- **`betterado_serviceendpoint_github_enterprise` resource migrated to terraform-plugin-framework.**
  Supports PAT and OAuth authentication; schema unchanged (`project_id`, `service_endpoint_name`,
  `github_enterprise_url`, `personal_access_token`, `oauth_configuration_id`, `description`).
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_gitlab` resource migrated to terraform-plugin-framework.**
  Username + password basic authentication. Deregistered from the SDKv2 provider; served through
  the mux provider.

- **`betterado_serviceendpoint_bitbucket` resource and data source migrated to terraform-plugin-framework.**
  Username + password basic authentication. Deregistered from the SDKv2 provider; served through
  the mux provider.

- **`betterado_serviceendpoint_jenkins` resource migrated to terraform-plugin-framework.**
  Supports username + password authentication with optional `accept_untrusted_certs` flag.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_argocd` resource migrated to terraform-plugin-framework.**
  Supports both token (`authentication_token`) and basic (`authentication_basic`) authentication modes.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_incomingwebhook` resource migrated to terraform-plugin-framework.**
  Supports `webhook_name`, optional `secret`, and optional `http_header` attributes.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_externaltfs` resource migrated to terraform-plugin-framework.**
  Uses `auth_personal` block with `personal_access_token` for Token authentication.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_azuredevops` resource migrated to terraform-plugin-framework.**
  Supports `org_url`, `release_api_url`, and `personal_access_token` attributes.
  Deprecated: use `betterado_serviceendpoint_runpipeline` instead.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_black_duck` resource migrated to terraform-plugin-framework.**
  Supports `server_url` and `api_token` attributes for Black Duck security scanning.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_checkmarx_one` resource migrated to terraform-plugin-framework.**
  Supports `server_url`, `authentication_url`, `tenant`, `client_id`, and `client_secret` attributes.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_checkmarx_sca` resource migrated to terraform-plugin-framework.**
  Supports `server_url`, `access_control_url`, `tenant`, `username`, and `password` attributes.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_checkmarx_sast` resource migrated to terraform-plugin-framework.**
  Supports `server_url`, `username`, and `password` attributes.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_artifactory` resource migrated to terraform-plugin-framework.**
  Supports both token (`authentication_token`) and basic (`authentication_basic`) authentication modes.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_dynamics_lifecycle_services` resource migrated to terraform-plugin-framework.**
  Uses `authorization_endpoint`, `lifecycle_services_api_endpoint`, `client_id`, `username`, and `password`
  attributes with `UsernamePassword` authentication scheme.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_npm` data source migrated to terraform-plugin-framework.**
  Reads an npm service endpoint by `project_id` + `service_endpoint_name`; returns connection details.
  Deregistered from the SDKv2 provider; served through the mux provider.

- **`betterado_serviceendpoint_sonarcloud` data source migrated to terraform-plugin-framework.**
  Reads a SonarCloud service endpoint by `project_id` + `service_endpoint_name`.
  Deregistered from the SDKv2 provider; served through the mux provider.
### FEATURES

- **`betterado_wiki` migrated to terraform-plugin-framework.** The resource now
  uses the terraform-plugin-framework implementation served through the mux
  provider. CRUD operations target the Azure DevOps Wiki API; the schema is
  unchanged (`project_id`, `name`, `type`, `repository_id`, `version`,
  `mapped_path`, `remote_url`, `url`). Project wikis and code wikis are both
  supported. Delete for project wikis now uses the `DeleteWiki` API directly
  (previously attempted to delete the underlying git repository, which was
  unreliable). Verified by live acceptance tests `TestAccWikiResource_projectWiki`
  and `TestAccWikiResource_codeWiki`.

- **`betterado_wiki_page` migrated to terraform-plugin-framework.** The resource
  now uses the terraform-plugin-framework implementation served through the mux
  provider. Schema is unchanged (`project_id`, `wiki_id`, `path`, `content`,
  `etag`). Concurrent page operations are serialised with a mutex to avoid
  ADO's "page already updated by another client" conflict. Verified by live
  acceptance tests `TestAccWikiPageResource_basic` and
  `TestAccWikiPageResource_update`.
### Changed

- Migrated `betterado_servicehook_storage_queue_pipelines` to terraform-plugin-framework (schema and behaviour unchanged).
- Migrated `betterado_servicehook_webhook_tfs` to terraform-plugin-framework (schema and behaviour unchanged).
### ENHANCEMENTS

- **`betterado_git_repository` resource migrated to terraform-plugin-framework.**
  The resource now uses the terraform-plugin-framework implementation served through
  the mux provider. Schema is unchanged (`project_id`, `name`, `default_branch`,
  `disabled`, `parent_repository_id`, `initialization`). Read-only attributes
  `is_fork`, `remote_url`, `size`, `ssh_url`, `url`, `web_url` remain available.
  Verified by live acceptance test `TestAccGitRepositoryFramework`.

- **`betterado_git_repository` data source migrated to terraform-plugin-framework.**
  Reads a single Git repository by `project_id` + `name`; exposes all repository
  attributes. Served through the mux provider. Verified by live acceptance test
  `TestAccDataGitRepositoryFramework`.

- **`betterado_git_repositories` data source migrated to terraform-plugin-framework.**
  Lists Git repositories in an Azure DevOps project; exposes a `repositories` list
  with full repository details per entry. Served through the mux provider.
  Verified by live acceptance test `TestAccDataGitRepositoriesFramework`.

- **`betterado_git_repository_branch` resource migrated to terraform-plugin-framework.**
  Manages a Git branch in an Azure DevOps repository. Schema is unchanged
  (`repository_id`, `name`, `ref_branch`, `is_default`). Served through the mux
  provider. Verified by live acceptance test `TestAccGitRepositoryBranchFramework`.

- **`betterado_git_repository_file` resource migrated to terraform-plugin-framework.**
  Manages a file in an Azure DevOps Git repository. Schema is unchanged
  (`repository_id`, `file`, `content`, `branch`, `commit_message`,
  `overwrite_on_create`). Served through the mux provider. Verified by live
  acceptance test `TestAccGitRepositoryFileFramework`.

- **`betterado_git_repository_file` data source migrated to terraform-plugin-framework.**
  Reads a file from an Azure DevOps Git repository by `repository_id`, `file`, and
  optional `branch`. Served through the mux provider. Verified by live acceptance
  test `TestAccDataGitRepositoryFileFramework`.
## [1.3.0] - 2026-07-01

### Added

- `betterado_feature_flag` resource and data source — manage ADO Feature Management API state at project or host scope.
### Changed

- `betterado_dashboard`: migrated to terraform-plugin-framework (parity with existing SDKv2 behaviour; no schema changes)
- `betterado_extension`: migrated to terraform-plugin-framework (parity with existing SDKv2 behaviour; no schema changes)

### Added

- `docs/dashboard-gap-matrix.md`: field-by-field coverage analysis of the ADO Dashboard API v7.1
- `docs/extension-gap-matrix.md`: field-by-field coverage analysis of the ADO Extension Management API v7.1

### FEATURES

- **`betterado_extension` migrated to terraform-plugin-framework.** The resource
  now uses the terraform-plugin-framework implementation served through the mux
  provider. Schema is unchanged (`extension_id`, `publisher_id`, `disabled`,
  `version`, `extension_name`, `publisher_name`, `scope`); ForceNew behaviour on
  `extension_id`/`publisher_id` is preserved. Verified by live acceptance tests
  `TestAccExtension_basic`, `TestAccExtension_complete`, `TestAccExtension_update`.

- **`betterado_dashboard` migrated to terraform-plugin-framework.** The resource
  now uses the terraform-plugin-framework implementation served through the mux
  provider. Schema is unchanged (`project_id`, `team_id`, `name`, `description`,
  `refresh_interval`, `owner_id`); supports both project-scoped and team-scoped
  dashboards. Live acceptance tests `TestAccDashboard_project_basic`,
  `TestAccDashboard_project_update`, `TestAccDashboard_team_basic`,
  `TestAccDashboard_team_update` verified against ADO with the betterado-standing-demo
  fixture project; live read-back evidence captured under `dashboard-acceptance-resource`
  label in `.forge/live-evidence/`.

## [1.3.0] - 2026-07-03

### FEATURES

- **`betterado_user_entitlement` migrated to terraform-plugin-framework.**
  The resource now uses the terraform-plugin-framework implementation served through
  the mux provider. CRUD operations target the Member Entitlement Management API at
  `{org}/_apis/memberentitlementmanagement/userentitlements`; the schema is unchanged
  (`principal_name`, `origin_id`, `origin`, `account_license_type`, `licensing_source`,
  `descriptor`). Verified by live acceptance test `TestAccUserEntitlement_Create`.

- **`betterado_group_entitlement` migrated to terraform-plugin-framework.**
  The resource now uses the terraform-plugin-framework implementation served through
  the mux provider. CRUD operations target the Member Entitlement Management API at
  `{org}/_apis/memberentitlementmanagement/groupentitlements`; the schema is unchanged
  (`display_name`, `origin_id`, `origin`, `account_license_type`, `licensing_source`,
  `principal_name`, `descriptor`). Verified by live acceptance test
  `TestAccGroupEntitlement_Create`.

- **`betterado_service_principal_entitlement` migrated to terraform-plugin-framework.**
  The resource now uses the terraform-plugin-framework implementation served through
  the mux provider. CRUD operations target the Member Entitlement Management API at
  `{org}/_apis/memberentitlementmanagement/serviceprincipals`; the schema is unchanged
  (`origin_id`, `origin`, `account_license_type`, `licensing_source`, `display_name`,
  `descriptor`). Verified by live acceptance test `TestAccServicePrincipalEntitlement_create`.
- **`betterado_build_definition` resource migrated to terraform-plugin-framework.**
  The resource now uses the framework implementation served through the mux provider.
  CRUD operations continue to target the Build API. Implemented attributes: `project_id`,
  `name`, `path`, `repository` (with `repo_type` validator), `variable` (fully wired in
  expand and read), `ci_trigger` (with `override` sub-block: batch, branch_filter,
  path_filter, max_concurrent_builds_per_branch, polling_interval), `pull_request_trigger`
  (with `override` sub-block and `forks` Required sub-block for SDKv2 parity),
  `agent_pool_name`, `agent_specification` (schema-present, not wired to API — see
  `docs/build-gap-matrix.md`), `job_authorization_scope`, `queue_status`, `skip_first_run`.
  Framework validators added for `name` (StringIsNotWhiteSpace), `path` (path format),
  `job_authorization_scope` (enum), `queue_status` (enum), `comment_required` (enum),
  `repo_type` (enum). Cross-attribute conflict validator added: setting both
  `repository.github_enterprise_url` and `repository.url` raises a plan-time error.
  **CI/PR trigger read-back wired:** `readIntoModel` now parses `def.Triggers` back into
  `model.CITrigger`/`model.PullRequestTrigger` via `flattenTriggersIntoModel`, so
  ADO-side trigger changes are surfaced as drift on `terraform plan`.
  **`skip_first_run` defaults to `true`** (skip — SDKv2 parity: absent `features` block means no auto-run).
  Set `skip_first_run = false` to opt-in to an immediate `PipelinesClient.RunPipeline` call on `Create`;
  a warning is emitted if the run fails (e.g. no YAML file yet) but does not prevent resource creation.
  **Deliberately NOT migrated this iteration** (documented in `docs/build-gap-matrix.md`):
  `variable_groups`, `build_completion_trigger`, `schedules`, `jobs` (OtherGit only).
  Verified by unit tests `TestBuildDefinitionFramework_Schema`,
  `TestBuildDefinitionFramework_FlattenCITrigger_UseYAML`,
  `TestBuildDefinitionFramework_FlattenCITrigger_Override`,
  `TestBuildDefinitionFramework_FlattenPRTrigger`,
  `TestBuildDefinitionFramework_FlattenFilters`,
  `TestBuildDefinitionFramework_ValidateConfig_Conflict`,
  `TestBuildDefinitionFramework_SkipFirstRunDefault`.

- **`betterado_build_folder` resource migrated to terraform-plugin-framework.**
  The resource now uses the framework implementation served through the mux provider.
  Schema: `project_id`, `path`, `description`. Framework validator added for `path`
  (path format: must start with `\`, no invalid characters — SDKv2 parity).
  Verified by unit test `TestBuildFolderFramework_PathValidator` and live acceptance
  test `TestAccBuildFolder_Framework_basic`.

- **`betterado_pipeline_authorization` resource migrated to terraform-plugin-framework.**
  The resource now uses the framework implementation. Schema: `project_id`,
  `pipeline_project_id`, `resource_id`, `type`, `pipeline_id`. Validators: `resource_id`
  (StringIsNotWhiteSpace), `type` (enum), `pipeline_id` (IntAtLeast(1)).
  Verified by live acceptance test `TestAccPipelineAuthorization_Framework_allPipeline_queue`.

- **`betterado_resource_authorization` resource migrated to terraform-plugin-framework.**
  The resource now uses the framework implementation. Schema: `project_id`,
  `resource_id`, `definition_id`, `type`, `authorized`. Verified by unit test.

- **`betterado_build_definition` data source migrated to terraform-plugin-framework.**
  The data source now uses the framework implementation. Reads a build definition by
  `project_id`, `name`, and optional `path`; returns `id`, `revision`, `repository`,
  `ci_trigger`, `pull_request_trigger`, `variable`, `agent_pool_name`,
  `agent_specification`, `job_authorization_scope`, `queue_status`, `skip_first_run`.
  **Deliberately NOT migrated this iteration** (documented in `docs/build-gap-matrix.md`):
  `variable_groups`, `schedules`, `jobs`, `build_completion_trigger`.
  Verified by live acceptance test `TestAccBuildDefinition_Framework_DataSource`.


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
