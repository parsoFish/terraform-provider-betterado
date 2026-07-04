package workitemtrackingprocess

// resource_inherited_control_framework.go — terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_inherited_control.

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &inheritedControlResource{}
	_ resource.ResourceWithConfigure   = &inheritedControlResource{}
	_ resource.ResourceWithImportState = &inheritedControlResource{}
)

// ── Inline plan modifiers ─────────────────────────────────────────────────────

type inheritedControlUseStateForUnknownString struct{}

func (inheritedControlUseStateForUnknownString) Description(_ context.Context) string {
	return "use prior state"
}

func (inheritedControlUseStateForUnknownString) MarkdownDescription(_ context.Context) string {
	return "use prior state"
}

func (inheritedControlUseStateForUnknownString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type inheritedControlRequiresReplaceString struct{}

func (inheritedControlRequiresReplaceString) Description(_ context.Context) string {
	return "requires replacement if changed"
}

func (inheritedControlRequiresReplaceString) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}

func (inheritedControlRequiresReplaceString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ── resource struct ────────────────────────────────────────────────────────────

type inheritedControlResource struct {
	client *client.AggregatedClient
}

// NewInheritedControlResource returns a new resource.Resource for betterado_workitemtrackingprocess_inherited_control.
func NewInheritedControlResource() resource.Resource {
	return &inheritedControlResource{}
}

// ── Model ──────────────────────────────────────────────────────────────────────

type inheritedControlResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProcessID      types.String `tfsdk:"process_id"`
	WorkItemTypeID types.String `tfsdk:"work_item_type_id"`
	GroupID        types.String `tfsdk:"group_id"`
	ControlID      types.String `tfsdk:"control_id"`
	Label          types.String `tfsdk:"label"`
	Visible        types.Bool   `tfsdk:"visible"`
}

// ── Metadata ───────────────────────────────────────────────────────────────────

func (r *inheritedControlResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_inherited_control"
}

// ── Schema ─────────────────────────────────────────────────────────────────────

func (r *inheritedControlResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the customisation of an inherited control in an Azure DevOps process work item type layout.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the control (same as control_id).",
				PlanModifiers: []planmodifier.String{
					inheritedControlUseStateForUnknownString{},
				},
			},
			"process_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the process.",
				PlanModifiers: []planmodifier.String{
					inheritedControlRequiresReplaceString{},
				},
			},
			"work_item_type_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID (reference name) of the work item type.",
				PlanModifiers: []planmodifier.String{
					inheritedControlRequiresReplaceString{},
				},
			},
			"group_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the group containing the control.",
				PlanModifiers: []planmodifier.String{
					inheritedControlRequiresReplaceString{},
				},
			},
			"control_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the inherited control to customize.",
				PlanModifiers: []planmodifier.String{
					inheritedControlRequiresReplaceString{},
				},
			},
			"label": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Label for the control.",
				PlanModifiers: []planmodifier.String{
					inheritedControlUseStateForUnknownString{},
				},
			},
			"visible": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the control should be visible.",
			},
		},
	}
}

// ── Configure ──────────────────────────────────────────────────────────────────

func (r *inheritedControlResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *inheritedControlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_inherited_control Create: provider client not configured")
		return
	}

	var model inheritedControlResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	controlID := model.ControlID.ValueString()
	processID := model.ProcessID.ValueString()
	witRefName := model.WorkItemTypeID.ValueString()
	groupID := model.GroupID.ValueString()

	// Verify control is inherited
	workItemType, err := r.client.WorkItemTrackingProcessClient.GetProcessWorkItemType(ctx, workitemtrackingprocess.GetProcessWorkItemTypeArgs{
		ProcessId:  converter.UUID(processID),
		WitRefName: &witRefName,
		Expand:     &workitemtrackingprocess.GetWorkItemTypeExpandValues.Layout,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("getting work item type: %s", err))
		return
	}

	group := findGroupById(workItemType.Layout, groupID)
	if group == nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("group %s not found in layout", groupID))
		return
	}

	existingControl := findControlInGroup(group, controlID)
	if existingControl == nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("control %s not found in group %s", controlID, groupID))
		return
	}

	if existingControl.Inherited == nil || !*existingControl.Inherited {
		resp.Diagnostics.AddError("Create error",
			fmt.Sprintf("control %s is not inherited, use betterado_workitemtrackingprocess_control to manage custom controls", controlID))
		return
	}

	model.ID = types.StringValue(controlID)

	// Apply customisation
	if err := r.applyCustomisation(ctx, model); err != nil {
		resp.Diagnostics.AddError("Create error (update)", err.Error())
		return
	}

	if err := r.readIntoModel(ctx, &model); err != nil {
		resp.Diagnostics.AddError("Create error (read-back)", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ───────────────────────────────────────────────────────────────────────

func (r *inheritedControlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_inherited_control Read: provider client not configured")
		return
	}

	var model inheritedControlResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readIntoModel(ctx, &model); err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ─────────────────────────────────────────────────────────────────────

func (r *inheritedControlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_inherited_control Update: provider client not configured")
		return
	}

	var plan, currentState inheritedControlResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &currentState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = currentState.ID

	if err := r.applyCustomisation(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Update error", err.Error())
		return
	}

	if err := r.readIntoModel(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Update error (read-back)", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ─────────────────────────────────────────────────────────────────────

// Delete reverts the inherited control customisation by removing it from the group.
func (r *inheritedControlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_inherited_control Delete: provider client not configured")
		return
	}

	var model inheritedControlResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	controlID := model.ID.ValueString()
	processID := converter.UUID(model.ProcessID.ValueString())
	witRefName := model.WorkItemTypeID.ValueString()
	groupID := model.GroupID.ValueString()

	err := r.client.WorkItemTrackingProcessClient.RemoveControlFromGroup(ctx, workitemtrackingprocess.RemoveControlFromGroupArgs{
		ProcessId:  processID,
		WitRefName: &witRefName,
		GroupId:    &groupID,
		ControlId:  &controlID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("reverting inherited control: %s", err))
	}
}

// ── ImportState ────────────────────────────────────────────────────────────────

func (r *inheritedControlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 4)
	if len(parts) != 4 {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Expected process_id/work_item_type_id/group_id/control_id, got %q", req.ID))
		return
	}
	if _, err := uuid.Parse(parts[0]); err != nil {
		resp.Diagnostics.AddError("Invalid process_id", fmt.Sprintf("process_id must be a UUID: %s", err))
		return
	}

	model := inheritedControlResourceModel{
		ProcessID:      types.StringValue(parts[0]),
		WorkItemTypeID: types.StringValue(parts[1]),
		GroupID:        types.StringValue(parts[2]),
		ControlID:      types.StringValue(parts[3]),
		ID:             types.StringValue(parts[3]),
	}

	if err := r.readIntoModel(ctx, &model); err != nil {
		resp.Diagnostics.AddError("Import error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ────────────────────────────────────────────────────────────────────

func (r *inheritedControlResource) readIntoModel(ctx context.Context, model *inheritedControlResourceModel) error {
	controlID := model.ID.ValueString()
	processID := model.ProcessID.ValueString()
	witRefName := model.WorkItemTypeID.ValueString()
	groupID := model.GroupID.ValueString()

	workItemType, err := r.client.WorkItemTrackingProcessClient.GetProcessWorkItemType(ctx, workitemtrackingprocess.GetProcessWorkItemTypeArgs{
		ProcessId:  converter.UUID(processID),
		WitRefName: &witRefName,
		Expand:     &workitemtrackingprocess.GetWorkItemTypeExpandValues.Layout,
	})
	if err != nil {
		return fmt.Errorf("getting work item type: %w", err)
	}

	group := findGroupById(workItemType.Layout, groupID)
	if group == nil {
		return fmt.Errorf("not found")
	}

	control := findControlInGroup(group, controlID)
	if control == nil {
		return fmt.Errorf("not found")
	}

	if control.Label != nil {
		model.Label = types.StringValue(*control.Label)
	}
	if control.Visible != nil {
		model.Visible = types.BoolValue(*control.Visible)
	}

	return nil
}

func (r *inheritedControlResource) applyCustomisation(ctx context.Context, model inheritedControlResourceModel) error {
	controlID := model.ID.ValueString()
	processID := converter.UUID(model.ProcessID.ValueString())
	witRefName := model.WorkItemTypeID.ValueString()
	groupID := model.GroupID.ValueString()

	control := workitemtrackingprocess.Control{}

	if !model.Visible.IsNull() && !model.Visible.IsUnknown() {
		control.Visible = converter.Bool(model.Visible.ValueBool())
	}
	if !model.Label.IsNull() && !model.Label.IsUnknown() {
		control.Label = converter.String(model.Label.ValueString())
	}

	_, err := r.client.WorkItemTrackingProcessClient.UpdateControl(ctx, workitemtrackingprocess.UpdateControlArgs{
		ProcessId:  processID,
		WitRefName: &witRefName,
		GroupId:    &groupID,
		ControlId:  &controlID,
		Control:    &control,
	})
	return err
}
