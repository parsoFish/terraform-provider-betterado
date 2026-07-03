package workitemtrackingprocess

// resource_work_item_type_framework.go — terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_workitemtype.

import (
	"context"
	"fmt"
	"regexp"
	"strings"

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
	_ resource.Resource                = &workItemTypeResource{}
	_ resource.ResourceWithConfigure   = &workItemTypeResource{}
	_ resource.ResourceWithImportState = &workItemTypeResource{}
)

// ── Inline defaults ───────────────────────────────────────────────────────────

type witStaticStringDefault struct{ value string }

func witStaticStringDef(v string) defaults.String { return witStaticStringDefault{value: v} }

func (s witStaticStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", s.value)
}
func (s witStaticStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%s`", s.value)
}
func (s witStaticStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(s.value)
}

type witStaticBoolDefault struct{ value bool }

func witStaticBoolDef(v bool) defaults.Bool { return witStaticBoolDefault{value: v} }

func (s witStaticBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", s.value)
}
func (s witStaticBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", s.value)
}
func (s witStaticBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(s.value)
}

// ── Inline plan modifiers ─────────────────────────────────────────────────────

type witUseStateForUnknown struct{}

func witUseStateForUnknownMod() planmodifier.String { return witUseStateForUnknown{} }

func (witUseStateForUnknown) Description(_ context.Context) string         { return "use prior state" }
func (witUseStateForUnknown) MarkdownDescription(_ context.Context) string { return "use prior state" }
func (witUseStateForUnknown) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type witRequiresReplace struct{}

func witRequiresReplaceMod() planmodifier.String { return witRequiresReplace{} }

func (witRequiresReplace) Description(_ context.Context) string         { return "requires replace" }
func (witRequiresReplace) MarkdownDescription(_ context.Context) string { return "requires replace" }
func (witRequiresReplace) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ── Resource struct ───────────────────────────────────────────────────────────

type workItemTypeResource struct {
	client *client.AggregatedClient
}

// NewWorkItemTypeResource returns a new resource.Resource.
func NewWorkItemTypeResource() resource.Resource {
	return &workItemTypeResource{}
}

// ── Model ─────────────────────────────────────────────────────────────────────

type workItemTypeResourceModel struct {
	ID                          types.String `tfsdk:"id"`
	ProcessID                   types.String `tfsdk:"process_id"`
	Name                        types.String `tfsdk:"name"`
	Color                       types.String `tfsdk:"color"`
	Description                 types.String `tfsdk:"description"`
	Icon                        types.String `tfsdk:"icon"`
	ParentWorkItemReferenceName types.String `tfsdk:"parent_work_item_reference_name"`
	IsEnabled                   types.Bool   `tfsdk:"is_enabled"`
	ReferenceName               types.String `tfsdk:"reference_name"`
	URL                         types.String `tfsdk:"url"`
	Pages                       types.List   `tfsdk:"pages"`
}

// ── Nested attr types ─────────────────────────────────────────────────────────

var witControlAttrTypes = map[string]attr.Type{
	"id": types.StringType,
}
var witControlObjectType = types.ObjectType{AttrTypes: witControlAttrTypes}

var witGroupAttrTypes = map[string]attr.Type{
	"id":       types.StringType,
	"controls": types.ListType{ElemType: witControlObjectType},
}
var witGroupObjectType = types.ObjectType{AttrTypes: witGroupAttrTypes}

var witSectionAttrTypes = map[string]attr.Type{
	"id":     types.StringType,
	"groups": types.ListType{ElemType: witGroupObjectType},
}
var witSectionObjectType = types.ObjectType{AttrTypes: witSectionAttrTypes}

var witPageAttrTypes = map[string]attr.Type{
	"id":        types.StringType,
	"page_type": types.StringType,
	"sections":  types.ListType{ElemType: witSectionObjectType},
}
var witPageObjectType = types.ObjectType{AttrTypes: witPageAttrTypes}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *workItemTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_workitemtype"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *workItemTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a work item type in an Azure DevOps process.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The reference name of the work item type (used as ID).",
				PlanModifiers: []planmodifier.String{
					witUseStateForUnknownMod(),
				},
			},
			"process_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the process the work item type belongs to.",
				PlanModifiers: []planmodifier.String{
					witRequiresReplaceMod(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of work item type.",
				PlanModifiers: []planmodifier.String{
					witRequiresReplaceMod(),
				},
			},
			"color": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     witStaticStringDef("#009ccc"),
				Description: "Color hexadecimal code to represent the work item type (e.g. #009ccc).",
				Validators: []validator.String{
					witHexColorValidator{},
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     witStaticStringDef(""),
				Description: "Description of the work item type.",
			},
			"icon": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     witStaticStringDef("icon_clipboard"),
				Description: "Icon to represent the work item type.",
			},
			"parent_work_item_reference_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Reference name of the parent work item type.",
				PlanModifiers: []planmodifier.String{
					witRequiresReplaceMod(),
					witUseStateForUnknownMod(),
				},
			},
			"is_enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     witStaticBoolDef(true),
				Description: "True if the work item type is enabled.",
			},
			"reference_name": schema.StringAttribute{
				Computed:    true,
				Description: "Reference name of the work item type.",
				PlanModifiers: []planmodifier.String{
					witUseStateForUnknownMod(),
				},
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "URL of the work item type.",
				PlanModifiers: []planmodifier.String{
					witUseStateForUnknownMod(),
				},
			},
			"pages": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of pages for the work item type.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the page.",
						},
						"page_type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of the page.",
						},
						"sections": schema.ListNestedAttribute{
							Computed:    true,
							Description: "List of sections in the page.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Computed:    true,
										Description: "The ID of the section.",
									},
									"groups": schema.ListNestedAttribute{
										Computed:    true,
										Description: "List of groups in the section.",
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"id": schema.StringAttribute{
													Computed:    true,
													Description: "The ID of the group.",
												},
												"controls": schema.ListNestedAttribute{
													Computed:    true,
													Description: "List of controls in the group.",
													NestedObject: schema.NestedAttributeObject{
														Attributes: map[string]schema.Attribute{
															"id": schema.StringAttribute{
																Computed:    true,
																Description: "The ID of the control.",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *workItemTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ── Create ────────────────────────────────────────────────────────────────────

func (r *workItemTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_workitemtype Create: provider client not configured")
		return
	}

	var model workItemTypeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	colorAPI := witColorToAPI(model.Color.ValueString())
	createReq := workitemtrackingprocess.CreateProcessWorkItemTypeRequest{
		Name:       converter.String(model.Name.ValueString()),
		IsDisabled: converter.Bool(!model.IsEnabled.ValueBool()),
		Color:      &colorAPI,
		Icon:       converter.String(model.Icon.ValueString()),
	}
	if !model.Description.IsNull() && !model.Description.IsUnknown() && model.Description.ValueString() != "" {
		createReq.Description = converter.String(model.Description.ValueString())
	}
	if !model.ParentWorkItemReferenceName.IsNull() && !model.ParentWorkItemReferenceName.IsUnknown() && model.ParentWorkItemReferenceName.ValueString() != "" {
		createReq.InheritsFrom = converter.String(model.ParentWorkItemReferenceName.ValueString())
	}

	created, err := r.client.WorkItemTrackingProcessClient.CreateProcessWorkItemType(ctx, workitemtrackingprocess.CreateProcessWorkItemTypeArgs{
		ProcessId:    converter.UUID(model.ProcessID.ValueString()),
		WorkItemType: &createReq,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("creating work item type: %s", err))
		return
	}
	if created.ReferenceName == nil {
		resp.Diagnostics.AddError("Create error", "creating work item type: reference name is nil")
		return
	}

	model.ID = types.StringValue(*created.ReferenceName)

	resp.Diagnostics.Append(r.refreshModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *workItemTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_workitemtype Read: provider client not configured")
		return
	}

	var model workItemTypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.refreshModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if model.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *workItemTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_workitemtype Update: provider client not configured")
		return
	}

	var plan workItemTypeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state workItemTypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	referenceName := state.ID.ValueString()
	colorAPI := witColorToAPI(plan.Color.ValueString())
	updateReq := &workitemtrackingprocess.UpdateProcessWorkItemTypeRequest{
		IsDisabled: converter.Bool(!plan.IsEnabled.ValueBool()),
		Color:      &colorAPI,
		Icon:       converter.String(plan.Icon.ValueString()),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		updateReq.Description = converter.String(plan.Description.ValueString())
	}

	_, err := r.client.WorkItemTrackingProcessClient.UpdateProcessWorkItemType(ctx, workitemtrackingprocess.UpdateProcessWorkItemTypeArgs{
		ProcessId:          converter.UUID(state.ProcessID.ValueString()),
		WitRefName:         &referenceName,
		WorkItemTypeUpdate: updateReq,
	})
	if err != nil {
		resp.Diagnostics.AddError("Update error", fmt.Sprintf("updating work item type: %s", err))
		return
	}

	plan.ID = state.ID
	plan.ProcessID = state.ProcessID
	plan.ReferenceName = state.ReferenceName

	resp.Diagnostics.Append(r.refreshModel(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *workItemTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_workitemtype Delete: provider client not configured")
		return
	}

	var model workItemTypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	referenceName := model.ID.ValueString()
	err := r.client.WorkItemTrackingProcessClient.DeleteProcessWorkItemType(ctx, workitemtrackingprocess.DeleteProcessWorkItemTypeArgs{
		ProcessId:  converter.UUID(model.ProcessID.ValueString()),
		WitRefName: &referenceName,
	})
	if err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("deleting work item type: %s", err))
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

func (r *workItemTypeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected <processId>/<referenceName>, got %q", req.ID))
		return
	}
	model := workItemTypeResourceModel{
		ProcessID: types.StringValue(parts[0]),
		ID:        types.StringValue(parts[1]),
	}
	resp.Diagnostics.Append(r.refreshModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── refreshModel ─────────────────────────────────────────────────────────────

func (r *workItemTypeResource) refreshModel(ctx context.Context, model *workItemTypeResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	referenceName := model.ID.ValueString()
	processID := model.ProcessID.ValueString()

	wit, err := r.client.WorkItemTrackingProcessClient.GetProcessWorkItemType(ctx, workitemtrackingprocess.GetProcessWorkItemTypeArgs{
		ProcessId:  converter.UUID(processID),
		WitRefName: &referenceName,
		Expand:     &workitemtrackingprocess.GetWorkItemTypeExpandValues.Layout,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			model.ID = types.StringValue("")
			return diags
		}
		diags.AddError("Read error", fmt.Sprintf("reading work item type %s for process %s: %s", referenceName, processID, err))
		return diags
	}

	diags.Append(witPopulateModel(model, wit)...)
	return diags
}

// ── witPopulateModel ──────────────────────────────────────────────────────────

func witPopulateModel(model *workItemTypeResourceModel, wit *workitemtrackingprocess.ProcessWorkItemType) diag.Diagnostics {
	var diags diag.Diagnostics
	if wit.ReferenceName != nil {
		model.ID = types.StringValue(*wit.ReferenceName)
		model.ReferenceName = types.StringValue(*wit.ReferenceName)
	}
	if wit.Name != nil {
		model.Name = types.StringValue(*wit.Name)
	}
	if wit.Description != nil {
		model.Description = types.StringValue(*wit.Description)
	} else {
		model.Description = types.StringValue("")
	}
	if wit.Color != nil {
		model.Color = types.StringValue(witColorToResource(*wit.Color))
	}
	if wit.Icon != nil {
		model.Icon = types.StringValue(*wit.Icon)
	}
	if wit.Inherits != nil {
		model.ParentWorkItemReferenceName = types.StringValue(*wit.Inherits)
	} else {
		model.ParentWorkItemReferenceName = types.StringValue("")
	}
	if wit.IsDisabled != nil {
		model.IsEnabled = types.BoolValue(!*wit.IsDisabled)
	} else {
		model.IsEnabled = types.BoolValue(true)
	}
	if wit.Url != nil {
		model.URL = types.StringValue(*wit.Url)
	}

	pages, d := witBuildPagesValue(wit)
	diags.Append(d...)
	model.Pages = pages
	return diags
}

// ── Page builders ─────────────────────────────────────────────────────────────

func witBuildPagesValue(wit *workitemtrackingprocess.ProcessWorkItemType) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	empty := types.ListValueMust(witPageObjectType, []attr.Value{})
	if wit.Layout == nil || wit.Layout.Pages == nil {
		return empty, diags
	}
	objs := make([]attr.Value, 0, len(*wit.Layout.Pages))
	for _, page := range *wit.Layout.Pages {
		obj, d := witBuildPageValue(page)
		diags.Append(d...)
		if diags.HasError() {
			return empty, diags
		}
		objs = append(objs, obj)
	}
	list, d := types.ListValue(witPageObjectType, objs)
	diags.Append(d...)
	return list, diags
}

func witBuildPageValue(page workitemtrackingprocess.Page) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	sectionsVal, d := witBuildSectionsValue(page.Sections)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}
	pageType := ""
	if page.PageType != nil {
		pageType = string(*page.PageType)
	}
	attrs := map[string]attr.Value{
		"id":        types.StringValue(witStrVal(page.Id)),
		"page_type": types.StringValue(pageType),
		"sections":  sectionsVal,
	}
	obj, d2 := types.ObjectValue(witPageAttrTypes, attrs)
	diags.Append(d2...)
	return obj, diags
}

func witBuildSectionsValue(sections *[]workitemtrackingprocess.Section) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	empty := types.ListValueMust(witSectionObjectType, []attr.Value{})
	if sections == nil {
		return empty, diags
	}
	objs := make([]attr.Value, 0, len(*sections))
	for _, sec := range *sections {
		groupsVal, d := witBuildGroupsValue(sec.Groups)
		diags.Append(d...)
		if diags.HasError() {
			return empty, diags
		}
		attrs := map[string]attr.Value{
			"id":     types.StringValue(witStrVal(sec.Id)),
			"groups": groupsVal,
		}
		obj, d2 := types.ObjectValue(witSectionAttrTypes, attrs)
		diags.Append(d2...)
		if diags.HasError() {
			return empty, diags
		}
		objs = append(objs, obj)
	}
	list, d := types.ListValue(witSectionObjectType, objs)
	diags.Append(d...)
	return list, diags
}

func witBuildGroupsValue(groups *[]workitemtrackingprocess.Group) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	empty := types.ListValueMust(witGroupObjectType, []attr.Value{})
	if groups == nil {
		return empty, diags
	}
	objs := make([]attr.Value, 0, len(*groups))
	for _, grp := range *groups {
		ctrlsVal, d := witBuildControlsValue(grp.Controls)
		diags.Append(d...)
		if diags.HasError() {
			return empty, diags
		}
		attrs := map[string]attr.Value{
			"id":       types.StringValue(witStrVal(grp.Id)),
			"controls": ctrlsVal,
		}
		obj, d2 := types.ObjectValue(witGroupAttrTypes, attrs)
		diags.Append(d2...)
		if diags.HasError() {
			return empty, diags
		}
		objs = append(objs, obj)
	}
	list, d := types.ListValue(witGroupObjectType, objs)
	diags.Append(d...)
	return list, diags
}

func witBuildControlsValue(controls *[]workitemtrackingprocess.Control) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	empty := types.ListValueMust(witControlObjectType, []attr.Value{})
	if controls == nil {
		return empty, diags
	}
	objs := make([]attr.Value, 0, len(*controls))
	for _, ctrl := range *controls {
		attrs := map[string]attr.Value{
			"id": types.StringValue(witStrVal(ctrl.Id)),
		}
		obj, d := types.ObjectValue(witControlAttrTypes, attrs)
		diags.Append(d...)
		if diags.HasError() {
			return empty, diags
		}
		objs = append(objs, obj)
	}
	list, d := types.ListValue(witControlObjectType, objs)
	diags.Append(d...)
	return list, diags
}

// ── Color helpers ─────────────────────────────────────────────────────────────

func witColorToAPI(hex string) string {
	return strings.TrimPrefix(hex, "#")
}

func witColorToResource(api string) string {
	if strings.HasPrefix(api, "#") {
		return api
	}
	return "#" + api
}

func witStrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ── Hex color validator ───────────────────────────────────────────────────────

type witHexColorValidator struct{}

func (v witHexColorValidator) Description(_ context.Context) string {
	return "must be a hex color in #RRGGBB format"
}
func (v witHexColorValidator) MarkdownDescription(_ context.Context) string {
	return "must be a hex color in `#RRGGBB` format"
}

var witHexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func (v witHexColorValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	if !witHexColorRe.MatchString(val) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid color", fmt.Sprintf("Value %q is not a valid hex color (#RRGGBB).", val))
	}
}
