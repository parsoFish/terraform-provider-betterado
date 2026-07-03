package serviceendpoint

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/serviceendpoint"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = (*ServiceEndpointAzureRMDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ServiceEndpointAzureRMDataSource)(nil)
)

// NewServiceEndpointAzureRMDataSource is the constructor for the framework provider.
func NewServiceEndpointAzureRMDataSource() datasource.DataSource {
	return &ServiceEndpointAzureRMDataSource{}
}

// ServiceEndpointAzureRMDataSource is the terraform-plugin-framework implementation
// of the betterado_serviceendpoint_azurerm data source.
type ServiceEndpointAzureRMDataSource struct {
	client *client.AggregatedClient
}

// serviceEndpointAzureRMDataModel is the data source state model.
type serviceEndpointAzureRMDataModel struct {
	ProjectID                           types.String `tfsdk:"project_id"`
	ServiceEndpointID                   types.String `tfsdk:"service_endpoint_id"`
	ServiceEndpointName                 types.String `tfsdk:"service_endpoint_name"`
	AzureRMSpnTenantID                  types.String `tfsdk:"azurerm_spn_tenantid"`
	AzureRMSubscriptionID               types.String `tfsdk:"azurerm_subscription_id"`
	AzureRMSubscriptionName             types.String `tfsdk:"azurerm_subscription_name"`
	AzureRMManagementGroupID            types.String `tfsdk:"azurerm_management_group_id"`
	AzureRMManagementGroupName          types.String `tfsdk:"azurerm_management_group_name"`
	ResourceGroup                       types.String `tfsdk:"resource_group"`
	Environment                         types.String `tfsdk:"environment"`
	ServiceEndpointAuthenticationScheme types.String `tfsdk:"service_endpoint_authentication_scheme"`
	ServerURL                           types.String `tfsdk:"server_url"`
	ServicePrincipalID                  types.String `tfsdk:"service_principal_id"`
	WorkloadIdentityFederationIssuer    types.String `tfsdk:"workload_identity_federation_issuer"`
	WorkloadIdentityFederationSubject   types.String `tfsdk:"workload_identity_federation_subject"`
}

func (d *ServiceEndpointAzureRMDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_azurerm"
}

func (d *ServiceEndpointAzureRMDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to access information about an existing Azure RM service endpoint.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The project ID or project name.",
			},
			"service_endpoint_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The service endpoint ID.",
			},
			"service_endpoint_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The service endpoint name.",
			},
			"azurerm_spn_tenantid": schema.StringAttribute{
				Computed:    true,
				Description: "The service principal tenant id.",
			},
			"azurerm_subscription_id": schema.StringAttribute{
				Computed:    true,
				Description: "The Azure subscription Id.",
			},
			"azurerm_subscription_name": schema.StringAttribute{
				Computed:    true,
				Description: "The Azure subscription name.",
			},
			"azurerm_management_group_id": schema.StringAttribute{
				Computed:    true,
				Description: "The Azure managementGroup Id.",
			},
			"azurerm_management_group_name": schema.StringAttribute{
				Computed:    true,
				Description: "The Azure managementGroup name.",
			},
			"resource_group": schema.StringAttribute{
				Computed:    true,
				Description: "Scope Resource Group.",
			},
			"environment": schema.StringAttribute{
				Computed:    true,
				Description: "Environment (Azure Cloud type).",
			},
			"service_endpoint_authentication_scheme": schema.StringAttribute{
				Computed:    true,
				Description: "The AzureRM Service Endpoint Authentication Scheme.",
			},
			"server_url": schema.StringAttribute{
				Computed:    true,
				Description: "A URL to the server (for AzureStack environments).",
			},
			"service_principal_id": schema.StringAttribute{
				Computed:    true,
				Description: "The service principal ID.",
			},
			"workload_identity_federation_issuer": schema.StringAttribute{
				Computed:    true,
				Description: "The issuer of the workload identity federation service principal.",
			},
			"workload_identity_federation_subject": schema.StringAttribute{
				Computed:    true,
				Description: "The subject of the workload identity federation service principal.",
			},
		},
	}
}

func (d *ServiceEndpointAzureRMDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *client.AggregatedClient, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *ServiceEndpointAzureRMDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var config serviceEndpointAzureRMDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := config.ProjectID.ValueString()

	var ep *serviceendpoint.ServiceEndpoint
	var lookupErr error

	// Lookup by ID if provided, otherwise by name.
	if !config.ServiceEndpointID.IsNull() && !config.ServiceEndpointID.IsUnknown() && config.ServiceEndpointID.ValueString() != "" {
		endpointID, err := uuid.Parse(config.ServiceEndpointID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid service endpoint ID", err.Error())
			return
		}
		ep, lookupErr = d.client.ServiceEndpointClient.GetServiceEndpointDetails(ctx, serviceendpoint.GetServiceEndpointDetailsArgs{
			EndpointId: &endpointID,
			Project:    &projectID,
		})
	} else {
		name := config.ServiceEndpointName.ValueString()
		eps, err := d.client.ServiceEndpointClient.GetServiceEndpointsByNames(ctx, serviceendpoint.GetServiceEndpointsByNamesArgs{
			Project:       &projectID,
			EndpointNames: &[]string{name},
		})
		if err != nil {
			resp.Diagnostics.AddError("Error looking up service endpoint by name", err.Error())
			return
		}
		if eps == nil || len(*eps) == 0 {
			resp.Diagnostics.AddError("Service endpoint not found",
				fmt.Sprintf("No service endpoint with name %q found in project %q", name, projectID))
			return
		}
		ep = &(*eps)[0]
		lookupErr = nil
	}

	if lookupErr != nil {
		resp.Diagnostics.AddError("Error reading service endpoint", lookupErr.Error())
		return
	}
	if ep == nil || ep.Id == nil {
		resp.Diagnostics.AddError("Service endpoint not found", "The service endpoint was not found.")
		return
	}

	config.ServiceEndpointID = types.StringValue(ep.Id.String())
	if ep.Name != nil {
		config.ServiceEndpointName = types.StringValue(*ep.Name)
	}

	if ep.Authorization != nil && ep.Authorization.Scheme != nil {
		config.ServiceEndpointAuthenticationScheme = types.StringValue(*ep.Authorization.Scheme)
	}

	if ep.Authorization != nil && ep.Authorization.Parameters != nil {
		params := *ep.Authorization.Parameters
		if v, ok := params["tenantid"]; ok {
			config.AzureRMSpnTenantID = types.StringValue(v)
		}
		if v, ok := params["serviceprincipalid"]; ok {
			config.ServicePrincipalID = types.StringValue(v)
		}
		if v, ok := params["workloadIdentityFederationIssuer"]; ok {
			config.WorkloadIdentityFederationIssuer = types.StringValue(v)
		}
		if v, ok := params["workloadIdentityFederationSubject"]; ok {
			config.WorkloadIdentityFederationSubject = types.StringValue(v)
		}
		if scopeVal, ok := params["scope"]; ok {
			parts := strings.Split(scopeVal, "/")
			if len(parts) == 5 {
				config.ResourceGroup = types.StringValue(parts[4])
			}
		}
	}

	if ep.Data != nil {
		data := *ep.Data
		if v, ok := data["environment"]; ok {
			config.Environment = types.StringValue(v)
		}
		if v, ok := data["subscriptionId"]; ok {
			config.AzureRMSubscriptionID = types.StringValue(v)
		}
		if v, ok := data["subscriptionName"]; ok {
			config.AzureRMSubscriptionName = types.StringValue(v)
		}
		if v, ok := data["managementGroupId"]; ok {
			config.AzureRMManagementGroupID = types.StringValue(v)
		}
		if v, ok := data["managementGroupName"]; ok {
			config.AzureRMManagementGroupName = types.StringValue(v)
		}
	}

	if ep.Url != nil {
		config.ServerURL = types.StringValue(*ep.Url)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
