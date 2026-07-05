package workitemtrackingprocess

// resource_system_control_framework.go — terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_system_control.

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &systemControlResource{}
	_ resource.ResourceWithConfigure   = &systemControlResource{}
	_ resource.ResourceWithImportState = &systemControlResource{}
)

// ── Inline plan modifiers ─────────────────────────────────────────────────────

type systemControlUseStateForUnknownString struct{}

func (systemControlUseStateForUnknownString) Description(_ context.Context) string {
	return "use prior state"
}

func (systemControlUseStateForUnknownString) MarkdownDescription(_ context.Context) string {
	return "use prior state"
}

func (systemControlUseStateForUnknownString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type systemControlRequiresReplaceString struct{}

func (systemControlRequiresReplaceString) Description(_ context.Context) string {
	return "requires replacement if changed"
}

func (systemControlRequiresReplaceString) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}

func (systemControlRequiresReplaceString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ── inline bool default ────────────────────────────────────────────────────────

type systemControlStaticBoolDefault struct{ value bool }

func (d systemControlStaticBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", d.value)
}

func (d systemControlStaticBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%v`", d.value)
}

func (d systemControlStaticBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.value)
}

// ── resource struct ────────────────────────────────────────────────────────────

type systemControlResource struct {
	client *client.AggregatedClient
}

// NewSystemControlResource returns a new resource.Resource for betterado_workitemtrackingprocess_system_control.
func NewSystemControlResource() resource.Resource {
	return &systemControlResource{}
}

// ── Model ──────────────────────────────────────────────────────────────────────

type systemControlResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProcessID      types.String `tfsdk:"process_id"`
	WorkItemTypeID types.String `tfsdk:"work_item_type_id"`
	ControlID      types.String `tfsdk:"control_id"`
	Label          types.String `tfsdk:"label"`
	Visible        types.Bool   `tfsdk:"visible"`
	ControlType    types.String `tfsdk:"control_type"`
	ReadOnly       types.Bool   `tfsdk:"read_only"`
}

// ── Metadata ───────────────────────────────────────────────────────────────────

func (r *systemControlResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_system_control"
}

// ── Schema ─────────────────────────────────────────────────────────────────────

func (r *systemControlResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the customisation of a system control in an Azure DevOps process work item type.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the control (same as control_id).",
				PlanModifiers: []planmodifier.String{
					systemControlUseStateForUnknownString{},
				},
			},
			"process_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the process.",
				PlanModifiers: []planmodifier.String{
					systemControlRequiresReplaceString{},
				},
			},
			"work_item_type_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID (reference name) of the work item type.",
				PlanModifiers: []planmodifier.String{
					systemControlRequiresReplaceString{},
				},
			},
			"control_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the system control (e.g., System.AreaPath, System.IterationPath).",
				PlanModifiers: []planmodifier.String{
					systemControlRequiresReplaceString{},
				},
			},
			"label": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Label for the control.",
				PlanModifiers: []planmodifier.String{
					systemControlUseStateForUnknownString{},
				},
			},
			"visible": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     systemControlStaticBoolDefault{value: true},
				Description: "Whether the control should be visible.",
			},
			"control_type": schema.StringAttribute{
				Computed:    true,
				Description: "Type of the control.",
				PlanModifiers: []planmodifier.String{
					systemControlUseStateForUnknownString{},
				},
			},
			"read_only": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the control is read-only.",
			},
		},
	}
}

// ── Configure ──────────────────────────────────────────────────────────────────

func (r *systemControlResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *systemControlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_system_control Create: provider client not configured")
		return
	}

	var model systemControlResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	controlID := model.ControlID.ValueString()

	control := workitemtrackingprocess.Control{
		Visible: converter.Bool(model.Visible.ValueBool()),
	}
	if !model.Label.IsNull() && !model.Label.IsUnknown() {
		control.Label = converter.String(model.Label.ValueString())
	}

	updatedControl, err := r.client.WorkItemTrackingProcessClient.UpdateSystemControl(ctx, workitemtrackingprocess.UpdateSystemControlArgs{
		ProcessId:  converter.UUID(model.ProcessID.ValueString()),
		WitRefName: converter.String(model.WorkItemTypeID.ValueString()),
		ControlId:  &controlID,
		Control:    &control,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("creating system control customization: %s", err))
		return
	}
	if updatedControl == nil {
		resp.Diagnostics.AddError("Create error", "updated system control is nil")
		return
	}

	model.ID = types.StringValue(controlID)

	if err := r.readIntoModel(ctx, &model); err != nil {
		resp.Diagnostics.AddError("Create error (read-back)", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ───────────────────────────────────────────────────────────────────────

func (r *systemControlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_system_control Read: provider client not configured")
		return
	}

	var model systemControlResourceModel
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

func (r *systemControlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_system_control Update: provider client not configured")
		return
	}

	var plan, currentState systemControlResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &currentState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = currentState.ID
	controlID := plan.ID.ValueString()

	control := workitemtrackingprocess.Control{
		Visible: converter.Bool(plan.Visible.ValueBool()),
	}
	if !plan.Label.IsNull() && !plan.Label.IsUnknown() {
		control.Label = converter.String(plan.Label.ValueString())
	}

	_, err := r.client.WorkItemTrackingProcessClient.UpdateSystemControl(ctx, workitemtrackingprocess.UpdateSystemControlArgs{
		ProcessId:  converter.UUID(plan.ProcessID.ValueString()),
		WitRefName: converter.String(plan.WorkItemTypeID.ValueString()),
		ControlId:  &controlID,
		Control:    &control,
	})
	if err != nil {
		resp.Diagnostics.AddError("Update error", fmt.Sprintf("updating system control: %s", err))
		return
	}

	if err := r.readIntoModel(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Update error (read-back)", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ─────────────────────────────────────────────────────────────────────

func (r *systemControlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_system_control Delete: provider client not configured")
		return
	}

	var model systemControlResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	controlID := model.ID.ValueString()

	_, err := r.client.WorkItemTrackingProcessClient.DeleteSystemControl(ctx, workitemtrackingprocess.DeleteSystemControlArgs{
		ProcessId:  converter.UUID(model.ProcessID.ValueString()),
		WitRefName: converter.String(model.WorkItemTypeID.ValueString()),
		ControlId:  &controlID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("deleting system control customization: %s", err))
	}
}

// ── ImportState ────────────────────────────────────────────────────────────────

func (r *systemControlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Expected process_id/work_item_type_id/control_id, got %q", req.ID))
		return
	}
	if _, err := uuid.Parse(parts[0]); err != nil {
		resp.Diagnostics.AddError("Invalid process_id", fmt.Sprintf("process_id must be a UUID: %s", err))
		return
	}

	model := systemControlResourceModel{
		ProcessID:      types.StringValue(parts[0]),
		WorkItemTypeID: types.StringValue(parts[1]),
		ControlID:      types.StringValue(parts[2]),
		ID:             types.StringValue(parts[2]),
	}

	if err := r.readIntoModel(ctx, &model); err != nil {
		resp.Diagnostics.AddError("Import error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ────────────────────────────────────────────────────────────────────

func (r *systemControlResource) readIntoModel(ctx context.Context, model *systemControlResourceModel) error {
	controlID := model.ID.ValueString()
	processID := converter.UUID(model.ProcessID.ValueString())
	witRefName := model.WorkItemTypeID.ValueString()

	controls, err := r.client.WorkItemTrackingProcessClient.GetSystemControls(ctx, workitemtrackingprocess.GetSystemControlsArgs{
		ProcessId:  processID,
		WitRefName: &witRefName,
	})
	if err != nil {
		return fmt.Errorf("getting system controls: %w", err)
	}

	var foundControl *workitemtrackingprocess.Control
	if controls != nil {
		for _, c := range *controls {
			if c.Id != nil && *c.Id == controlID {
				cc := c
				foundControl = &cc
				break
			}
		}
	}

	if foundControl == nil {
		// Not in the edited list — reverted to default
		return fmt.Errorf("not found")
	}

	if foundControl.Label != nil {
		model.Label = types.StringValue(*foundControl.Label)
	}
	if foundControl.Visible != nil {
		model.Visible = types.BoolValue(*foundControl.Visible)
	}
	if foundControl.ControlType != nil {
		model.ControlType = types.StringValue(*foundControl.ControlType)
	}
	if foundControl.ReadOnly != nil {
		model.ReadOnly = types.BoolValue(*foundControl.ReadOnly)
	}

	return nil
}
