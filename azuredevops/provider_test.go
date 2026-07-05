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
		// betterado_area_permissions is now in the framework provider (framework_provider.go)
		// betterado_branch_policy_auto_reviewers is now in the framework provider (framework_provider.go)
		// betterado_branch_policy_build_validation is now in the framework provider (framework_provider.go)
		// betterado_branch_policy_comment_resolution is now in the framework provider (framework_provider.go)
		// betterado_branch_policy_merge_types is now in the framework provider (framework_provider.go)
		// betterado_branch_policy_min_reviewers is now in the framework provider (framework_provider.go)
		// betterado_branch_policy_status_check is now in the framework provider (framework_provider.go)
		// betterado_branch_policy_work_item_linking is now in the framework provider (framework_provider.go)
		// betterado_build_definition is now a framework resource (registered in framework_provider.go)
		// and is no longer in the SDKv2 provider resource map.
		// betterado_build_definition_permissions is now in the framework provider (framework_provider.go)
		// betterado_build_folder is now a framework resource (registered in framework_provider.go)
		// and is no longer in the SDKv2 provider resource map.
		// betterado_build_folder_permissions is now in the framework provider (framework_provider.go)
		// betterado_check_approval is now in the framework provider (framework_provider.go)
		// betterado_check_branch_control is now in the framework provider (framework_provider.go)
		// betterado_check_business_hours is now in the framework provider (framework_provider.go)
		// betterado_check_exclusive_lock is now in the framework provider (framework_provider.go)
		// betterado_check_required_template is now in the framework provider (framework_provider.go)
		// betterado_check_rest_api is now in the framework provider (framework_provider.go)
		// betterado_dashboard is now a framework resource (registered in framework_provider.go)
		// and is no longer in the SDKv2 provider resource map.
		"betterado_deployment_group",
		"betterado_elastic_pool",
		"betterado_environment",
		"betterado_environment_resource_kubernetes",
		// betterado_extension is now a framework resource (registered in framework_provider.go)
		// and is no longer in the SDKv2 provider resource map.
		// betterado_git_permissions is now in the framework provider (framework_provider.go)
		// "betterado_git_repository" migrated to terraform-plugin-framework provider
		// "betterado_git_repository_branch" migrated to terraform-plugin-framework provider
		// "betterado_git_repository_file" migrated to terraform-plugin-framework provider
		// betterado_group is now a framework resource (registered in framework_provider.go)
		// betterado_group_entitlement is now a framework resource (registered in framework_provider.go)
		// and is no longer in the SDKv2 provider resource map.
		// betterado_group_membership is now a framework resource (registered in framework_provider.go)
		// betterado_iteration_permissions is now in the framework provider (framework_provider.go)
		// betterado_library_permissions is now in the framework provider (framework_provider.go)
		// betterado_pipeline_authorization is now a framework resource (registered in framework_provider.go)
		// and is no longer in the SDKv2 provider resource map.
		// betterado_project is now in the framework provider (framework_provider.go)
		// betterado_project_features is now in the framework provider (framework_provider.go)
		// betterado_project_permissions is now in the framework provider (framework_provider.go)
		// betterado_project_pipeline_settings is now in the framework provider (framework_provider.go)
		// betterado_project_tags is now in the framework provider (framework_provider.go)
		// betterado_release_definition_permissions is now a framework resource (registered in framework_provider.go)
		// and is no longer in the SDKv2 provider resource map.
		// betterado_release_folder is now a framework resource (registered in framework_provider.go)
		// and is no longer in the SDKv2 provider resource map.
		// betterado_repository_policy_author_email_pattern is now in the framework provider (framework_provider.go)
		// betterado_repository_policy_case_enforcement is now in the framework provider (framework_provider.go)
		// betterado_repository_policy_check_credentials is now in the framework provider (framework_provider.go)
		// betterado_repository_policy_file_path_pattern is now in the framework provider (framework_provider.go)
		// betterado_repository_policy_max_file_size is now in the framework provider (framework_provider.go)
		// betterado_repository_policy_max_path_length is now in the framework provider (framework_provider.go)
		// betterado_repository_policy_reserved_names is now in the framework provider (framework_provider.go)
		// betterado_resource_authorization is now a framework resource (registered in framework_provider.go)
		// and is no longer in the SDKv2 provider resource map.
		// betterado_security_permissions is now in the framework provider (framework_provider.go)
		// betterado_securityrole_assignment is now in the framework provider (framework_provider.go)
		"betterado_serviceendpoint_jfrog_artifactory_v2",
		"betterado_serviceendpoint_jfrog_distribution_v2",
		"betterado_serviceendpoint_jfrog_platform_v2",
		"betterado_serviceendpoint_jfrog_xray_v2",
		"betterado_serviceendpoint_kubernetes",
		"betterado_serviceendpoint_maven",
		"betterado_serviceendpoint_nexus",
		"betterado_serviceendpoint_nuget",
		"betterado_serviceendpoint_octopusdeploy",
		"betterado_serviceendpoint_openshift",
		// betterado_serviceendpoint_permissions is now in the framework provider (framework_provider.go)
		"betterado_serviceendpoint_runpipeline",
		"betterado_serviceendpoint_servicefabric",
		"betterado_serviceendpoint_snyk",
		"betterado_serviceendpoint_sonarqube",
		"betterado_serviceendpoint_ssh",
		"betterado_serviceendpoint_visualstudiomarketplace",
		// betterado_servicehook_permissions is now in the framework provider (framework_provider.go)
		// betterado_service_principal_entitlement is now a framework resource (registered in framework_provider.go)
		// and is no longer in the SDKv2 provider resource map.
		// betterado_tagging_permissions is now in the framework provider (framework_provider.go)
		// betterado_team is now in the framework provider (framework_provider.go)
		// betterado_team_administrators is now in the framework provider (framework_provider.go)
		// betterado_team_members is now in the framework provider (framework_provider.go)
		"betterado_variable_group",
		// betterado_variable_group_permissions is now in the framework provider (framework_provider.go)
		"betterado_variable_group_variable",
		// betterado_workitem is now a framework resource (registered in framework_provider.go)
		// betterado_workitemtracking_field is now a framework resource (registered in framework_provider.go)
		// betterado_workitemquery is now a framework resource (registered in framework_provider.go)
		// betterado_workitemquery_folder is now a framework resource (registered in framework_provider.go)
		// betterado_workitemquery_permissions is now in the framework provider (framework_provider.go)
		"betterado_workitemtrackingprocess_control",
		"betterado_workitemtrackingprocess_field",
		"betterado_workitemtrackingprocess_group",
		"betterado_workitemtrackingprocess_inherited_control",
		"betterado_workitemtrackingprocess_inherited_page",
		"betterado_workitemtrackingprocess_inherited_state",
		"betterado_workitemtrackingprocess_list",
		"betterado_workitemtrackingprocess_page",
		"betterado_workitemtrackingprocess_process",
		// betterado_workitemtrackingprocess_process_permissions is now in the framework provider (framework_provider.go)
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
		// betterado_build_definition (data source) has also been migrated to the
		// framework provider and is no longer listed here.
		"betterado_agent_pool",
		"betterado_agent_pools",
		"betterado_agent_queue",
		// betterado_area is now a framework data source (registered in framework_provider.go)
		// betterado_client_config is now in the framework provider (framework_provider.go)
		// betterado_descriptor is now a framework data source (registered in framework_provider.go)
		"betterado_environment",
		// "betterado_git_repositories" data source migrated to terraform-plugin-framework provider
		// "betterado_git_repository" data source migrated to terraform-plugin-framework provider
		// "betterado_git_repository_file" data source migrated to terraform-plugin-framework provider
		// betterado_group is now a framework data source (registered in framework_provider.go)
		// betterado_group_membership is now a framework data source (registered in framework_provider.go)
		// betterado_groups is now a framework data source (registered in framework_provider.go)
		// betterado_identity_group is now a framework data source (registered in framework_provider.go)
		// betterado_identity_groups is now a framework data source (registered in framework_provider.go)
		// betterado_identity_user is now a framework data source (registered in framework_provider.go)
		// betterado_iteration is now a framework data source (registered in framework_provider.go)
		// betterado_project is now in the framework provider (framework_provider.go)
		// betterado_projects is now in the framework provider (framework_provider.go)
		// betterado_security_namespace is now in the framework provider (framework_provider.go)
		// betterado_security_namespace_token is now in the framework provider (framework_provider.go)
		// betterado_security_namespaces is now in the framework provider (framework_provider.go)
		// betterado_securityrole_definitions is now in the framework provider (framework_provider.go)
		// betterado_storage_key is now a framework data source (registered in framework_provider.go)
		// betterado_service_principal is now a framework data source (registered in framework_provider.go)
		// betterado_team is now in the framework provider (framework_provider.go)
		"betterado_task_group",
		// betterado_teams is now in the framework provider (framework_provider.go)
		// betterado_user is now a framework data source (registered in framework_provider.go)
		// betterado_users is now a framework data source (registered in framework_provider.go)
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
