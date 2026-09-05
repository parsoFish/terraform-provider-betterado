# Project profile — terraform-provider-betterado

> Brain entry for the `betterado` Terraform provider (fork of `microsoft/terraform-provider-azuredevops`).
> Updated by WI-5 (gap-registry consolidation initiative, 2026-09-05).

## One-liner

A Terraform provider for Azure DevOps that adds classic release pipeline management,
task groups, and a long-tail of ADO API surface not covered by the upstream
`microsoft/azuredevops` provider.

## Repository

`github.com/parsoFish/terraform-provider-betterado`

## Fork relationship

Forked from `microsoft/terraform-provider-azuredevops`. betterado inherits all upstream
resources/data-sources and adds net-new surface the upstream does not cover.

## Net-new resources and data sources

> Types registered in betterado but NOT present in `microsoft/terraform-provider-azuredevops`.
> Derived from `azuredevops/internal/provider/framework_provider.go` registrations.

| Resource / data source | File | Host |
|---|---|---|
| `betterado_release_definition` (resource) | `azuredevops/internal/service/release/` | `vsrm.dev.azure.com` |
| `betterado_release_folder` (resource) | `azuredevops/internal/service/release/` | `vsrm.dev.azure.com` |
| `betterado_release_definition_permissions` (resource) | `azuredevops/internal/service/permissions/` | `dev.azure.com` |
| `betterado_task_group` (resource) | `azuredevops/internal/service/taskagent/` | `dev.azure.com` |
| `betterado_check_approval` (resource) | `azuredevops/internal/service/approvalsandchecks/` | `dev.azure.com` |
| `betterado_check_branch_control` (resource) | `azuredevops/internal/service/approvalsandchecks/` | `dev.azure.com` |
| `betterado_check_business_hours` (resource) | `azuredevops/internal/service/approvalsandchecks/` | `dev.azure.com` |
| `betterado_check_exclusive_lock` (resource) | `azuredevops/internal/service/approvalsandchecks/` | `dev.azure.com` |
| `betterado_check_required_template` (resource) | `azuredevops/internal/service/approvalsandchecks/` | `dev.azure.com` |
| `betterado_check_rest_api` (resource) | `azuredevops/internal/service/approvalsandchecks/` | `dev.azure.com` |
| `betterado_securityrole_assignment` (resource) | `azuredevops/internal/service/securityroles/` | `dev.azure.com` |
| `betterado_dashboard` (resource) | `azuredevops/internal/service/dashboard/` | `dev.azure.com` |
| `betterado_notification_subscription` (resource) | `azuredevops/internal/service/notification/` | `dev.azure.com` |
| `betterado_feature_flag` (resource) | `azuredevops/internal/service/featuremanagement/` | `dev.azure.com` |
| `betterado_extension` (resource) | `azuredevops/internal/service/extension/` | `dev.azure.com` |
| `betterado_extension_install` (resource) | `azuredevops/internal/service/extensionmanagement/` | `extmgmt.dev.azure.com` |
| `betterado_pipeline_authorization` (resource) | `azuredevops/internal/service/build/` | `dev.azure.com` |
| `betterado_workitem` (resource) | `azuredevops/internal/service/workitemtracking/` | `dev.azure.com` |
| `betterado_workitemtracking_field` (resource) | `azuredevops/internal/service/workitemtracking/` | `dev.azure.com` |
| `betterado_workitemquery` (resource) | `azuredevops/internal/service/workitemtracking/` | `dev.azure.com` |
| `betterado_workitemquery_folder` (resource) | `azuredevops/internal/service/workitemtracking/` | `dev.azure.com` |
| `betterado_workitemtrackingprocess_process` (resource) | `azuredevops/internal/service/workitemtrackingprocess/` | `dev.azure.com` |
| `betterado_workitemtrackingprocess_workitemtype` (resource) | `azuredevops/internal/service/workitemtrackingprocess/` | `dev.azure.com` |
| `betterado_workitemtrackingprocess_state` (resource) | `azuredevops/internal/service/workitemtrackingprocess/` | `dev.azure.com` |
| `betterado_workitemtrackingprocess_page` (resource) | `azuredevops/internal/service/workitemtrackingprocess/` | `dev.azure.com` |
| `betterado_workitemtrackingprocess_group` (resource) | `azuredevops/internal/service/workitemtrackingprocess/` | `dev.azure.com` |
| `betterado_workitemtrackingprocess_control` (resource) | `azuredevops/internal/service/workitemtrackingprocess/` | `dev.azure.com` |
| `betterado_workitemtrackingprocess_list` (resource) | `azuredevops/internal/service/workitemtrackingprocess/` | `dev.azure.com` |
| `betterado_workitemtrackingprocess_field` (resource) | `azuredevops/internal/service/workitemtrackingprocess/` | `dev.azure.com` |
| `betterado_workitemtrackingprocess_rule` (resource) | `azuredevops/internal/service/workitemtrackingprocess/` | `dev.azure.com` |
| `betterado_variable_group_variable` (resource) | `azuredevops/internal/service/taskagent/` | `dev.azure.com` |
| `betterado_deployment_group` (resource) | `azuredevops/internal/service/taskagent/` | `dev.azure.com` |
| `betterado_elastic_pool` (resource) | `azuredevops/internal/service/taskagent/` | `dev.azure.com` |
| `betterado_environment_resource_kubernetes` (resource) | `azuredevops/internal/service/taskagent/` | `dev.azure.com` |
| `betterado_pipeline_approval` (resource) | `azuredevops/internal/service/pipelinesapproval/` | `dev.azure.com` |
| `data.betterado_release_definition` (data source) | `azuredevops/internal/service/release/` | `vsrm.dev.azure.com` |
| `data.betterado_release_definitions` (data source) | `azuredevops/internal/service/release/` | `vsrm.dev.azure.com` |
| `data.betterado_release_definition_history` (data source) | `azuredevops/internal/service/release/` | `vsrm.dev.azure.com` |
| `data.betterado_release_definition_revision` (data source) | `azuredevops/internal/service/release/` | `vsrm.dev.azure.com` |
| `data.betterado_release_folder` (data source) | `azuredevops/internal/service/release/` | `vsrm.dev.azure.com` |
| `data.betterado_feature_flag` (data source) | `azuredevops/internal/service/featuremanagement/` | `dev.azure.com` |
| `data.betterado_task_group` (data source) | `azuredevops/internal/service/taskagent/` | `dev.azure.com` |
| `data.betterado_pipeline_approvals` (data source) | `azuredevops/internal/service/pipelinesapproval/` | `dev.azure.com` |
| `data.betterado_accounts` (data source) | `azuredevops/internal/service/accounts/` | `app.vssps.visualstudio.com` |
| `data.betterado_profile` (data source) | `azuredevops/internal/service/profile/` | `app.vssps.visualstudio.com` |
| `data.betterado_securityrole_definitions` (data source) | `azuredevops/internal/service/securityroles/` | `dev.azure.com` |
| `data.betterado_notification_subscription` (data source) | `azuredevops/internal/service/notification/` | `dev.azure.com` |

## API-coverage discipline

> Gap-open count derived from `docs/gap-registry.md` Coverage index (WI-5, 2026-09-05).

**91 mapped / 91 writable+computed gaps open** across all 31 ADO API areas.
See `docs/gap-registry.md` for the full per-area breakdown and Priority backlog.

## Key properties

- **Provider mux:** `tf6muxserver` multiplexing SDKv2 (`tf5to6server`) + terraform-plugin-framework
- **Net-new surface:** classic release pipelines, task groups, approvals/checks, security roles, dashboards, notification subscriptions, feature flags, extension management, work item tracking process, WI queries, WI fields, deployment groups, elastic pools
- **Inherited surface:** all upstream `microsoft/azuredevops` resources/data-sources (~132 resources, ~44 data sources) maintained at parity
- **Test discipline:** canonical 5-mock unit tests (azdosdkmocks+gomock) + live TF_ACC acceptance tests per feature
- **State compatibility:** schema version upgraders for `betterado_release_definition` (v0→v1 on framework migration)
