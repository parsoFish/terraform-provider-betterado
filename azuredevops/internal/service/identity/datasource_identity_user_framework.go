package identity

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var _ datasource.DataSource = &identityUserDataSource{}

// identityUserDataSource is the terraform-plugin-framework data source for
// betterado_identity_user.
type identityUserDataSource struct {
	client *client.AggregatedClient
}

// NewIdentityUserDataSource returns a new framework data source for betterado_identity_user.
func NewIdentityUserDataSource() datasource.DataSource {
	return &identityUserDataSource{}
}

// identityUserDataSourceModel is the tfsdk model for the data source.
type identityUserDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	SearchFilter      types.String `tfsdk:"search_filter"`
	Descriptor        types.String `tfsdk:"descriptor"`
	SubjectDescriptor types.String `tfsdk:"subject_descriptor"`
}

func (d *identityUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_identity_user"
}

func (d *identityUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up an Azure DevOps identity user by name with an optional search filter.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The identity ID of the user.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name to search for the identity user.",
			},
			"search_filter": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The search filter to use. One of: AccountName, DisplayName, MailAddress, General. Defaults to General.",
			},
			"descriptor": schema.StringAttribute{
				Computed:    true,
				Description: "The descriptor of the identity user.",
			},
			"subject_descriptor": schema.StringAttribute{
				Computed:    true,
				Description: "The subject descriptor of the identity user.",
			},
		},
	}
}

func (d *identityUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *identityUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_identity_user data source Read: provider client not configured")
		return
	}

	var model identityUserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userName := model.Name.ValueString()
	searchFilter := model.SearchFilter.ValueString()
	if searchFilter == "" {
		searchFilter = "General"
	}

	// Query ADO for list of identity users with filter.
	filterUser, err := getIdentityUsersWithFilterValue(d.client, searchFilter, userName)
	if err != nil {
		resp.Diagnostics.AddError("Read error",
			fmt.Sprintf("Finding user with filter %s. Error: %v", searchFilter, err))
		return
	}

	// Filter for the desired user in the results.
	targetUser := validateIdentityUser(filterUser, userName, searchFilter)
	if targetUser == nil {
		resp.Diagnostics.AddError("Not found",
			fmt.Sprintf("Could not find user with name: %s, with filter: %s", userName, searchFilter))
		return
	}

	model.ID = types.StringValue(targetUser.Id.String())
	model.SearchFilter = types.StringValue(searchFilter)
	if targetUser.Descriptor != nil {
		model.Descriptor = types.StringValue(*targetUser.Descriptor)
	} else {
		model.Descriptor = types.StringValue("")
	}
	if targetUser.SubjectDescriptor != nil {
		model.SubjectDescriptor = types.StringValue(*targetUser.SubjectDescriptor)
	} else {
		model.SubjectDescriptor = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
