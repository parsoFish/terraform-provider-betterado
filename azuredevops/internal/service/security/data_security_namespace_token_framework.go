package security

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/security"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/permissions/utils"
)

// Ensure interface compliance.
var _ datasource.DataSource = &securityNamespaceTokenDataSource{}

// securityNamespaceTokenDataSource is the terraform-plugin-framework data source for
// betterado_security_namespace_token.
type securityNamespaceTokenDataSource struct {
	client *client.AggregatedClient
}

// NewSecurityNamespaceTokenDataSource returns a new framework data source for
// betterado_security_namespace_token.
func NewSecurityNamespaceTokenDataSource() datasource.DataSource {
	return &securityNamespaceTokenDataSource{}
}

// ── Model ─────────────────────────────────────────────────────────────────────

type securityNamespaceTokenModel struct {
	ID                   types.String `tfsdk:"id"`
	NamespaceID          types.String `tfsdk:"namespace_id"`
	NamespaceName        types.String `tfsdk:"namespace_name"`
	Identifiers          types.Map    `tfsdk:"identifiers"`
	ReturnIdentifierInfo types.Bool   `tfsdk:"return_identifier_info"`
	Token                types.String `tfsdk:"token"`
	RequiredIdentifiers  types.List   `tfsdk:"required_identifiers"`
	OptionalIdentifiers  types.List   `tfsdk:"optional_identifiers"`
}

// ── Metadata / Schema ─────────────────────────────────────────────────────────

func (d *securityNamespaceTokenDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_namespace_token"
}

func (d *securityNamespaceTokenDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Generates a security token for a given Azure DevOps security namespace and set of identifiers.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The computed ID of this data source.",
			},
			"namespace_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the security namespace.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
						"must be a valid UUID",
					),
				},
			},
			"namespace_name": schema.StringAttribute{
				Optional:    true,
				Description: "The name of the security namespace (e.g., 'Git Repositories', 'Project').",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"identifiers": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Map of identifiers required for token generation (e.g., project_id, repository_id).",
			},
			"return_identifier_info": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "When true, returns the required and optional identifiers for the namespace instead of generating a token.",
			},
			"token": schema.StringAttribute{
				Computed:    true,
				Description: "The generated security token for the namespace.",
			},
			"required_identifiers": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of required identifiers for this namespace (only populated when return_identifier_info is true).",
			},
			"optional_identifiers": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "List of optional identifiers for this namespace (only populated when return_identifier_info is true).",
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *securityNamespaceTokenDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// ── Read ──────────────────────────────────────────────────────────────────────

func (d *securityNamespaceTokenDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_security_namespace_token data source Read: provider client not configured")
		return
	}

	var model securityNamespaceTokenModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve namespace ID.
	var namespaceID uuid.UUID
	var parseErr error

	hasNsID := !model.NamespaceID.IsNull() && !model.NamespaceID.IsUnknown() && model.NamespaceID.ValueString() != ""
	hasNsName := !model.NamespaceName.IsNull() && !model.NamespaceName.IsUnknown() && model.NamespaceName.ValueString() != ""

	switch {
	case hasNsID:
		namespaceID, parseErr = uuid.Parse(model.NamespaceID.ValueString())
		if parseErr != nil {
			resp.Diagnostics.AddError("Read error", fmt.Sprintf("invalid namespace_id: %v", parseErr))
			return
		}
	case hasNsName:
		nsName := model.NamespaceName.ValueString()
		namespaces, err := d.client.SecurityClient.QuerySecurityNamespaces(d.client.Ctx, security.QuerySecurityNamespacesArgs{})
		if err != nil {
			resp.Diagnostics.AddError("Read error", fmt.Sprintf("querying security namespaces: %v", err))
			return
		}
		found := false
		for _, ns := range *namespaces {
			if ns.Name != nil && *ns.Name == nsName {
				if ns.NamespaceId != nil {
					namespaceID = *ns.NamespaceId
					found = true
					break
				}
			}
		}
		if !found {
			resp.Diagnostics.AddError("Read error", fmt.Sprintf("namespace with name '%s' not found", nsName))
			return
		}
	default:
		resp.Diagnostics.AddError("Configuration error", "either 'namespace_id' or 'namespace_name' must be specified")
		return
	}

	// Determine if we're returning identifier info or generating a token.
	returnIdentifierInfo := false
	if !model.ReturnIdentifierInfo.IsNull() && !model.ReturnIdentifierInfo.IsUnknown() {
		returnIdentifierInfo = model.ReturnIdentifierInfo.ValueBool()
	}

	if returnIdentifierInfo {
		template, exists := namespaceTokenTemplates[utils.SecurityNamespaceID(namespaceID)]
		if !exists {
			resp.Diagnostics.AddError("Read error", fmt.Sprintf("no template information available for namespace %s", namespaceID.String()))
			return
		}

		reqIDs, diags := types.ListValueFrom(ctx, types.StringType, template.RequiredIdentifiers)
		resp.Diagnostics.Append(diags...)
		optIDs, diags2 := types.ListValueFrom(ctx, types.StringType, template.OptionalIdentifiers)
		resp.Diagnostics.Append(diags2...)
		if resp.Diagnostics.HasError() {
			return
		}

		model.RequiredIdentifiers = reqIDs
		model.OptionalIdentifiers = optIDs
		model.Token = types.StringValue("")
		model.ReturnIdentifierInfo = types.BoolValue(true)
		model.ID = types.StringValue(fmt.Sprintf("ns-info-%s", namespaceID.String()))
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
		return
	}

	// Build identifiers map from config.
	identifiers := make(map[string]string)
	if !model.Identifiers.IsNull() && !model.Identifiers.IsUnknown() {
		ids := map[string]string{}
		resp.Diagnostics.Append(model.Identifiers.ElementsAs(ctx, &ids, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		for k, v := range ids {
			identifiers[k] = v
		}
	}

	// Generate token.
	template, exists := namespaceTokenTemplates[utils.SecurityNamespaceID(namespaceID)]
	if !exists {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("unable to generate token for namespace %s", namespaceID.String()))
		return
	}

	token, err := template.BuildFunc(identifiers, d.client)
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("generating token: %v", err))
		return
	}

	model.Token = types.StringValue(token)
	model.ReturnIdentifierInfo = types.BoolValue(false)
	emptyList := types.ListValueMust(types.StringType, []attr.Value{})
	model.RequiredIdentifiers = emptyList
	model.OptionalIdentifiers = emptyList
	model.ID = types.StringValue(fmt.Sprintf("ns-token-%s-%s", namespaceID.String(), token))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
