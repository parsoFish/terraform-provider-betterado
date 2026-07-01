package release

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var _ datasource.DataSource = &releaseDefinitionHistoryDataSource{}

// releaseDefinitionHistoryDataSource is the terraform-plugin-framework data source for
// betterado_release_definition_history.
type releaseDefinitionHistoryDataSource struct {
	client *client.AggregatedClient
}

// NewReleaseDefinitionHistoryDataSource returns a new framework data source for
// betterado_release_definition_history.
func NewReleaseDefinitionHistoryDataSource() datasource.DataSource {
	return &releaseDefinitionHistoryDataSource{}
}

// releaseRevisionItem is one element in the revisions list.
type releaseRevisionItem struct {
	Revision    types.Int64  `tfsdk:"revision"`
	ChangedBy   types.String `tfsdk:"changed_by"`
	ChangedDate types.String `tfsdk:"changed_date"`
	ChangeType  types.String `tfsdk:"change_type"`
	Comment     types.String `tfsdk:"comment"`
}

// releaseDefinitionHistoryDataModel is the tfsdk model for the data source.
type releaseDefinitionHistoryDataModel struct {
	ID                  types.String          `tfsdk:"id"`
	ProjectID           types.String          `tfsdk:"project_id"`
	ReleaseDefinitionID types.Int64           `tfsdk:"release_definition_id"`
	Revisions           []releaseRevisionItem `tfsdk:"revisions"`
}

func (d *releaseDefinitionHistoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_release_definition_history"
}

func (d *releaseDefinitionHistoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the revision history of a release definition from Azure DevOps.",
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
			"revisions": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of revisions for the release definition.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"revision": schema.Int64Attribute{
							Computed:    true,
							Description: "The revision number.",
						},
						"changed_by": schema.StringAttribute{
							Computed:    true,
							Description: "The display name of the user who made the change.",
						},
						"changed_date": schema.StringAttribute{
							Computed:    true,
							Description: "The date/time of the change in RFC3339 format.",
						},
						"change_type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of change (add, update, delete).",
						},
						"comment": schema.StringAttribute{
							Computed:    true,
							Description: "The comment associated with the revision.",
						},
					},
				},
			},
		},
	}
}

func (d *releaseDefinitionHistoryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *releaseDefinitionHistoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_release_definition_history data source Read: provider client not configured")
		return
	}

	var model releaseDefinitionHistoryDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	defID := int(model.ReleaseDefinitionID.ValueInt64())

	revisions, err := d.client.ReleaseClient.GetReleaseDefinitionHistory(d.client.Ctx, releaseapi.GetReleaseDefinitionHistoryArgs{
		Project:      &projectID,
		DefinitionId: &defID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading release definition history (project: %s, def: %d): %+v", projectID, defID, err))
		return
	}

	items := make([]releaseRevisionItem, 0)
	if revisions != nil {
		for _, rev := range *revisions {
			item := releaseRevisionItem{
				Revision:    types.Int64Value(0),
				ChangedBy:   types.StringValue(""),
				ChangedDate: types.StringValue(""),
				ChangeType:  types.StringValue(""),
				Comment:     types.StringValue(""),
			}
			if rev.Revision != nil {
				item.Revision = types.Int64Value(int64(*rev.Revision))
			}
			if rev.ChangedBy != nil && rev.ChangedBy.DisplayName != nil {
				item.ChangedBy = types.StringValue(*rev.ChangedBy.DisplayName)
			}
			if rev.ChangedDate != nil {
				item.ChangedDate = types.StringValue(rev.ChangedDate.Time.Format(time.RFC3339))
			}
			if rev.ChangeType != nil {
				item.ChangeType = types.StringValue(string(*rev.ChangeType))
			}
			if rev.Comment != nil {
				item.Comment = types.StringValue(*rev.Comment)
			}
			items = append(items, item)
		}
	}

	model.Revisions = items
	model.ID = types.StringValue(fmt.Sprintf("release-definition-history-%s-%d", projectID, defID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
