package security

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/security"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var _ datasource.DataSource = &securityNamespacesDataSource{}

// securityNamespacesDataSource is the terraform-plugin-framework data source for
// betterado_security_namespaces.
type securityNamespacesDataSource struct {
	client *client.AggregatedClient
}

// NewSecurityNamespacesDataSource returns a new framework data source for
// betterado_security_namespaces.
func NewSecurityNamespacesDataSource() datasource.DataSource {
	return &securityNamespacesDataSource{}
}

// ── Model ─────────────────────────────────────────────────────────────────────

type securityNamespacesModel struct {
	Namespaces types.List `tfsdk:"namespaces"`
}

var namespaceAttrTypes = map[string]attr.Type{
	"id":           types.StringType,
	"name":         types.StringType,
	"display_name": types.StringType,
	"actions": types.ListType{
		ElemType: types.ObjectType{AttrTypes: actionAttrTypes},
	},
}

// ── Metadata / Schema ─────────────────────────────────────────────────────────

func (d *securityNamespacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_namespaces"
}

func (d *securityNamespacesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads all Azure DevOps security namespaces.",
		Attributes: map[string]schema.Attribute{
			"namespaces": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of all security namespaces.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the security namespace.",
						},
						"name": schema.StringAttribute{
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
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *securityNamespacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *securityNamespacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_security_namespaces data source Read: provider client not configured")
		return
	}

	var model securityNamespacesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	allNamespaces, err := d.client.SecurityClient.QuerySecurityNamespaces(d.client.Ctx, security.QuerySecurityNamespacesArgs{})
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("querying security namespaces: %v", err))
		return
	}

	if allNamespaces == nil || len(*allNamespaces) == 0 {
		resp.Diagnostics.AddError("Read error", "no security namespaces found")
		return
	}

	nsElemType := types.ObjectType{AttrTypes: namespaceAttrTypes}
	actionElemType := types.ObjectType{AttrTypes: actionAttrTypes}

	var nsValues []attr.Value

	for _, ns := range *allNamespaces {
		nsID := ""
		nsName := ""
		nsDisplay := ""
		if ns.NamespaceId != nil {
			nsID = ns.NamespaceId.String()
		}
		if ns.Name != nil {
			nsName = *ns.Name
		}
		if ns.DisplayName != nil {
			nsDisplay = *ns.DisplayName
		}

		// Build actions list.
		var actionValues []attr.Value
		if ns.Actions != nil {
			for _, action := range *ns.Actions {
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
				actionObj, diags := types.ObjectValue(actionAttrTypes, map[string]attr.Value{
					"name":         types.StringValue(aName),
					"display_name": types.StringValue(aDisplay),
					"bit":          types.Int64Value(aBit),
				})
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				actionValues = append(actionValues, actionObj)
			}
		}

		actionsList, diags := types.ListValue(actionElemType, actionValues)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		nsObj, diags := types.ObjectValue(namespaceAttrTypes, map[string]attr.Value{
			"id":           types.StringValue(nsID),
			"name":         types.StringValue(nsName),
			"display_name": types.StringValue(nsDisplay),
			"actions":      actionsList,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		nsValues = append(nsValues, nsObj)
	}

	nsList, diags := types.ListValue(nsElemType, nsValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Namespaces = nsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
