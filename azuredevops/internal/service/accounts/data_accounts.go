package accounts

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var (
	_ datasource.DataSource              = &accountsDataSource{}
	_ datasource.DataSourceWithConfigure = &accountsDataSource{}
)

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

// vspsAccount is a wire-format account from the VSSPS REST API.
type vspsAccount struct {
	AccountID        string `json:"accountId"`
	AccountName      string `json:"accountName"`
	AccountURI       string `json:"accountUri"`
	AccountType      string `json:"accountType"`
	OrganizationName string `json:"organizationName"`
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

	// Build query params for the VSSPS accounts endpoint.
	// The accounts REST API is hosted on the global VSSPS endpoint:
	//   https://app.vssps.visualstudio.com/_apis/accounts
	// The org-specific VSSPS URL (https://vssps.dev.azure.com/<org>/_apis/accounts)
	// returns 404 "The controller for path '/_apis/accounts' was not found" because
	// the accounts controller is not registered in the org-scoped VSSPS routing table.
	//
	// For org-scoped PATs: the global VSSPS endpoint accepts org-scoped PATs when
	// a memberId filter is provided. Without memberId, the API requires a full-org PAT.
	// To handle org-scoped PATs transparently, when no memberId is given, we first
	// resolve the current user's profile UUID and use it as memberId automatically.
	params := url.Values{}
	params.Set("api-version", "7.1-preview.1")

	if !model.MemberID.IsNull() && !model.MemberID.IsUnknown() && model.MemberID.ValueString() != "" {
		memberID := model.MemberID.ValueString()
		if _, err := uuid.Parse(memberID); err != nil {
			resp.Diagnostics.AddError("Invalid member_id", fmt.Sprintf("member_id must be a valid UUID: %s", err))
			return
		}
		params.Set("memberId", memberID)
	} else {
		// No memberId supplied — auto-resolve the current user's UUID via the profile
		// endpoint so org-scoped PATs (which require memberId on the global accounts
		// endpoint) also work. Build the org-specific profile URL for resolution.
		orgName := extractOrgName(d.client.OrganizationURL)
		var profileURL string
		if orgName != "" {
			profileURL = fmt.Sprintf(
				"https://vssps.dev.azure.com/%s/_apis/profile/profiles/me?api-version=7.1-preview.3",
				orgName,
			)
		} else {
			profileURL = "https://app.vssps.visualstudio.com/_apis/profile/profiles/me?api-version=7.1-preview.3"
		}
		memberID, err := resolveCurrentUserID(ctx, profileURL, d.client.BasicAuth)
		if err == nil && memberID != "" {
			params.Set("memberId", memberID)
		}
		// If profile lookup fails, proceed without memberId (works for full-org PATs).
	}

	endpointURL := "https://app.vssps.visualstudio.com/_apis/accounts?" + params.Encode()
	accts, err := fetchAccounts(ctx, endpointURL, d.client.BasicAuth)
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading accounts: %s", err))
		return
	}

	var accountItems []accountModel
	for _, a := range accts {
		accountItems = append(accountItems, accountModel{
			AccountID:        types.StringValue(a.AccountID),
			AccountName:      types.StringValue(a.AccountName),
			AccountURI:       types.StringValue(a.AccountURI),
			AccountType:      types.StringValue(a.AccountType),
			OrganizationName: types.StringValue(a.OrganizationName),
		})
	}
	if accountItems == nil {
		accountItems = []accountModel{}
	}

	model.ID = types.StringValue("accounts")
	model.Accounts = accountItems
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// extractOrgName returns the ADO organization name from an org service URL of
// the form https://dev.azure.com/<orgname>[/...].  Returns "" when the URL
// does not match that pattern (e.g. on-prem Server URLs).
func extractOrgName(orgURL string) string {
	// Normalise: strip scheme, lowercase, trim trailing slashes.
	s := strings.ToLower(strings.TrimRight(orgURL, "/"))
	const prefix = "https://dev.azure.com/"
	if !strings.HasPrefix(s, prefix) {
		return ""
	}
	rest := s[len(prefix):]
	// rest may be "<orgname>" or "<orgname>/..."
	if idx := strings.Index(rest, "/"); idx >= 0 {
		return rest[:idx]
	}
	return rest
}

// resolveCurrentUserID fetches the authenticated user's profile and returns
// their UUID (subject ID). This is used to auto-populate memberId when querying
// the accounts endpoint with an org-scoped PAT.
func resolveCurrentUserID(ctx context.Context, profileURL, basicAuth string) (string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, profileURL, nil)
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", basicAuth)
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer httpResp.Body.Close() //nolint:errcheck

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", fmt.Errorf("profile lookup HTTP %d", httpResp.StatusCode)
	}

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", err
	}

	var p struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return "", err
	}
	return p.ID, nil
}

// fetchAccounts makes a direct REST call to the VSSPS accounts endpoint and
// returns the list of accounts. It uses the supplied basicAuth header (e.g.
// "Basic <base64>") to authenticate, bypassing the SDK's location-service
// discovery which issues an OPTIONS /_apis probe that can 401 on vssps.
func fetchAccounts(ctx context.Context, endpointURL, basicAuth string) ([]vspsAccount, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Authorization", basicAuth)
	httpReq.Header.Set("Accept", "application/json")

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer httpResp.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		// Trim body to avoid leaking large HTML error pages in diagnostics.
		snippet := string(body)
		if len(snippet) > 200 {
			snippet = snippet[:200]
		}
		return nil, fmt.Errorf("HTTP %d: %s", httpResp.StatusCode, snippet)
	}

	// The VSSPS accounts endpoint returns a collection object: {"count": N, "value": [...]}
	var wrapper struct {
		Count int           `json:"count"`
		Value []vspsAccount `json:"value"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		// Fallback: try parsing as a plain array.
		var direct []vspsAccount
		if err2 := json.Unmarshal(body, &direct); err2 != nil {
			return nil, fmt.Errorf("decoding response: %w (body snippet: %.200s)", err, body)
		}
		return direct, nil
	}
	return wrapper.Value, nil
}
