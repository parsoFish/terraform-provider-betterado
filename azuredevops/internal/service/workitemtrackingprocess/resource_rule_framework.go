package workitemtrackingprocess

// resource_rule_framework.go — terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_rule.

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &ruleResource{}
	_ resource.ResourceWithConfigure   = &ruleResource{}
	_ resource.ResourceWithImportState = &ruleResource{}
)

// ── Inline plan modifiers and defaults ────────────────────────────────────────

type ruleUseStateForUnknownString struct{}

func (ruleUseStateForUnknownString) Description(_ context.Context) string { return "use prior state" }

func (ruleUseStateForUnknownString) MarkdownDescription(_ context.Context) string {
	return "use prior state"
}

func (ruleUseStateForUnknownString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type ruleRequiresReplaceString struct{}

func (ruleRequiresReplaceString) Description(_ context.Context) string {
	return "requires replacement if changed"
}

func (ruleRequiresReplaceString) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}

func (ruleRequiresReplaceString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type ruleStaticBoolDefault struct{ value bool }

func (s ruleStaticBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("default value %v", s.value)
}

func (s ruleStaticBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("default value `%v`", s.value)
}

func (s ruleStaticBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(s.value)
}

// ── Validators ────────────────────────────────────────────────────────────────

var validConditionTypes = []string{
	"when",
	"whenNot",
	"whenChanged",
	"whenNotChanged",
	"whenWas",
	"whenCurrentUserIsMemberOfGroup",
	"whenCurrentUserIsNotMemberOfGroup",
}

var validActionTypes = []string{
	"makeRequired",
	"makeReadOnly",
	"setDefaultValue",
	"setDefaultFromClock",
	"setDefaultFromField",
	"copyValue",
	"copyFromClock",
	"copyFromCurrentUser",
	"copyFromField",
	"setValueToEmpty",
	"copyFromServerClock",
	"copyFromServerCurrentUser",
	"hideTargetField",
	"disallowValue",
}

type ruleConditionTypeValidator struct{}

func (v ruleConditionTypeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Valid condition types: %s", strings.Join(validConditionTypes, ", "))
}

func (v ruleConditionTypeValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Valid condition types: `%s`", strings.Join(validConditionTypes, "`, `"))
}

func (v ruleConditionTypeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	for _, valid := range validConditionTypes {
		if val == valid {
			return
		}
	}
	resp.Diagnostics.AddAttributeError(req.Path, "Invalid condition_type",
		fmt.Sprintf("Got %q. Valid values: %s", val, strings.Join(validConditionTypes, ", ")))
}

type ruleActionTypeValidator struct{}

func (v ruleActionTypeValidator) Description(_ context.Context) string {
	return fmt.Sprintf("Valid action types: %s", strings.Join(validActionTypes, ", "))
}

func (v ruleActionTypeValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Valid action types: `%s`", strings.Join(validActionTypes, "`, `"))
}

func (v ruleActionTypeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	for _, valid := range validActionTypes {
		if val == valid {
			return
		}
	}
	resp.Diagnostics.AddAttributeError(req.Path, "Invalid action_type",
		fmt.Sprintf("Got %q. Valid values: %s", val, strings.Join(validActionTypes, ", ")))
}

// ── Resource struct ────────────────────────────────────────────────────────────

type ruleResource struct {
	client *client.AggregatedClient
}

// NewRuleResource returns a new resource.Resource for betterado_workitemtrackingprocess_rule.
func NewRuleResource() resource.Resource {
	return &ruleResource{}
}

// ── Model ──────────────────────────────────────────────────────────────────────

type ruleResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProcessID      types.String `tfsdk:"process_id"`
	WorkItemTypeID types.String `tfsdk:"work_item_type_id"`
	Name           types.String `tfsdk:"name"`
	IsEnabled      types.Bool   `tfsdk:"is_enabled"`
	Conditions     types.Set    `tfsdk:"condition"`
	Actions        types.Set    `tfsdk:"action"`
	URL            types.String `tfsdk:"url"`
}

type ruleConditionModel struct {
	ConditionType types.String `tfsdk:"condition_type"`
	Field         types.String `tfsdk:"field"`
	Value         types.String `tfsdk:"value"`
}

type ruleActionModel struct {
	ActionType  types.String `tfsdk:"action_type"`
	TargetField types.String `tfsdk:"target_field"`
	Value       types.String `tfsdk:"value"`
}

var conditionAttrTypes = map[string]attr.Type{
	"condition_type": types.StringType,
	"field":          types.StringType,
	"value":          types.StringType,
}

var actionAttrTypes = map[string]attr.Type{
	"action_type":  types.StringType,
	"target_field": types.StringType,
	"value":        types.StringType,
}

// ── Metadata ───────────────────────────────────────────────────────────────────

func (r *ruleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_rule"
}

// ── Schema ─────────────────────────────────────────────────────────────────────

func (r *ruleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a rule on a work item type in an Azure DevOps process.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the rule.",
				PlanModifiers: []planmodifier.String{
					ruleUseStateForUnknownString{},
				},
			},
			"process_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the process.",
				PlanModifiers: []planmodifier.String{
					ruleRequiresReplaceString{},
				},
			},
			"work_item_type_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID (reference name) of the work item type.",
				PlanModifiers: []planmodifier.String{
					ruleRequiresReplaceString{},
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the rule.",
			},
			"is_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     ruleStaticBoolDefault{value: true},
				Description: "Indicates if the rule is enabled.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "URL of the rule resource.",
				PlanModifiers: []planmodifier.String{
					ruleUseStateForUnknownString{},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"condition": schema.SetNestedBlock{
				Description: "Set of conditions when the rule should be triggered.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"condition_type": schema.StringAttribute{
							Required:    true,
							Description: "Type of condition.",
							Validators: []validator.String{
								ruleConditionTypeValidator{},
							},
						},
						"field": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Field reference name for the condition.",
						},
						"value": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Value to match for the condition.",
						},
					},
				},
			},
			"action": schema.SetNestedBlock{
				Description: "Set of actions to take when the rule is triggered.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"action_type": schema.StringAttribute{
							Required:    true,
							Description: "Type of action.",
							Validators: []validator.String{
								ruleActionTypeValidator{},
							},
						},
						"target_field": schema.StringAttribute{
							Required:    true,
							Description: "Field (reference name) to act on.",
						},
						"value": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Value to set on the target field.",
						},
					},
				},
			},
		},
	}
}

// ── Configure ──────────────────────────────────────────────────────────────────

func (r *ruleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *client.AggregatedClient, got %T", req.ProviderData))
		return
	}
	r.client = c
}

// ── Create ─────────────────────────────────────────────────────────────────────

func (r *ruleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured",
			"betterado_workitemtrackingprocess_rule Create: provider client not configured")
		return
	}

	var model ruleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	processID := model.ProcessID.ValueString()
	witRefName := model.WorkItemTypeID.ValueString()

	conditions, diags := expandRuleConditions(ctx, model.Conditions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	actions, diags := expandRuleActions(ctx, model.Actions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleRequest := &workitemtrackingprocess.CreateProcessRuleRequest{
		Name:       converter.String(model.Name.ValueString()),
		IsDisabled: converter.Bool(!model.IsEnabled.ValueBool()),
		Conditions: conditions,
		Actions:    actions,
	}

	args := workitemtrackingprocess.AddProcessWorkItemTypeRuleArgs{
		ProcessId:         converter.UUID(processID),
		WitRefName:        &witRefName,
		ProcessRuleCreate: ruleRequest,
	}

	createdRule, err := r.client.WorkItemTrackingProcessClient.AddProcessWorkItemTypeRule(ctx, args)
	if err != nil {
		resp.Diagnostics.AddError("Create error",
			fmt.Sprintf("creating rule for work item type %s: %s", witRefName, err))
		return
	}
	if createdRule == nil {
		resp.Diagnostics.AddError("Create error", "created rule is nil")
		return
	}
	if createdRule.Id == nil {
		resp.Diagnostics.AddError("Create error", "created rule has no ID")
		return
	}

	model.ID = types.StringValue(createdRule.Id.String())
	diags = r.flattenRule(ctx, &model, createdRule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ───────────────────────────────────────────────────────────────────────

func (r *ruleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured",
			"betterado_workitemtrackingprocess_rule Read: provider client not configured")
		return
	}

	var model ruleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	processID := model.ProcessID.ValueString()
	witRefName := model.WorkItemTypeID.ValueString()
	ruleID := model.ID.ValueString()

	ruleUUID, err := uuid.Parse(ruleID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid rule ID", fmt.Sprintf("Cannot parse rule ID %q as UUID: %s", ruleID, err))
		return
	}

	args := workitemtrackingprocess.GetProcessWorkItemTypeRuleArgs{
		ProcessId:  converter.UUID(processID),
		WitRefName: &witRefName,
		RuleId:     &ruleUUID,
	}

	rule, err := r.client.WorkItemTrackingProcessClient.GetProcessWorkItemTypeRule(ctx, args)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read error",
			fmt.Sprintf("reading rule %s: %s", ruleID, err))
		return
	}
	if rule == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	diags := r.flattenRule(ctx, &model, rule)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ─────────────────────────────────────────────────────────────────────

func (r *ruleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured",
			"betterado_workitemtrackingprocess_rule Update: provider client not configured")
		return
	}

	var plan, state ruleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	processID := plan.ProcessID.ValueString()
	witRefName := plan.WorkItemTypeID.ValueString()
	ruleID := state.ID.ValueString()

	ruleUUID, err := uuid.Parse(ruleID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid rule ID", fmt.Sprintf("Cannot parse rule ID %q as UUID: %s", ruleID, err))
		return
	}

	conditions, diags := expandRuleConditions(ctx, plan.Conditions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	actions, diags := expandRuleActions(ctx, plan.Actions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleUpdate := &workitemtrackingprocess.UpdateProcessRuleRequest{
		Id:         &ruleUUID,
		Name:       converter.String(plan.Name.ValueString()),
		IsDisabled: converter.Bool(!plan.IsEnabled.ValueBool()),
		Conditions: conditions,
		Actions:    actions,
	}

	args := workitemtrackingprocess.UpdateProcessWorkItemTypeRuleArgs{
		ProcessId:   converter.UUID(processID),
		WitRefName:  &witRefName,
		RuleId:      &ruleUUID,
		ProcessRule: ruleUpdate,
	}

	updatedRule, err := r.client.WorkItemTrackingProcessClient.UpdateProcessWorkItemTypeRule(ctx, args)
	if err != nil {
		resp.Diagnostics.AddError("Update error",
			fmt.Sprintf("updating rule %s: %s", ruleID, err))
		return
	}

	plan.ID = state.ID
	if updatedRule != nil {
		diags = r.flattenRule(ctx, &plan, updatedRule)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ─────────────────────────────────────────────────────────────────────

func (r *ruleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured",
			"betterado_workitemtrackingprocess_rule Delete: provider client not configured")
		return
	}

	var model ruleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	processID := model.ProcessID.ValueString()
	witRefName := model.WorkItemTypeID.ValueString()
	ruleID := model.ID.ValueString()

	ruleUUID, err := uuid.Parse(ruleID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid rule ID", fmt.Sprintf("Cannot parse rule ID %q as UUID: %s", ruleID, err))
		return
	}

	args := workitemtrackingprocess.DeleteProcessWorkItemTypeRuleArgs{
		ProcessId:  converter.UUID(processID),
		WitRefName: &witRefName,
		RuleId:     &ruleUUID,
	}

	err = r.client.WorkItemTrackingProcessClient.DeleteProcessWorkItemTypeRule(ctx, args)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Delete error",
			fmt.Sprintf("deleting rule %s: %s", ruleID, err))
	}
}

// ── ImportState ────────────────────────────────────────────────────────────────

// ImportState imports an existing rule by "process_id/work_item_type_id/rule_id".
func (r *ruleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Expected process_id/work_item_type_id/rule_id, got %q", req.ID))
		return
	}

	var model ruleResourceModel
	model.ID = types.StringValue(parts[2])
	model.ProcessID = types.StringValue(parts[0])
	model.WorkItemTypeID = types.StringValue(parts[1])

	// Set required computed-but-not-read-at-import fields to null/empty so
	// the subsequent Read populates them correctly.
	model.Name = types.StringNull()
	model.IsEnabled = types.BoolNull()
	model.URL = types.StringNull()
	model.Conditions = types.SetValueMust(
		types.ObjectType{AttrTypes: conditionAttrTypes},
		[]attr.Value{},
	)
	model.Actions = types.SetValueMust(
		types.ObjectType{AttrTypes: actionAttrTypes},
		[]attr.Value{},
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ────────────────────────────────────────────────────────────────────

func expandRuleConditions(ctx context.Context, conditionsSet types.Set) (*[]workitemtrackingprocess.RuleCondition, diag.Diagnostics) {
	var ds diag.Diagnostics
	var conditionModels []ruleConditionModel
	ds.Append(conditionsSet.ElementsAs(ctx, &conditionModels, false)...)
	if ds.HasError() {
		return nil, ds
	}

	conditions := make([]workitemtrackingprocess.RuleCondition, len(conditionModels))
	for i, cm := range conditionModels {
		c := workitemtrackingprocess.RuleCondition{}
		if !cm.ConditionType.IsNull() && !cm.ConditionType.IsUnknown() {
			ct := workitemtrackingprocess.RuleConditionType(cm.ConditionType.ValueString())
			c.ConditionType = &ct
		}
		if !cm.Field.IsNull() && !cm.Field.IsUnknown() && cm.Field.ValueString() != "" {
			f := cm.Field.ValueString()
			c.Field = &f
		}
		if !cm.Value.IsNull() && !cm.Value.IsUnknown() && cm.Value.ValueString() != "" {
			v := cm.Value.ValueString()
			c.Value = &v
		}
		conditions[i] = c
	}
	return &conditions, ds
}

func expandRuleActions(ctx context.Context, actionsSet types.Set) (*[]workitemtrackingprocess.RuleAction, diag.Diagnostics) {
	var ds diag.Diagnostics
	var actionModels []ruleActionModel
	ds.Append(actionsSet.ElementsAs(ctx, &actionModels, false)...)
	if ds.HasError() {
		return nil, ds
	}

	actions := make([]workitemtrackingprocess.RuleAction, len(actionModels))
	for i, am := range actionModels {
		a := workitemtrackingprocess.RuleAction{}
		if !am.ActionType.IsNull() && !am.ActionType.IsUnknown() {
			at := workitemtrackingprocess.RuleActionType(am.ActionType.ValueString())
			a.ActionType = &at
		}
		if !am.TargetField.IsNull() && !am.TargetField.IsUnknown() {
			tf := am.TargetField.ValueString()
			a.TargetField = &tf
		}
		if !am.Value.IsNull() && !am.Value.IsUnknown() && am.Value.ValueString() != "" {
			v := am.Value.ValueString()
			a.Value = &v
		}
		actions[i] = a
	}
	return &actions, ds
}

func (r *ruleResource) flattenRule(ctx context.Context, model *ruleResourceModel, rule *workitemtrackingprocess.ProcessRule) diag.Diagnostics {
	var diags diag.Diagnostics

	if rule.Name != nil {
		model.Name = types.StringValue(*rule.Name)
	}
	if rule.IsDisabled != nil {
		model.IsEnabled = types.BoolValue(!*rule.IsDisabled)
	}
	if rule.Url != nil {
		model.URL = types.StringValue(*rule.Url)
	}

	// Flatten conditions
	if rule.Conditions != nil {
		condObjects := make([]attr.Value, len(*rule.Conditions))
		for i, c := range *rule.Conditions {
			condMap := map[string]attr.Value{
				"condition_type": types.StringNull(),
				"field":          types.StringValue(""),
				"value":          types.StringValue(""),
			}
			if c.ConditionType != nil {
				condMap["condition_type"] = types.StringValue(string(*c.ConditionType))
			}
			if c.Field != nil {
				condMap["field"] = types.StringValue(*c.Field)
			}
			if c.Value != nil {
				condMap["value"] = types.StringValue(*c.Value)
			}
			obj, d := types.ObjectValue(conditionAttrTypes, condMap)
			diags.Append(d...)
			condObjects[i] = obj
		}
		condSet, d := types.SetValue(types.ObjectType{AttrTypes: conditionAttrTypes}, condObjects)
		diags.Append(d...)
		model.Conditions = condSet
	}

	// Flatten actions
	if rule.Actions != nil {
		actObjects := make([]attr.Value, len(*rule.Actions))
		for i, a := range *rule.Actions {
			actMap := map[string]attr.Value{
				"action_type":  types.StringNull(),
				"target_field": types.StringValue(""),
				"value":        types.StringValue(""),
			}
			if a.ActionType != nil {
				actMap["action_type"] = types.StringValue(string(*a.ActionType))
			}
			if a.TargetField != nil {
				actMap["target_field"] = types.StringValue(*a.TargetField)
			}
			if a.Value != nil {
				actMap["value"] = types.StringValue(*a.Value)
			}
			obj, d := types.ObjectValue(actionAttrTypes, actMap)
			diags.Append(d...)
			actObjects[i] = obj
		}
		actSet, d := types.SetValue(types.ObjectType{AttrTypes: actionAttrTypes}, actObjects)
		diags.Append(d...)
		model.Actions = actSet
	}

	return diags
}
