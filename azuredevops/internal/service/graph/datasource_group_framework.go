package graph

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	graphapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var _ datasource.DataSource = &groupDataSource{}

// groupDataSource is the terraform-plugin-framework data source for
// betterado_group.
type groupDataSource struct {
	client *client.AggregatedClient
}

// NewGroupDataSource returns a new framework data source for betterado_group.
func NewGroupDataSource() datasource.DataSource {
	return &groupDataSource{}
}

// groupDataSourceModel is the tfsdk model for the data source.
type groupDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	ProjectID  types.String `tfsdk:"project_id"`
	Descriptor types.String `tfsdk:"descriptor"`
	Origin     types.String `tfsdk:"origin"`
	OriginID   types.String `tfsdk:"origin_id"`
	GroupID    types.String `tfsdk:"group_id"`
}

func (d *groupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *groupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up an Azure DevOps group by name (and optionally project).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The descriptor of the group (used as the data source ID).",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The display name of the group.",
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the project the group belongs to. If omitted, collection-level groups are searched.",
			},
			"descriptor": schema.StringAttribute{
				Computed:    true,
				Description: "The descriptor of the group.",
			},
			"origin": schema.StringAttribute{
				Computed:    true,
				Description: "The origin of the group (e.g. vsts, aad).",
			},
			"origin_id": schema.StringAttribute{
				Computed:    true,
				Description: "The origin-specific identifier of the group.",
			},
			"group_id": schema.StringAttribute{
				Computed:    true,
				Description: "The storage key UUID of the group.",
			},
		},
	}
}

func (d *groupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *groupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_group data source Read: provider client not configured")
		return
	}

	var model groupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupName := model.Name.ValueString()
	projectID := model.ProjectID.ValueString()

	// Reuse the package-level helpers from data_group.go
	projectDescriptor, err := getProjectDescriptor(d.client, projectID)
	if err != nil {
		resp.Diagnostics.AddError("Read error",
			fmt.Sprintf("getting descriptor for project %q: %s", projectID, err))
		return
	}

	projectGroups, err := getGroupsForDescriptor(d.client, projectDescriptor)
	if err != nil {
		errMsg := "finding groups"
		if projectID != "" {
			errMsg = fmt.Sprintf("finding groups for project %q", projectID)
		}
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("%s: %s", errMsg, err))
		return
	}

	targetGroup := selectGroup(projectGroups, groupName)
	if targetGroup == nil {
		if projectID != "" {
			resp.Diagnostics.AddError("Not found",
				fmt.Sprintf("could not find group %q in project %q", groupName, projectID))
		} else {
			resp.Diagnostics.AddError("Not found",
				fmt.Sprintf("could not find group %q", groupName))
		}
		return
	}

	storageKey, err := d.client.GraphClient.GetStorageKey(d.client.Ctx, graphapi.GetStorageKeyArgs{
		SubjectDescriptor: targetGroup.Descriptor,
	})
	if err != nil {
		resp.Diagnostics.AddError("Read error",
			fmt.Sprintf("getting storage key for group %q: %s", groupName, err))
		return
	}

	model.ID = types.StringValue(*targetGroup.Descriptor)
	model.Descriptor = types.StringValue(*targetGroup.Descriptor)
	if targetGroup.Origin != nil {
		model.Origin = types.StringValue(*targetGroup.Origin)
	} else {
		model.Origin = types.StringValue("")
	}
	if targetGroup.OriginId != nil {
		model.OriginID = types.StringValue(*targetGroup.OriginId)
	} else {
		model.OriginID = types.StringValue("")
	}
	if storageKey.Value != nil {
		model.GroupID = types.StringValue(storageKey.Value.String())
	} else {
		model.GroupID = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
