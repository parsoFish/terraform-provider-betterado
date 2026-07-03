package workitemtrackingprocess

// resource_inherited_page_framework.go — terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_inherited_page.
//
// An "inherited page" is a page that exists on an inherited work item type. It
// cannot be created or destroyed — it can only be customised (label changed) and
// reverted.  "Create" means: record the page_id and apply the label customisation.
// "Delete" means: call RemovePage which reverts the page back to its inherited
// defaults (the page still exists in ADO; only the customisation is gone).

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
	_ resource.Resource                = &inheritedPageResource{}
	_ resource.ResourceWithConfigure   = &inheritedPageResource{}
	_ resource.ResourceWithImportState = &inheritedPageResource{}
)

// ── Inline plan modifiers ─────────────────────────────────────────────────────
// (stringplanmodifier sub-package is not vendored in this project.)

type inheritedPageUseStateForUnknownString struct{}

func (inheritedPageUseStateForUnknownString) Description(_ context.Context) string {
	return "use prior state"
}
func (inheritedPageUseStateForUnknownString) MarkdownDescription(_ context.Context) string {
	return "use prior state"
}
func (inheritedPageUseStateForUnknownString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type inheritedPageRequiresReplaceString struct{}

func (inheritedPageRequiresReplaceString) Description(_ context.Context) string {
	return "requires replacement if changed"
}
func (inheritedPageRequiresReplaceString) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}
func (inheritedPageRequiresReplaceString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ── resource struct ────────────────────────────────────────────────────────────

type inheritedPageResource struct {
	client *client.AggregatedClient
}

// NewInheritedPageResource returns a new resource.Resource for betterado_workitemtrackingprocess_inherited_page.
func NewInheritedPageResource() resource.Resource {
	return &inheritedPageResource{}
}

// ── Model ──────────────────────────────────────────────────────────────────────

type inheritedPageResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProcessID      types.String `tfsdk:"process_id"`
	WorkItemTypeID types.String `tfsdk:"work_item_type_id"`
	PageID         types.String `tfsdk:"page_id"`
	Label          types.String `tfsdk:"label"`
}

// ── Metadata ───────────────────────────────────────────────────────────────────

func (r *inheritedPageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_inherited_page"
}

// ── Schema ─────────────────────────────────────────────────────────────────────

func (r *inheritedPageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the customisation of an inherited page in an Azure DevOps process work item type layout.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the page (same as page_id).",
				PlanModifiers: []planmodifier.String{
					inheritedPageUseStateForUnknownString{},
				},
			},
			"process_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the process.",
				PlanModifiers: []planmodifier.String{
					inheritedPageRequiresReplaceString{},
				},
			},
			"work_item_type_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID (reference name) of the work item type.",
				PlanModifiers: []planmodifier.String{
					inheritedPageRequiresReplaceString{},
				},
			},
			"page_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the inherited page to customise.",
				PlanModifiers: []planmodifier.String{
					inheritedPageRequiresReplaceString{},
				},
			},
			"label": schema.StringAttribute{
				Required:    true,
				Description: "Label for the page.",
			},
		},
	}
}

// ── Configure ──────────────────────────────────────────────────────────────────

func (r *inheritedPageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create sets the label on an existing inherited page and records it in state.
func (r *inheritedPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_inherited_page Create: provider client not configured")
		return
	}

	var model inheritedPageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pageID := model.PageID.ValueString()

	// Verify the page exists and is inherited before applying customisation.
	page, err := r.getPage(ctx, model.ProcessID.ValueString(), model.WorkItemTypeID.ValueString(), pageID)
	if err != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("finding page %q: %s", pageID, err))
		return
	}
	if page == nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("page %q not found in work item type layout", pageID))
		return
	}
	if page.Inherited == nil || !*page.Inherited {
		resp.Diagnostics.AddError("Create error",
			fmt.Sprintf("page %q is not inherited; use betterado_workitemtrackingprocess_page to manage custom pages", pageID))
		return
	}

	// The ID is the page_id.
	model.ID = types.StringValue(pageID)

	// Apply the label customisation.
	if err := r.applyLabel(ctx, model); err != nil {
		resp.Diagnostics.AddError("Create error (update label)", err.Error())
		return
	}

	// Read back for authoritative state.
	if err := r.readIntoModel(ctx, &model); err != nil {
		resp.Diagnostics.AddError("Create error (read-back)", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ───────────────────────────────────────────────────────────────────────

func (r *inheritedPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_inherited_page Read: provider client not configured")
		return
	}

	var model inheritedPageResourceModel
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

func (r *inheritedPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_inherited_page Update: provider client not configured")
		return
	}

	var plan, currentState inheritedPageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &currentState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = currentState.ID

	if err := r.applyLabel(ctx, plan); err != nil {
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

// Delete reverts the inherited page customisation back to default by calling
// RemovePage. The page still exists in ADO; only the custom label is removed.
func (r *inheritedPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_inherited_page Delete: provider client not configured")
		return
	}

	var model inheritedPageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pageID := model.ID.ValueString()
	processID := converter.UUID(model.ProcessID.ValueString())
	witRefName := model.WorkItemTypeID.ValueString()

	err := r.client.WorkItemTrackingProcessClient.RemovePage(ctx, workitemtrackingprocess.RemovePageArgs{
		ProcessId:  processID,
		WitRefName: &witRefName,
		PageId:     &pageID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("reverting inherited page: %s", err))
	}
}

// ── ImportState ────────────────────────────────────────────────────────────────

// ImportState imports by "process_id/work_item_type_id/page_id".
func (r *inheritedPageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected process_id/work_item_type_id/page_id, got %q", req.ID))
		return
	}
	if _, err := uuid.Parse(parts[0]); err != nil {
		resp.Diagnostics.AddError("Invalid process_id", fmt.Sprintf("process_id must be a UUID: %s", err))
		return
	}

	model := inheritedPageResourceModel{
		ProcessID:      types.StringValue(parts[0]),
		WorkItemTypeID: types.StringValue(parts[1]),
		PageID:         types.StringValue(parts[2]),
		ID:             types.StringValue(parts[2]),
	}

	if err := r.readIntoModel(ctx, &model); err != nil {
		resp.Diagnostics.AddError("Import error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ────────────────────────────────────────────────────────────────────

// getPage fetches the work item type layout and returns the page with the given ID (or nil).
func (r *inheritedPageResource) getPage(ctx context.Context, processID, witRefName, pageID string) (*workitemtrackingprocess.Page, error) {
	workItemType, err := r.client.WorkItemTrackingProcessClient.GetProcessWorkItemType(ctx, workitemtrackingprocess.GetProcessWorkItemTypeArgs{
		ProcessId:  converter.UUID(processID),
		WitRefName: &witRefName,
		Expand:     &workitemtrackingprocess.GetWorkItemTypeExpandValues.Layout,
	})
	if err != nil {
		return nil, err
	}
	return findPageByIDFramework(workItemType.Layout, pageID), nil
}

// readIntoModel fetches the page and writes its fields into model.
func (r *inheritedPageResource) readIntoModel(ctx context.Context, model *inheritedPageResourceModel) error {
	pageID := model.ID.ValueString()

	page, err := r.getPage(ctx, model.ProcessID.ValueString(), model.WorkItemTypeID.ValueString(), pageID)
	if err != nil {
		return fmt.Errorf("getting page %q: %w", pageID, err)
	}
	if page == nil {
		return fmt.Errorf("not found")
	}

	if page.Label != nil {
		model.Label = types.StringValue(*page.Label)
	}
	// Ensure page_id matches id in state.
	model.PageID = types.StringValue(pageID)

	return nil
}

// applyLabel sends an UpdatePage call to apply the desired label on the inherited page.
func (r *inheritedPageResource) applyLabel(ctx context.Context, model inheritedPageResourceModel) error {
	pageID := model.ID.ValueString()
	processID := converter.UUID(model.ProcessID.ValueString())
	witRefName := model.WorkItemTypeID.ValueString()

	_, err := r.client.WorkItemTrackingProcessClient.UpdatePage(ctx, workitemtrackingprocess.UpdatePageArgs{
		ProcessId:  processID,
		WitRefName: &witRefName,
		Page: &workitemtrackingprocess.Page{
			Id:    &pageID,
			Label: converter.String(model.Label.ValueString()),
		},
	})
	return err
}
