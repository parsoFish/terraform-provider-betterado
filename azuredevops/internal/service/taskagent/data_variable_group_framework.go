package taskagent

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = (*VariableGroupDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*VariableGroupDataSource)(nil)
)

// VariableGroupDataSource is the terraform-plugin-framework implementation of betterado_variable_group data source.
type VariableGroupDataSource struct {
	client *client.AggregatedClient
}

// NewVariableGroupDataSource returns a new datasource.DataSource for betterado_variable_group.
func NewVariableGroupDataSource() datasource.DataSource {
	return &VariableGroupDataSource{}
}

// variableGroupDataModel is the Terraform state model for the betterado_variable_group data source.
type variableGroupDataModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	AllowAccess types.Bool   `tfsdk:"allow_access"`
	Variable    types.List   `tfsdk:"variable"`
	KeyVault    types.List   `tfsdk:"key_vault"`
}

// variableDataAttrTypes mirrors variableAttrTypes but for the data source (no secret_value sensitivity difference in schema type).
var variableDataAttrTypes = map[string]attr.Type{
	"name":         types.StringType,
	"value":        types.StringType,
	"secret_value": types.StringType,
	"is_secret":    types.BoolType,
	"content_type": types.StringType,
	"enabled":      types.BoolType,
	"expires":      types.StringType,
}

// keyVaultDataAttrTypes for the data source key_vault block (no search_depth in data source).
var keyVaultDataAttrTypes = map[string]attr.Type{
	"name":                types.StringType,
	"service_endpoint_id": types.StringType,
}

func (d *VariableGroupDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "betterado_variable_group"
}

func (d *VariableGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to access information about an existing Variable Group within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the variable group.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the variable group to look up.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the variable group.",
			},
			"allow_access": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this variable group is shared by all pipelines in the project.",
			},
			// variable uses ListNestedAttribute (not Set) for consistency with the resource
			// schema and to avoid the set-hash/sensitive-attribute bug.  Since all attributes
			// are Computed-only there is no Default-in-set issue, but using List also gives
			// a stable alphabetical ordering that users can rely on.
			"variable": schema.ListNestedAttribute{
				Computed:    true,
				Description: "One or more variable blocks.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The key or name of the variable.",
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Description: "The value of the variable.",
						},
						"secret_value": schema.StringAttribute{
							Computed:    true,
							Sensitive:   true,
							Description: "The secret value of the variable.",
						},
						"is_secret": schema.BoolAttribute{
							Computed:    true,
							Description: "If `true`, the variable is a secret.",
						},
						"content_type": schema.StringAttribute{
							Computed:    true,
							Description: "The content type of the variable (for key vault variables).",
						},
						"enabled": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the variable is enabled (for key vault variables).",
						},
						"expires": schema.StringAttribute{
							Computed:    true,
							Description: "The expiry date of the variable (for key vault variables).",
						},
					},
				},
			},
			"key_vault": schema.ListNestedAttribute{
				Computed:    true,
				Description: "A key_vault block if this variable group is backed by an Azure Key Vault.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the Azure Key Vault.",
						},
						"service_endpoint_id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the Azure RM service endpoint.",
						},
					},
				},
			},
		},
	}
}

func (d *VariableGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.AggregatedClient, got: %T", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *VariableGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config variableGroupDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := config.ProjectID.ValueString()
	name := config.Name.ValueString()

	variableGroups, err := d.client.TaskAgentClient.GetVariableGroups(ctx, taskagent.GetVariableGroupsArgs{
		Project:   &projectID,
		GroupName: &name,
		Top:       converter.Int(1),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading variable group", err.Error())
		return
	}
	if variableGroups == nil || len(*variableGroups) == 0 {
		resp.Diagnostics.AddError("Variable group not found", fmt.Sprintf("Unable to find variable group with name: %s", name))
		return
	}

	vg := &(*variableGroups)[0]
	state, diags := d.flattenToDataModel(ctx, vg, projectID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// flattenToDataModel converts an AzDO VariableGroup into the data source state model.
// Variables are sorted by name for a stable canonical list ordering.
func (d *VariableGroupDataSource) flattenToDataModel(ctx context.Context, vg *taskagent.VariableGroup, projectID string) (variableGroupDataModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	model := variableGroupDataModel{
		ID:          types.StringValue(strconv.Itoa(*vg.Id)),
		ProjectID:   types.StringValue(projectID),
		Name:        types.StringValue(converter.ToString(vg.Name, "")),
		Description: types.StringValue(converter.ToString(vg.Description, "")),
		AllowAccess: types.BoolValue(false),
	}

	// Sort variable names for stable list ordering.
	varNames := make([]string, 0, len(*vg.Variables))
	for varName := range *vg.Variables {
		varNames = append(varNames, varName)
	}
	sort.Strings(varNames)

	isKeyVault := isKeyVaultVariableGroupType(vg.Type)
	varObjects := make([]attr.Value, 0, len(*vg.Variables))
	for _, varName := range varNames {
		varVal := (*vg.Variables)[varName]
		varJSON, err := json.Marshal(varVal)
		if err != nil {
			diags.AddError("Error marshaling variable", err.Error())
			return model, diags
		}

		var vm variableModel
		if isKeyVault {
			var kvVar taskagent.AzureKeyVaultVariableValue
			if err := json.Unmarshal(varJSON, &kvVar); err != nil {
				diags.AddError("Error unmarshaling key vault variable", err.Error())
				return model, diags
			}
			vm = variableModel{
				Name:        types.StringValue(varName),
				Value:       types.StringValue(""),
				SecretValue: types.StringValue(""),
				IsSecret:    types.BoolValue(false),
				ContentType: types.StringValue(converter.ToString(kvVar.ContentType, "")),
				Enabled:     types.BoolValue(converter.ToBool(kvVar.Enabled, false)),
				Expires:     types.StringValue(""),
			}
			if kvVar.Expires != nil {
				vm.Expires = types.StringValue(kvVar.Expires.String())
			}
		} else {
			var apiVar struct {
				Value    *string `json:"value"`
				IsSecret *bool   `json:"isSecret"`
			}
			if err := json.Unmarshal(varJSON, &apiVar); err != nil {
				diags.AddError("Error unmarshaling variable", err.Error())
				return model, diags
			}
			isSecret := apiVar.IsSecret != nil && *apiVar.IsSecret
			valStr := ""
			if !isSecret && apiVar.Value != nil {
				valStr = *apiVar.Value
			}
			vm = variableModel{
				Name:        types.StringValue(varName),
				Value:       types.StringValue(valStr),
				SecretValue: types.StringValue(""),
				IsSecret:    types.BoolValue(isSecret),
				ContentType: types.StringValue(""),
				Enabled:     types.BoolValue(false),
				Expires:     types.StringValue(""),
			}
		}

		obj, d := types.ObjectValueFrom(ctx, variableDataAttrTypes, vm)
		diags.Append(d...)
		if diags.HasError() {
			return model, diags
		}
		varObjects = append(varObjects, obj)
	}

	varList, dList := types.ListValue(types.ObjectType{AttrTypes: variableDataAttrTypes}, varObjects)
	diags.Append(dList...)
	if diags.HasError() {
		return model, diags
	}
	model.Variable = varList

	// Key vault block
	if isKeyVault {
		providerDataJSON, err := json.Marshal(vg.ProviderData)
		if err != nil {
			diags.AddError("Error marshaling key vault provider data", err.Error())
			return model, diags
		}
		var pd taskagent.AzureKeyVaultVariableGroupProviderData
		if err := json.Unmarshal(providerDataJSON, &pd); err != nil {
			diags.AddError("Error unmarshaling key vault provider data", err.Error())
			return model, diags
		}
		seID := ""
		if pd.ServiceEndpointId != nil {
			seID = pd.ServiceEndpointId.String()
		}
		vaultName := converter.ToString(pd.Vault, "")

		kvObj, d2 := types.ObjectValueFrom(ctx, keyVaultDataAttrTypes, struct {
			Name              types.String `tfsdk:"name"`
			ServiceEndpointID types.String `tfsdk:"service_endpoint_id"`
		}{
			Name:              types.StringValue(vaultName),
			ServiceEndpointID: types.StringValue(seID),
		})
		diags.Append(d2...)
		if diags.HasError() {
			return model, diags
		}
		kvList, d3 := types.ListValue(types.ObjectType{AttrTypes: keyVaultDataAttrTypes}, []attr.Value{kvObj})
		diags.Append(d3...)
		if diags.HasError() {
			return model, diags
		}
		model.KeyVault = kvList
	} else {
		// Data source key_vault is Computed, so empty list is fine here
		// (no user-planned value to mismatch against).
		model.KeyVault = types.ListValueMust(types.ObjectType{AttrTypes: keyVaultDataAttrTypes}, []attr.Value{})
	}

	return model, diags
}
