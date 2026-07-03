package profile

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	profileapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/profile"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var _ datasource.DataSource = &profileDataSource{}
var _ datasource.DataSourceWithConfigure = &profileDataSource{}

// profileDataSource implements the betterado_profile framework data source.
type profileDataSource struct {
	profileClient profileapi.Client
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
	d.profileClient = c.ProfileClient
}

func (d *profileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.profileClient == nil {
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

	details := true
	coreAttrs := "Email,Avatar,DisplayName,ContactWithOffers"
	p, err := d.profileClient.GetProfile(ctx, profileapi.GetProfileArgs{
		Id:             &id,
		Details:        &details,
		CoreAttributes: &coreAttrs,
	})
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

	// Extract the profile ID (UUID).
	profileID := ""
	if p.Id != nil {
		profileID = p.Id.String()
	}
	model.ID = types.StringValue(profileID)

	// Extract core attributes from the CoreAttributes map.
	model.DisplayName = types.StringValue("")
	model.EmailAddress = types.StringValue("")
	model.PublicAlias = types.StringValue("")
	model.AvatarURL = types.StringValue("")

	if p.CoreAttributes != nil {
		attrs := *p.CoreAttributes

		if attr, ok := attrs["DisplayName"]; ok {
			model.DisplayName = types.StringValue(extractStringValue(attr.Value))
		}
		if attr, ok := attrs["Email"]; ok {
			model.EmailAddress = types.StringValue(extractStringValue(attr.Value))
		}
		if attr, ok := attrs["PublicAlias"]; ok {
			model.PublicAlias = types.StringValue(extractStringValue(attr.Value))
		}
		if attr, ok := attrs["Avatar"]; ok {
			model.AvatarURL = types.StringValue(extractAvatarValue(attr.Value))
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// extractStringValue converts an interface{} value from CoreProfileAttribute.Value to a string.
// The value may be a string, a JSON-encoded string, or a map with a "value" key.
func extractStringValue(v interface{}) string {
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
	// Try JSON-marshalled representation.
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

// extractAvatarValue extracts and base64-encodes the avatar bytes from CoreProfileAttribute.Value.
// The Avatar value is typically a JSON object: {"value": "<base64>", ...} or raw bytes.
func extractAvatarValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case []byte:
		return base64.StdEncoding.EncodeToString(val)
	case string:
		return val
	case map[string]interface{}:
		// The value field may be a base64 string directly.
		if s, ok := val["value"].(string); ok {
			return s
		}
	}
	return ""
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
