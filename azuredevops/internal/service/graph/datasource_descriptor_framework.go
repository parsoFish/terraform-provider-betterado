package graph

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	graphapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var _ datasource.DataSource = &descriptorDataSource{}

// descriptorDataSource is the terraform-plugin-framework data source for
// betterado_descriptor.
type descriptorDataSource struct {
	client *client.AggregatedClient
}

// NewDescriptorDataSource returns a new framework data source for betterado_descriptor.
func NewDescriptorDataSource() datasource.DataSource {
	return &descriptorDataSource{}
}

// descriptorDataModel is the tfsdk model for the data source.
type descriptorDataModel struct {
	ID         types.String `tfsdk:"id"`
	StorageKey types.String `tfsdk:"storage_key"`
	Descriptor types.String `tfsdk:"descriptor"`
}

func (d *descriptorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_descriptor"
}

func (d *descriptorDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Gets the descriptor for an Azure DevOps object by its storage key (UUID).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The storage key (used as the data source ID).",
			},
			"storage_key": schema.StringAttribute{
				Required:    true,
				Description: "The storage key (UUID) of the Azure DevOps object.",
			},
			"descriptor": schema.StringAttribute{
				Computed:    true,
				Description: "The descriptor for the Azure DevOps object.",
			},
		},
	}
}

func (d *descriptorDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *descriptorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_descriptor data source Read: provider client not configured")
		return
	}

	var model descriptorDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storageKey := model.StorageKey.ValueString()
	storageKeyUUID, err := uuid.Parse(storageKey)
	if err != nil {
		resp.Diagnostics.AddError("Invalid storage key", fmt.Sprintf("Invalid storage key %q: %s", storageKey, err))
		return
	}

	descriptor, err := d.client.GraphClient.GetDescriptor(d.client.Ctx, graphapi.GetDescriptorArgs{
		StorageKey: &storageKeyUUID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading descriptor for storage key %q: %s", storageKey, err))
		return
	}

	model.ID = types.StringValue(storageKey)
	model.Descriptor = types.StringValue(*descriptor.Value)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
