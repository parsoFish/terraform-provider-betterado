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
	_ datasource.DataSource              = (*ServiceEndpointDockerRegistryDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ServiceEndpointDockerRegistryDataSource)(nil)
)

// NewServiceEndpointDockerRegistryDataSource is the constructor for the framework provider.
func NewServiceEndpointDockerRegistryDataSource() datasource.DataSource {
	return &ServiceEndpointDockerRegistryDataSource{}
}

// ServiceEndpointDockerRegistryDataSource is the terraform-plugin-framework implementation
// of the betterado_serviceendpoint_dockerregistry data source.
type ServiceEndpointDockerRegistryDataSource struct {
	client *client.AggregatedClient
}

// serviceEndpointDockerRegistryDataModel is the data source state model.
type serviceEndpointDockerRegistryDataModel struct {
	ProjectID           types.String `tfsdk:"project_id"`
	ServiceEndpointID   types.String `tfsdk:"service_endpoint_id"`
	ServiceEndpointName types.String `tfsdk:"service_endpoint_name"`
	DockerRegistry      types.String `tfsdk:"docker_registry"`
	DockerUsername      types.String `tfsdk:"docker_username"`
	DockerPassword      types.String `tfsdk:"docker_password"`
	DockerEmail         types.String `tfsdk:"docker_email"`
	RegistryType        types.String `tfsdk:"registry_type"`
}

func (d *ServiceEndpointDockerRegistryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serviceendpoint_dockerregistry"
}

func (d *ServiceEndpointDockerRegistryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to access information about an existing Docker Registry service endpoint.",
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
			"docker_registry": schema.StringAttribute{
				Computed:    true,
				Description: "The DockerRegistry registry.",
			},
			"docker_username": schema.StringAttribute{
				Computed:    true,
				Description: "The DockerRegistry username.",
			},
			"docker_password": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The DockerRegistry password.",
			},
			"docker_email": schema.StringAttribute{
				Computed:    true,
				Description: "The DockerRegistry email address.",
			},
			"registry_type": schema.StringAttribute{
				Computed:    true,
				Description: "The registry type (DockerHub or Others).",
			},
		},
	}
}

func (d *ServiceEndpointDockerRegistryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceEndpointDockerRegistryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured.")
		return
	}
	var config serviceEndpointDockerRegistryDataModel
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

	if ep.Authorization != nil && ep.Authorization.Parameters != nil {
		params := *ep.Authorization.Parameters
		if v, ok := params["registry"]; ok {
			config.DockerRegistry = types.StringValue(v)
		}
		if v, ok := params["email"]; ok {
			config.DockerEmail = types.StringValue(v)
		}
		if v, ok := params["username"]; ok {
			config.DockerUsername = types.StringValue(v)
		}
	}

	if ep.Data != nil {
		if v, ok := (*ep.Data)["registrytype"]; ok {
			config.RegistryType = types.StringValue(v)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
