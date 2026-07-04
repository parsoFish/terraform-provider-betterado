package identity

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var _ datasource.DataSource = &identityGroupsDataSource{}

// identityGroupsDataSource is the terraform-plugin-framework data source for
// betterado_identity_groups.
type identityGroupsDataSource struct {
	client *client.AggregatedClient
}

// NewIdentityGroupsDataSource returns a new framework data source for betterado_identity_groups.
func NewIdentityGroupsDataSource() datasource.DataSource {
	return &identityGroupsDataSource{}
}

// identityGroupsDataSourceModel is the tfsdk model for the data source.
type identityGroupsDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Groups    types.Set    `tfsdk:"groups"`
}

// identityGroupItemAttrTypes defines the attribute types for each group item in the set.
var identityGroupItemAttrTypes = map[string]attr.Type{
	"id":                 types.StringType,
	"name":               types.StringType,
	"descriptor":         types.StringType,
	"subject_descriptor": types.StringType,
}

// identityGroupItemModel represents one group entry in the groups set.
type identityGroupItemModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Descriptor        types.String `tfsdk:"descriptor"`
	SubjectDescriptor types.String `tfsdk:"subject_descriptor"`
}

func (d *identityGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity_groups"
}

func (d *identityGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up Azure DevOps identity groups with an optional project scope.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "A unique identifier for this data source read.",
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the project to scope the group lookup.",
			},
			"groups": schema.SetNestedAttribute{
				Computed:    true,
				Description: "The set of identity groups.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The UUID of the identity group.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The display name of the group.",
						},
						"descriptor": schema.StringAttribute{
							Computed:    true,
							Description: "The descriptor of the identity group.",
						},
						"subject_descriptor": schema.StringAttribute{
							Computed:    true,
							Description: "The subject descriptor of the identity group.",
						},
					},
				},
			},
		},
	}
}

func (d *identityGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *identityGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_identity_groups data source Read: provider client not configured")
		return
	}

	var model identityGroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()

	// Get groups in the specified project.
	groups, err := getIdentityGroupsWithProjectID(d.client, projectID)
	if err != nil {
		resp.Diagnostics.AddError("Read error",
			fmt.Sprintf("Failed to get groups for project with ID %s. Error: %v", projectID, err))
		return
	}

	// Build comma-separated identity IDs for batch ReadIdentities call.
	identityIds := ""
	for _, group := range groups {
		identityIds += group.Id.String() + ","
	}

	var identityGroups *[]identity.Identity
	if identityIds != "" {
		identityGroups, err = d.client.IdentityClient.ReadIdentities(d.client.Ctx, identity.ReadIdentitiesArgs{
			IdentityIds: &identityIds,
		})
		if err != nil {
			resp.Diagnostics.AddError("Read error",
				fmt.Sprintf("Failed to get Identity Groups for project with ID %s. Error: %v", projectID, err))
			return
		}
	}

	// Build the group items list.
	groupItems := make([]identityGroupItemModel, 0)
	if identityGroups != nil {
		for _, grp := range *identityGroups {
			item := identityGroupItemModel{}
			if grp.Id != nil {
				item.ID = types.StringValue(grp.Id.String())
			} else {
				item.ID = types.StringValue("")
			}
			if grp.ProviderDisplayName != nil {
				item.Name = types.StringValue(*grp.ProviderDisplayName)
			} else {
				item.Name = types.StringValue("")
			}
			if grp.Descriptor != nil {
				item.Descriptor = types.StringValue(*grp.Descriptor)
			} else {
				item.Descriptor = types.StringValue("")
			}
			if grp.SubjectDescriptor != nil {
				item.SubjectDescriptor = types.StringValue(*grp.SubjectDescriptor)
			} else {
				item.SubjectDescriptor = types.StringValue("")
			}
			groupItems = append(groupItems, item)
		}
	}

	groupsSet, diags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: identityGroupItemAttrTypes}, groupItems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.ID = types.StringValue("identity-groups-" + uuid.New().String())
	model.Groups = groupsSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
