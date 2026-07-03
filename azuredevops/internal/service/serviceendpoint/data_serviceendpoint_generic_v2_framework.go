package serviceendpoint

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/serviceendpoint"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = (*ServiceEndpointGenericV2DataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ServiceEndpointGenericV2DataSource)(nil)
)

// NewServiceEndpointGenericV2DataSource is the constructor for the framework provider.
func NewServiceEndpointGenericV2DataSource() datasource.DataSource {
	return &ServiceEndpointGenericV2DataSource{}
}

// ServiceEndpointGenericV2DataSource is the terraform-plugin-framework implementation
// of the betterado_serviceendpoint_generic_v2 data source.
type ServiceEndpointGenericV2DataSource struct {
	client *client.AggregatedClient
}

// serviceEndpointGenericV2DataModel is the data source state model.
type serviceEndpointGenericV2DataModel struct {
	ProjectID           types.String `tfsdk:"project_id"`
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Type                types.String `tfsdk:"type"`
	Description         types.String `tfsdk:"description"`
	ServerURL           types.String `tfsdk:"server_url"`
	AuthorizationScheme types.String `tfsdk:"authorization_scheme"`
	AuthorizationParams types.Map    `tfsdk:"authorization_parameters"`
	Data                types.Map    `tfsdk:"data"`
}

func (d *ServiceEndpointGenericV2DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_generic_v2"
}

func (d *ServiceEndpointGenericV2DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to access information about an existing Generic V2 Service Connection within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The ID of the service endpoint. Exactly one of id or name must be specified.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the service endpoint. Exactly one of id or name must be specified.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the service endpoint.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The Service Endpoint description.",
			},
			"server_url": schema.StringAttribute{
				Computed:    true,
				Description: "The server URL of the service connection.",
			},
			"authorization_scheme": schema.StringAttribute{
				Computed:    true,
				Description: "The authorization scheme used for the service connection.",
			},
			"authorization_parameters": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "A map of authorization parameters for the service connection.",
			},
			"data": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "A map of endpoint-specific data parameters.",
			},
		},
	}
}

func (d *ServiceEndpointGenericV2DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceEndpointGenericV2DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}

	var config serviceEndpointGenericV2DataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := config.ProjectID.ValueString()

	// Validate: exactly one of id or name must be set
	hasID := !config.ID.IsNull() && !config.ID.IsUnknown() && config.ID.ValueString() != ""
	hasName := !config.Name.IsNull() && !config.Name.IsUnknown() && config.Name.ValueString() != ""
	if !hasID && !hasName {
		resp.Diagnostics.AddError("Configuration error", "Exactly one of 'id' or 'name' must be specified.")
		return
	}

	var ep *serviceendpoint.ServiceEndpoint
	var lookupErr error

	if hasID {
		endpointID, parseErr := uuid.Parse(config.ID.ValueString())
		if parseErr != nil {
			resp.Diagnostics.AddError("Invalid service endpoint ID", parseErr.Error())
			return
		}
		ep, lookupErr = d.client.ServiceEndpointClient.GetServiceEndpointDetails(ctx,
			serviceendpoint.GetServiceEndpointDetailsArgs{
				EndpointId: &endpointID,
				Project:    &projectID,
			})
	} else {
		// Lookup by name
		endpointName := config.Name.ValueString()
		eps, findErr := d.client.ServiceEndpointClient.GetServiceEndpointsByNames(ctx,
			serviceendpoint.GetServiceEndpointsByNamesArgs{
				Project:       &projectID,
				EndpointNames: &[]string{endpointName},
			})
		if findErr != nil {
			resp.Diagnostics.AddError("Error finding service endpoint by name", findErr.Error())
			return
		}
		if eps == nil || len(*eps) == 0 {
			resp.Diagnostics.AddError("Service endpoint not found",
				fmt.Sprintf("No service endpoint with name %q found in project %q", endpointName, projectID))
			return
		}
		ep = &(*eps)[0]
	}

	if lookupErr != nil {
		resp.Diagnostics.AddError("Error reading service endpoint", lookupErr.Error())
		return
	}
	if ep == nil || ep.Id == nil {
		resp.Diagnostics.AddError("Service endpoint not found", "The specified service endpoint does not exist.")
		return
	}

	config.ID = types.StringValue(ep.Id.String())
	if ep.Name != nil {
		config.Name = types.StringValue(*ep.Name)
	}
	if ep.Type != nil {
		config.Type = types.StringValue(*ep.Type)
	}
	if ep.Description != nil {
		config.Description = types.StringValue(*ep.Description)
	}
	if ep.Url != nil {
		config.ServerURL = types.StringValue(*ep.Url)
	}
	if ep.Authorization != nil {
		if ep.Authorization.Scheme != nil {
			config.AuthorizationScheme = types.StringValue(*ep.Authorization.Scheme)
		}
		if ep.Authorization.Parameters != nil {
			m, diags := types.MapValueFrom(ctx, types.StringType, *ep.Authorization.Parameters)
			resp.Diagnostics.Append(diags...)
			if !diags.HasError() {
				config.AuthorizationParams = m
			}
		}
	}
	if config.AuthorizationParams.IsNull() || config.AuthorizationParams.IsUnknown() {
		m, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{})
		config.AuthorizationParams = m
	}

	if ep.Data != nil && len(*ep.Data) > 0 {
		m, diags := types.MapValueFrom(ctx, types.StringType, *ep.Data)
		resp.Diagnostics.Append(diags...)
		if !diags.HasError() {
			config.Data = m
		}
	}
	if config.Data.IsNull() || config.Data.IsUnknown() {
		m, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{})
		config.Data = m
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
