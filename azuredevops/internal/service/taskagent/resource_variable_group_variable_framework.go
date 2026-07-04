package taskagent

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = (*VariableGroupVariableResource)(nil)
	_ resource.ResourceWithConfigure   = (*VariableGroupVariableResource)(nil)
	_ resource.ResourceWithImportState = (*VariableGroupVariableResource)(nil)
)

// forEachLockFramework serialises concurrent Create/Update/Delete operations
// on the same variable group to avoid race conditions on the group-level PUT.
var forEachLockFramework = new(sync.Mutex)

// VariableGroupVariableResource is the terraform-plugin-framework implementation
// of betterado_variable_group_variable.
type VariableGroupVariableResource struct {
	client *client.AggregatedClient
}

// NewVariableGroupVariableResource returns a new resource.Resource.
func NewVariableGroupVariableResource() resource.Resource {
	return &VariableGroupVariableResource{}
}

// variableGroupVariableModel is the Terraform state model.
type variableGroupVariableModel struct {
	ID              types.String `tfsdk:"id"`
	ProjectID       types.String `tfsdk:"project_id"`
	VariableGroupID types.String `tfsdk:"variable_group_id"`
	Name            types.String `tfsdk:"name"`
	Value           types.String `tfsdk:"value"`
	SecretValue     types.String `tfsdk:"secret_value"`
}

func (r *VariableGroupVariableResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_variable_group_variable"
}

func (r *VariableGroupVariableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a single variable within an Azure DevOps Variable Group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the variable group variable, in the format `project_id/variable_group_id/name`.",
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
			"variable_group_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the variable group.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the variable.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Optional:    true,
				Description: "The value of the variable. Conflicts with `secret_value`.",
			},
			"secret_value": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The secret value of the variable. Conflicts with `value`.",
				PlanModifiers: []planmodifier.String{
					useStateForUnknown(),
				},
			},
		},
	}
}

func (r *VariableGroupVariableResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = c
}

func (r *VariableGroupVariableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan variableGroupVariableModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	forEachLockFramework.Lock()
	defer forEachLockFramework.Unlock()

	projectID := plan.ProjectID.ValueString()
	variableGroupIDStr := plan.VariableGroupID.ValueString()
	variableGroupID, err := strconv.Atoi(variableGroupIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid variable_group_id", fmt.Sprintf("parsing variable_group_id as integer: %v", err))
		return
	}

	vg, err := r.client.TaskAgentClient.GetVariableGroup(ctx, taskagent.GetVariableGroupArgs{
		GroupId: &variableGroupID,
		Project: &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading variable group", fmt.Sprintf("GetVariableGroup: %v", err))
		return
	}
	if vg.Variables == nil {
		resp.Diagnostics.AddError("Unexpected null variables", "variable group returned nil Variables map")
		return
	}

	vars := *vg.Variables
	name := plan.Name.ValueString()
	id := fmt.Sprintf("%s/%d/%s", projectID, variableGroupID, name)

	// Existence check
	if _, ok := vars[name]; ok {
		resp.Diagnostics.AddError(
			"Variable already exists",
			fmt.Sprintf("A variable named %q already exists in the variable group. To manage it, use `terraform import %s %s`.", name, "betterado_variable_group_variable.<name>", id),
		)
		return
	}

	var (
		value    string
		isSecret bool
	)
	if !plan.Value.IsNull() && !plan.Value.IsUnknown() {
		value = plan.Value.ValueString()
	} else if !plan.SecretValue.IsNull() && !plan.SecretValue.IsUnknown() {
		value = plan.SecretValue.ValueString()
		isSecret = true
	}

	vars[name] = map[string]any{
		"value":    value,
		"isSecret": isSecret,
	}

	params := taskagent.VariableGroupParameters{
		Description:                    vg.Description,
		Name:                           vg.Name,
		ProviderData:                   vg.ProviderData,
		Type:                           vg.Type,
		VariableGroupProjectReferences: vg.VariableGroupProjectReferences,
		Variables:                      &vars,
	}

	if _, err := updateVariableGroup(r.client, &params, &variableGroupID); err != nil {
		resp.Diagnostics.AddError("Error updating variable group", fmt.Sprintf("updateVariableGroup: %v", err))
		return
	}

	plan.ID = types.StringValue(id)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VariableGroupVariableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state variableGroupVariableModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID, variableGroupID, varName, err := ResourceVariableGroupVariableParseId(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error parsing resource ID", err.Error())
		return
	}

	vg, err := r.client.TaskAgentClient.GetVariableGroup(ctx, taskagent.GetVariableGroupArgs{
		GroupId: &variableGroupID,
		Project: &projectID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading variable group", fmt.Sprintf("GetVariableGroup: %v", err))
		return
	}

	if vg.Variables == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	vars := *vg.Variables
	varVal, ok := vars[varName]
	if !ok {
		resp.State.RemoveResource(ctx)
		return
	}

	var (
		value    string
		isSecret bool
	)
	if varMap, ok := varVal.(map[string]any); ok {
		if v, ok := varMap["value"].(string); ok {
			value = v
		}
		if v, ok := varMap["isSecret"].(bool); ok {
			isSecret = v
		}
	}

	state.ProjectID = types.StringValue(projectID)
	state.VariableGroupID = types.StringValue(strconv.Itoa(variableGroupID))
	state.Name = types.StringValue(varName)

	if isSecret {
		// ADO doesn't return secret values; keep what's in state.
		state.Value = types.StringNull()
		// secret_value stays as-is from state (UseStateForUnknown handles this)
	} else {
		state.Value = types.StringValue(value)
		state.SecretValue = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *VariableGroupVariableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan variableGroupVariableModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state variableGroupVariableModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	forEachLockFramework.Lock()
	defer forEachLockFramework.Unlock()

	projectID := plan.ProjectID.ValueString()
	variableGroupIDStr := plan.VariableGroupID.ValueString()
	variableGroupID, err := strconv.Atoi(variableGroupIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid variable_group_id", fmt.Sprintf("parsing variable_group_id as integer: %v", err))
		return
	}

	vg, err := r.client.TaskAgentClient.GetVariableGroup(ctx, taskagent.GetVariableGroupArgs{
		GroupId: &variableGroupID,
		Project: &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading variable group", fmt.Sprintf("GetVariableGroup: %v", err))
		return
	}
	if vg.Variables == nil {
		resp.Diagnostics.AddError("Unexpected null variables", "variable group returned nil Variables map")
		return
	}

	vars := *vg.Variables
	name := plan.Name.ValueString()

	var (
		value    string
		isSecret bool
	)
	if !plan.Value.IsNull() && !plan.Value.IsUnknown() {
		value = plan.Value.ValueString()
	} else if !plan.SecretValue.IsNull() && !plan.SecretValue.IsUnknown() {
		value = plan.SecretValue.ValueString()
		isSecret = true
	}

	vars[name] = map[string]any{
		"value":    value,
		"isSecret": isSecret,
	}

	params := taskagent.VariableGroupParameters{
		Description:                    vg.Description,
		Name:                           vg.Name,
		ProviderData:                   vg.ProviderData,
		Type:                           vg.Type,
		VariableGroupProjectReferences: vg.VariableGroupProjectReferences,
		Variables:                      &vars,
	}

	if _, err := updateVariableGroup(r.client, &params, &variableGroupID); err != nil {
		resp.Diagnostics.AddError("Error updating variable group", fmt.Sprintf("updateVariableGroup: %v", err))
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *VariableGroupVariableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state variableGroupVariableModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	forEachLockFramework.Lock()
	defer forEachLockFramework.Unlock()

	projectID := state.ProjectID.ValueString()
	variableGroupIDStr := state.VariableGroupID.ValueString()
	variableGroupID, err := strconv.Atoi(variableGroupIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid variable_group_id", fmt.Sprintf("parsing variable_group_id as integer: %v", err))
		return
	}

	vg, err := r.client.TaskAgentClient.GetVariableGroup(ctx, taskagent.GetVariableGroupArgs{
		GroupId: &variableGroupID,
		Project: &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading variable group", fmt.Sprintf("GetVariableGroup: %v", err))
		return
	}
	if vg.Variables == nil {
		// Nothing to delete
		return
	}

	vars := *vg.Variables
	name := state.Name.ValueString()

	if _, ok := vars[name]; !ok {
		// Variable already gone
		return
	}

	delete(vars, name)

	params := taskagent.VariableGroupParameters{
		Description:                    vg.Description,
		Name:                           vg.Name,
		ProviderData:                   vg.ProviderData,
		Type:                           vg.Type,
		VariableGroupProjectReferences: vg.VariableGroupProjectReferences,
		Variables:                      &vars,
	}

	if _, err := updateVariableGroup(r.client, &params, &variableGroupID); err != nil {
		resp.Diagnostics.AddError("Error updating variable group", fmt.Sprintf("updateVariableGroup: %v", err))
		return
	}
}

func (r *VariableGroupVariableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	segs := strings.SplitN(id, "/", 3)
	if len(segs) != 3 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format: <project_id>/<variable_group_id>/<variable_name>, got: %q", id),
		)
		return
	}

	state := variableGroupVariableModel{
		ID:              types.StringValue(id),
		ProjectID:       types.StringValue(segs[0]),
		VariableGroupID: types.StringValue(segs[1]),
		Name:            types.StringValue(segs[2]),
		Value:           types.StringNull(),
		SecretValue:     types.StringNull(),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ResourceVariableGroupVariableParseId parses the resource ID string
// "<project_id>/<variable_group_id>/<variable_name>" and returns its components.
// This function is exported so acceptance tests can use it directly.
func ResourceVariableGroupVariableParseId(id string) (string, int, string, error) {
	segs := strings.SplitN(id, "/", 3)
	if len(segs) != 3 {
		return "", 0, "", fmt.Errorf("invalid resource id, expect length=3, got=%d", len(segs))
	}
	projectID, variableGroupIDStr, varName := segs[0], segs[1], segs[2]
	variableGroupID, err := strconv.Atoi(variableGroupIDStr)
	if err != nil {
		return "", 0, "", fmt.Errorf("converting the variable group id as integer: %v", err)
	}
	return projectID, variableGroupID, varName, nil
}
