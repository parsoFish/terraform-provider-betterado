package workitemtrackingprocess

// resource_state_framework.go — terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_state.

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
	_ resource.Resource                = &stateResource{}
	_ resource.ResourceWithConfigure   = &stateResource{}
	_ resource.ResourceWithImportState = &stateResource{}
)

// ── inline plan modifiers ─────────────────────────────────────────────────────

type stateUseStateForUnknownString struct{}

func (stateUseStateForUnknownString) Description(_ context.Context) string        { return "use prior state" }
func (stateUseStateForUnknownString) MarkdownDescription(_ context.Context) string { return "use prior state" }
func (stateUseStateForUnknownString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type stateUseStateForUnknownInt64 struct{}

func (stateUseStateForUnknownInt64) Description(_ context.Context) string        { return "use prior state" }
func (stateUseStateForUnknownInt64) MarkdownDescription(_ context.Context) string { return "use prior state" }
func (stateUseStateForUnknownInt64) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type stateRequiresReplaceString struct{}

func (stateRequiresReplaceString) Description(_ context.Context) string        { return "requires replacement if changed" }
func (stateRequiresReplaceString) MarkdownDescription(_ context.Context) string { return "requires replacement if changed" }
func (stateRequiresReplaceString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ── validators ─────────────────────────────────────────────────────────────────

type stateColorValidator struct{}

func (v stateColorValidator) Description(_ context.Context) string { return "Must be a hexadecimal color code, e.g. #b2b2b2" }
func (v stateColorValidator) MarkdownDescription(_ context.Context) string { return "Must be a hexadecimal color code, e.g. `#b2b2b2`" }
func (v stateColorValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if !regexp.MustCompile(`^#[0-9a-fA-F]{6}$`).MatchString(req.ConfigValue.ValueString()) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid color", "Must be a hexadecimal color code, i.e. #b2b2b2")
	}
}

type stateCategoryValidator struct{}

func (v stateCategoryValidator) Description(_ context.Context) string { return "Valid values: Proposed, InProgress, Resolved, Completed, Removed" }
func (v stateCategoryValidator) MarkdownDescription(_ context.Context) string { return "Valid values: Proposed, InProgress, Resolved, Completed, Removed" }
func (v stateCategoryValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	valid := map[string]bool{"Proposed": true, "InProgress": true, "Resolved": true, "Completed": true, "Removed": true}
	if !valid[req.ConfigValue.ValueString()] {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid state_category", "Valid values: Proposed, InProgress, Resolved, Completed, Removed")
	}
}

// ── resource struct ────────────────────────────────────────────────────────────

type stateResource struct {
	client *client.AggregatedClient
}

// NewStateResource returns a new resource.Resource for betterado_workitemtrackingprocess_state.
func NewStateResource() resource.Resource {
	return &stateResource{}
}

// ── Model ──────────────────────────────────────────────────────────────────────

type stateResourceModel struct {
	ID            types.String `tfsdk:"id"`
	ProcessID     types.String `tfsdk:"process_id"`
	WorkItemType  types.String `tfsdk:"work_item_type_id"`
	Name          types.String `tfsdk:"name"`
	Color         types.String `tfsdk:"color"`
	StateCategory types.String `tfsdk:"state_category"`
	Order         types.Int64  `tfsdk:"order"`
	URL           types.String `tfsdk:"url"`
}

// ── Metadata ───────────────────────────────────────────────────────────────────

func (r *stateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_state"
}

// ── Schema ─────────────────────────────────────────────────────────────────────

func (r *stateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a custom work item state in an Azure DevOps process.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the state.",
				PlanModifiers: []planmodifier.String{
					stateUseStateForUnknownString{},
				},
			},
			"process_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the process.",
				PlanModifiers: []planmodifier.String{
					stateRequiresReplaceString{},
				},
			},
			"work_item_type_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID (reference name) of the work item type.",
				PlanModifiers: []planmodifier.String{
					stateRequiresReplaceString{},
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the state.",
				PlanModifiers: []planmodifier.String{
					stateRequiresReplaceString{},
				},
			},
			"color": schema.StringAttribute{
				Required:    true,
				Description: "Color hexadecimal code to represent the state, e.g. #b2b2b2.",
				Validators: []validator.String{
					stateColorValidator{},
				},
			},
			"state_category": schema.StringAttribute{
				Required:    true,
				Description: "Category of the state. Valid values: Proposed, InProgress, Resolved, Completed, Removed.",
				Validators: []validator.String{
					stateCategoryValidator{},
				},
			},
			"order": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Order within the category where the state should appear.",
				PlanModifiers: []planmodifier.Int64{
					stateUseStateForUnknownInt64{},
				},
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "URL of the state.",
				PlanModifiers: []planmodifier.String{
					stateUseStateForUnknownString{},
				},
			},
		},
	}
}

// ── Configure ──────────────────────────────────────────────────────────────────

func (r *stateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *stateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_state Create: provider client not configured")
		return
	}

	var model stateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateModel := &workitemtrackingprocess.WorkItemStateInputModel{
		Name:          converter.String(model.Name.ValueString()),
		Color:         converter.String(stateColorToAPI(model.Color.ValueString())),
		StateCategory: converter.String(model.StateCategory.ValueString()),
	}
	if !model.Order.IsNull() && !model.Order.IsUnknown() {
		order := int(model.Order.ValueInt64())
		stateModel.Order = &order
	}

	created, err := r.client.WorkItemTrackingProcessClient.CreateStateDefinition(ctx, workitemtrackingprocess.CreateStateDefinitionArgs{
		ProcessId:  converter.UUID(model.ProcessID.ValueString()),
		WitRefName: converter.String(model.WorkItemType.ValueString()),
		StateModel: stateModel,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("creating state: %s", err))
		return
	}
	if created == nil {
		resp.Diagnostics.AddError("Create error", "created state is nil")
		return
	}
	if created.Id == nil {
		resp.Diagnostics.AddError("Create error", "created state has no ID")
		return
	}

	r.flattenState(&model, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ───────────────────────────────────────────────────────────────────────

func (r *stateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_state Read: provider client not configured")
		return
	}

	var model stateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state, err := r.client.WorkItemTrackingProcessClient.GetStateDefinition(ctx, workitemtrackingprocess.GetStateDefinitionArgs{
		ProcessId:  converter.UUID(model.ProcessID.ValueString()),
		WitRefName: converter.String(model.WorkItemType.ValueString()),
		StateId:    converter.UUID(model.ID.ValueString()),
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("getting state %s: %s", model.ID.ValueString(), err))
		return
	}
	if state == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.flattenState(&model, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ─────────────────────────────────────────────────────────────────────

func (r *stateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_state Update: provider client not configured")
		return
	}

	var plan, stateModel stateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	inputModel := &workitemtrackingprocess.WorkItemStateInputModel{
		Color:         converter.String(stateColorToAPI(plan.Color.ValueString())),
		StateCategory: converter.String(plan.StateCategory.ValueString()),
	}
	if !plan.Order.IsNull() && !plan.Order.IsUnknown() {
		order := int(plan.Order.ValueInt64())
		inputModel.Order = &order
	}

	updated, err := r.client.WorkItemTrackingProcessClient.UpdateStateDefinition(ctx, workitemtrackingprocess.UpdateStateDefinitionArgs{
		ProcessId:  converter.UUID(plan.ProcessID.ValueString()),
		WitRefName: converter.String(plan.WorkItemType.ValueString()),
		StateId:    converter.UUID(stateModel.ID.ValueString()),
		StateModel: inputModel,
	})
	if err != nil {
		resp.Diagnostics.AddError("Update error", fmt.Sprintf("updating state: %s", err))
		return
	}

	plan.ID = stateModel.ID
	if updated != nil {
		r.flattenState(&plan, updated)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ─────────────────────────────────────────────────────────────────────

func (r *stateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_state Delete: provider client not configured")
		return
	}

	var model stateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.WorkItemTrackingProcessClient.DeleteStateDefinition(ctx, workitemtrackingprocess.DeleteStateDefinitionArgs{
		ProcessId:  converter.UUID(model.ProcessID.ValueString()),
		WitRefName: converter.String(model.WorkItemType.ValueString()),
		StateId:    converter.UUID(model.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("deleting state: %s", err))
	}
}

// ── ImportState ────────────────────────────────────────────────────────────────

// ImportState imports an existing state by "process_id/work_item_type_id/state_id".
func (r *stateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected process_id/work_item_type_id/state_id, got %q", req.ID))
		return
	}
	if _, err := uuid.Parse(parts[0]); err != nil {
		resp.Diagnostics.AddError("Invalid process_id", fmt.Sprintf("process_id must be a UUID: %s", err))
		return
	}
	if _, err := uuid.Parse(parts[2]); err != nil {
		resp.Diagnostics.AddError("Invalid state_id", fmt.Sprintf("state_id must be a UUID: %s", err))
		return
	}

	var model stateResourceModel
	model.ProcessID = types.StringValue(parts[0])
	model.WorkItemType = types.StringValue(parts[1])
	model.ID = types.StringValue(parts[2])

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ────────────────────────────────────────────────────────────────────

// stateColorToAPI strips the leading '#' for the API.
func stateColorToAPI(color string) string {
	return strings.TrimPrefix(color, "#")
}

// stateColorFromAPI adds the leading '#' from the API response.
func stateColorFromAPI(color string) string {
	if color == "" {
		return ""
	}
	if strings.HasPrefix(color, "#") {
		return color
	}
	return "#" + color
}

func (r *stateResource) flattenState(model *stateResourceModel, s *workitemtrackingprocess.WorkItemStateResultModel) {
	if s.Id != nil {
		model.ID = types.StringValue(s.Id.String())
	}
	if s.Name != nil {
		model.Name = types.StringValue(*s.Name)
	}
	if s.Color != nil {
		model.Color = types.StringValue(stateColorFromAPI(*s.Color))
	}
	if s.StateCategory != nil {
		model.StateCategory = types.StringValue(*s.StateCategory)
	}
	if s.Order != nil {
		model.Order = types.Int64Value(int64(*s.Order))
	}
	if s.Url != nil {
		model.URL = types.StringValue(*s.Url)
	}
}
