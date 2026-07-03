package azuredevops_test

import (
	"fmt"
	"testing"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops"
	"github.com/stretchr/testify/require"
)

func TestProvider_HasChildResources(t *testing.T) {
	expectedResources := []string{
		"betterado_agent_pool",
		"betterado_agent_queue",
		"betterado_area_permissions",
		"betterado_branch_policy_auto_reviewers",
		"betterado_branch_policy_build_validation",
		"betterado_branch_policy_comment_resolution",
		"betterado_branch_policy_merge_types",
		"betterado_branch_policy_min_reviewers",
		"betterado_branch_policy_status_check",
		"betterado_branch_policy_work_item_linking",
		"betterado_build_definition",
		"betterado_build_definition_permissions",
		"betterado_build_folder",
		"betterado_build_folder_permissions",
		"betterado_check_approval",
		"betterado_check_branch_control",
		"betterado_check_business_hours",
		"betterado_check_exclusive_lock",
		"betterado_check_required_template",
		"betterado_check_rest_api",
		"betterado_dashboard",
		"betterado_deployment_group",
		"betterado_elastic_pool",
		"betterado_environment",
		"betterado_environment_resource_kubernetes",
		"betterado_extension",
		"betterado_feed",
		"betterado_feed_permission",
		"betterado_feed_retention_policy",
		"betterado_git_permissions",
		"betterado_git_repository",
		"betterado_git_repository_branch",
		"betterado_git_repository_file",
		"betterado_group",
		"betterado_group_entitlement",
		"betterado_group_membership",
		"betterado_iteration_permissions",
		"betterado_library_permissions",
		"betterado_pipeline_authorization",
		"betterado_project",
		"betterado_project_features",
		"betterado_project_permissions",
		"betterado_project_pipeline_settings",
		"betterado_project_tags",
		// betterado_release_definition_permissions is now a framework resource (registered in framework_provider.go)
		// and is no longer in the SDKv2 provider resource map.
		// betterado_release_folder is now a framework resource (registered in framework_provider.go)
		// and is no longer in the SDKv2 provider resource map.
		"betterado_repository_policy_author_email_pattern",
		"betterado_repository_policy_case_enforcement",
		"betterado_repository_policy_check_credentials",
		"betterado_repository_policy_file_path_pattern",
		"betterado_repository_policy_max_file_size",
		"betterado_repository_policy_max_path_length",
		"betterado_repository_policy_reserved_names",
		"betterado_resource_authorization",
		"betterado_security_permissions",
		"betterado_securityrole_assignment",
		"betterado_serviceendpoint_generic_v2",
		"betterado_serviceendpoint_argocd",
		"betterado_serviceendpoint_artifactory",
		"betterado_serviceendpoint_aws",
		"betterado_serviceendpoint_azure_service_bus",
		"betterado_serviceendpoint_azurecr",
		"betterado_serviceendpoint_azuredevops",
		"betterado_serviceendpoint_azurerm",
		"betterado_serviceendpoint_bitbucket",
		"betterado_serviceendpoint_black_duck",
		"betterado_serviceendpoint_checkmarx_one",
		"betterado_serviceendpoint_checkmarx_sca",
		"betterado_serviceendpoint_checkmarx_sast",
		"betterado_serviceendpoint_dockerregistry",
		"betterado_serviceendpoint_dynamics_lifecycle_services",
		"betterado_serviceendpoint_externaltfs",
		"betterado_serviceendpoint_gcp_terraform",
		"betterado_serviceendpoint_generic",
		"betterado_serviceendpoint_generic_git",
		"betterado_serviceendpoint_github",
		"betterado_serviceendpoint_github_enterprise",
		"betterado_serviceendpoint_gitlab",
		"betterado_serviceendpoint_incomingwebhook",
		"betterado_serviceendpoint_jenkins",
		"betterado_serviceendpoint_jfrog_artifactory_v2",
		"betterado_serviceendpoint_jfrog_distribution_v2",
		"betterado_serviceendpoint_jfrog_platform_v2",
		"betterado_serviceendpoint_jfrog_xray_v2",
		"betterado_serviceendpoint_kubernetes",
		"betterado_serviceendpoint_maven",
		"betterado_serviceendpoint_nexus",
		"betterado_serviceendpoint_npm",
		"betterado_serviceendpoint_nuget",
		"betterado_serviceendpoint_octopusdeploy",
		"betterado_serviceendpoint_openshift",
		"betterado_serviceendpoint_permissions",
		"betterado_serviceendpoint_runpipeline",
		"betterado_serviceendpoint_servicefabric",
		"betterado_serviceendpoint_snyk",
		"betterado_serviceendpoint_sonarcloud",
		"betterado_serviceendpoint_sonarqube",
		"betterado_serviceendpoint_ssh",
		"betterado_serviceendpoint_visualstudiomarketplace",
		"betterado_servicehook_permissions",
		"betterado_servicehook_storage_queue_pipelines",
		"betterado_servicehook_webhook_tfs",
		"betterado_service_principal_entitlement",
		"betterado_tagging_permissions",
		"betterado_team",
		"betterado_team_administrators",
		"betterado_team_members",
		"betterado_variable_group",
		"betterado_variable_group_permissions",
		"betterado_variable_group_variable",
		"betterado_wiki",
		"betterado_wiki_page",
		"betterado_workitem",
		"betterado_workitemtracking_field",
		"betterado_workitemquery",
		"betterado_workitemquery_folder",
		"betterado_workitemquery_permissions",
		"betterado_workitemtrackingprocess_control",
		"betterado_workitemtrackingprocess_field",
		"betterado_workitemtrackingprocess_group",
		"betterado_workitemtrackingprocess_inherited_control",
		"betterado_workitemtrackingprocess_inherited_page",
		"betterado_workitemtrackingprocess_inherited_state",
		"betterado_workitemtrackingprocess_list",
		"betterado_workitemtrackingprocess_page",
		"betterado_workitemtrackingprocess_process",
		"betterado_workitemtrackingprocess_process_permissions",
		"betterado_workitemtrackingprocess_rule",
		"betterado_workitemtrackingprocess_state",
		"betterado_workitemtrackingprocess_system_control",
		"betterado_workitemtrackingprocess_workitemtype",
	}

	resources := azuredevops.Provider().ResourcesMap

	for _, resource := range expectedResources {
		require.Contains(t, resources, resource, fmt.Sprintf("An expected resource (%s) was not registered", resource))
		require.NotNil(t, resources[resource], "A resource cannot have a nil schema")
	}
	require.Equal(t, len(expectedResources), len(resources), "There are an unexpected number of registered resources")
}

func TestProvider_HasChildDataSources(t *testing.T) {
	expectedDataSources := []string{
		// NOTE: betterado_release_definition, betterado_release_definition_history,
		// betterado_release_definition_revision, betterado_release_definitions, and
		// betterado_release_folder have been migrated to the framework provider —
		// they are no longer registered in the SDKv2 DataSourcesMap.
		"betterado_agent_pool",
		"betterado_agent_pools",
		"betterado_agent_queue",
		"betterado_area",
		"betterado_build_definition",
		"betterado_client_config",
		"betterado_descriptor",
		"betterado_environment",
		"betterado_feed",
		"betterado_git_repositories",
		"betterado_git_repository",
		"betterado_git_repository_file",
		"betterado_group",
		"betterado_group_membership",
		"betterado_groups",
		"betterado_identity_group",
		"betterado_identity_groups",
		"betterado_identity_user",
		"betterado_iteration",
		"betterado_project",
		"betterado_projects",
		"betterado_security_namespace",
		"betterado_security_namespace_token",
		"betterado_security_namespaces",
		"betterado_securityrole_definitions",
		"betterado_serviceendpoint_generic_v2",
		"betterado_serviceendpoint_azurecr",
		"betterado_serviceendpoint_azurerm",
		"betterado_serviceendpoint_bitbucket",
		"betterado_serviceendpoint_dockerregistry",
		"betterado_serviceendpoint_github",
		"betterado_serviceendpoint_npm",
		"betterado_serviceendpoint_sonarcloud",
		"betterado_storage_key",
		"betterado_service_principal",
		"betterado_team",
		"betterado_task_group",
		"betterado_teams",
		"betterado_user",
		"betterado_users",
		"betterado_variable_group",
		"betterado_workitemtrackingprocess_process",
		"betterado_workitemtrackingprocess_processes",
		"betterado_workitemtrackingprocess_workitemtype",
		"betterado_workitemtrackingprocess_workitemtypes",
	}

	dataSources := azuredevops.Provider().DataSourcesMap

	for _, resource := range expectedDataSources {
		require.Contains(t, dataSources, resource, "An expected data source was not registered")
		require.NotNil(t, dataSources[resource], "A data source cannot have a nil schema")
	}
	require.Equal(t, len(expectedDataSources), len(dataSources), "There are an unexpected number of registered data sources")
}

func TestProvider_SchemaIsValid(t *testing.T) {
	type testParams struct {
		name      string
		required  bool
		sensitive bool
	}

	tests := []testParams{
		{"org_service_url", false, false},
		{"personal_access_token", false, true},

		{"client_id", false, false},
		{"client_id_file_path", false, false},
		{"tenant_id", false, false},
		{"auxiliary_tenant_ids", false, false},
		{"client_certificate_path", false, false},
		{"client_certificate", false, true},
		{"client_certificate_password", false, true},
		{"client_secret", false, true},
		{"client_secret_path", false, false},
		{"oidc_request_token", false, false},
		{"oidc_request_url", false, false},
		{"oidc_token", false, true},
		{"oidc_token_file_path", false, false},
		{"oidc_azure_service_connection_id", false, false},
		{"use_oidc", false, false},
		{"use_msi", false, false},
		{"use_cli", false, false},
	}

	schema := azuredevops.Provider().Schema
	require.Equal(t, len(tests), len(schema), "There are an unexpected number of properties in the schema")

	for _, test := range tests {
		require.Contains(t, schema, test.name, "An expected property was not found in the schema")
		require.NotNil(t, schema[test.name], "A property in the schema cannot have a nil value")
		require.Equal(t, test.sensitive, schema[test.name].Sensitive, "A property in the schema has an incorrect sensitivity value")
		require.Equal(t, test.required, schema[test.name].Required, "A property in the schema has an incorrect required value")
	}
}
