package release

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var _ datasource.DataSource = &releaseFolderDataSource{}

// releaseFolderDataSource is the terraform-plugin-framework data source for
// betterado_release_folder.
type releaseFolderDataSource struct {
	client *client.AggregatedClient
}

// NewReleaseFolderDataSource returns a new framework data source for betterado_release_folder.
func NewReleaseFolderDataSource() datasource.DataSource {
	return &releaseFolderDataSource{}
}

// releaseFolderDataModel is the tfsdk model for the data source.
type releaseFolderDataModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Path        types.String `tfsdk:"path"`
	Description types.String `tfsdk:"description"`
}

func (d *releaseFolderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_release_folder"
}

func (d *releaseFolderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a release folder from an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The path of the release folder (used as the data source ID).",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project the release folder belongs to.",
			},
			"path": schema.StringAttribute{
				Required:    true,
				Description: "The path of the release folder.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the release folder.",
			},
		},
	}
}

func (d *releaseFolderDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *releaseFolderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_release_folder data source Read: provider client not configured")
		return
	}

	var model releaseFolderDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	path := model.Path.ValueString()

	folders, err := d.client.ReleaseClient.GetFolders(d.client.Ctx, releaseapi.GetFoldersArgs{
		Project: &projectID,
		Path:    &path,
	})
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading release folder (path: %s): %s", path, err))
		return
	}
	if folders == nil || len(*folders) == 0 {
		resp.Diagnostics.AddError("Not found", fmt.Sprintf("release folder not found: path %q in project %s", path, projectID))
		return
	}

	folder := (*folders)[0]
	model.ID = types.StringValue(*folder.Path)
	model.Path = types.StringValue(*folder.Path)
	if folder.Description != nil {
		model.Description = types.StringValue(*folder.Description)
	} else {
		model.Description = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
