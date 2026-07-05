package graph

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	graphapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Ensure interface compliance.
var _ datasource.DataSource = &userDataSource{}

// userDataSource is the terraform-plugin-framework data source for betterado_user.
type userDataSource struct {
	client *client.AggregatedClient
}

// NewUserDataSource returns a new framework data source for betterado_user.
func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

// userDataSourceModel is the tfsdk model for the data source.
type userDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Descriptor    types.String `tfsdk:"descriptor"`
	DisplayName   types.String `tfsdk:"display_name"`
	Domain        types.String `tfsdk:"domain"`
	MailAddress   types.String `tfsdk:"mail_address"`
	Origin        types.String `tfsdk:"origin"`
	OriginID      types.String `tfsdk:"origin_id"`
	PrincipalName types.String `tfsdk:"principal_name"`
	SubjectKind   types.String `tfsdk:"subject_kind"`
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Looks up an Azure DevOps user by descriptor.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The descriptor of the user (used as the data source ID).",
			},
			"descriptor": schema.StringAttribute{
				Required:    true,
				Description: "The descriptor of the user to look up.",
			},
			"display_name": schema.StringAttribute{
				Computed:    true,
				Description: "The display name of the user.",
			},
			"domain": schema.StringAttribute{
				Computed:    true,
				Description: "The domain the user belongs to.",
			},
			"mail_address": schema.StringAttribute{
				Computed:    true,
				Description: "The mail address of the user.",
			},
			"origin": schema.StringAttribute{
				Computed:    true,
				Description: "The origin of the user (e.g. aad, msa).",
			},
			"origin_id": schema.StringAttribute{
				Computed:    true,
				Description: "The origin-specific identifier of the user.",
			},
			"principal_name": schema.StringAttribute{
				Computed:    true,
				Description: "The principal name of the user.",
			},
			"subject_kind": schema.StringAttribute{
				Computed:    true,
				Description: "The subject kind (e.g. User).",
			},
		},
	}
}

func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_user data source Read: provider client not configured")
		return
	}

	var model userDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	descriptor := model.Descriptor.ValueString()

	user, err := d.client.GraphClient.GetUser(d.client.Ctx, graphapi.GetUserArgs{
		UserDescriptor: converter.String(descriptor),
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.Diagnostics.AddError("Not found",
				fmt.Sprintf("User does not exist with descriptor: %s", descriptor))
			return
		}
		resp.Diagnostics.AddError("Read error",
			fmt.Sprintf("Reading User: %+v", err))
		return
	}

	if user == nil {
		resp.Diagnostics.AddError("Not found",
			fmt.Sprintf("User does not exist with descriptor: %s", descriptor))
		return
	}

	model.ID = types.StringValue(*user.Descriptor)
	model.Descriptor = types.StringValue(*user.Descriptor)
	model.SubjectKind = strPtrToStringValue(user.SubjectKind)
	model.PrincipalName = strPtrToStringValue(user.PrincipalName)
	model.MailAddress = strPtrToStringValue(user.MailAddress)
	model.Origin = strPtrToStringValue(user.Origin)
	model.OriginID = strPtrToStringValue(user.OriginId)
	model.DisplayName = strPtrToStringValue(user.DisplayName)
	model.Domain = strPtrToStringValue(user.Domain)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// strPtrToStringValue converts *string to types.StringValue, defaulting to "".
func strPtrToStringValue(s *string) types.String {
	if s == nil {
		return types.StringValue("")
	}
	return types.StringValue(*s)
}
