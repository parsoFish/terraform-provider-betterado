package identity

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Ensure interface compliance.
var _ datasource.DataSource = &identityGroupDataSource{}

// identityGroupDataSource is the terraform-plugin-framework data source for
// betterado_identity_group.
type identityGroupDataSource struct {
	client *client.AggregatedClient
}

// NewIdentityGroupDataSource returns a new framework data source for betterado_identity_group.
func NewIdentityGroupDataSource() datasource.DataSource {
	return &identityGroupDataSource{}
}

// identityGroupDataSourceModel is the tfsdk model for the data source.
type identityGroupDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	ProjectID         types.String `tfsdk:"project_id"`
	Descriptor        types.String `tfsdk:"descriptor"`
	SubjectDescriptor types.String `tfsdk:"subject_descriptor"`
}

func (d *identityGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity_group"
}

func (d *identityGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up an Azure DevOps identity group by name within a project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The identity ID of the group.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The provider display name of the group (e.g. [ProjectName]\\GroupName).",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project the group belongs to.",
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
	}
}

func (d *identityGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *identityGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_identity_group data source Read: provider client not configured")
		return
	}

	var model identityGroupDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupName := model.Name.ValueString()
	projectID := model.ProjectID.ValueString()

	// Get groups in the specified project.
	projectGroups, err := getIdentityGroupsWithProjectID(d.client, projectID)
	if err != nil {
		resp.Diagnostics.AddError("Read error",
			fmt.Sprintf("Failed to get groups for project with ID: %s. Error: %v", projectID, err))
		return
	}

	// Select specific group by name.
	targetGroup := selectIdentityGroup(&projectGroups, groupName)
	if targetGroup == nil {
		resp.Diagnostics.AddError("Not found",
			fmt.Sprintf("Cannot find group with name %s in project with ID %s", groupName, projectID))
		return
	}

	identityGroup, err := d.client.IdentityClient.ReadIdentity(d.client.Ctx, identity.ReadIdentityArgs{
		IdentityId: converter.String(targetGroup.Id.String()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Read error",
			fmt.Sprintf("Failed to get Identity for group with ID: %s. Error: %v", targetGroup.Id.String(), err))
		return
	}

	model.ID = types.StringValue(targetGroup.Id.String())
	if targetGroup.Descriptor != nil {
		model.Descriptor = types.StringValue(*targetGroup.Descriptor)
	} else {
		model.Descriptor = types.StringValue("")
	}
	if identityGroup.SubjectDescriptor != nil {
		model.SubjectDescriptor = types.StringValue(*identityGroup.SubjectDescriptor)
	} else {
		model.SubjectDescriptor = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
