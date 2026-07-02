package securityroles

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	sdkroles "github.com/parsoFish/terraform-provider-betterado/azuredevops/utils/sdk/securityroles"
)

// Ensure interface compliance.
var _ datasource.DataSource = &securityRoleDefinitionsDataSource{}

// securityRoleDefinitionsDataSource is the terraform-plugin-framework
// implementation of betterado_securityrole_definitions.
type securityRoleDefinitionsDataSource struct {
	client *client.AggregatedClient
}

// NewSecurityRoleDefinitionsDataSource returns a new framework data source for
// betterado_securityrole_definitions.
func NewSecurityRoleDefinitionsDataSource() datasource.DataSource {
	return &securityRoleDefinitionsDataSource{}
}

// ── Model ─────────────────────────────────────────────────────────────────────

type securityRoleDefinitionsModel struct {
	ID          types.String `tfsdk:"id"`
	Scope       types.String `tfsdk:"scope"`
	Definitions types.Set    `tfsdk:"definitions"`
}

// definitionObject describes the object type stored in the definitions set.
var definitionObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"name":              types.StringType,
		"display_name":      types.StringType,
		"allow_permissions": types.Int64Type,
		"deny_permissions":  types.Int64Type,
		"identifier":        types.StringType,
		"description":       types.StringType,
		"scope":             types.StringType,
	},
}

// ── Metadata / Schema ─────────────────────────────────────────────────────────

func (d *securityRoleDefinitionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_securityrole_definitions"
}

func (d *securityRoleDefinitionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to access information about Security Role Definitions in a given scope within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of this resource.",
			},
			"scope": schema.StringAttribute{
				Required:    true,
				Description: "The scope of the security role definitions to look up.",
			},
			"definitions": schema.SetNestedAttribute{
				Computed:    true,
				Description: "Set of security role definitions found at the given scope.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The internal name of the role.",
						},
						"display_name": schema.StringAttribute{
							Computed:    true,
							Description: "The human-readable display name of the role.",
						},
						"allow_permissions": schema.Int64Attribute{
							Computed:    true,
							Description: "Bitmask of permissions the role allows.",
						},
						"deny_permissions": schema.Int64Attribute{
							Computed:    true,
							Description: "Bitmask of permissions the role denies.",
						},
						"identifier": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the role definition.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "A description of the role.",
						},
						"scope": schema.StringAttribute{
							Computed:    true,
							Description: "The scope of the role definition.",
						},
					},
				},
			},
		},
	}
}

// ── Provider data injection ───────────────────────────────────────────────────

func (d *securityRoleDefinitionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	agg, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.AggregatedClient, got %T", req.ProviderData),
		)
		return
	}
	d.client = agg
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (d *securityRoleDefinitionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config securityRoleDefinitionsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider was not configured with valid credentials.")
		return
	}

	scope := config.Scope.ValueString()
	defs, err := d.client.SecurityRolesClient.ListSecurityRoleDefinitions(ctx, &sdkroles.ListSecurityRoleDefinitionsArgs{
		Scope: &scope,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading security role definitions",
			fmt.Sprintf("Finding security role definitions for scope: %s. Error: %v", scope, err),
		)
		return
	}

	if defs == nil || len(*defs) == 0 {
		resp.Diagnostics.AddError(
			"No definitions found",
			fmt.Sprintf("No role definition found at scope: %s", scope),
		)
		return
	}

	definitionElems := make([]attr.Value, 0, len(*defs))
	for _, def := range *defs {
		name := ""
		if def.Name != nil {
			name = *def.Name
		}
		displayName := ""
		if def.DisplayName != nil {
			displayName = *def.DisplayName
		}
		identifier := ""
		if def.Identifier != nil {
			identifier = *def.Identifier
		}
		description := ""
		if def.Description != nil {
			description = *def.Description
		}
		defScope := ""
		if def.Scope != nil {
			defScope = *def.Scope
		}
		var allowPerms int64
		if def.AllowPermissions != nil {
			allowPerms = int64(*def.AllowPermissions)
		}
		var denyPerms int64
		if def.DenyPermissions != nil {
			denyPerms = int64(*def.DenyPermissions)
		}

		obj, diags := types.ObjectValue(definitionObjectType.AttrTypes, map[string]attr.Value{
			"name":              types.StringValue(name),
			"display_name":      types.StringValue(displayName),
			"allow_permissions": types.Int64Value(allowPerms),
			"deny_permissions":  types.Int64Value(denyPerms),
			"identifier":        types.StringValue(identifier),
			"description":       types.StringValue(description),
			"scope":             types.StringValue(defScope),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		definitionElems = append(definitionElems, obj)
	}

	definitionsSet, diags := types.SetValue(definitionObjectType, definitionElems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ID = types.StringValue("secroledefs-" + uuid.New().String())
	config.Definitions = definitionsSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
