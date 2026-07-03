package workitemtrackingprocess

// resource_page_framework.go — terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_page.

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &pageResource{}
	_ resource.ResourceWithConfigure   = &pageResource{}
	_ resource.ResourceWithImportState = &pageResource{}
)

// ── Inline plan modifiers and defaults ────────────────────────────────────────
// (stringplanmodifier / booldefault sub-packages are not vendored in this project.)

type pageUseStateForUnknownString struct{}

func (pageUseStateForUnknownString) Description(_ context.Context) string { return "use prior state" }
func (pageUseStateForUnknownString) MarkdownDescription(_ context.Context) string {
	return "use prior state"
}
func (pageUseStateForUnknownString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type pageRequiresReplaceString struct{}

func (pageRequiresReplaceString) Description(_ context.Context) string {
	return "requires replacement if changed"
}
func (pageRequiresReplaceString) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}
func (pageRequiresReplaceString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type pageStaticBoolDefault struct{ value bool }

func (d pageStaticBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", d.value)
}
func (d pageStaticBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%v`", d.value)
}
func (d pageStaticBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.value)
}

// ── resource struct ────────────────────────────────────────────────────────────

type pageResource struct {
	client *client.AggregatedClient
}

// NewPageResource returns a new resource.Resource for betterado_workitemtrackingprocess_page.
func NewPageResource() resource.Resource {
	return &pageResource{}
}

// ── Model ──────────────────────────────────────────────────────────────────────

type pageResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProcessID      types.String `tfsdk:"process_id"`
	WorkItemTypeID types.String `tfsdk:"work_item_type_id"`
	Label          types.String `tfsdk:"label"`
	Order          types.Int64  `tfsdk:"order"`
	Visible        types.Bool   `tfsdk:"visible"`
	Sections       types.List   `tfsdk:"sections"`
}

// sectionObjectType is the object type for the sections list elements.
var sectionObjectType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"id": types.StringType,
	},
}

// ── Metadata ───────────────────────────────────────────────────────────────────

func (r *pageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_page"
}

// ── Schema ─────────────────────────────────────────────────────────────────────

func (r *pageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a page in the layout of a work item type in an Azure DevOps process.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the page.",
				PlanModifiers: []planmodifier.String{
					pageUseStateForUnknownString{},
				},
			},
			"process_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the process.",
				PlanModifiers: []planmodifier.String{
					pageRequiresReplaceString{},
				},
			},
			"work_item_type_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID (reference name) of the work item type.",
				PlanModifiers: []planmodifier.String{
					pageRequiresReplaceString{},
				},
			},
			"label": schema.StringAttribute{
				Required:    true,
				Description: "The label for the page.",
			},
			"order": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Order in which the page should appear in the layout.",
			},
			"visible": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     pageStaticBoolDefault{value: true},
				Description: "A value indicating if the page should be visible or not.",
			},
			"sections": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The sections of the page.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the section.",
						},
					},
				},
			},
		},
	}
}

// ── Configure ──────────────────────────────────────────────────────────────────

func (r *pageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *pageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_page Create: provider client not configured")
		return
	}

	var model pageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	page := workitemtrackingprocess.Page{
		Label:    converter.String(model.Label.ValueString()),
		Visible:  converter.Bool(model.Visible.ValueBool()),
		PageType: &workitemtrackingprocess.PageTypeValues.Custom,
	}

	if !model.Order.IsNull() && !model.Order.IsUnknown() {
		order := int(model.Order.ValueInt64())
		page.Order = &order
	}

	processID := converter.UUID(model.ProcessID.ValueString())
	witRefName := model.WorkItemTypeID.ValueString()

	createdPage, err := r.client.WorkItemTrackingProcessClient.AddPage(ctx, workitemtrackingprocess.AddPageArgs{
		ProcessId:  processID,
		WitRefName: &witRefName,
		Page:       &page,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("creating page: %s", err))
		return
	}
	if createdPage.Id == nil {
		resp.Diagnostics.AddError("Create error", "created page has no ID")
		return
	}

	model.ID = types.StringValue(*createdPage.Id)

	// Read back for authoritative state (sections etc.).
	if err := r.readPageIntoModel(ctx, &model); err != nil {
		resp.Diagnostics.AddError("Create error (read-back)", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ───────────────────────────────────────────────────────────────────────

func (r *pageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_page Read: provider client not configured")
		return
	}

	var model pageResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readPageIntoModel(ctx, &model); err != nil {
		// If the resource is gone treat it as destroyed.
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ─────────────────────────────────────────────────────────────────────

func (r *pageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_page Update: provider client not configured")
		return
	}

	var plan, currentState pageResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &currentState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = currentState.ID
	pageID := plan.ID.ValueString()

	updatePage := &workitemtrackingprocess.Page{
		Id:      &pageID,
		Label:   converter.String(plan.Label.ValueString()),
		Visible: converter.Bool(plan.Visible.ValueBool()),
	}

	if !plan.Order.IsNull() && !plan.Order.IsUnknown() {
		order := int(plan.Order.ValueInt64())
		updatePage.Order = &order
	}

	processID := converter.UUID(plan.ProcessID.ValueString())
	witRefName := plan.WorkItemTypeID.ValueString()

	_, err := r.client.WorkItemTrackingProcessClient.UpdatePage(ctx, workitemtrackingprocess.UpdatePageArgs{
		ProcessId:  processID,
		WitRefName: &witRefName,
		Page:       updatePage,
	})
	if err != nil {
		resp.Diagnostics.AddError("Update error", fmt.Sprintf("updating page: %s", err))
		return
	}

	if err := r.readPageIntoModel(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Update error (read-back)", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ─────────────────────────────────────────────────────────────────────

func (r *pageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_page Delete: provider client not configured")
		return
	}

	var model pageResourceModel
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
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("deleting page: %s", err))
	}
}

// ── ImportState ────────────────────────────────────────────────────────────────

// ImportState imports by "process_id/work_item_type_id/page_id".
func (r *pageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected process_id/work_item_type_id/page_id, got %q", req.ID))
		return
	}
	if _, err := uuid.Parse(parts[0]); err != nil {
		resp.Diagnostics.AddError("Invalid process_id", fmt.Sprintf("process_id must be a UUID: %s", err))
		return
	}

	model := pageResourceModel{
		ProcessID:      types.StringValue(parts[0]),
		WorkItemTypeID: types.StringValue(parts[1]),
		ID:             types.StringValue(parts[2]),
	}

	if err := r.readPageIntoModel(ctx, &model); err != nil {
		resp.Diagnostics.AddError("Import error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ────────────────────────────────────────────────────────────────────

// readPageIntoModel fetches the work item type layout and updates the model with the page data.
// Returns an error if the page is not found (caller should treat this as destroyed).
func (r *pageResource) readPageIntoModel(ctx context.Context, model *pageResourceModel) error {
	processID := converter.UUID(model.ProcessID.ValueString())
	witRefName := model.WorkItemTypeID.ValueString()
	pageID := model.ID.ValueString()

	workItemType, err := r.client.WorkItemTrackingProcessClient.GetProcessWorkItemType(ctx, workitemtrackingprocess.GetProcessWorkItemTypeArgs{
		ProcessId:  processID,
		WitRefName: &witRefName,
		Expand:     &workitemtrackingprocess.GetWorkItemTypeExpandValues.Layout,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			return fmt.Errorf("not found")
		}
		return fmt.Errorf("getting work item type %q: %w", witRefName, err)
	}

	page := findPageByIDFramework(workItemType.Layout, pageID)
	if page == nil {
		return fmt.Errorf("not found")
	}

	if page.Label != nil {
		model.Label = types.StringValue(*page.Label)
	}
	if page.Order != nil {
		model.Order = types.Int64Value(int64(*page.Order))
	}
	if page.Visible != nil {
		model.Visible = types.BoolValue(*page.Visible)
	} else {
		model.Visible = types.BoolValue(true)
	}

	// Flatten sections.
	if page.Sections != nil {
		sectionElems := make([]attr.Value, len(*page.Sections))
		for i, s := range *page.Sections {
			sectionID := ""
			if s.Id != nil {
				sectionID = *s.Id
			}
			obj, diags := types.ObjectValue(
				sectionObjectType.AttrTypes,
				map[string]attr.Value{
					"id": types.StringValue(sectionID),
				},
			)
			if diags.HasError() {
				return fmt.Errorf("building sections object: %s", diags)
			}
			sectionElems[i] = obj
		}
		listVal, diags := types.ListValue(sectionObjectType, sectionElems)
		if diags.HasError() {
			return fmt.Errorf("building sections list: %s", diags)
		}
		model.Sections = listVal
	} else {
		model.Sections = types.ListValueMust(sectionObjectType, []attr.Value{})
	}

	return nil
}

// findPageByIDFramework walks the layout pages and returns the one matching pageID.
func findPageByIDFramework(layout *workitemtrackingprocess.FormLayout, pageID string) *workitemtrackingprocess.Page {
	if layout == nil || layout.Pages == nil {
		return nil
	}
	for _, page := range *layout.Pages {
		if page.Id != nil && *page.Id == pageID {
			p := page
			return &p
		}
	}
	return nil
}
