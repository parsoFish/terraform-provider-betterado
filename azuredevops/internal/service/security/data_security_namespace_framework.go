package security

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/security"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var _ datasource.DataSource = &securityNamespaceDataSource{}

// securityNamespaceDataSource is the terraform-plugin-framework data source for
// betterado_security_namespace.
type securityNamespaceDataSource struct {
	client *client.AggregatedClient
}

// NewSecurityNamespaceDataSource returns a new framework data source for
// betterado_security_namespace.
func NewSecurityNamespaceDataSource() datasource.DataSource {
	return &securityNamespaceDataSource{}
}

// ── Model ─────────────────────────────────────────────────────────────────────

type securityNamespaceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	DisplayName types.String `tfsdk:"display_name"`
	Actions     types.List   `tfsdk:"actions"`
}

var actionAttrTypes = map[string]attr.Type{
	"name":         types.StringType,
	"display_name": types.StringType,
	"bit":          types.Int64Type,
}

// ── Metadata / Schema ─────────────────────────────────────────────────────────

func (d *securityNamespaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_namespace"
}

func (d *securityNamespaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads an Azure DevOps security namespace by name or ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The ID of the security namespace.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the security namespace.",
			},
			"display_name": schema.StringAttribute{
				Computed:    true,
				Description: "The display name of the security namespace.",
			},
			"actions": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Available actions (permissions) in this namespace.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the action/permission.",
						},
						"display_name": schema.StringAttribute{
							Computed:    true,
							Description: "The display name of the action/permission.",
						},
						"bit": schema.Int64Attribute{
							Computed:    true,
							Description: "The bit value for this permission.",
						},
					},
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *securityNamespaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *securityNamespaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_security_namespace data source Read: provider client not configured")
		return
	}

	var model securityNamespaceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := model.Name.ValueString()
	namespaceID := model.ID.ValueString()

	if name == "" && namespaceID == "" {
		resp.Diagnostics.AddError("Configuration error", "either 'name' or 'id' must be specified")
		return
	}

	// Fetch all namespaces.
	allNamespaces, err := d.client.SecurityClient.QuerySecurityNamespaces(d.client.Ctx, security.QuerySecurityNamespacesArgs{})
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("querying security namespaces: %v", err))
		return
	}

	if allNamespaces == nil || len(*allNamespaces) == 0 {
		resp.Diagnostics.AddError("Read error", "no security namespaces found")
		return
	}

	var found *security.SecurityNamespaceDescription
	for i, ns := range *allNamespaces {
		if name != "" && ns.Name != nil && strings.EqualFold(*ns.Name, name) {
			found = &(*allNamespaces)[i]
			break
		}
		if namespaceID != "" && ns.NamespaceId != nil && strings.EqualFold(ns.NamespaceId.String(), namespaceID) {
			found = &(*allNamespaces)[i]
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("security namespace not found with name=%q or id=%q", name, namespaceID))
		return
	}

	// Populate model.
	if found.NamespaceId != nil {
		model.ID = types.StringValue(found.NamespaceId.String())
	}
	if found.Name != nil {
		model.Name = types.StringValue(*found.Name)
	}
	if found.DisplayName != nil {
		model.DisplayName = types.StringValue(*found.DisplayName)
	}

	// Build actions list.
	actionElemType := types.ObjectType{AttrTypes: actionAttrTypes}
	actionsVal := []attr.Value{}
	if found.Actions != nil {
		for _, action := range *found.Actions {
			aName := ""
			aDisplay := ""
			var aBit int64
			if action.Name != nil {
				aName = *action.Name
			}
			if action.DisplayName != nil {
				aDisplay = *action.DisplayName
			}
			if action.Bit != nil {
				aBit = int64(*action.Bit)
			}
			obj, diags := types.ObjectValue(actionAttrTypes, map[string]attr.Value{
				"name":         types.StringValue(aName),
				"display_name": types.StringValue(aDisplay),
				"bit":          types.Int64Value(aBit),
			})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			actionsVal = append(actionsVal, obj)
		}
	}

	actionsList, diags := types.ListValue(actionElemType, actionsVal)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Actions = actionsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
