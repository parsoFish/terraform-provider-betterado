package graph

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var _ datasource.DataSource = &groupMembershipDataSource{}

// groupMembershipDataSource is the terraform-plugin-framework data source for
// betterado_group_membership.
type groupMembershipDataSource struct {
	client *client.AggregatedClient
}

// NewGroupMembershipDataSource returns a new framework data source for betterado_group_membership.
func NewGroupMembershipDataSource() datasource.DataSource {
	return &groupMembershipDataSource{}
}

// groupMembershipDataSourceModel is the tfsdk model for the data source.
type groupMembershipDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	GroupDescriptor types.String `tfsdk:"group_descriptor"`
	Members         types.List   `tfsdk:"members"`
}

func (d *groupMembershipDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_membership"
}

func (d *groupMembershipDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists the members of an Azure DevOps group by its descriptor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The group descriptor (used as the data source ID).",
			},
			"group_descriptor": schema.StringAttribute{
				Required:    true,
				Description: "The descriptor of the group whose members to list.",
			},
			"members": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The list of member descriptors.",
			},
		},
	}
}

func (d *groupMembershipDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *groupMembershipDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_group_membership data source Read: provider client not configured")
		return
	}

	var model groupMembershipDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupDescriptor := model.GroupDescriptor.ValueString()

	// Reuse the package-level helper from resource_group_membership.go
	memberships, err := getGroupMemberships(d.client, groupDescriptor)
	if err != nil {
		resp.Diagnostics.AddError("Read error",
			fmt.Sprintf("reading group memberships for descriptor %q: %s", groupDescriptor, err))
		return
	}

	memberDescriptors := make([]string, 0, len(*memberships))
	for _, membership := range *memberships {
		memberDescriptors = append(memberDescriptors, *membership.MemberDescriptor)
	}

	members, diags := types.ListValueFrom(ctx, types.StringType, memberDescriptors)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.ID = types.StringValue(groupDescriptor)
	model.Members = members

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
