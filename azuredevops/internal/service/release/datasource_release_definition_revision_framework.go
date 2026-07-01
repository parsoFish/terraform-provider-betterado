package release

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var _ datasource.DataSource = &releaseDefinitionRevisionDataSource{}

// releaseDefinitionRevisionDataSource is the terraform-plugin-framework data source for
// betterado_release_definition_revision.
type releaseDefinitionRevisionDataSource struct {
	client *client.AggregatedClient
}

// NewReleaseDefinitionRevisionDataSource returns a new framework data source for
// betterado_release_definition_revision.
func NewReleaseDefinitionRevisionDataSource() datasource.DataSource {
	return &releaseDefinitionRevisionDataSource{}
}

// releaseDefinitionRevisionDataModel is the tfsdk model for the data source.
type releaseDefinitionRevisionDataModel struct {
	ID                  types.String `tfsdk:"id"`
	ProjectID           types.String `tfsdk:"project_id"`
	ReleaseDefinitionID types.Int64  `tfsdk:"release_definition_id"`
	Revision            types.Int64  `tfsdk:"revision"`
	JSONContent         types.String `tfsdk:"json_content"`
}

func (d *releaseDefinitionRevisionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_release_definition_revision"
}

func (d *releaseDefinitionRevisionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a specific revision of a release definition from Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "A stable ID for this data source.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
			},
			"release_definition_id": schema.Int64Attribute{
				Required:    true,
				Description: "The numeric ID of the release definition.",
			},
			"revision": schema.Int64Attribute{
				Required:    true,
				Description: "The revision number to fetch.",
			},
			"json_content": schema.StringAttribute{
				Computed:    true,
				Description: "The JSON content of the release definition at the specified revision.",
			},
		},
	}
}

func (d *releaseDefinitionRevisionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *releaseDefinitionRevisionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_release_definition_revision data source Read: provider client not configured")
		return
	}

	var model releaseDefinitionRevisionDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	defID := int(model.ReleaseDefinitionID.ValueInt64())
	rev := int(model.Revision.ValueInt64())

	rc, err := d.client.ReleaseClient.GetDefinitionRevision(d.client.Ctx, releaseapi.GetDefinitionRevisionArgs{
		Project:      &projectID,
		DefinitionId: &defID,
		Revision:     &rev,
	})
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading release definition revision (def: %d, rev: %d): %+v", defID, rev, err))
		return
	}
	defer rc.Close()

	raw, err := io.ReadAll(rc)
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("draining release definition revision response (def: %d, rev: %d): %+v", defID, rev, err))
		return
	}

	model.ID = types.StringValue(fmt.Sprintf("%s-%d-%d", projectID, defID, rev))
	model.JSONContent = types.StringValue(string(raw))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
