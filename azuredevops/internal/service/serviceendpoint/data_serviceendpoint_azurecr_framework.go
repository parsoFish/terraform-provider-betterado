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
	_ datasource.DataSource              = (*ServiceEndpointAzureCRDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ServiceEndpointAzureCRDataSource)(nil)
)

// NewServiceEndpointAzureCRDataSource is the constructor for the framework provider.
func NewServiceEndpointAzureCRDataSource() datasource.DataSource {
	return &ServiceEndpointAzureCRDataSource{}
}

// ServiceEndpointAzureCRDataSource is the terraform-plugin-framework implementation
// of the betterado_serviceendpoint_azurecr data source.
type ServiceEndpointAzureCRDataSource struct {
	client *client.AggregatedClient
}

// serviceEndpointAzureCRDataModel is the data source state model.
type serviceEndpointAzureCRDataModel struct {
	ProjectID                           types.String `tfsdk:"project_id"`
	ServiceEndpointID                   types.String `tfsdk:"service_endpoint_id"`
	ServiceEndpointName                 types.String `tfsdk:"service_endpoint_name"`
	AzureCRSpnTenantID                  types.String `tfsdk:"azurecr_spn_tenantid"`
	AzureCRSubscriptionID               types.String `tfsdk:"azurecr_subscription_id"`
	AzureCRSubscriptionName             types.String `tfsdk:"azurecr_subscription_name"`
	ResourceGroup                       types.String `tfsdk:"resource_group"`
	AzureCRName                         types.String `tfsdk:"azurecr_name"`
	AppObjectID                         types.String `tfsdk:"app_object_id"`
	SpnObjectID                         types.String `tfsdk:"spn_object_id"`
	AzSpnRoleAssignmentID               types.String `tfsdk:"az_spn_role_assignment_id"`
	AzSpnRolePermissions                types.String `tfsdk:"az_spn_role_permissions"`
	ServicePrincipalID                  types.String `tfsdk:"service_principal_id"`
	ServiceEndpointAuthenticationScheme types.String `tfsdk:"service_endpoint_authentication_scheme"`
}

func (d *ServiceEndpointAzureCRDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_azurecr"
}

func (d *ServiceEndpointAzureCRDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to access information about an existing Azure Container Registry service endpoint.",
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
			"azurecr_spn_tenantid": schema.StringAttribute{
				Computed:    true,
				Description: "The service principal tenant id.",
			},
			"azurecr_subscription_id": schema.StringAttribute{
				Computed:    true,
				Description: "The Azure subscription Id.",
			},
			"azurecr_subscription_name": schema.StringAttribute{
				Computed:    true,
				Description: "The Azure subscription name.",
			},
			"resource_group": schema.StringAttribute{
				Computed:    true,
				Description: "Scope Resource Group.",
			},
			"azurecr_name": schema.StringAttribute{
				Computed:    true,
				Description: "The AzureContainerRegistry name.",
			},
			"app_object_id": schema.StringAttribute{
				Computed:    true,
				Description: "The object ID of the service principal's application.",
			},
			"spn_object_id": schema.StringAttribute{
				Computed:    true,
				Description: "The object ID of the service principal.",
			},
			"az_spn_role_assignment_id": schema.StringAttribute{
				Computed:    true,
				Description: "The service principal role assignment ID.",
			},
			"az_spn_role_permissions": schema.StringAttribute{
				Computed:    true,
				Description: "The service principal role permissions.",
			},
			"service_principal_id": schema.StringAttribute{
				Computed:    true,
				Description: "The service principal ID.",
			},
			"service_endpoint_authentication_scheme": schema.StringAttribute{
				Computed:    true,
				Description: "The authentication scheme.",
			},
		},
	}
}

func (d *ServiceEndpointAzureCRDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceEndpointAzureCRDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var config serviceEndpointAzureCRDataModel
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
		if v, ok := params["tenantId"]; ok {
			config.AzureCRSpnTenantID = types.StringValue(v)
		}
		if v, ok := params["serviceprincipalid"]; ok {
			config.ServicePrincipalID = types.StringValue(v)
		}
		if scopeVal, ok := params["scope"]; ok {
			s := strings.Split(scopeVal, "/")
			if len(s) >= 9 {
				config.ResourceGroup = types.StringValue(s[4])
				config.AzureCRName = types.StringValue(s[8])
			}
		}
	}

	if ep.Data != nil {
		data := *ep.Data
		if v, ok := data["subscriptionId"]; ok {
			config.AzureCRSubscriptionID = types.StringValue(v)
		}
		if v, ok := data["subscriptionName"]; ok {
			config.AzureCRSubscriptionName = types.StringValue(v)
		}
		if v, ok := data["appObjectId"]; ok {
			config.AppObjectID = types.StringValue(v)
		}
		if v, ok := data["spnObjectId"]; ok {
			config.SpnObjectID = types.StringValue(v)
		}
		if v, ok := data["azureSpnPermissions"]; ok {
			config.AzSpnRolePermissions = types.StringValue(v)
		}
		if v, ok := data["azureSpnRoleAssignmentId"]; ok {
			config.AzSpnRoleAssignmentID = types.StringValue(v)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
