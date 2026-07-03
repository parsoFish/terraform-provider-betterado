package taskagent

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = (*VariableGroupResource)(nil)
	_ resource.ResourceWithConfigure   = (*VariableGroupResource)(nil)
	_ resource.ResourceWithImportState = (*VariableGroupResource)(nil)
)

// VariableGroupResource is the terraform-plugin-framework implementation of betterado_variable_group.
type VariableGroupResource struct {
	client *client.AggregatedClient
}

// NewVariableGroupResource returns a new resource.Resource for betterado_variable_group.
func NewVariableGroupResource() resource.Resource {
	return &VariableGroupResource{}
}

// variableModel is the state model for one variable entry inside the list.
type variableModel struct {
	Name        types.String `tfsdk:"name"`
	Value       types.String `tfsdk:"value"`
	SecretValue types.String `tfsdk:"secret_value"`
	IsSecret    types.Bool   `tfsdk:"is_secret"`
	ContentType types.String `tfsdk:"content_type"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	Expires     types.String `tfsdk:"expires"`
}

// keyVaultModel is the state model for the optional key_vault block.
type keyVaultModel struct {
	Name              types.String `tfsdk:"name"`
	ServiceEndpointID types.String `tfsdk:"service_endpoint_id"`
	SearchDepth       types.Int64  `tfsdk:"search_depth"`
}

// variableGroupModel is the Terraform state model for betterado_variable_group.
type variableGroupModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	AllowAccess types.Bool   `tfsdk:"allow_access"`
	Variable    types.List   `tfsdk:"variable"`
	KeyVault    types.List   `tfsdk:"key_vault"`
}

// variableAttrTypes is the schema attribute type map for the variable nested object.
var variableAttrTypes = map[string]attr.Type{
	"name":         types.StringType,
	"value":        types.StringType,
	"secret_value": types.StringType,
	"is_secret":    types.BoolType,
	"content_type": types.StringType,
	"enabled":      types.BoolType,
	"expires":      types.StringType,
}

// keyVaultAttrTypes is the schema attribute type map for key_vault nested object.
var keyVaultAttrTypes = map[string]attr.Type{
	"name":                types.StringType,
	"service_endpoint_id": types.StringType,
	"search_depth":        types.Int64Type,
}

func (r *VariableGroupResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_variable_group"
}

func (r *VariableGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Variable Group within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					useStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the variable group.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     defaultString(""),
				Description: "The description of the variable group.",
			},
			"allow_access": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     defaultBool(false),
				Description: "Boolean that indicate if this variable group is shared by all pipelines of this project.",
			},
			// variable is a ListNestedAttribute (not Set) so that:
			//   1. HCL syntax is unchanged: variable = [{ name = "..." ... }]
			//   2. List uses positional indexing which avoids the set-hash bug where
			//      Default values applied to nested sensitive attributes change the
			//      element hash mid-transform, causing sibling defaults to fail and
			//      producing "inconsistent values for sensitive attribute" errors.
			// Elements are returned sorted by name to give a stable canonical ordering.
			"variable": schema.ListNestedAttribute{
				Required:    true,
				Description: "One or more variable blocks as documented below.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The key or name of the variable.",
						},
						"value": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     defaultString(""),
							Description: "The value of the variable.",
						},
						"secret_value": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Sensitive:   true,
							Default:     defaultString(""),
							Description: "The secret value of the variable. Used when `is_secret` is `true`.",
						},
						"is_secret": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     defaultBool(false),
							Description: "If `true`, the variable will be treated as a secret.",
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
				Optional:    true,
				Description: "A key_vault block as documented below.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the Azure Key Vault.",
						},
						"service_endpoint_id": schema.StringAttribute{
							Required:    true,
							Description: "The ID of the Azure RM service endpoint.",
						},
						"search_depth": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The number of pages to search for the secrets. Defaults to `20`.",
						},
					},
				},
			},
		},
	}
}

func (r *VariableGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.AggregatedClient, got: %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *VariableGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: <projectID>/<variableGroupID>
	parts := splitImportID(req.ID)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format <projectID>/<variableGroupID>, got: %s", req.ID),
		)
		return
	}
	projectID := parts[0]
	idStr := parts[1]

	// Resolve numeric ID
	vgID, err := strconv.Atoi(idStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid variable group ID", fmt.Sprintf("%s is not an integer: %v", idStr, err))
		return
	}

	vg, err := r.client.TaskAgentClient.GetVariableGroup(ctx, taskagent.GetVariableGroupArgs{
		GroupId: &vgID,
		Project: &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading variable group", err.Error())
		return
	}
	if vg == nil || vg.Id == nil {
		resp.Diagnostics.AddError("Variable group not found", fmt.Sprintf("Variable group %s not found in project %s", idStr, projectID))
		return
	}

	state, diags := r.flattenToModel(ctx, vg, projectID, nil)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read allow_access
	resourceRefType := "variablegroup"
	varGroupIDStr := strconv.Itoa(*vg.Id)
	projectResources, err := r.client.BuildClient.GetProjectResources(ctx, build.GetProjectResourcesArgs{
		Project: &projectID,
		Type:    &resourceRefType,
		Id:      &varGroupIDStr,
	})
	if err == nil {
		state.AllowAccess = types.BoolValue(extractAllowAccess(projectResources, varGroupIDStr))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *VariableGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan variableGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ProjectID.ValueString()
	params, diags := r.expandToParams(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.createVariableGroupWithRetry(ctx, params, &projectID)
	if err != nil {
		resp.Diagnostics.AddError("Error creating variable group", err.Error())
		return
	}

	// Update allow_access
	allowAccess := plan.AllowAccess.ValueBool()
	vgIDStr := strconv.Itoa(*created.Id)
	resourceRefType := "variablegroup"
	defRef := []build.DefinitionResourceReference{{
		Type:       &resourceRefType,
		Authorized: &allowAccess,
		Name:       created.Name,
		Id:         &vgIDStr,
	}}
	projectResources, err := r.client.BuildClient.AuthorizeProjectResources(ctx, build.AuthorizeProjectResourcesArgs{
		Resources: &defRef,
		Project:   &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error setting allow_access", err.Error())
		return
	}

	// Build state using plan as prior-state so that secret values and ordering
	// are preserved faithfully (the API never returns secret values, and the
	// plan already has the correct user-supplied values).
	state, diags2 := r.flattenToModel(ctx, created, projectID, &plan)
	resp.Diagnostics.Append(diags2...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.AllowAccess = types.BoolValue(extractAllowAccess(projectResources, vgIDStr))

	// Preserve the plan's variable list order and sensitive values verbatim.
	// The API never returns secret values and may iterate map keys in a
	// different order; using the plan list avoids "inconsistent values for
	// sensitive attribute" errors from Terraform core's post-apply check.
	mergedVars, mergeDiags := r.mergeVariableListWithPlan(ctx, state.Variable, plan.Variable)
	resp.Diagnostics.Append(mergeDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Variable = mergedVars

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *VariableGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state variableGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := state.ProjectID.ValueString()
	idStr := state.ID.ValueString()
	vgID, err := strconv.Atoi(idStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid variable group ID in state", err.Error())
		return
	}

	vg, err := r.client.TaskAgentClient.GetVariableGroup(ctx, taskagent.GetVariableGroupArgs{
		GroupId: &vgID,
		Project: &projectID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading variable group", err.Error())
		return
	}
	if vg == nil || vg.Id == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState, diags := r.flattenToModel(ctx, vg, projectID, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read allow_access
	resourceRefType := "variablegroup"
	varGroupIDStr := strconv.Itoa(*vg.Id)
	projectResources, err := r.client.BuildClient.GetProjectResources(ctx, build.GetProjectResourcesArgs{
		Project: &projectID,
		Type:    &resourceRefType,
		Id:      &varGroupIDStr,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading allow_access", err.Error())
		return
	}
	newState.AllowAccess = types.BoolValue(extractAllowAccess(projectResources, varGroupIDStr))

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *VariableGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state variableGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ProjectID.ValueString()
	idStr := state.ID.ValueString()
	vgID, err := strconv.Atoi(idStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid variable group ID in state", err.Error())
		return
	}

	params, diags := r.expandToParams(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err = r.client.TaskAgentClient.UpdateVariableGroup(ctx, taskagent.UpdateVariableGroupArgs{
		GroupId:                 &vgID,
		VariableGroupParameters: params,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating variable group", err.Error())
		return
	}

	// ADO's UpdateVariableGroup may return stale data from the previous state.
	// Do a fresh Get to ensure the post-apply state reflects the actual API state.
	updated, err := r.client.TaskAgentClient.GetVariableGroup(ctx, taskagent.GetVariableGroupArgs{
		GroupId: &vgID,
		Project: &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading variable group after update", err.Error())
		return
	}
	if updated == nil || updated.Id == nil {
		resp.Diagnostics.AddError("Variable group not found after update", fmt.Sprintf("Variable group %d not found in project %s after update", vgID, projectID))
		return
	}

	// Update allow_access
	allowAccess := plan.AllowAccess.ValueBool()
	vgIDStr := strconv.Itoa(*updated.Id)
	resourceRefType := "variablegroup"
	defRef := []build.DefinitionResourceReference{{
		Type:       &resourceRefType,
		Authorized: &allowAccess,
		Name:       updated.Name,
		Id:         &vgIDStr,
	}}
	projectResources, err := r.client.BuildClient.AuthorizeProjectResources(ctx, build.AuthorizeProjectResourcesArgs{
		Resources: &defRef,
		Project:   &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error setting allow_access", err.Error())
		return
	}

	// Build state using plan as prior-state so that secret values and ordering
	// are preserved faithfully.
	newState, diags2 := r.flattenToModel(ctx, updated, projectID, &plan)
	resp.Diagnostics.Append(diags2...)
	if resp.Diagnostics.HasError() {
		return
	}
	newState.AllowAccess = types.BoolValue(extractAllowAccess(projectResources, vgIDStr))

	// For non-computed user-configured attributes (name, description), always
	// use the plan values to avoid "inconsistent result after apply" errors from
	// ADO's eventual consistency on updates (the fresh Get may still return the
	// prior values for a brief window after the PUT).
	newState.Name = plan.Name
	newState.Description = plan.Description

	// Preserve plan order and sensitive values verbatim.
	mergedVarsU, mergeDiagsU := r.mergeVariableListWithPlan(ctx, newState.Variable, plan.Variable)
	resp.Diagnostics.Append(mergeDiagsU...)
	if resp.Diagnostics.HasError() {
		return
	}
	newState.Variable = mergedVarsU

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *VariableGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state variableGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := state.ProjectID.ValueString()
	idStr := state.ID.ValueString()
	vgID, err := strconv.Atoi(idStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid variable group ID in state", err.Error())
		return
	}

	// Remove allow_access first (set to unauthorized)
	auth := false
	emptyName := ""
	resourceRefType := "variablegroup"
	defRef := []build.DefinitionResourceReference{{
		Type:       &resourceRefType,
		Authorized: &auth,
		Name:       &emptyName,
		Id:         &idStr,
	}}
	//nolint:errcheck // best-effort: unauthorize before delete; ignore failure
	_, _ = r.client.BuildClient.AuthorizeProjectResources(ctx, build.AuthorizeProjectResourcesArgs{
		Resources: &defRef,
		Project:   &projectID,
	})

	err = r.client.TaskAgentClient.DeleteVariableGroup(ctx, taskagent.DeleteVariableGroupArgs{
		ProjectIds: &[]string{projectID},
		GroupId:    &vgID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting variable group", err.Error())
		return
	}

	// ADO variable group deletion is eventually consistent — wait until the
	// resource is no longer found before returning, so that CheckDestroy in
	// acceptance tests does not race against the API.
	//
	// ContinuousTargetOccurence: 3 requires three consecutive "not found"
	// results before we declare deletion confirmed. This prevents a transient
	// 404 / error flash (common in ADO's caching layer) from being interpreted
	// as a successful deletion while the VG is actually still being removed.
	deleteConf := &retry.StateChangeConf{
		Pending:                   []string{"deleting"},
		Target:                    []string{"deleted"},
		ContinuousTargetOccurence: 3,
		Delay:                     5 * time.Second,
		MinTimeout:                5 * time.Second,
		Timeout:                   5 * time.Minute,
		Refresh: func() (interface{}, string, error) {
			vg, getErr := r.client.TaskAgentClient.GetVariableGroup(ctx, taskagent.GetVariableGroupArgs{
				GroupId: &vgID,
				Project: &projectID,
			})
			if getErr != nil {
				// Any API error (404, 400 "not found", etc.) means the VG is
				// no longer accessible — count as one "deleted" occurrence.
				// We intentionally discard the error: the VG being gone is the
				// success condition, not a failure.
				return vgID, "deleted", nil //nolint:nilerr
			}
			if vg == nil || vg.Id == nil {
				return vgID, "deleted", nil
			}
			return vg, "deleting", nil
		},
	}
	if _, waitErr := deleteConf.WaitForStateContext(ctx); waitErr != nil {
		resp.Diagnostics.AddError("Timeout waiting for variable group deletion", waitErr.Error())
	}
}

// ── Expand / Flatten helpers ─────────────────────────────────────────────────

// expandToParams converts the Terraform plan model to an AzDO VariableGroupParameters.
func (r *VariableGroupResource) expandToParams(ctx context.Context, plan *variableGroupModel) (*taskagent.VariableGroupParameters, diag.Diagnostics) {
	var diags diag.Diagnostics

	projectID := plan.ProjectID.ValueString()
	name := plan.Name.ValueString()
	description := plan.Description.ValueString()

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		diags.AddError("Invalid project_id", fmt.Sprintf("Cannot parse project_id %q as UUID: %v", projectID, err))
		return nil, diags
	}

	// Expand variables from list
	var variableElems []variableModel
	diags.Append(plan.Variable.ElementsAs(ctx, &variableElems, false)...)
	if diags.HasError() {
		return nil, diags
	}

	variableMap := make(map[string]interface{})
	for _, v := range variableElems {
		varName := v.Name.ValueString()
		isSecret := v.IsSecret.ValueBool()
		if isSecret {
			variableMap[varName] = taskagent.VariableValue{
				Value:    converter.String(v.SecretValue.ValueString()),
				IsSecret: converter.Bool(true),
			}
		} else {
			variableMap[varName] = taskagent.VariableValue{
				Value:    converter.String(v.Value.ValueString()),
				IsSecret: converter.Bool(false),
			}
		}
	}

	params := &taskagent.VariableGroupParameters{
		Name:        &name,
		Description: &description,
		Variables:   &variableMap,
		VariableGroupProjectReferences: &[]taskagent.VariableGroupProjectReference{
			{
				Description: &description,
				Name:        &name,
				ProjectReference: &taskagent.ProjectReference{
					Id: &projectUUID,
				},
			},
		},
	}

	// Key vault
	if !plan.KeyVault.IsNull() && !plan.KeyVault.IsUnknown() {
		var kvElems []keyVaultModel
		diags.Append(plan.KeyVault.ElementsAs(ctx, &kvElems, false)...)
		if diags.HasError() {
			return nil, diags
		}
		if len(kvElems) == 1 {
			kv := kvElems[0]
			kvName := kv.Name.ValueString()
			seID := kv.ServiceEndpointID.ValueString()
			depth := int(kv.SearchDepth.ValueInt64())
			if depth == 0 {
				depth = 20
			}

			seUUID, err := uuid.Parse(seID)
			if err != nil {
				diags.AddError("Invalid service_endpoint_id", fmt.Sprintf("Cannot parse %q as UUID: %v", seID, err))
				return nil, diags
			}

			params.ProviderData = taskagent.AzureKeyVaultVariableGroupProviderData{
				ServiceEndpointId: &seUUID,
				Vault:             &kvName,
				LastRefreshedOn:   &azuredevops.Time{Time: time.Now()},
			}
			params.Type = converter.String(azureKeyVaultType)

			// searchAzureKVSecrets expects variables as []interface{} with a "name" key
			// in each element (matches the SDKv2 TypeSet element format).
			kvVarList := make([]interface{}, 0, len(variableElems))
			for _, v := range variableElems {
				kvVarList = append(kvVarList, map[string]interface{}{
					"name": v.Name.ValueString(),
				})
			}
			kvSecrets, invalidSecrets, err := searchAzureKVSecrets(r.client, projectID, kvName, seID, kvVarList, depth)
			if err != nil {
				diags.AddError("Error fetching Key Vault secrets", err.Error())
				return nil, diags
			}
			if len(invalidSecrets) > 0 {
				diags.AddError("Invalid Key Vault secrets", fmt.Sprintf("Cannot find secrets %v in vault %s", invalidSecrets, kvName))
				return nil, diags
			}
			params.Variables = &kvSecrets
		}
	}

	return params, diags
}

// flattenToModel converts an AzDO VariableGroup into the Terraform state model.
//
// priorState is used to:
//   - recover secret_value (API never returns secret values), and
//   - determine the ordering of the variable list. Variables are emitted in
//     the same order as they appear in priorState (to avoid spurious list-
//     index diffs on subsequent plans); API variables not present in priorState
//     are appended alphabetically at the end.
//
// When priorState is nil (import) variables are returned sorted alphabetically.
func (r *VariableGroupResource) flattenToModel(ctx context.Context, vg *taskagent.VariableGroup, projectID string, priorState *variableGroupModel) (variableGroupModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	model := variableGroupModel{
		ID:          types.StringValue(strconv.Itoa(*vg.Id)),
		ProjectID:   types.StringValue(projectID),
		Name:        types.StringValue(converter.ToString(vg.Name, "")),
		Description: types.StringValue(converter.ToString(vg.Description, "")),
		AllowAccess: types.BoolValue(false), // filled by caller
	}

	// Build prior-state maps for secret recovery and ordering.
	priorSecrets := map[string]string{}
	priorOrder := []string{} // ordered names from prior state
	priorSet := map[string]bool{}
	if priorState != nil && !priorState.Variable.IsNull() {
		var priorVars []variableModel
		diags.Append(priorState.Variable.ElementsAs(ctx, &priorVars, false)...)
		if diags.HasError() {
			return model, diags
		}
		for _, pv := range priorVars {
			n := pv.Name.ValueString()
			priorSecrets[n] = pv.SecretValue.ValueString()
			priorOrder = append(priorOrder, n)
			priorSet[n] = true
		}
	}

	// Build the ordered variable name list:
	//   1. Names present in prior state, in prior state order (skip API-deleted ones).
	//   2. Names in the API result not present in prior state, alphabetically appended.
	apiVarSet := make(map[string]bool, len(*vg.Variables))
	for varName := range *vg.Variables {
		apiVarSet[varName] = true
	}

	varNames := make([]string, 0, len(*vg.Variables))
	for _, n := range priorOrder {
		if apiVarSet[n] {
			varNames = append(varNames, n)
		}
	}
	// Append new names (in API result but not prior state) alphabetically.
	newNames := make([]string, 0)
	for n := range apiVarSet {
		if !priorSet[n] {
			newNames = append(newNames, n)
		}
	}
	sort.Strings(newNames)
	varNames = append(varNames, newNames...)

	// Build variables list (sorted by name).
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
			secretValStr := ""
			if isSecret {
				// API never returns secret values — recover from prior state
				secretValStr = priorSecrets[varName]
			} else if apiVar.Value != nil {
				valStr = *apiVar.Value
			}
			vm = variableModel{
				Name:        types.StringValue(varName),
				Value:       types.StringValue(valStr),
				SecretValue: types.StringValue(secretValStr),
				IsSecret:    types.BoolValue(isSecret),
				ContentType: types.StringValue(""),
				Enabled:     types.BoolValue(false),
				Expires:     types.StringValue(""),
			}
		}

		obj, d := types.ObjectValueFrom(ctx, variableAttrTypes, vm)
		diags.Append(d...)
		if diags.HasError() {
			return model, diags
		}
		varObjects = append(varObjects, obj)
	}

	varList, d := types.ListValue(types.ObjectType{AttrTypes: variableAttrTypes}, varObjects)
	diags.Append(d...)
	if diags.HasError() {
		return model, diags
	}
	model.Variable = varList

	// Key vault block — return null when not a key-vault VG so that the planned
	// null matches the state null (returning an empty list would cause Terraform
	// to report "was null, but now cty.ListValEmpty(...)").
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

		// Recover search_depth from prior state
		searchDepth := int64(20)
		if priorState != nil && !priorState.KeyVault.IsNull() {
			var priorKV []keyVaultModel
			diags.Append(priorState.KeyVault.ElementsAs(ctx, &priorKV, false)...)
			if len(priorKV) == 1 {
				searchDepth = priorKV[0].SearchDepth.ValueInt64()
			}
		}

		kvObj, d := types.ObjectValueFrom(ctx, keyVaultAttrTypes, keyVaultModel{
			Name:              types.StringValue(vaultName),
			ServiceEndpointID: types.StringValue(seID),
			SearchDepth:       types.Int64Value(searchDepth),
		})
		diags.Append(d...)
		if diags.HasError() {
			return model, diags
		}
		kvList, d := types.ListValue(types.ObjectType{AttrTypes: keyVaultAttrTypes}, []attr.Value{kvObj})
		diags.Append(d...)
		if diags.HasError() {
			return model, diags
		}
		model.KeyVault = kvList
	} else {
		// Return null (not empty list) for key_vault when the VG is not a
		// key-vault type.  The schema marks key_vault as Optional (not
		// Computed), so the plan value is null when the user omits it; an
		// empty list would mismatch and trigger an "inconsistent result"
		// error from Terraform core.
		model.KeyVault = types.ListNull(types.ObjectType{AttrTypes: keyVaultAttrTypes})
	}

	return model, diags
}

// mergeVariableListWithPlan returns the API-derived list but re-ordered and
// re-valued to match the plan's variable list.  This is necessary for two
// reasons:
//
//  1. Plan order: the plan list is in user-config order; the API-derived list
//     is sorted alphabetically by name (flattenToModel's stable ordering).
//     Terraform's post-apply consistency check compares by list index, so the
//     orders must match.
//
//  2. Sensitive values: the API never returns secret values. flattenToModel
//     recovers them from priorState, but for Create the priorState is the
//     plan (secret_value is known in plan, unknown in "prior state").  By
//     copying plan elements for variables that exist in the plan, we ensure
//     the returned state has the exact sensitive values Terraform planned,
//     which satisfies the "inconsistent values for sensitive attribute" check.
func (r *VariableGroupResource) mergeVariableListWithPlan(ctx context.Context, apiList, planList types.List) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	var planVars []variableModel
	if d := planList.ElementsAs(ctx, &planVars, false); d.HasError() {
		diags.Append(d...)
		return apiList, diags
	}

	var apiVars []variableModel
	if d := apiList.ElementsAs(ctx, &apiVars, false); d.HasError() {
		diags.Append(d...)
		return apiList, diags
	}

	// Build name→apiVar lookup for computed fields (content_type, enabled, expires).
	apiByName := make(map[string]variableModel, len(apiVars))
	for _, v := range apiVars {
		apiByName[v.Name.ValueString()] = v
	}

	// Build result in plan order: prefer plan values for configured attributes,
	// take computed attributes from the API result.
	merged := make([]attr.Value, 0, len(planVars))
	for _, pv := range planVars {
		name := pv.Name.ValueString()
		av, inAPI := apiByName[name]

		m := variableModel{
			Name:        pv.Name,
			Value:       pv.Value,
			SecretValue: pv.SecretValue,
			IsSecret:    pv.IsSecret,
		}
		// Computed-only fields come from the API result when available.
		if inAPI {
			m.ContentType = av.ContentType
			m.Enabled = av.Enabled
			m.Expires = av.Expires
		} else {
			m.ContentType = types.StringValue("")
			m.Enabled = types.BoolValue(false)
			m.Expires = types.StringValue("")
		}

		obj, d := types.ObjectValueFrom(ctx, variableAttrTypes, m)
		if d.HasError() {
			diags.Append(d...)
			return apiList, diags
		}
		merged = append(merged, obj)
	}

	result, d := types.ListValue(types.ObjectType{AttrTypes: variableAttrTypes}, merged)
	if d.HasError() {
		diags.Append(d...)
		return apiList, diags
	}
	return result, diags
}

// createVariableGroupWithRetry creates a variable group and waits until it is ready.
func (r *VariableGroupResource) createVariableGroupWithRetry(ctx context.Context, params *taskagent.VariableGroupParameters, project *string) (*taskagent.VariableGroup, error) {
	created, err := r.client.TaskAgentClient.AddVariableGroup(ctx, taskagent.AddVariableGroupArgs{
		VariableGroupParameters: params,
	})
	if err != nil {
		return nil, err
	}

	stateConf := &retry.StateChangeConf{
		ContinuousTargetOccurence: 2,
		Delay:                     5 * time.Second,
		MinTimeout:                10 * time.Second,
		Pending:                   []string{"inProgress"},
		Target:                    []string{"succeeded", "failed"},
		Refresh: func() (interface{}, string, error) {
			vg, err := r.client.TaskAgentClient.GetVariableGroup(ctx, taskagent.GetVariableGroupArgs{
				GroupId: created.Id,
				Project: project,
			})
			if err != nil {
				return vg, "failed", fmt.Errorf("getting variable group: %w", err)
			}
			if vg != nil && vg.Variables != nil && params.Variables != nil && len(*params.Variables) == len(*vg.Variables) {
				return vg, "succeeded", nil
			}
			return vg, "inProgress", nil
		},
		Timeout: 10 * time.Minute,
	}

	result, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("waiting for variable group to be ready: %w", err)
	}
	return result.(*taskagent.VariableGroup), nil
}

// extractAllowAccess extracts the allow_access value from the project resources.
func extractAllowAccess(refs *[]build.DefinitionResourceReference, vgID string) bool {
	if refs == nil {
		return false
	}
	for _, ref := range *refs {
		if ref.Id != nil && *ref.Id == vgID && ref.Authorized != nil {
			return *ref.Authorized
		}
	}
	return false
}

// splitImportID splits an import ID "a/b" into ["a", "b"] by the last slash.
func splitImportID(id string) []string {
	for i := len(id) - 1; i >= 0; i-- {
		if id[i] == '/' {
			return []string{id[:i], id[i+1:]}
		}
	}
	return []string{id}
}
