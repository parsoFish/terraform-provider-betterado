package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/accounts"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/approvalsandchecks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/build"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/core"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/dashboard"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/extension"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/extensionmanagement"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/featuremanagement"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/feed"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/git"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/graph"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/identity"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/memberentitlementmanagement"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/notification"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/permissions"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/pipelines"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/pipelinesapproval"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/policy/branch"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/policy/repository"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/profile"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/release"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/security"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/securityroles"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/serviceendpoint"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/servicehook"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/testplan"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/wiki"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/workitemtracking"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/workitemtrackingprocess"
)

// BetteradoFrameworkProvider is the terraform-plugin-framework provider stub.
// It is multiplexed with the existing SDKv2 provider in main.go so that
// framework resources can be registered here without touching main.go.
type BetteradoFrameworkProvider struct{}

// NewFrameworkProvider returns a provider.Provider for use in the mux setup.
func NewFrameworkProvider() provider.Provider {
	return &BetteradoFrameworkProvider{}
}

func (p *BetteradoFrameworkProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "betterado"
}

// Schema mirrors every attribute from the SDKv2 provider schema exactly.
// The tf6muxserver mux requires both providers to expose an identical schema;
// any mismatch causes "Invalid Provider Server Combination" at plan time.
// This schema is intentionally read-only here — Configure() reads credentials
// from env vars directly; the SDKv2 provider handles the actual HCL block.
func (p *BetteradoFrameworkProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"org_service_url": schema.StringAttribute{
				Optional:    true,
				Description: "The url of the Azure DevOps instance which should be used.",
			},
			"personal_access_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The personal access token which should be used.",
			},
			"client_id": schema.StringAttribute{
				Optional:    true,
				Description: "The service principal client id which should be used for AAD auth.",
			},
			"client_id_file_path": schema.StringAttribute{
				Optional:    true,
				Description: "The path to a file containing the Client ID which should be used.",
			},
			"tenant_id": schema.StringAttribute{
				Optional:    true,
				Description: "The service principal tenant id which should be used for AAD auth.",
			},
			"auxiliary_tenant_ids": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of auxiliary Tenant IDs required for multi-tenancy and cross-tenant scenarios.",
			},
			"client_certificate_path": schema.StringAttribute{
				Optional:    true,
				Description: "Path to a certificate to use to authenticate to the service principal.",
			},
			"client_certificate": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Base64 encoded certificate to use to authenticate to the service principal.",
			},
			"client_certificate_password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password for a client certificate password.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Client secret for authenticating to  a service principal.",
			},
			"client_secret_path": schema.StringAttribute{
				Optional:    true,
				Description: "Path to a file containing a client secret for authenticating to  a service principal.",
			},
			"oidc_request_token": schema.StringAttribute{
				Optional:    true,
				Description: "The bearer token for the request to the OIDC provider. For use when authenticating as a Service Principal using OpenID Connect.",
			},
			"oidc_request_url": schema.StringAttribute{
				Optional:    true,
				Description: "The URL for the OIDC provider from which to request an ID token. For use when authenticating as a Service Principal using OpenID Connect.",
			},
			"oidc_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "OIDC token to authenticate as a service principal.",
			},
			"oidc_token_file_path": schema.StringAttribute{
				Optional:    true,
				Description: "OIDC token from file to authenticate as a service principal.",
			},
			"oidc_azure_service_connection_id": schema.StringAttribute{
				Optional:    true,
				Description: "The Azure Pipelines Service Connection ID to use for authentication.",
			},
			"use_oidc": schema.BoolAttribute{
				Optional:    true,
				Description: "Use an OIDC token to authenticate to a service principal. Defaults to `false`.",
			},
			"use_cli": schema.BoolAttribute{
				Optional:    true,
				Description: "Use Azure CLI to authenticate. Defaults to `true`.",
			},
			"use_msi": schema.BoolAttribute{
				Optional:    true,
				Description: "Use an Azure Managed Service Identity. Defaults to `false`.",
			},
		},
	}
}

// Configure builds the AggregatedClient and stores it as resp.ResourceData /
// resp.DataSourceData so framework resources (e.g. TaskGroupResource,
// ReleaseDefinitionResource) can retrieve it via ConfigureRequest.ProviderData.
//
// The mux multiplexes the provider Configure call across both the SDKv2 provider
// and this framework provider, passing the SAME parsed HCL provider config to
// each. Credentials are read from the HCL provider block (org_service_url +
// personal_access_token) first, falling back to the canonical AZDO_* environment
// variables when an attribute is null/unset — matching the SDKv2 provider's
// EnvDefaultFunc behaviour. The two providers share no runtime state, so this one
// builds its own client.
func (p *BetteradoFrameworkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Read credentials from the HCL provider config. GetAttribute leaves the
	// value null when the attribute is unset/unknown, in which case ValueString
	// returns "" and we fall back to the environment variable.
	var orgURLCfg, patCfg types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("org_service_url"), &orgURLCfg)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("personal_access_token"), &patCfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgURL := orgURLCfg.ValueString()
	if orgURL == "" {
		orgURL = os.Getenv("AZDO_ORG_SERVICE_URL")
	}
	pat := patCfg.ValueString()
	if pat == "" {
		pat = os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	}

	// If PAT credentials are unavailable (e.g. an unknown value during plan, or
	// AAD/OIDC auth handled by the SDKv2 provider) leave resp.ResourceData nil
	// and let an individual resource's Configure report the error only when it
	// actually needs the client. This keeps the provider Configure step from
	// failing for non-PAT auth flows.
	if orgURL == "" || pat == "" {
		return
	}

	authProvider := azuredevops.NewAuthProviderPAT(pat)
	agg, err := client.GetAzdoClient(authProvider, orgURL)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to configure Azure DevOps client",
			err.Error(),
		)
		return
	}

	resp.ResourceData = agg
	resp.DataSourceData = agg
}

// FRAMEWORK EXTENSION POINT
//
// To register a new terraform-plugin-framework resource, add it to the slice
// returned by Resources() below:
//
//	func (p *BetteradoFrameworkProvider) Resources(_ context.Context) []func() resource.Resource {
//	    return []func() resource.Resource{
//	        newMyFrameworkResource,   // ← add here
//	    }
//	}
//
// To register a new framework data source, add it to DataSources():
//
//	func (p *BetteradoFrameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
//	    return []func() datasource.DataSource{
//	        newMyFrameworkDataSource, // ← add here
//	    }
//	}
//
// No changes to main.go are needed after the mux is wired in this file.

func (p *BetteradoFrameworkProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		dashboard.NewDashboardResource,
		extension.NewExtensionResource,
		taskagent.NewTaskGroupResource,
		release.NewReleaseDefinitionResource,
		release.NewReleaseFolderResource,
		permissions.NewReleaseDefinitionPermissionsResource,
		build.NewBuildFolderResource,
		build.NewBuildDefinitionResource,
		build.NewPipelineAuthorizationResource,
		build.NewResourceAuthorizationResource,
		memberentitlementmanagement.NewUserEntitlementResource,
		memberentitlementmanagement.NewGroupEntitlementResource,
		memberentitlementmanagement.NewServicePrincipalEntitlementResource,
		featuremanagement.NewFeatureFlagResource,
		git.NewGitRepositoryResource,
		git.NewGitRepositoryBranchResource,
		git.NewGitRepositoryFileResource,
		servicehook.NewServicehookStorageQueuePipelinesResource,
		servicehook.NewServicehookWebhookTfsResource,
		wiki.NewWikiResource,
		wiki.NewWikiPageResource,
		serviceendpoint.NewServiceEndpointArgoCDResource,
		serviceendpoint.NewServiceEndpointArtifactoryResource,
		serviceendpoint.NewServiceEndpointAwsResource,
		serviceendpoint.NewServiceEndpointAzureServiceBusResource,
		serviceendpoint.NewServiceEndpointAzureCRResource,
		serviceendpoint.NewServiceEndpointAzureDevOpsResource,
		serviceendpoint.NewServiceEndpointAzureRMResource,
		serviceendpoint.NewServiceEndpointBitBucketResource,
		serviceendpoint.NewServiceEndpointBlackDuckResource,
		serviceendpoint.NewServiceEndpointCheckMarxOneResource,
		serviceendpoint.NewServiceEndpointCheckMarxSASTResource,
		serviceendpoint.NewServiceEndpointCheckMarxSCAResource,
		serviceendpoint.NewServiceEndpointDockerRegistryResource,
		serviceendpoint.NewServiceEndpointDynamicsLifecycleServicesResource,
		serviceendpoint.NewServiceEndpointExternalTFSResource,
		serviceendpoint.NewServiceEndpointGcpTerraformResource,
		serviceendpoint.NewServiceEndpointGenericResource,
		serviceendpoint.NewServiceEndpointGenericGitResource,
		serviceendpoint.NewServiceEndpointGenericV2Resource,
		serviceendpoint.NewServiceEndpointGitHubEnterpriseResource,
		serviceendpoint.NewServiceEndpointGitHubResource,
		serviceendpoint.NewServiceEndpointGitLabResource,
		serviceendpoint.NewServiceEndpointIncomingWebhookResource,
		serviceendpoint.NewServiceEndpointJenkinsResource,
		serviceendpoint.NewServiceEndpointJFrogArtifactoryV2Resource,
		serviceendpoint.NewServiceEndpointJFrogDistributionV2Resource,
		serviceendpoint.NewServiceEndpointJFrogPlatformV2Resource,
		serviceendpoint.NewServiceEndpointJFrogXRayV2Resource,
		serviceendpoint.NewServiceEndpointKubernetesResource,
		serviceendpoint.NewServiceEndpointMavenResource,
		serviceendpoint.NewServiceEndpointNexusResource,
		serviceendpoint.NewServiceEndpointNuGetResource,
		serviceendpoint.NewServiceEndpointOctopusDeployResource,
		serviceendpoint.NewServiceEndpointOpenshiftResource,
		serviceendpoint.NewServiceEndpointRunPipelineResource,
		serviceendpoint.NewServiceEndpointServiceFabricResource,
		serviceendpoint.NewServiceEndpointSnykResource,
		serviceendpoint.NewServiceEndpointSonarQubeResource,
		serviceendpoint.NewServiceEndpointSSHResource,
		serviceendpoint.NewServiceEndpointVisualStudioMarketplaceResource,
		feed.NewFeedResource,
		feed.NewFeedPermissionResource,
		feed.NewFeedRetentionPolicyResource,
		graph.NewGroupResource,
		graph.NewGroupMembershipResource,
		workitemtracking.NewWorkItemResource,
		workitemtracking.NewFieldResource,
		workitemtracking.NewWorkItemQueryResource,
		workitemtracking.NewWorkItemQueryFolderResource,
		notification.NewNotificationSubscriptionResource,
		pipelines.NewPipelineResource,
		branch.NewAutoReviewersResource,
		branch.NewBuildValidationResource,
		branch.NewCommentResolutionResource,
		branch.NewMergeTypesResource,
		branch.NewMinReviewersResource,
		branch.NewStatusCheckResource,
		branch.NewWorkItemLinkingResource,
		repository.NewAuthorEmailPatternsResource,
		repository.NewFilePathPatternsResource,
		repository.NewEnforceConsistentCaseResource,
		repository.NewCheckCredentialsResource,
		repository.NewReservedNamesResource,
		repository.NewMaxPathLengthResource,
		repository.NewMaxFileSizeResource,
		approvalsandchecks.NewApprovalResource,
		approvalsandchecks.NewBranchControlResource,
		approvalsandchecks.NewBusinessHoursResource,
		approvalsandchecks.NewExclusiveLockResource,
		approvalsandchecks.NewRequiredTemplateResource,
		approvalsandchecks.NewRestAPIResource,
		permissions.NewAreaPermissionsResource,
		permissions.NewBuildDefinitionPermissionsResource,
		permissions.NewBuildFolderPermissionsResource,
		permissions.NewGitPermissionsResource,
		permissions.NewIterationPermissionsResource,
		permissions.NewLibraryPermissionsResource,
		permissions.NewServiceEndpointPermissionsResource,
		permissions.NewServiceHookPermissionsResource,
		permissions.NewTaggingPermissionsResource,
		permissions.NewVariableGroupPermissionsResource,
		permissions.NewWorkItemQueryPermissionsResource,
		permissions.NewWorkItemTrackingProcessPermissionsResource,
		permissions.NewProjectPermissionsResource,
		security.NewSecurityPermissionsResource,
		securityroles.NewSecurityRoleAssignmentResource,
		testplan.NewTestPlanResource,
		testplan.NewTestSuiteResource,
		testplan.NewTestConfigurationResource,
		testplan.NewTestVariableResource,
		pipelinesapproval.NewPipelineApprovalResource,
		extensionmanagement.NewExtensionInstallResource,
		core.NewProjectResource,
		core.NewProjectFeaturesResource,
		core.NewProjectPipelineSettingsResource,
		core.NewProjectTagsResource,
		core.NewTeamResource,
		core.NewTeamAdministratorsResource,
		core.NewTeamMembersResource,
		workitemtrackingprocess.NewProcessResource,
		workitemtrackingprocess.NewWorkItemTypeResource,
		workitemtrackingprocess.NewStateResource,
		workitemtrackingprocess.NewInheritedStateResource,
		workitemtrackingprocess.NewPageResource,
		workitemtrackingprocess.NewInheritedPageResource,
		workitemtrackingprocess.NewGroupResource,
		workitemtrackingprocess.NewControlResource,
		workitemtrackingprocess.NewInheritedControlResource,
		workitemtrackingprocess.NewSystemControlResource,
		workitemtrackingprocess.NewListResource,
		workitemtrackingprocess.NewFieldResource,
		workitemtrackingprocess.NewRuleResource,
		taskagent.NewAgentPoolResource,
		taskagent.NewAgentQueueResource,
		taskagent.NewEnvironmentResource,
		taskagent.NewEnvironmentResourceKubernetesResource,
		taskagent.NewDeploymentGroupResource,
		taskagent.NewElasticPoolResource,
		taskagent.NewVariableGroupResource,
		taskagent.NewVariableGroupVariableResource,
	}
}

func (p *BetteradoFrameworkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		release.NewReleaseDefinitionDataSource,
		release.NewReleaseDefinitionHistoryDataSource,
		release.NewReleaseDefinitionRevisionDataSource,
		release.NewReleaseDefinitionsDataSource,
		release.NewReleaseFolderDataSource,
		build.NewBuildDefinitionDataSource,
		featuremanagement.NewFeatureFlagDataSource,
		serviceendpoint.NewServiceEndpointAzureCRDataSource,
		serviceendpoint.NewServiceEndpointAzureRMDataSource,
		serviceendpoint.NewServiceEndpointBitBucketDataSource,
		serviceendpoint.NewServiceEndpointDockerRegistryDataSource,
		serviceendpoint.NewServiceEndpointGenericV2DataSource,
		serviceendpoint.NewServiceEndpointGitHubDataSource,
		serviceendpoint.NewServiceEndpointNpmDataSource,
		serviceendpoint.NewServiceEndpointSonarCloudDataSource,
		git.NewGitRepositoryDataSource,
		git.NewGitRepositoriesDataSource,
		git.NewGitRepositoryFileDataSource,
		feed.NewFeedDataSource,
		graph.NewDescriptorDataSource,
		graph.NewStorageKeyDataSource,
		graph.NewGroupDataSource,
		graph.NewGroupMembershipDataSource,
		graph.NewUserDataSource,
		graph.NewUsersDataSource,
		graph.NewGroupsDataSource,
		graph.NewServicePrincipalDataSource,
		identity.NewIdentityGroupDataSource,
		identity.NewIdentityGroupsDataSource,
		identity.NewIdentityUserDataSource,
		workitemtracking.NewAreaDataSource,
		workitemtracking.NewIterationDataSource,
		notification.NewNotificationSubscriptionDataSource,
		pipelines.NewPipelineDataSource,
		pipelines.NewPipelineRunDataSource,
		security.NewSecurityNamespaceDataSource,
		security.NewSecurityNamespaceTokenDataSource,
		security.NewSecurityNamespacesDataSource,
		securityroles.NewSecurityRoleDefinitionsDataSource,
		accounts.NewAccountsDataSource,
		profile.NewProfileDataSource,
		testplan.NewTestPlanDataSource,
		testplan.NewTestRunDataSource,
		testplan.NewTestResultDataSource,
		pipelinesapproval.NewPipelineApprovalsDataSource,
		core.NewProjectDataSource,
		core.NewProjectsDataSource,
		core.NewTeamDataSource,
		core.NewTeamsDataSource,
		service.NewClientConfigDataSource,
		workitemtrackingprocess.NewProcessDataSource,
		workitemtrackingprocess.NewProcessesDataSource,
		workitemtrackingprocess.NewWorkItemTypeDataSource,
		workitemtrackingprocess.NewWorkItemTypesDataSource,
		taskagent.NewTaskGroupDataSource,
		taskagent.NewAgentPoolDataSource,
		taskagent.NewAgentPoolsDataSource,
		taskagent.NewAgentQueueDataSource,
		taskagent.NewEnvironmentDataSource,
		taskagent.NewVariableGroupDataSource,
	}
}
