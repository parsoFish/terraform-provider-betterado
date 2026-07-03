package graph

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	graphapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Ensure interface compliance.
var _ datasource.DataSource = &groupsDataSource{}

// groupsDataSource is the terraform-plugin-framework data source for betterado_groups.
type groupsDataSource struct {
	client *client.AggregatedClient
}

// NewGroupsDataSource returns a new framework data source for betterado_groups.
func NewGroupsDataSource() datasource.DataSource {
	return &groupsDataSource{}
}

// groupsDataSourceModel is the tfsdk model for the data source.
type groupsDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Groups    types.Set    `tfsdk:"groups"`
}

// groupItemModel represents one group entry in the groups set.
type groupItemModel struct {
	ID            types.String `tfsdk:"id"`
	Descriptor    types.String `tfsdk:"descriptor"`
	DisplayName   types.String `tfsdk:"display_name"`
	Origin        types.String `tfsdk:"origin"`
	OriginID      types.String `tfsdk:"origin_id"`
	Domain        types.String `tfsdk:"domain"`
	MailAddress   types.String `tfsdk:"mail_address"`
	PrincipalName types.String `tfsdk:"principal_name"`
}

// groupItemAttrTypes returns the attr.Type map for groupItemModel.
var groupItemAttrTypes = map[string]attr.Type{
	"id":             types.StringType,
	"descriptor":     types.StringType,
	"display_name":   types.StringType,
	"origin":         types.StringType,
	"origin_id":      types.StringType,
	"domain":         types.StringType,
	"mail_address":   types.StringType,
	"principal_name": types.StringType,
}

func (d *groupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *groupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up Azure DevOps groups with an optional project scope.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "A unique identifier for this data source read.",
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the project to scope the group lookup. If omitted, collection-level groups are returned.",
			},
			"groups": schema.SetNestedAttribute{
				Computed:    true,
				Description: "The set of matching groups.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The storage key UUID of the group.",
						},
						"descriptor": schema.StringAttribute{
							Computed:    true,
							Description: "The descriptor of the group.",
						},
						"display_name": schema.StringAttribute{
							Computed:    true,
							Description: "The display name of the group.",
						},
						"origin": schema.StringAttribute{
							Computed:    true,
							Description: "The origin of the group (e.g. vsts, aad).",
						},
						"origin_id": schema.StringAttribute{
							Computed:    true,
							Description: "The origin-specific identifier of the group.",
						},
						"domain": schema.StringAttribute{
							Computed:    true,
							Description: "The domain of the group.",
						},
						"mail_address": schema.StringAttribute{
							Computed:    true,
							Description: "The mail address of the group.",
						},
						"principal_name": schema.StringAttribute{
							Computed:    true,
							Description: "The principal name of the group.",
						},
					},
				},
			},
		},
	}
}

func (d *groupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *groupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_groups data source Read: provider client not configured")
		return
	}

	var model groupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()

	projectDescriptor, err := getProjectDescriptor(d.client, projectID)
	if err != nil {
		resp.Diagnostics.AddError("Read error",
			fmt.Sprintf("getting descriptor for project %q: %s", projectID, err))
		return
	}

	groups, err := getGroupsForDescriptor(d.client, projectDescriptor)
	if err != nil {
		errMsg := "finding groups"
		if projectID != "" {
			errMsg = fmt.Sprintf("finding groups for project %q", projectID)
		}
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("%s: %s", errMsg, err))
		return
	}

	groupItems := make([]groupItemModel, 0)
	if groups != nil {
		for _, grp := range *groups {
			if grp.Descriptor == nil {
				continue
			}

			storageKey, err := d.client.GraphClient.GetStorageKey(d.client.Ctx, graphapi.GetStorageKeyArgs{
				SubjectDescriptor: converter.String(*grp.Descriptor),
			})
			if err != nil {
				resp.Diagnostics.AddError("Read error",
					fmt.Sprintf("getting storage key for group %q: %s", *grp.Descriptor, err))
				return
			}

			item := groupItemModel{
				Descriptor:    strPtrToStringValue(grp.Descriptor),
				DisplayName:   strPtrToStringValue(grp.DisplayName),
				Origin:        strPtrToStringValue(grp.Origin),
				OriginID:      strPtrToStringValue(grp.OriginId),
				Domain:        strPtrToStringValue(grp.Domain),
				MailAddress:   strPtrToStringValue(grp.MailAddress),
				PrincipalName: strPtrToStringValue(grp.PrincipalName),
			}
			if storageKey.Value != nil {
				item.ID = types.StringValue(storageKey.Value.String())
			} else {
				item.ID = types.StringValue("")
			}
			groupItems = append(groupItems, item)
		}
	}

	groupsSet, diags := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: groupItemAttrTypes}, groupItems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.ID = types.StringValue("groups-" + uuid.New().String())
	model.Groups = groupsSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
