package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var (
	_ datasource.DataSource              = &profileDataSource{}
	_ datasource.DataSourceWithConfigure = &profileDataSource{}
)

// profileDataSource implements the betterado_profile framework data source.
type profileDataSource struct {
	client *client.AggregatedClient
}

// NewProfileDataSource returns a new framework data source for betterado_profile.
func NewProfileDataSource() datasource.DataSource {
	return &profileDataSource{}
}

// profileDataModel is the tfsdk model for the betterado_profile data source.
type profileDataModel struct {
	ID           types.String `tfsdk:"id"`
	DisplayName  types.String `tfsdk:"display_name"`
	EmailAddress types.String `tfsdk:"email_address"`
	PublicAlias  types.String `tfsdk:"public_alias"`
	AvatarURL    types.String `tfsdk:"avatar_url"`
}

// vspsProfile is the wire-format response from the VSSPS profile REST endpoint.
type vspsProfile struct {
	ID             string                     `json:"id"`
	DisplayName    string                     `json:"displayName"`
	EmailAddress   string                     `json:"emailAddress"`
	PublicAlias    string                     `json:"publicAlias"`
	CoreAttributes map[string]vspsProfileAttr `json:"coreAttributes"`
}

// vspsProfileAttr holds one entry from the coreAttributes map.
type vspsProfileAttr struct {
	Descriptor string      `json:"descriptor"`
	Value      interface{} `json:"value"`
}

func (d *profileDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_profile"
}

func (d *profileDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to look up an Azure DevOps user profile. Use `id = \"me\"` to fetch the authenticated user's own profile.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The profile identity to look up. Use `\"me\"` for the authenticated user.",
			},
			"display_name": schema.StringAttribute{
				Computed:    true,
				Description: "The display name of the profile.",
			},
			"email_address": schema.StringAttribute{
				Computed:    true,
				Description: "The email address of the profile.",
			},
			"public_alias": schema.StringAttribute{
				Computed:    true,
				Description: "The public alias of the profile.",
			},
			"avatar_url": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Base64-encoded PNG avatar of the profile. May be empty if the user has no custom avatar.",
			},
		},
	}
}

func (d *profileDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *profileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_profile data source Read: provider client not configured")
		return
	}

	var model profileDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := model.ID.ValueString()
	if id == "" {
		resp.Diagnostics.AddError("Invalid id", "id must be a non-empty string (use \"me\" for the authenticated user)")
		return
	}

	// Call the VSSPS profile endpoint directly, bypassing the SDK's
	// location-service discovery (OPTIONS /_apis) which can return 401 when
	// the PAT is scoped to a single org.
	//
	// Org-scoped PATs only work against the org-specific VSSPS host:
	//   https://vssps.dev.azure.com/<orgname>/_apis/profile/profiles/{id}
	orgName := extractOrgName(d.client.OrganizationURL)
	var endpointURL string
	if orgName != "" {
		endpointURL = fmt.Sprintf(
			"https://vssps.dev.azure.com/%s/_apis/profile/profiles/%s?api-version=7.1-preview.3&details=true&coreAttributes=Email,Avatar,DisplayName,ContactWithOffers",
			orgName, id,
		)
	} else {
		endpointURL = fmt.Sprintf(
			"https://app.vssps.visualstudio.com/_apis/profile/profiles/%s?api-version=7.1-preview.3&details=true&coreAttributes=Email,Avatar,DisplayName,ContactWithOffers",
			id,
		)
	}
	p, err := fetchProfile(ctx, endpointURL, d.client.BasicAuth)
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading profile %q: %s", id, err))
		return
	}

	if p == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	model.ID = types.StringValue(p.ID)
	model.DisplayName = types.StringValue(p.DisplayName)
	model.EmailAddress = types.StringValue(p.EmailAddress)
	model.PublicAlias = types.StringValue(p.PublicAlias)

	// Avatar is a coreAttribute — extract if present.
	avatarURL := ""
	if p.CoreAttributes != nil {
		if attr, ok := p.CoreAttributes["Avatar"]; ok {
			avatarURL = extractStringAttr(attr.Value)
		}
	}
	model.AvatarURL = types.StringValue(avatarURL)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// extractOrgName returns the ADO organization name from an org service URL of
// the form https://dev.azure.com/<orgname>[/...].  Returns "" when the URL
// does not match that pattern (e.g. on-prem Server URLs).
func extractOrgName(orgURL string) string {
	s := strings.ToLower(strings.TrimRight(orgURL, "/"))
	const prefix = "https://dev.azure.com/"
	if !strings.HasPrefix(s, prefix) {
		return ""
	}
	rest := s[len(prefix):]
	if idx := strings.Index(rest, "/"); idx >= 0 {
		return rest[:idx]
	}
	return rest
}

// fetchProfile makes a direct REST call to the VSSPS profile endpoint and
// returns the profile object. It uses the supplied basicAuth header (e.g.
// "Basic <base64>") to authenticate, bypassing the SDK's location-service
// discovery which issues an OPTIONS /_apis probe that can 401 on vssps.
func fetchProfile(ctx context.Context, endpointURL, basicAuth string) (*vspsProfile, error) {
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

	if httpResp.StatusCode == 404 {
		return nil, fmt.Errorf("404 Not Found")
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		snippet := string(body)
		if len(snippet) > 200 {
			snippet = snippet[:200]
		}
		return nil, fmt.Errorf("HTTP %d: %s", httpResp.StatusCode, snippet)
	}

	var profile vspsProfile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &profile, nil
}

// extractStringAttr converts an interface{} value from a core profile attribute
// to a string.
func extractStringAttr(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case map[string]interface{}:
		if s, ok := val["value"].(string); ok {
			return s
		}
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		return s
	}
	return string(b)
}

// isNotFound returns true when the ADO API responded with a 404.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return containsStr(msg, "404") || containsStr(msg, "Not Found")
}

func containsStr(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
