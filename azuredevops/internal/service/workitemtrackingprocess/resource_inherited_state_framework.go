package workitemtrackingprocess

// resource_inherited_state_framework.go — terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_inherited_state.

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
	_ resource.Resource                = &inheritedStateResource{}
	_ resource.ResourceWithConfigure   = &inheritedStateResource{}
	_ resource.ResourceWithImportState = &inheritedStateResource{}
)

// ── inline plan modifiers ─────────────────────────────────────────────────────

type inheritedStateUseStateForUnknownString struct{}

func (inheritedStateUseStateForUnknownString) Description(_ context.Context) string {
	return "use prior state"
}

func (inheritedStateUseStateForUnknownString) MarkdownDescription(_ context.Context) string {
	return "use prior state"
}

func (inheritedStateUseStateForUnknownString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type inheritedStateUseStateForUnknownBool struct{}

func (inheritedStateUseStateForUnknownBool) Description(_ context.Context) string {
	return "use prior state"
}

func (inheritedStateUseStateForUnknownBool) MarkdownDescription(_ context.Context) string {
	return "use prior state"
}

func (inheritedStateUseStateForUnknownBool) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type inheritedStateRequiresReplaceString struct{}

func (inheritedStateRequiresReplaceString) Description(_ context.Context) string {
	return "requires replacement if changed"
}

func (inheritedStateRequiresReplaceString) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}

func (inheritedStateRequiresReplaceString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ── resource struct ────────────────────────────────────────────────────────────

type inheritedStateResource struct {
	client *client.AggregatedClient
}

// NewInheritedStateResource returns a new resource.Resource for betterado_workitemtrackingprocess_inherited_state.
func NewInheritedStateResource() resource.Resource {
	return &inheritedStateResource{}
}

// ── Model ──────────────────────────────────────────────────────────────────────

type inheritedStateResourceModel struct {
	ID           types.String `tfsdk:"id"`
	ProcessID    types.String `tfsdk:"process_id"`
	WorkItemType types.String `tfsdk:"work_item_type_id"`
	Name         types.String `tfsdk:"name"`
	Visible      types.Bool   `tfsdk:"visible"`
}

// ── Metadata ───────────────────────────────────────────────────────────────────

func (r *inheritedStateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_inherited_state"
}

// ── Schema ─────────────────────────────────────────────────────────────────────

func (r *inheritedStateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the visibility of an inherited work item state in an Azure DevOps process.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the state.",
				PlanModifiers: []planmodifier.String{
					inheritedStateUseStateForUnknownString{},
				},
			},
			"process_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the process.",
				PlanModifiers: []planmodifier.String{
					inheritedStateRequiresReplaceString{},
				},
			},
			"work_item_type_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID (reference name) of the work item type.",
				PlanModifiers: []planmodifier.String{
					inheritedStateRequiresReplaceString{},
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the inherited state to manage.",
				PlanModifiers: []planmodifier.String{
					inheritedStateRequiresReplaceString{},
				},
			},
			"visible": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the state should be visible.",
				PlanModifiers: []planmodifier.Bool{
					inheritedStateUseStateForUnknownBool{},
				},
			},
		},
	}
}

// ── Configure ──────────────────────────────────────────────────────────────────

func (r *inheritedStateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create for inherited_state means: find the state by name, then set its visibility.
// Inherited states already exist — they cannot be created; they can only be toggled.
func (r *inheritedStateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_inherited_state Create: provider client not configured")
		return
	}

	var model inheritedStateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, err := r.findInheritedStateByName(ctx, model.ProcessID.ValueString(), model.WorkItemType.ValueString(), model.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("finding inherited state %q: %s", model.Name.ValueString(), err))
		return
	}
	if state.Id == nil {
		resp.Diagnostics.AddError("Create error", "inherited state ID is nil")
		return
	}

	model.ID = types.StringValue(state.Id.String())

	// Apply visibility setting.
	if !model.Visible.IsNull() && !model.Visible.IsUnknown() {
		if err := r.setVisibility(ctx, model); err != nil {
			resp.Diagnostics.AddError("Create error (visibility)", fmt.Sprintf("setting visibility: %s", err))
			return
		}
	}

	// Read back to get authoritative state.
	updated, err := r.readStateByID(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("Create error (read-back)", fmt.Sprintf("reading back state: %s", err))
		return
	}
	r.flattenInheritedState(&model, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ───────────────────────────────────────────────────────────────────────

func (r *inheritedStateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_inherited_state Read: provider client not configured")
		return
	}

	var model inheritedStateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, err := r.readStateByID(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("getting inherited state: %s", err))
		return
	}
	if state == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.flattenInheritedState(&model, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ─────────────────────────────────────────────────────────────────────

func (r *inheritedStateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_inherited_state Update: provider client not configured")
		return
	}

	var plan, currentState inheritedStateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &currentState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = currentState.ID

	if err := r.setVisibility(ctx, plan); err != nil {
		resp.Diagnostics.AddError("Update error", fmt.Sprintf("setting visibility: %s", err))
		return
	}

	updated, err := r.readStateByID(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("Update error (read-back)", fmt.Sprintf("reading back state: %s", err))
		return
	}
	r.flattenInheritedState(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ─────────────────────────────────────────────────────────────────────

// Delete for inherited_state means: restore the state to its default visibility (visible).
// The resource is "gone" from Terraform but the state itself persists in Azure DevOps.
func (r *inheritedStateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_inherited_state Delete: provider client not configured")
		return
	}

	var model inheritedStateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// "Delete" means unhide (restore to visible) by calling DeleteStateDefinition
	// which removes the hidden override but leaves the inherited state in place.
	witRefName := model.WorkItemType.ValueString()
	err := r.client.WorkItemTrackingProcessClient.DeleteStateDefinition(ctx, workitemtrackingprocess.DeleteStateDefinitionArgs{
		ProcessId:  converter.UUID(model.ProcessID.ValueString()),
		WitRefName: &witRefName,
		StateId:    converter.UUID(model.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("deleting (unhiding) inherited state: %s", err))
	}
}

// ── ImportState ────────────────────────────────────────────────────────────────

// ImportState imports by "process_id/work_item_type_id/state_name".
func (r *inheritedStateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected process_id/work_item_type_id/state_name, got %q", req.ID))
		return
	}
	if _, err := uuid.Parse(parts[0]); err != nil {
		resp.Diagnostics.AddError("Invalid process_id", fmt.Sprintf("process_id must be a UUID: %s", err))
		return
	}

	var model inheritedStateResourceModel
	model.ProcessID = types.StringValue(parts[0])
	model.WorkItemType = types.StringValue(parts[1])
	model.Name = types.StringValue(parts[2])

	// Look up the state by name to get the ID.
	state, err := r.findInheritedStateByName(ctx, parts[0], parts[1], parts[2])
	if err != nil {
		resp.Diagnostics.AddError("Import error", fmt.Sprintf("finding inherited state %q: %s", parts[2], err))
		return
	}
	if state.Id == nil {
		resp.Diagnostics.AddError("Import error", "state ID is nil")
		return
	}

	model.ID = types.StringValue(state.Id.String())
	r.flattenInheritedState(&model, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ────────────────────────────────────────────────────────────────────

// findInheritedStateByName lists all states and returns the one matching name, if it's inherited/system.
func (r *inheritedStateResource) findInheritedStateByName(ctx context.Context, processID, witRefName, name string) (*workitemtrackingprocess.WorkItemStateResultModel, error) {
	states, err := r.client.WorkItemTrackingProcessClient.GetStateDefinitions(ctx, workitemtrackingprocess.GetStateDefinitionsArgs{
		ProcessId:  converter.UUID(processID),
		WitRefName: &witRefName,
	})
	if err != nil {
		return nil, err
	}

	if states != nil {
		for _, s := range *states {
			if s.Name != nil && *s.Name == name {
				if s.CustomizationType == nil {
					return nil, fmt.Errorf("state %q has no customization type", name)
				}
				if *s.CustomizationType == workitemtrackingprocess.CustomizationTypeValues.Custom {
					return nil, fmt.Errorf("state %q is a custom state; use betterado_workitemtrackingprocess_state for custom states", name)
				}
				return &s, nil
			}
		}
	}

	return nil, fmt.Errorf("inherited state %q not found", name)
}

// readStateByID reads a state definition by its UUID.
func (r *inheritedStateResource) readStateByID(ctx context.Context, model inheritedStateResourceModel) (*workitemtrackingprocess.WorkItemStateResultModel, error) {
	witRefName := model.WorkItemType.ValueString()
	stateID := converter.UUID(model.ID.ValueString())

	return r.client.WorkItemTrackingProcessClient.GetStateDefinition(ctx, workitemtrackingprocess.GetStateDefinitionArgs{
		ProcessId:  converter.UUID(model.ProcessID.ValueString()),
		WitRefName: &witRefName,
		StateId:    stateID,
	})
}

// setVisibility hides or unhides the inherited state based on the model's Visible flag.
func (r *inheritedStateResource) setVisibility(ctx context.Context, model inheritedStateResourceModel) error {
	if model.Visible.IsNull() || model.Visible.IsUnknown() {
		return nil
	}

	witRefName := model.WorkItemType.ValueString()
	stateID := converter.UUID(model.ID.ValueString())
	processID := converter.UUID(model.ProcessID.ValueString())

	if !model.Visible.ValueBool() {
		// Hide the state.
		hidden := true
		_, err := r.client.WorkItemTrackingProcessClient.HideStateDefinition(ctx, workitemtrackingprocess.HideStateDefinitionArgs{
			ProcessId:      processID,
			WitRefName:     &witRefName,
			StateId:        stateID,
			HideStateModel: &workitemtrackingprocess.HideStateModel{Hidden: &hidden},
		})
		return err
	}

	// Unhide: DELETE the hidden override to restore visibility.
	return r.client.WorkItemTrackingProcessClient.DeleteStateDefinition(ctx, workitemtrackingprocess.DeleteStateDefinitionArgs{
		ProcessId:  processID,
		WitRefName: &witRefName,
		StateId:    stateID,
	})
}

// flattenInheritedState writes API response fields into the model.
func (r *inheritedStateResource) flattenInheritedState(model *inheritedStateResourceModel, s *workitemtrackingprocess.WorkItemStateResultModel) {
	if s.Id != nil {
		model.ID = types.StringValue(s.Id.String())
	}
	if s.Name != nil {
		model.Name = types.StringValue(*s.Name)
	}

	// hidden=nil or hidden=false → visible=true; hidden=true → visible=false.
	if s.Hidden != nil {
		model.Visible = types.BoolValue(!*s.Hidden)
	} else {
		/*
		   Since visible/hidden is never explicitly sent to the downstream API
		   we must assume that when hidden is not set the resource is visible
		   or else there might be a diff if visibility was configured.
		*/
		model.Visible = types.BoolValue(true)
	}
}
