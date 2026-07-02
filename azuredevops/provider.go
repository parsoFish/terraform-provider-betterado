package azuredevops

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/entrauth/aztfauth"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/approvalsandchecks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/build"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/core"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/feed"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/git"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/graph"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/identity"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/memberentitlementmanagement"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/permissions"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/policy/branch"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/policy/repository"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/security"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/securityroles"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/serviceendpoint"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/servicehook"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/wiki"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/workitemtracking"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/workitemtrackingprocess"
)

// Provider - The top level Azure DevOps Provider definition.
func Provider() *schema.Provider {
	p := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"betterado_agent_pool":                       taskagent.ResourceAgentPool(),
			"betterado_agent_queue":                      taskagent.ResourceAgentQueue(),
			"betterado_area_permissions":                 permissions.ResourceAreaPermissions(),
			"betterado_branch_policy_auto_reviewers":     branch.ResourceBranchPolicyAutoReviewers(),
			"betterado_branch_policy_build_validation":   branch.ResourceBranchPolicyBuildValidation(),
			"betterado_branch_policy_comment_resolution": branch.ResourceBranchPolicyCommentResolution(),
			"betterado_branch_policy_merge_types":        branch.ResourceBranchPolicyMergeTypes(),
			"betterado_branch_policy_min_reviewers":      branch.ResourceBranchPolicyMinReviewers(),
			"betterado_branch_policy_status_check":       branch.ResourceBranchPolicyStatusCheck(),
			"betterado_branch_policy_work_item_linking":  branch.ResourceBranchPolicyWorkItemLinking(),
			"betterado_build_definition":                 build.ResourceBuildDefinition(),
			"betterado_build_definition_permissions":     permissions.ResourceBuildDefinitionPermissions(),
			"betterado_build_folder":                     build.ResourceBuildFolder(),
			"betterado_build_folder_permissions":         permissions.ResourceBuildFolderPermissions(),
			"betterado_check_approval":                   approvalsandchecks.ResourceCheckApproval(),
			"betterado_check_branch_control":             approvalsandchecks.ResourceCheckBranchControl(),
			"betterado_check_business_hours":             approvalsandchecks.ResourceCheckBusinessHours(),
			"betterado_check_exclusive_lock":             approvalsandchecks.ResourceCheckExclusiveLock(),
			"betterado_check_required_template":          approvalsandchecks.ResourceCheckRequiredTemplate(),
			"betterado_check_rest_api":                   approvalsandchecks.ResourceCheckRestAPI(),
			// betterado_dashboard is now registered in the framework provider (framework_provider.go)
			// and must NOT be listed here — duplicating a resource type across mux providers causes
			// "Invalid Provider Server Combination" at plan time.
			"betterado_deployment_group":                taskagent.ResourceDeploymentGroup(),
			"betterado_elastic_pool":                    taskagent.ResourceAgentPoolVMSS(),
			"betterado_environment":                     taskagent.ResourceEnvironment(),
			"betterado_environment_resource_kubernetes": taskagent.ResourceEnvironmentKubernetes(),
			// betterado_extension is now registered in the framework provider (framework_provider.go)
			// and must NOT be listed here — duplicating a resource type across mux providers causes
			// "Invalid Provider Server Combination" at plan time.
			"betterado_feed":                      feed.ResourceFeed(),
			"betterado_feed_permission":           feed.ResourceFeedPermission(),
			"betterado_feed_retention_policy":     feed.ResourceFeedRetentionPolicy(),
			"betterado_git_permissions":           permissions.ResourceGitPermissions(),
			"betterado_git_repository":            git.ResourceGitRepository(),
			"betterado_git_repository_branch":     git.ResourceGitRepositoryBranch(),
			"betterado_git_repository_file":       git.ResourceGitRepositoryFile(),
			"betterado_group":                     graph.ResourceGroup(),
			"betterado_group_entitlement":         memberentitlementmanagement.ResourceGroupEntitlement(),
			"betterado_group_membership":          graph.ResourceGroupMembership(),
			"betterado_iteration_permissions":     permissions.ResourceIterationPermissions(),
			"betterado_library_permissions":       permissions.ResourceLibraryPermissions(),
			"betterado_pipeline_authorization":    build.ResourcePipelineAuthorization(),
			"betterado_project":                   core.ResourceProject(),
			"betterado_project_features":          core.ResourceProjectFeatures(),
			"betterado_project_permissions":       permissions.ResourceProjectPermissions(),
			"betterado_project_pipeline_settings": core.ResourceProjectPipelineSettings(),
			"betterado_project_tags":              core.ResourceProjectTag(),
			// betterado_release_definition_permissions is now registered in the framework provider (framework_provider.go)
			// and must NOT be listed here — duplicating a resource type across mux providers causes
			// "Invalid Provider Server Combination" at plan time.
			// betterado_release_folder is now registered in the framework provider (framework_provider.go)
			// and must NOT be listed here — duplicating a resource type across mux providers causes
			// "Invalid Provider Server Combination" at plan time.
			"betterado_repository_policy_author_email_pattern":      repository.ResourceRepositoryPolicyAuthorEmailPatterns(),
			"betterado_repository_policy_case_enforcement":          repository.ResourceRepositoryEnforceConsistentCase(),
			"betterado_repository_policy_check_credentials":         repository.ResourceRepositoryPolicyCheckCredentials(),
			"betterado_repository_policy_file_path_pattern":         repository.ResourceRepositoryFilePathPatterns(),
			"betterado_repository_policy_max_file_size":             repository.ResourceRepositoryMaxFileSize(),
			"betterado_repository_policy_max_path_length":           repository.ResourceRepositoryMaxPathLength(),
			"betterado_repository_policy_reserved_names":            repository.ResourceRepositoryReservedNames(),
			"betterado_resource_authorization":                      build.ResourceResourceAuthorization(),
			"betterado_security_permissions":                        security.ResourceGenericPermissions(),
			"betterado_securityrole_assignment":                     securityroles.ResourceSecurityRoleAssignment(),
			"betterado_serviceendpoint_generic_v2":                  serviceendpoint.ResourceServiceEndpointGenericV2(),
			"betterado_serviceendpoint_argocd":                      serviceendpoint.ResourceServiceEndpointArgoCD(),
			"betterado_serviceendpoint_artifactory":                 serviceendpoint.ResourceServiceEndpointArtifactory(),
			"betterado_serviceendpoint_aws":                         serviceendpoint.ResourceServiceEndpointAws(),
			"betterado_serviceendpoint_azure_service_bus":           serviceendpoint.ResourceServiceEndpointAzureServiceBus(),
			"betterado_serviceendpoint_azurecr":                     serviceendpoint.ResourceServiceEndpointAzureCR(),
			"betterado_serviceendpoint_azuredevops":                 serviceendpoint.ResourceServiceEndpointAzureDevOps(),
			"betterado_serviceendpoint_azurerm":                     serviceendpoint.ResourceServiceEndpointAzureRM(),
			"betterado_serviceendpoint_bitbucket":                   serviceendpoint.ResourceServiceEndpointBitBucket(),
			"betterado_serviceendpoint_black_duck":                  serviceendpoint.ResourceServiceEndpointBlackDuck(),
			"betterado_serviceendpoint_checkmarx_one":               serviceendpoint.ResourceServiceEndpointCheckMarxOneService(),
			"betterado_serviceendpoint_checkmarx_sca":               serviceendpoint.ResourceServiceEndpointCheckMarxSCA(),
			"betterado_serviceendpoint_checkmarx_sast":              serviceendpoint.ResourceServiceEndpointCheckMarxSAST(),
			"betterado_serviceendpoint_dockerregistry":              serviceendpoint.ResourceServiceEndpointDockerRegistry(),
			"betterado_serviceendpoint_dynamics_lifecycle_services": serviceendpoint.ResourceServiceEndpointDynamicsLifecycleServices(),
			"betterado_serviceendpoint_externaltfs":                 serviceendpoint.ResourceServiceEndpointExternalTFS(),
			"betterado_serviceendpoint_gcp_terraform":               serviceendpoint.ResourceServiceEndpointGcp(),
			"betterado_serviceendpoint_generic":                     serviceendpoint.ResourceServiceEndpointGeneric(),
			"betterado_serviceendpoint_generic_git":                 serviceendpoint.ResourceServiceEndpointGenericGit(),
			"betterado_serviceendpoint_github":                      serviceendpoint.ResourceServiceEndpointGitHub(),
			"betterado_serviceendpoint_github_enterprise":           serviceendpoint.ResourceServiceEndpointGitHubEnterprise(),
			"betterado_serviceendpoint_gitlab":                      serviceendpoint.ResourceServiceEndpointGitLab(),
			"betterado_serviceendpoint_incomingwebhook":             serviceendpoint.ResourceServiceEndpointIncomingWebhook(),
			"betterado_serviceendpoint_jenkins":                     serviceendpoint.ResourceServiceEndpointJenkins(),
			"betterado_serviceendpoint_jfrog_artifactory_v2":        serviceendpoint.ResourceServiceEndpointJFrogArtifactoryV2(),
			"betterado_serviceendpoint_jfrog_distribution_v2":       serviceendpoint.ResourceServiceEndpointJFrogDistributionV2(),
			"betterado_serviceendpoint_jfrog_platform_v2":           serviceendpoint.ResourceServiceEndpointJFrogPlatformV2(),
			"betterado_serviceendpoint_jfrog_xray_v2":               serviceendpoint.ResourceServiceEndpointJFrogXRayV2(),
			"betterado_serviceendpoint_kubernetes":                  serviceendpoint.ResourceServiceEndpointKubernetes(),
			"betterado_serviceendpoint_maven":                       serviceendpoint.ResourceServiceEndpointMaven(),
			"betterado_serviceendpoint_nexus":                       serviceendpoint.ResourceServiceEndpointNexus(),
			"betterado_serviceendpoint_npm":                         serviceendpoint.ResourceServiceEndpointNpm(),
			"betterado_serviceendpoint_nuget":                       serviceendpoint.ResourceServiceEndpointNuGet(),
			"betterado_serviceendpoint_octopusdeploy":               serviceendpoint.ResourceServiceEndpointOctopusDeploy(),
			"betterado_serviceendpoint_openshift":                   serviceendpoint.ResourceServiceEndpointOpenshift(),
			"betterado_serviceendpoint_permissions":                 permissions.ResourceServiceEndpointPermissions(),
			"betterado_serviceendpoint_runpipeline":                 serviceendpoint.ResourceServiceEndpointRunPipeline(),
			"betterado_serviceendpoint_servicefabric":               serviceendpoint.ResourceServiceEndpointServiceFabric(),
			"betterado_serviceendpoint_snyk":                        serviceendpoint.ResourceServiceEndpointSnyk(),
			"betterado_serviceendpoint_sonarcloud":                  serviceendpoint.ResourceServiceEndpointSonarCloud(),
			"betterado_serviceendpoint_sonarqube":                   serviceendpoint.ResourceServiceEndpointSonarQube(),
			"betterado_serviceendpoint_ssh":                         serviceendpoint.ResourceServiceEndpointSSH(),
			"betterado_serviceendpoint_visualstudiomarketplace":     serviceendpoint.ResourceServiceEndpointMarketplace(),
			"betterado_servicehook_permissions":                     permissions.ResourceServiceHookPermissions(),
			"betterado_servicehook_storage_queue_pipelines":         servicehook.ResourceServicehookStorageQueuePipelines(),
			"betterado_servicehook_webhook_tfs":                     servicehook.ResourceServicehookWebhookTfs(),
			"betterado_service_principal_entitlement":               memberentitlementmanagement.ResourceServicePrincipalEntitlement(),
			"betterado_tagging_permissions":                         permissions.ResourceTaggingPermissions(),
			"betterado_team":                                        core.ResourceTeam(),
			"betterado_team_administrators":                         core.ResourceTeamAdministrators(),
			"betterado_team_members":                                core.ResourceTeamMembers(),
			"betterado_user_entitlement":                            memberentitlementmanagement.ResourceUserEntitlement(),
			"betterado_variable_group":                              taskagent.ResourceVariableGroup(),
			"betterado_variable_group_permissions":                  permissions.ResourceVariableGroupPermissions(),
			"betterado_variable_group_variable":                     taskagent.ResourceVariableGroupVariable(),
			"betterado_wiki":                                        wiki.ResourceWiki(),
			"betterado_wiki_page":                                   wiki.ResourceWikiPage(),
			"betterado_workitem":                                    workitemtracking.ResourceWorkItem(),
			"betterado_workitemtracking_field":                      workitemtracking.ResourceField(),
			"betterado_workitemquery_permissions":                   permissions.ResourceWorkItemQueryPermissions(),
			"betterado_workitemquery":                               workitemtracking.ResourceQuery(),
			"betterado_workitemquery_folder":                        workitemtracking.ResourceQueryFolder(),
			"betterado_workitemtrackingprocess_control":             workitemtrackingprocess.ResourceControl(),
			"betterado_workitemtrackingprocess_group":               workitemtrackingprocess.ResourceGroup(),
			"betterado_workitemtrackingprocess_inherited_control":   workitemtrackingprocess.ResourceInheritedControl(),
			"betterado_workitemtrackingprocess_inherited_page":      workitemtrackingprocess.ResourceInheritedPage(),
			"betterado_workitemtrackingprocess_inherited_state":     workitemtrackingprocess.ResourceInheritedState(),
			"betterado_workitemtrackingprocess_list":                workitemtrackingprocess.ResourceList(),
			"betterado_workitemtrackingprocess_page":                workitemtrackingprocess.ResourcePage(),
			"betterado_workitemtrackingprocess_process":             workitemtrackingprocess.ResourceProcess(),
			"betterado_workitemtrackingprocess_process_permissions": permissions.ResourceWorkItemTrackingProcessPermissions(),
			"betterado_workitemtrackingprocess_state":               workitemtrackingprocess.ResourceState(),
			"betterado_workitemtrackingprocess_system_control":      workitemtrackingprocess.ResourceSystemControl(),
			"betterado_workitemtrackingprocess_workitemtype":        workitemtrackingprocess.ResourceWorkItemType(),
			"betterado_workitemtrackingprocess_field":               workitemtrackingprocess.ResourceField(),
			"betterado_workitemtrackingprocess_rule":                workitemtrackingprocess.ResourceRule(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			// NOTE: betterado_release_definition, betterado_release_definition_history,
			// betterado_release_definition_revision, betterado_release_definitions, and
			// betterado_release_folder have been migrated to the terraform-plugin-framework
			// provider (framework_provider.go) and are no longer registered here.
			"betterado_agent_pool":                            taskagent.DataAgentPool(),
			"betterado_agent_pools":                           taskagent.DataAgentPools(),
			"betterado_agent_queue":                           taskagent.DataAgentQueue(),
			"betterado_area":                                  workitemtracking.DataArea(),
			"betterado_build_definition":                      build.DataBuildDefinition(),
			"betterado_client_config":                         service.DataClientConfig(),
			"betterado_descriptor":                            graph.DataDescriptor(),
			"betterado_environment":                           taskagent.DataEnvironment(),
			"betterado_feed":                                  feed.DataFeed(),
			"betterado_git_repositories":                      git.DataGitRepositories(),
			"betterado_git_repository":                        git.DataGitRepository(),
			"betterado_git_repository_file":                   git.DataGitRepositoryFile(),
			"betterado_group":                                 graph.DataGroup(),
			"betterado_group_membership":                      graph.DataGroupMembership(),
			"betterado_groups":                                graph.DataGroups(),
			"betterado_identity_group":                        identity.DataIdentityGroup(),
			"betterado_identity_groups":                       identity.DataIdentityGroups(),
			"betterado_identity_user":                         identity.DataIdentityUser(),
			"betterado_iteration":                             workitemtracking.DataIteration(),
			"betterado_project":                               core.DataProject(),
			"betterado_projects":                              core.DataProjects(),
			"betterado_security_namespace":                    security.DataSecurityNamespace(),
			"betterado_security_namespace_token":              security.DataSecurityNamespaceToken(),
			"betterado_security_namespaces":                   security.DataSecurityNamespaces(),
			"betterado_securityrole_definitions":              securityroles.DataSecurityRoleDefinitions(),
			"betterado_serviceendpoint_generic_v2":            serviceendpoint.DataServiceEndpointGenericV2(),
			"betterado_serviceendpoint_azurecr":               serviceendpoint.DataResourceServiceEndpointAzureCR(),
			"betterado_serviceendpoint_azurerm":               serviceendpoint.DataServiceEndpointAzureRM(),
			"betterado_serviceendpoint_bitbucket":             serviceendpoint.DataResourceServiceEndpointBitbucket(),
			"betterado_serviceendpoint_dockerregistry":        serviceendpoint.DataResourceServiceEndpointDockerRegistry(),
			"betterado_serviceendpoint_github":                serviceendpoint.DataServiceEndpointGithub(),
			"betterado_serviceendpoint_npm":                   serviceendpoint.DataResourceServiceEndpointNpm(),
			"betterado_serviceendpoint_sonarcloud":            serviceendpoint.DataResourceServiceEndpointSonarCloud(),
			"betterado_service_principal":                     graph.DataServicePrincipal(),
			"betterado_storage_key":                           graph.DataStorageKey(),
			"betterado_team":                                  core.DataTeam(),
			"betterado_task_group":                            taskagent.DataTaskGroup(),
			"betterado_teams":                                 core.DataTeams(),
			"betterado_user":                                  graph.DataUser(),
			"betterado_users":                                 graph.DataUsers(),
			"betterado_variable_group":                        taskagent.DataVariableGroup(),
			"betterado_workitemtrackingprocess_process":       workitemtrackingprocess.DataProcess(),
			"betterado_workitemtrackingprocess_processes":     workitemtrackingprocess.DataProcesses(),
			"betterado_workitemtrackingprocess_workitemtype":  workitemtrackingprocess.DataWorkItemType(),
			"betterado_workitemtrackingprocess_workitemtypes": workitemtrackingprocess.DataWorkItemTypes(),
		},
		Schema: map[string]*schema.Schema{
			"org_service_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("AZDO_ORG_SERVICE_URL", nil),
				Description: "The url of the Azure DevOps instance which should be used.",
			},
			"personal_access_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("AZDO_PERSONAL_ACCESS_TOKEN", nil),
				Description: "The personal access token which should be used.",
				Sensitive:   true,
			},
			"client_id": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"ARM_CLIENT_ID", "AZURE_CLIENT_ID"}, nil),
				Description:  "The service principal client id which should be used for AAD auth.",
				ValidateFunc: validation.IsUUID,
			},
			"client_id_file_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_ID_FILE_PATH", nil),
				Description: "The path to a file containing the Client ID which should be used.",
			},
			"tenant_id": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("ARM_TENANT_ID", nil),
				Description:  "The service principal tenant id which should be used for AAD auth.",
				ValidateFunc: validation.IsUUID,
			},
			"auxiliary_tenant_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_AUXILIARY_TENANT_IDS", nil),
				Description: "List of auxiliary Tenant IDs required for multi-tenancy and cross-tenant scenarios.",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsUUID,
				},
			},
			"client_certificate_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_CERTIFICATE_PATH", nil),
				Description: "Path to a certificate to use to authenticate to the service principal.",
			},
			"client_certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_CERTIFICATE", nil),
				Description: "Base64 encoded certificate to use to authenticate to the service principal.",
			},
			"client_certificate_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_CERTIFICATE_PASSWORD", nil),
				Description: "Password for a client certificate password.",
			},
			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_SECRET", nil),
				Description: "Client secret for authenticating to  a service principal.",
			},
			"client_secret_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ARM_CLIENT_SECRET_PATH", "ARM_CLIENT_SECRET_FILE_PATH"}, nil),
				Description: "Path to a file containing a client secret for authenticating to  a service principal.",
			},
			"oidc_request_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ARM_OIDC_REQUEST_TOKEN", "ACTIONS_ID_TOKEN_REQUEST_TOKEN", "SYSTEM_ACCESSTOKEN"}, nil),
				Description: "The bearer token for the request to the OIDC provider. For use when authenticating as a Service Principal using OpenID Connect.",
			},
			"oidc_request_url": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ARM_OIDC_REQUEST_URL", "ACTIONS_ID_TOKEN_REQUEST_URL", "SYSTEM_OIDCREQUESTURI"}, nil),
				Description: "The URL for the OIDC provider from which to request an ID token. For use when authenticating as a Service Principal using OpenID Connect.",
			},
			"oidc_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_OIDC_TOKEN", nil),
				Description: "OIDC token to authenticate as a service principal.",
			},
			"oidc_token_file_path": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ARM_OIDC_TOKEN_FILE_PATH", "AZURE_FEDERATED_TOKEN_FILE"}, nil),
				Description: "OIDC token from file to authenticate as a service principal.",
			},
			"oidc_azure_service_connection_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{"ARM_ADO_PIPELINE_SERVICE_CONNECTION_ID", "ARM_OIDC_AZURE_SERVICE_CONNECTION_ID", "AZURESUBSCRIPTION_SERVICE_CONNECTION_ID"}, nil),
				Description: "The Azure Pipelines Service Connection ID to use for authentication.",
			},
			"use_oidc": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_USE_OIDC", nil),
				Description: "Use an OIDC token to authenticate to a service principal. Defaults to `false`.",
			},
			"use_cli": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_USE_CLI", true),
				Description: "Use Azure CLI to authenticate. Defaults to `true`.",
			},
			"use_msi": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ARM_USE_MSI", nil),
				Description: "Use an Azure Managed Service Identity. Defaults to `false`.",
			},
		},
	}

	p.ConfigureContextFunc = providerConfigure()

	return p
}

func providerConfigure() schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		authProvider, err := GetAuthProvider(ctx, d)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		organizationUrl := d.Get("org_service_url").(string)
		azdoClient, err := client.GetAzdoClient(authProvider, organizationUrl)
		if err != nil {
			return nil, diag.FromErr(clientErrorHandle(err, organizationUrl))
		}

		return azdoClient, diag.FromErr(err)
	}
}

func clientErrorHandle(err error, orgUrl string) error {
	switch err.(type) {
	case azuredevops.WrappedError:
		var wrapperError azuredevops.WrappedError
		if errors.As(err, &wrapperError) {
			if clientError := buildError(*wrapperError.StatusCode, orgUrl); clientError != nil {
				return clientError
			}
		}

	case *azuredevops.WrappedError:
		var wrapperError *azuredevops.WrappedError
		if errors.As(err, &wrapperError) {
			if clientError := buildError(*wrapperError.StatusCode, orgUrl); clientError != nil {
				return clientError
			}
		}
	}
	return err
}

func buildError(statusCode int, orgUrl string) error {
	if statusCode == http.StatusNotFound {
		return fmt.Errorf("Azure DevOps Organization: %s doesn't exist or can't be found. Make sure the URL is correct.", orgUrl)
	} else if statusCode == http.StatusUnauthorized {
		return fmt.Errorf("You are not authorized to access Azure DevOps Organization %s", orgUrl)
	}
	return nil
}

func GetAuthProvider(ctx context.Context, d *schema.ResourceData) (azuredevops.AuthProvider, error) {
	// Personal Access Token
	if personal_access_token, ok := d.GetOk("personal_access_token"); ok {
		return azuredevops.NewAuthProviderPAT(personal_access_token.(string)), nil
	}

	// AAD Authentication
	var auxTenants []string
	for _, tid := range d.Get("auxiliary_tenant_ids").([]any) {
		auxTenants = append(auxTenants, tid.(string))
	}

	cred, err := aztfauth.NewCredential(aztfauth.Option{
		Logger:                     log.New(log.Default().Writer(), "[DEBUG] ", log.LstdFlags|log.Lmsgprefix),
		TenantId:                   d.Get("tenant_id").(string),
		ClientId:                   d.Get("client_id").(string),
		ClientIdFile:               d.Get("client_id_file_path").(string),
		UseClientSecret:            true,
		ClientSecret:               d.Get("client_secret").(string),
		ClientSecretFile:           d.Get("client_secret_path").(string),
		UseClientCert:              true,
		ClientCertBase64:           d.Get("client_certificate").(string),
		ClientCertPfxFile:          d.Get("client_certificate_path").(string),
		ClientCertPassword:         []byte(d.Get("client_certificate_password").(string)),
		UseOIDCToken:               d.Get("use_oidc").(bool),
		OIDCToken:                  d.Get("oidc_token").(string),
		UseOIDCTokenFile:           d.Get("use_oidc").(bool),
		OIDCTokenFile:              d.Get("oidc_token_file_path").(string),
		UseOIDCTokenRequest:        d.Get("use_oidc").(bool),
		OIDCRequestToken:           d.Get("oidc_request_token").(string),
		OIDCRequestURL:             d.Get("oidc_request_url").(string),
		ADOServiceConnectionId:     d.Get("oidc_azure_service_connection_id").(string),
		UseMSI:                     d.Get("use_msi").(bool),
		UseAzureCLI:                d.Get("use_cli").(bool),
		AdditionallyAllowedTenants: auxTenants,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to new credential")
	}

	AzureDevOpsAppDefaultScope := "499b84ac-1321-427f-aa17-267ca6975798/.default"
	ap := azuredevops.NewAuthProviderAAD(cred, policy.TokenRequestOptions{
		Scopes: []string{AzureDevOpsAppDefaultScope},
	})
	return ap, nil
}
