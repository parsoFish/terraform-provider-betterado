package graph

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Ensure interface compliance.
var _ datasource.DataSource = &servicePrincipalDataSource{}

// servicePrincipalDataSource is the terraform-plugin-framework data source for
// betterado_service_principal.
type servicePrincipalDataSource struct {
	client *client.AggregatedClient
}

// NewServicePrincipalDataSource returns a new framework data source for
// betterado_service_principal.
func NewServicePrincipalDataSource() datasource.DataSource {
	return &servicePrincipalDataSource{}
}

// servicePrincipalDataSourceModel is the tfsdk model for the data source.
type servicePrincipalDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	OriginID    types.String `tfsdk:"origin_id"`
	Origin      types.String `tfsdk:"origin"`
	Descriptor  types.String `tfsdk:"descriptor"`
}

func (d *servicePrincipalDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_principal"
}

func (d *servicePrincipalDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up an Azure DevOps service principal by display name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The descriptor of the service principal (used as the data source ID).",
			},
			"display_name": schema.StringAttribute{
				Required:    true,
				Description: "The display name of the service principal to look up.",
			},
			"origin_id": schema.StringAttribute{
				Computed:    true,
				Description: "The origin-specific identifier of the service principal.",
			},
			"origin": schema.StringAttribute{
				Computed:    true,
				Description: "The origin of the service principal.",
			},
			"descriptor": schema.StringAttribute{
				Computed:    true,
				Description: "The descriptor of the service principal.",
			},
		},
	}
}

func (d *servicePrincipalDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.AggregatedClient, got %T", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *servicePrincipalDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_service_principal data source Read: provider client not configured")
		return
	}

	var model servicePrincipalDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	displayName := model.DisplayName.ValueString()
	searchFilter := converter.String("General")

	filteredServicePrincipals, err := getIdentityServicePrincipalsWithFilterValue(d.client, searchFilter, displayName)
	if err != nil || filteredServicePrincipals == nil || len(*filteredServicePrincipals) == 0 {
		resp.Diagnostics.AddError("Not found",
			fmt.Sprintf("Finding service principal with display name %q: %v", displayName, err))
		return
	}

	flattenedServicePrincipals, err := flattenIdentityServicePrincipals(filteredServicePrincipals)
	if err != nil {
		resp.Diagnostics.AddError("Read error",
			fmt.Sprintf("Flatten service principals: %v", err))
		return
	}

	targetServicePrincipal := validateIdentityServicePrincipal(flattenedServicePrincipals, displayName)
	if targetServicePrincipal == nil {
		resp.Diagnostics.AddError("Not found",
			fmt.Sprintf("Could not find service principal with name: %s", displayName))
		return
	}

	servicePrincipalDescriptor := targetServicePrincipal.SubjectDescriptor

	servicePrincipal, err := getServicePrincipal(d.client, servicePrincipalDescriptor)
	if err != nil {
		if servicePrincipalDescriptor != nil {
			resp.Diagnostics.AddError("Read error",
				fmt.Sprintf("Finding service principal with Descriptor %s", *servicePrincipalDescriptor))
		} else {
			resp.Diagnostics.AddError("Read error",
				fmt.Sprintf("Finding service principal: %v", err))
		}
		return
	}

	model.ID = strPtrToStringValue(servicePrincipal.Descriptor)
	model.Descriptor = strPtrToStringValue(servicePrincipal.Descriptor)
	model.DisplayName = strPtrToStringValue(servicePrincipal.DisplayName)
	model.OriginID = strPtrToStringValue(servicePrincipal.OriginId)
	model.Origin = strPtrToStringValue(servicePrincipal.Origin)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
