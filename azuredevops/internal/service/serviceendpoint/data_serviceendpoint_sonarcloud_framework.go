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
	_ datasource.DataSource              = (*ServiceEndpointSonarCloudDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ServiceEndpointSonarCloudDataSource)(nil)
)

// NewServiceEndpointSonarCloudDataSource is the constructor for the framework provider.
func NewServiceEndpointSonarCloudDataSource() datasource.DataSource {
	return &ServiceEndpointSonarCloudDataSource{}
}

// ServiceEndpointSonarCloudDataSource is the terraform-plugin-framework implementation
// of the betterado_serviceendpoint_sonarcloud data source.
type ServiceEndpointSonarCloudDataSource struct {
	client *client.AggregatedClient
}

// serviceEndpointSonarCloudDataModel is the data source state model.
type serviceEndpointSonarCloudDataModel struct {
	ProjectID           types.String `tfsdk:"project_id"`
	ServiceEndpointID   types.String `tfsdk:"service_endpoint_id"`
	ServiceEndpointName types.String `tfsdk:"service_endpoint_name"`
}

func (d *ServiceEndpointSonarCloudDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_sonarcloud"
}

func (d *ServiceEndpointSonarCloudDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to access information about an existing SonarCloud Service Connection.",
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
		},
	}
}

func (d *ServiceEndpointSonarCloudDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceEndpointSonarCloudDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var config serviceEndpointSonarCloudDataModel
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
