package accounts

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	accountsapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/accounts"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var _ datasource.DataSource = &accountsDataSource{}
var _ datasource.DataSourceWithConfigure = &accountsDataSource{}

// accountsDataSource implements the betterado_accounts framework data source.
type accountsDataSource struct {
	client *client.AggregatedClient
}

// NewAccountsDataSource returns a new framework data source for betterado_accounts.
func NewAccountsDataSource() datasource.DataSource {
	return &accountsDataSource{}
}

// accountModel represents a single ADO account in Terraform state.
type accountModel struct {
	AccountID        types.String `tfsdk:"account_id"`
	AccountName      types.String `tfsdk:"account_name"`
	AccountURI       types.String `tfsdk:"account_uri"`
	AccountType      types.String `tfsdk:"account_type"`
	OrganizationName types.String `tfsdk:"organization_name"`
}

// accountsDataModel is the tfsdk model for the betterado_accounts data source.
type accountsDataModel struct {
	ID       types.String   `tfsdk:"id"`
	MemberID types.String   `tfsdk:"member_id"`
	Accounts []accountModel `tfsdk:"accounts"`
}

func (d *accountsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_accounts"
}

func (d *accountsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list Azure DevOps accounts accessible to the authenticated user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The data source ID.",
			},
			"member_id": schema.StringAttribute{
				Optional:    true,
				Description: "Filter accounts by member subject descriptor (UUID). If omitted, all accounts accessible to the authenticated PAT are returned.",
			},
			"accounts": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of ADO accounts accessible to the authenticated user.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"account_id": schema.StringAttribute{
							Computed:    true,
							Description: "The UUID identifier of the account.",
						},
						"account_name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the account (organization name slug).",
						},
						"account_uri": schema.StringAttribute{
							Computed:    true,
							Description: "The URI of the account.",
						},
						"account_type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of account: personal or organization.",
						},
						"organization_name": schema.StringAttribute{
							Computed:    true,
							Description: "The organization name associated with the account.",
						},
					},
				},
			},
		},
	}
}

func (d *accountsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *accountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_accounts data source Read: provider client not configured")
		return
	}

	var model accountsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := accountsapi.GetAccountsArgs{}
	if !model.MemberID.IsNull() && !model.MemberID.IsUnknown() && model.MemberID.ValueString() != "" {
		memberUUID, err := uuid.Parse(model.MemberID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Invalid member_id", fmt.Sprintf("member_id must be a valid UUID: %s", err))
			return
		}
		args.MemberId = &memberUUID
	}

	accts, err := d.client.AccountsClient.GetAccounts(d.client.Ctx, args)
	if err != nil {
		// 404 → treat as empty list per WI spec.
		if isNotFound(err) {
			model.ID = types.StringValue("accounts")
			model.Accounts = []accountModel{}
			resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
			return
		}
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading accounts: %s", err))
		return
	}

	var accountItems []accountModel
	if accts != nil {
		for _, a := range *accts {
			item := accountModel{
				AccountID:        stringFromUUID(a.AccountId),
				AccountName:      stringFromPtr(a.AccountName),
				AccountURI:       stringFromPtr(a.AccountUri),
				AccountType:      accountTypeString(a.AccountType),
				OrganizationName: stringFromPtr(a.OrganizationName),
			}
			accountItems = append(accountItems, item)
		}
	}
	if accountItems == nil {
		accountItems = []accountModel{}
	}

	model.ID = types.StringValue("accounts")
	model.Accounts = accountItems
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func stringFromPtr(s *string) types.String {
	if s == nil {
		return types.StringValue("")
	}
	return types.StringValue(*s)
}

func stringFromUUID(u *uuid.UUID) types.String {
	if u == nil {
		return types.StringValue("")
	}
	return types.StringValue(u.String())
}

func accountTypeString(at *accountsapi.AccountType) types.String {
	if at == nil {
		return types.StringValue("")
	}
	return types.StringValue(string(*at))
}

// isNotFound returns true when the ADO API responded with a 404.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return fmt.Sprintf("%s", err) == "404 Not Found" || contains(err.Error(), "404")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
