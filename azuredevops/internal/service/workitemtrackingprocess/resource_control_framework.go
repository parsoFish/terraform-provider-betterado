package workitemtrackingprocess

// resource_control_framework.go — terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_control.

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &controlResource{}
	_ resource.ResourceWithConfigure   = &controlResource{}
	_ resource.ResourceWithImportState = &controlResource{}
)

// ── Inline plan modifiers ─────────────────────────────────────────────────────

type controlUseStateForUnknownString struct{}

func (controlUseStateForUnknownString) Description(_ context.Context) string {
	return "use prior state"
}

func (controlUseStateForUnknownString) MarkdownDescription(_ context.Context) string {
	return "use prior state"
}

func (controlUseStateForUnknownString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type controlRequiresReplaceString struct{}

func (controlRequiresReplaceString) Description(_ context.Context) string {
	return "requires replacement if changed"
}

func (controlRequiresReplaceString) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}

func (controlRequiresReplaceString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ── validators ─────────────────────────────────────────────────────────────────

type controlNonWhiteSpaceValidator struct{}

func (v controlNonWhiteSpaceValidator) Description(_ context.Context) string {
	return "value must not be empty or whitespace-only"
}

func (v controlNonWhiteSpaceValidator) MarkdownDescription(_ context.Context) string {
	return "value must not be empty or whitespace-only"
}

func (v controlNonWhiteSpaceValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if strings.TrimSpace(req.ConfigValue.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid value", "Value must not be empty or whitespace-only")
	}
}

// ── inline bool default ────────────────────────────────────────────────────────

type controlStaticBoolDefault struct{ value bool }

func (d controlStaticBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", d.value)
}

func (d controlStaticBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%v`", d.value)
}

func (d controlStaticBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.value)
}

// ── resource struct ────────────────────────────────────────────────────────────

type controlResource struct {
	client *client.AggregatedClient
}

// NewControlResource returns a new resource.Resource for betterado_workitemtrackingprocess_control.
func NewControlResource() resource.Resource {
	return &controlResource{}
}

// ── Models ─────────────────────────────────────────────────────────────────────

type controlContributionModel struct {
	ContributionID        types.String `tfsdk:"contribution_id"`
	Height                types.Int64  `tfsdk:"height"`
	Inputs                types.Map    `tfsdk:"inputs"`
	ShowOnDeletedWorkItem types.Bool   `tfsdk:"show_on_deleted_work_item"`
}

type controlResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProcessID      types.String `tfsdk:"process_id"`
	WorkItemTypeID types.String `tfsdk:"work_item_type_reference_name"`
	GroupID        types.String `tfsdk:"group_id"`
	ControlID      types.String `tfsdk:"control_id"`
	Label          types.String `tfsdk:"label"`
	Order          types.Int64  `tfsdk:"order"`
	Visible        types.Bool   `tfsdk:"visible"`
	ReadOnly       types.Bool   `tfsdk:"read_only"`
	Metadata       types.String `tfsdk:"metadata"`
	Watermark      types.String `tfsdk:"watermark"`
	ControlType    types.String `tfsdk:"control_type"`
	Inherited      types.Bool   `tfsdk:"inherited"`
	Overridden     types.Bool   `tfsdk:"overridden"`
	IsContribution types.Bool   `tfsdk:"is_contribution"`
	Contribution   types.List   `tfsdk:"contribution"`
}

// ── attr types ─────────────────────────────────────────────────────────────────

var controlContributionAttrTypes = map[string]attr.Type{
	"contribution_id":           types.StringType,
	"height":                    types.Int64Type,
	"inputs":                    types.MapType{ElemType: types.StringType},
	"show_on_deleted_work_item": types.BoolType,
}

var controlContributionObjectType = types.ObjectType{AttrTypes: controlContributionAttrTypes}

// ── Metadata ───────────────────────────────────────────────────────────────────

func (r *controlResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_control"
}

// ── Schema ─────────────────────────────────────────────────────────────────────

func (r *controlResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a control in an Azure DevOps process work item type layout group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the control.",
				PlanModifiers: []planmodifier.String{
					controlUseStateForUnknownString{},
				},
			},
			"process_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the process.",
				PlanModifiers: []planmodifier.String{
					controlRequiresReplaceString{},
				},
			},
			"work_item_type_reference_name": schema.StringAttribute{
				Required:    true,
				Description: "The reference name of the work item type.",
				PlanModifiers: []planmodifier.String{
					controlRequiresReplaceString{},
				},
				Validators: []validator.String{controlNonWhiteSpaceValidator{}},
			},
			"group_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the group to add the control to.",
				Validators:  []validator.String{controlNonWhiteSpaceValidator{}},
			},
			"control_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID for the control. For field controls, this is the field reference name.",
				PlanModifiers: []planmodifier.String{
					controlRequiresReplaceString{},
				},
				Validators: []validator.String{controlNonWhiteSpaceValidator{}},
			},
			"label": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Label for the field.",
				PlanModifiers: []planmodifier.String{
					controlUseStateForUnknownString{},
				},
			},
			"order": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Order in which the control should appear in its group.",
			},
			"visible": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     controlStaticBoolDefault{value: true},
				Description: "Whether the control is visible.",
			},
			"read_only": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     controlStaticBoolDefault{value: false},
				Description: "Whether the control is read-only.",
			},
			"metadata": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Inner text of the control.",
				PlanModifiers: []planmodifier.String{
					controlUseStateForUnknownString{},
				},
			},
			"watermark": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Watermark text for the textbox.",
				PlanModifiers: []planmodifier.String{
					controlUseStateForUnknownString{},
				},
			},
			"control_type": schema.StringAttribute{
				Computed:    true,
				Description: "Type of the control.",
				PlanModifiers: []planmodifier.String{
					controlUseStateForUnknownString{},
				},
			},
			"inherited": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this layout node has been inherited from a parent layout.",
			},
			"overridden": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this layout node has been overridden by a child layout.",
			},
			"is_contribution": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     controlStaticBoolDefault{value: false},
				Description: "Whether the layout node is contribution or not.",
			},
			"contribution": schema.ListNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Contribution for the control.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"contribution_id": schema.StringAttribute{
							Required:    true,
							Description: "The id for the contribution.",
						},
						"height": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The height for the contribution.",
						},
						"inputs": schema.MapAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Key value pairs for contribution inputs.",
						},
						"show_on_deleted_work_item": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     controlStaticBoolDefault{value: false},
							Description: "Whether the contribution shows on deleted work items.",
						},
					},
				},
			},
		},
	}
}

// ── Configure ──────────────────────────────────────────────────────────────────

func (r *controlResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *client.AggregatedClient, got %T", req.ProviderData))
		return
	}
	r.client = c
}

// ── Create ─────────────────────────────────────────────────────────────────────

func (r *controlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_control Create: provider client not configured")
		return
	}

	var model controlResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	control := workitemtrackingprocess.Control{
		Id:       converter.String(model.ControlID.ValueString()),
		Visible:  converter.Bool(model.Visible.ValueBool()),
		ReadOnly: converter.Bool(model.ReadOnly.ValueBool()),
	}

	if !model.Label.IsNull() && !model.Label.IsUnknown() {
		control.Label = converter.String(model.Label.ValueString())
	}
	if !model.Order.IsNull() && !model.Order.IsUnknown() {
		order := int(model.Order.ValueInt64())
		control.Order = &order
	}
	if !model.Metadata.IsNull() && !model.Metadata.IsUnknown() {
		control.Metadata = converter.String(model.Metadata.ValueString())
	}
	if !model.Watermark.IsNull() && !model.Watermark.IsUnknown() {
		control.Watermark = converter.String(model.Watermark.ValueString())
	}
	control.IsContribution = converter.Bool(model.IsContribution.ValueBool())
	if !model.Contribution.IsNull() && !model.Contribution.IsUnknown() {
		var contribModels []controlContributionModel
		resp.Diagnostics.Append(model.Contribution.ElementsAs(ctx, &contribModels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(contribModels) > 0 {
			control.Contribution = expandControlContribution(contribModels[0])
		}
	}

	processID := converter.UUID(model.ProcessID.ValueString())
	witRefName := model.WorkItemTypeID.ValueString()
	groupID := model.GroupID.ValueString()

	var createdControl *workitemtrackingprocess.Control
	err := utils.RetryOnContributionNotFound(ctx, 10*time.Minute, func() error {
		var createErr error
		createdControl, createErr = r.client.WorkItemTrackingProcessClient.CreateControlInGroup(ctx, workitemtrackingprocess.CreateControlInGroupArgs{
			ProcessId:  processID,
			WitRefName: &witRefName,
			GroupId:    &groupID,
			Control:    &control,
		})
		return createErr
	})
	if err != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("creating control: %s", err))
		return
	}
	if createdControl == nil || createdControl.Id == nil {
		resp.Diagnostics.AddError("Create error", "created control has no ID")
		return
	}

	model.ID = types.StringValue(*createdControl.Id)

	// Read back for eventual consistency
	readErr := utils.RetryOnNotFound(ctx, 10*time.Minute, func() error {
		return r.readControlIntoModel(ctx, &model)
	})
	if readErr != nil {
		resp.Diagnostics.AddError("Create error (read-back)", readErr.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ───────────────────────────────────────────────────────────────────────

func (r *controlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_control Read: provider client not configured")
		return
	}

	var model controlResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readControlIntoModel(ctx, &model); err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ─────────────────────────────────────────────────────────────────────

func (r *controlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_control Update: provider client not configured")
		return
	}

	var plan, currentState controlResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &currentState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = currentState.ID
	controlID := plan.ID.ValueString()
	processID := converter.UUID(plan.ProcessID.ValueString())
	witRefName := plan.WorkItemTypeID.ValueString()
	groupID := plan.GroupID.ValueString()

	control := &workitemtrackingprocess.Control{
		Visible:  converter.Bool(plan.Visible.ValueBool()),
		ReadOnly: converter.Bool(plan.ReadOnly.ValueBool()),
	}
	if !plan.Label.IsNull() && !plan.Label.IsUnknown() {
		control.Label = converter.String(plan.Label.ValueString())
	}
	if !plan.Order.IsNull() && !plan.Order.IsUnknown() {
		order := int(plan.Order.ValueInt64())
		control.Order = &order
	}
	if !plan.Metadata.IsNull() && !plan.Metadata.IsUnknown() {
		control.Metadata = converter.String(plan.Metadata.ValueString())
	}
	if !plan.Watermark.IsNull() && !plan.Watermark.IsUnknown() {
		control.Watermark = converter.String(plan.Watermark.ValueString())
	}
	control.IsContribution = converter.Bool(plan.IsContribution.ValueBool())
	if !plan.Contribution.IsNull() && !plan.Contribution.IsUnknown() {
		var contribModels []controlContributionModel
		resp.Diagnostics.Append(plan.Contribution.ElementsAs(ctx, &contribModels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(contribModels) > 0 {
			control.Contribution = expandControlContribution(contribModels[0])
		}
	}

	// If group_id changed, move the control
	oldGroupID := currentState.GroupID.ValueString()
	if groupID != oldGroupID {
		_, err := r.client.WorkItemTrackingProcessClient.MoveControlToGroup(ctx, workitemtrackingprocess.MoveControlToGroupArgs{
			ProcessId:         processID,
			WitRefName:        &witRefName,
			GroupId:           &groupID,
			ControlId:         &controlID,
			Control:           control,
			RemoveFromGroupId: &oldGroupID,
		})
		if err != nil {
			resp.Diagnostics.AddError("Update error", fmt.Sprintf("moving control: %s", err))
			return
		}
	} else {
		_, err := r.client.WorkItemTrackingProcessClient.UpdateControl(ctx, workitemtrackingprocess.UpdateControlArgs{
			ProcessId:  processID,
			WitRefName: &witRefName,
			GroupId:    &groupID,
			ControlId:  &controlID,
			Control:    control,
		})
		if err != nil {
			resp.Diagnostics.AddError("Update error", fmt.Sprintf("updating control: %s", err))
			return
		}
	}

	if err := r.readControlIntoModel(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Update error (read-back)", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ─────────────────────────────────────────────────────────────────────

func (r *controlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_control Delete: provider client not configured")
		return
	}

	var model controlResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	controlID := model.ID.ValueString()
	processID := converter.UUID(model.ProcessID.ValueString())
	witRefName := model.WorkItemTypeID.ValueString()
	groupID := model.GroupID.ValueString()

	err := utils.RetryOnUnexpectedException(ctx, 10*time.Minute, func() error {
		return r.client.WorkItemTrackingProcessClient.RemoveControlFromGroup(ctx, workitemtrackingprocess.RemoveControlFromGroupArgs{
			ProcessId:  processID,
			WitRefName: &witRefName,
			GroupId:    &groupID,
			ControlId:  &controlID,
		})
	})
	if err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("deleting control: %s", err))
	}
}

// ── ImportState ────────────────────────────────────────────────────────────────

func (r *controlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 4)
	if len(parts) != 4 {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Expected process_id/work_item_type_reference_name/group_id/control_id, got %q", req.ID))
		return
	}
	if _, err := uuid.Parse(parts[0]); err != nil {
		resp.Diagnostics.AddError("Invalid process_id", fmt.Sprintf("process_id must be a UUID: %s", err))
		return
	}

	model := controlResourceModel{
		ProcessID:      types.StringValue(parts[0]),
		WorkItemTypeID: types.StringValue(parts[1]),
		GroupID:        types.StringValue(parts[2]),
		ControlID:      types.StringValue(parts[3]),
		ID:             types.StringValue(parts[3]),
	}

	if err := r.readControlIntoModel(ctx, &model); err != nil {
		resp.Diagnostics.AddError("Import error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ────────────────────────────────────────────────────────────────────

func (r *controlResource) readControlIntoModel(ctx context.Context, model *controlResourceModel) error {
	controlID := model.ID.ValueString()
	processID := converter.UUID(model.ProcessID.ValueString())
	witRefName := model.WorkItemTypeID.ValueString()
	groupID := model.GroupID.ValueString()

	workItemType, err := r.client.WorkItemTrackingProcessClient.GetProcessWorkItemType(ctx, workitemtrackingprocess.GetProcessWorkItemTypeArgs{
		ProcessId:  processID,
		WitRefName: &witRefName,
		Expand:     &workitemtrackingprocess.GetWorkItemTypeExpandValues.Layout,
	})
	if err != nil {
		return err
	}

	group := findGroupById(workItemType.Layout, groupID)
	if group == nil {
		return azuredevops.WrappedError{
			StatusCode: converter.Int(404),
			Message:    converter.String(fmt.Sprintf("group %s not found in layout", groupID)),
		}
	}

	foundControl := findControlInGroup(group, controlID)
	if foundControl == nil {
		return azuredevops.WrappedError{
			StatusCode: converter.Int(404),
			Message:    converter.String(fmt.Sprintf("control %s not found in group %s", controlID, groupID)),
		}
	}

	model.GroupID = types.StringValue(groupID)
	if foundControl.Label != nil {
		model.Label = types.StringValue(*foundControl.Label)
	}
	if foundControl.Order != nil {
		model.Order = types.Int64Value(int64(*foundControl.Order))
	}
	if foundControl.Visible != nil {
		model.Visible = types.BoolValue(*foundControl.Visible)
	}
	if foundControl.ReadOnly != nil {
		model.ReadOnly = types.BoolValue(*foundControl.ReadOnly)
	}
	if foundControl.Metadata != nil {
		model.Metadata = types.StringValue(*foundControl.Metadata)
	}
	if foundControl.Watermark != nil {
		model.Watermark = types.StringValue(*foundControl.Watermark)
	}
	if foundControl.ControlType != nil {
		model.ControlType = types.StringValue(*foundControl.ControlType)
	}
	if foundControl.Inherited != nil {
		model.Inherited = types.BoolValue(*foundControl.Inherited)
	}
	if foundControl.Overridden != nil {
		model.Overridden = types.BoolValue(*foundControl.Overridden)
	}
	if foundControl.IsContribution != nil {
		model.IsContribution = types.BoolValue(*foundControl.IsContribution)
	}

	// Flatten contribution
	if foundControl.Contribution != nil {
		contrib := foundControl.Contribution
		inputsMap := types.MapValueMust(types.StringType, map[string]attr.Value{})
		if contrib.Inputs != nil {
			inputAttrs := map[string]attr.Value{}
			for k, v := range *contrib.Inputs {
				if sv, ok := v.(string); ok {
					inputAttrs[k] = types.StringValue(sv)
				}
			}
			m, d := types.MapValue(types.StringType, inputAttrs)
			if !d.HasError() {
				inputsMap = m
			}
		}
		contribAttrs := map[string]attr.Value{
			"contribution_id":           types.StringValue(converter.ToString(contrib.ContributionId, "")),
			"height":                    types.Int64Value(int64(converter.ToInt(contrib.Height, 0))),
			"inputs":                    inputsMap,
			"show_on_deleted_work_item": types.BoolValue(converter.ToBool(contrib.ShowOnDeletedWorkItem, false)),
		}
		contribObj, diags := types.ObjectValue(controlContributionAttrTypes, contribAttrs)
		if diags.HasError() {
			return fmt.Errorf("building contribution object")
		}
		model.Contribution = types.ListValueMust(controlContributionObjectType, []attr.Value{contribObj})
	} else {
		model.Contribution = types.ListValueMust(controlContributionObjectType, []attr.Value{})
	}

	return nil
}

func expandControlContribution(m controlContributionModel) *workitemtrackingprocess.WitContribution {
	contrib := &workitemtrackingprocess.WitContribution{}

	if !m.ContributionID.IsNull() && !m.ContributionID.IsUnknown() {
		contrib.ContributionId = converter.String(m.ContributionID.ValueString())
	}
	if !m.Height.IsNull() && !m.Height.IsUnknown() && m.Height.ValueInt64() != 0 {
		h := int(m.Height.ValueInt64())
		contrib.Height = &h
	}
	if !m.ShowOnDeletedWorkItem.IsNull() && !m.ShowOnDeletedWorkItem.IsUnknown() {
		contrib.ShowOnDeletedWorkItem = converter.Bool(m.ShowOnDeletedWorkItem.ValueBool())
	}
	if !m.Inputs.IsNull() && !m.Inputs.IsUnknown() {
		inputMap := map[string]interface{}{}
		for k, v := range m.Inputs.Elements() {
			if sv, ok := v.(types.String); ok {
				inputMap[k] = sv.ValueString()
			}
		}
		if len(inputMap) > 0 {
			contrib.Inputs = &inputMap
		}
	}

	return contrib
}
