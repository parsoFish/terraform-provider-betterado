package graph

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	graphapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Ensure interface compliance.
var _ datasource.DataSource = &storageKeyDataSource{}

// storageKeyDataSource is the terraform-plugin-framework data source for
// betterado_storage_key.
type storageKeyDataSource struct {
	client *client.AggregatedClient
}

// NewStorageKeyDataSource returns a new framework data source for betterado_storage_key.
func NewStorageKeyDataSource() datasource.DataSource {
	return &storageKeyDataSource{}
}

// storageKeyDataModel is the tfsdk model for the data source.
type storageKeyDataModel struct {
	ID         types.String `tfsdk:"id"`
	Descriptor types.String `tfsdk:"descriptor"`
	StorageKey types.String `tfsdk:"storage_key"`
}

func (d *storageKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_key"
}

func (d *storageKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Gets the storage key (UUID) for an Azure DevOps object by its descriptor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The storage key UUID (used as the data source ID).",
			},
			"descriptor": schema.StringAttribute{
				Required:    true,
				Description: "The descriptor of the Azure DevOps object.",
			},
			"storage_key": schema.StringAttribute{
				Computed:    true,
				Description: "The storage key (UUID) for the Azure DevOps object.",
			},
		},
	}
}

func (d *storageKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *storageKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_storage_key data source Read: provider client not configured")
		return
	}

	var model storageKeyDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	descriptor := model.Descriptor.ValueString()

	storageKey, err := d.client.GraphClient.GetStorageKey(d.client.Ctx, graphapi.GetStorageKeyArgs{
		SubjectDescriptor: converter.String(descriptor),
	})
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading storage key for descriptor %q: %s", descriptor, err))
		return
	}

	storageKeyStr := storageKey.Value.String()
	model.ID = types.StringValue(storageKeyStr)
	model.StorageKey = types.StringValue(storageKeyStr)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
