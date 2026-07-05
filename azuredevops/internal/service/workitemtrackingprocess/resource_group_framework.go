package workitemtrackingprocess

// resource_group_framework.go — terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_group.

import (
	"cmp"
	"context"
	"fmt"
	"math"
	"slices"
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
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &groupResource{}
	_ resource.ResourceWithConfigure   = &groupResource{}
	_ resource.ResourceWithImportState = &groupResource{}
)

// ── Inline plan modifiers ─────────────────────────────────────────────────────

type groupUseStateForUnknownString struct{}

func (groupUseStateForUnknownString) Description(_ context.Context) string { return "use prior state" }

func (groupUseStateForUnknownString) MarkdownDescription(_ context.Context) string {
	return "use prior state"
}

func (groupUseStateForUnknownString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type groupRequiresReplaceString struct{}

func (groupRequiresReplaceString) Description(_ context.Context) string {
	return "requires replacement if changed"
}

func (groupRequiresReplaceString) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}

func (groupRequiresReplaceString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ── validators ─────────────────────────────────────────────────────────────────

type groupNonWhiteSpaceValidator struct{}

func (v groupNonWhiteSpaceValidator) Description(_ context.Context) string {
	return "value must not be empty or whitespace-only"
}

func (v groupNonWhiteSpaceValidator) MarkdownDescription(_ context.Context) string {
	return "value must not be empty or whitespace-only"
}

func (v groupNonWhiteSpaceValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if strings.TrimSpace(req.ConfigValue.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid value", "Value must not be empty or whitespace-only")
	}
}

// ── inline bool default ────────────────────────────────────────────────────────

type groupStaticBoolDefault struct{ value bool }

func (d groupStaticBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", d.value)
}

func (d groupStaticBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%v`", d.value)
}

func (d groupStaticBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.value)
}

// ── resource struct ────────────────────────────────────────────────────────────

type groupResource struct {
	client *client.AggregatedClient
}

// NewGroupResource returns a new resource.Resource for betterado_workitemtrackingprocess_group.
func NewGroupResource() resource.Resource {
	return &groupResource{}
}

// ── Model ──────────────────────────────────────────────────────────────────────

// groupControlModel represents a control nested inside a group.
type groupControlModel struct {
	ID             types.String `tfsdk:"id"`
	Label          types.String `tfsdk:"label"`
	ControlType    types.String `tfsdk:"control_type"`
	Visible        types.Bool   `tfsdk:"visible"`
	ReadOnly       types.Bool   `tfsdk:"read_only"`
	Order          types.Int64  `tfsdk:"order"`
	Metadata       types.String `tfsdk:"metadata"`
	Watermark      types.String `tfsdk:"watermark"`
	Inherited      types.Bool   `tfsdk:"inherited"`
	Overridden     types.Bool   `tfsdk:"overridden"`
	IsContribution types.Bool   `tfsdk:"is_contribution"`
	Contribution   types.List   `tfsdk:"contribution"`
}

// groupContributionModel is the nested contribution block.
type groupContributionModel struct {
	ContributionID        types.String `tfsdk:"contribution_id"`
	Height                types.Int64  `tfsdk:"height"`
	Inputs                types.Map    `tfsdk:"inputs"`
	ShowOnDeletedWorkItem types.Bool   `tfsdk:"show_on_deleted_work_item"`
}

type groupResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProcessID      types.String `tfsdk:"process_id"`
	WorkItemTypeID types.String `tfsdk:"work_item_type_reference_name"`
	PageID         types.String `tfsdk:"page_id"`
	SectionID      types.String `tfsdk:"section_id"`
	Label          types.String `tfsdk:"label"`
	Order          types.Int64  `tfsdk:"order"`
	Visible        types.Bool   `tfsdk:"visible"`
	Controls       types.List   `tfsdk:"control"`
}

// ── contribution object type ───────────────────────────────────────────────────

var groupContributionAttrTypes = map[string]attr.Type{
	"contribution_id":           types.StringType,
	"height":                    types.Int64Type,
	"inputs":                    types.MapType{ElemType: types.StringType},
	"show_on_deleted_work_item": types.BoolType,
}

var groupContributionObjectType = types.ObjectType{AttrTypes: groupContributionAttrTypes}

// ── control object type ────────────────────────────────────────────────────────

var groupControlAttrTypes = map[string]attr.Type{
	"id":              types.StringType,
	"label":           types.StringType,
	"control_type":    types.StringType,
	"visible":         types.BoolType,
	"read_only":       types.BoolType,
	"order":           types.Int64Type,
	"metadata":        types.StringType,
	"watermark":       types.StringType,
	"inherited":       types.BoolType,
	"overridden":      types.BoolType,
	"is_contribution": types.BoolType,
	"contribution":    types.ListType{ElemType: groupContributionObjectType},
}

var groupControlObjectType = types.ObjectType{AttrTypes: groupControlAttrTypes}

// ── Metadata ───────────────────────────────────────────────────────────────────

func (r *groupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_group"
}

// ── Schema ─────────────────────────────────────────────────────────────────────

func (r *groupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	contributionSchema := schema.ListNestedAttribute{
		Optional:    true,
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
					Description: "A dictionary holding key value pairs for contribution inputs.",
				},
				"show_on_deleted_work_item": schema.BoolAttribute{
					Optional:    true,
					Computed:    true,
					Default:     groupStaticBoolDefault{value: false},
					Description: "A value indicating if the contribution should be shown on deleted work item.",
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		Description: "Manages a group in an Azure DevOps process work item type layout.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the group.",
				PlanModifiers: []planmodifier.String{
					groupUseStateForUnknownString{},
				},
			},
			"process_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the process.",
				PlanModifiers: []planmodifier.String{
					groupRequiresReplaceString{},
				},
			},
			"work_item_type_reference_name": schema.StringAttribute{
				Required:    true,
				Description: "The reference name of the work item type.",
				PlanModifiers: []planmodifier.String{
					groupRequiresReplaceString{},
				},
				Validators: []validator.String{
					groupNonWhiteSpaceValidator{},
				},
			},
			"page_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the page to add the group to.",
				Validators: []validator.String{
					groupNonWhiteSpaceValidator{},
				},
			},
			"section_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the section to add the group to.",
				Validators: []validator.String{
					groupNonWhiteSpaceValidator{},
				},
			},
			"label": schema.StringAttribute{
				Required:    true,
				Description: "Label for the group.",
				Validators: []validator.String{
					groupNonWhiteSpaceValidator{},
				},
			},
			"order": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Order in which the group should appear in the section.",
			},
			"visible": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     groupStaticBoolDefault{value: true},
				Description: "A value indicating if the group should be hidden or not.",
			},
			"control": schema.ListNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Controls to be created with the group. Required for HTML controls which cannot be added to existing groups.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Required:    true,
							Description: "The ID of the control (field reference name).",
						},
						"label": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Label for the control.",
						},
						"control_type": schema.StringAttribute{
							Computed:    true,
							Description: "Type of the control.",
						},
						"visible": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     groupStaticBoolDefault{value: true},
							Description: "Whether the control is visible.",
						},
						"read_only": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     groupStaticBoolDefault{value: false},
							Description: "Whether the control is read-only.",
						},
						"order": schema.Int64Attribute{
							Computed:    true,
							Description: "Order of the control in the group.",
						},
						"metadata": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Inner text of the control.",
						},
						"watermark": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Watermark text for the textbox.",
						},
						"inherited": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the control is inherited.",
						},
						"overridden": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the control is overridden.",
						},
						"is_contribution": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     groupStaticBoolDefault{value: false},
							Description: "Whether the control is a contribution.",
						},
						"contribution": contributionSchema,
					},
				},
			},
		},
	}
}

// ── Configure ──────────────────────────────────────────────────────────────────

func (r *groupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *groupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_group Create: provider client not configured")
		return
	}

	var model groupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	group := workitemtrackingprocess.Group{
		Label:   converter.String(model.Label.ValueString()),
		Visible: converter.Bool(model.Visible.ValueBool()),
	}
	if !model.Order.IsNull() && !model.Order.IsUnknown() {
		order := int(model.Order.ValueInt64())
		group.Order = &order
	}

	// Expand inline controls if provided
	if !model.Controls.IsNull() && !model.Controls.IsUnknown() {
		var controlModels []groupControlModel
		resp.Diagnostics.Append(model.Controls.ElementsAs(ctx, &controlModels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(controlModels) > 0 {
			controls := make([]workitemtrackingprocess.Control, len(controlModels))
			for i, cm := range controlModels {
				c := workitemtrackingprocess.Control{
					Id:       converter.String(cm.ID.ValueString()),
					Visible:  converter.Bool(cm.Visible.ValueBool()),
					ReadOnly: converter.Bool(cm.ReadOnly.ValueBool()),
					Order:    converter.Int(i),
				}
				if !cm.Label.IsNull() && !cm.Label.IsUnknown() {
					c.Label = converter.String(cm.Label.ValueString())
				}
				if !cm.Metadata.IsNull() && !cm.Metadata.IsUnknown() {
					c.Metadata = converter.String(cm.Metadata.ValueString())
				}
				if !cm.Watermark.IsNull() && !cm.Watermark.IsUnknown() {
					c.Watermark = converter.String(cm.Watermark.ValueString())
				}
				c.IsContribution = converter.Bool(cm.IsContribution.ValueBool())
				if !cm.Contribution.IsNull() && !cm.Contribution.IsUnknown() {
					var contribModels []groupContributionModel
					resp.Diagnostics.Append(cm.Contribution.ElementsAs(ctx, &contribModels, false)...)
					if resp.Diagnostics.HasError() {
						return
					}
					if len(contribModels) > 0 {
						c.Contribution = expandGroupContribution(contribModels[0])
					}
				}
				controls[i] = c
			}
			group.Controls = &controls
		}
	}

	processID := converter.UUID(model.ProcessID.ValueString())
	witRefName := model.WorkItemTypeID.ValueString()
	pageID := model.PageID.ValueString()
	sectionID := model.SectionID.ValueString()

	createdGroup, err := r.client.WorkItemTrackingProcessClient.AddGroup(ctx, workitemtrackingprocess.AddGroupArgs{
		ProcessId:  processID,
		WitRefName: &witRefName,
		PageId:     &pageID,
		SectionId:  &sectionID,
		Group:      &group,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("creating group: %s", err))
		return
	}
	if createdGroup == nil || createdGroup.Id == nil {
		resp.Diagnostics.AddError("Create error", "created group has no ID")
		return
	}

	model.ID = types.StringValue(*createdGroup.Id)

	// If controls were specified, poll until they are all present (eventual consistency).
	expectedControlCount := 0
	if group.Controls != nil {
		expectedControlCount = len(*group.Controls)
	}

	if expectedControlCount > 0 {
		_, err := pollUntilConsistent(ctx, 10*time.Minute, 1*time.Second, 4, func() (interface{}, string, error) {
			if err := r.readGroupIntoModel(ctx, &model); err != nil {
				return nil, "", err
			}
			var currentControls []groupControlModel
			if !model.Controls.IsNull() && !model.Controls.IsUnknown() {
				_ = model.Controls.ElementsAs(ctx, &currentControls, false)
			}
			if len(currentControls) < expectedControlCount {
				return nil, "inconsistent", nil
			}
			return model, "consistent", nil
		})
		if err != nil {
			resp.Diagnostics.AddError("Create error (waiting for controls)", err.Error())
			return
		}
	} else {
		if err := r.readGroupIntoModel(ctx, &model); err != nil {
			resp.Diagnostics.AddError("Create error (read-back)", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ───────────────────────────────────────────────────────────────────────

func (r *groupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_group Read: provider client not configured")
		return
	}

	var model groupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readGroupIntoModel(ctx, &model); err != nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ─────────────────────────────────────────────────────────────────────

func (r *groupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_group Update: provider client not configured")
		return
	}

	var plan, currentState groupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &currentState)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = currentState.ID
	groupID := plan.ID.ValueString()
	processID := converter.UUID(plan.ProcessID.ValueString())
	witRefName := plan.WorkItemTypeID.ValueString()
	pageID := plan.PageID.ValueString()
	sectionID := plan.SectionID.ValueString()

	updateGroup := &workitemtrackingprocess.Group{
		Label:   converter.String(plan.Label.ValueString()),
		Visible: converter.Bool(plan.Visible.ValueBool()),
	}
	if !plan.Order.IsNull() && !plan.Order.IsUnknown() {
		order := int(plan.Order.ValueInt64())
		updateGroup.Order = &order
	}

	// If page_id or section_id changed, we need to move the group
	oldPageID := currentState.PageID.ValueString()
	oldSectionID := currentState.SectionID.ValueString()

	if pageID != oldPageID || sectionID != oldSectionID {
		_, err := r.client.WorkItemTrackingProcessClient.MoveGroupToPage(ctx, workitemtrackingprocess.MoveGroupToPageArgs{
			ProcessId:           processID,
			WitRefName:          &witRefName,
			PageId:              &pageID,
			SectionId:           &sectionID,
			GroupId:             &groupID,
			Group:               updateGroup,
			RemoveFromPageId:    &oldPageID,
			RemoveFromSectionId: &oldSectionID,
		})
		if err != nil {
			resp.Diagnostics.AddError("Update error", fmt.Sprintf("moving group: %s", err))
			return
		}
	} else {
		_, err := r.client.WorkItemTrackingProcessClient.UpdateGroup(ctx, workitemtrackingprocess.UpdateGroupArgs{
			ProcessId:  processID,
			WitRefName: &witRefName,
			PageId:     &pageID,
			SectionId:  &sectionID,
			GroupId:    &groupID,
			Group:      updateGroup,
		})
		if err != nil {
			resp.Diagnostics.AddError("Update error", fmt.Sprintf("updating group: %s", err))
			return
		}
	}

	if err := r.readGroupIntoModel(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("Update error (read-back)", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ─────────────────────────────────────────────────────────────────────

func (r *groupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_group Delete: provider client not configured")
		return
	}

	var model groupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := model.ID.ValueString()
	processID := converter.UUID(model.ProcessID.ValueString())
	witRefName := model.WorkItemTypeID.ValueString()
	pageID := model.PageID.ValueString()
	sectionID := model.SectionID.ValueString()

	err := r.client.WorkItemTrackingProcessClient.RemoveGroup(ctx, workitemtrackingprocess.RemoveGroupArgs{
		ProcessId:  processID,
		WitRefName: &witRefName,
		PageId:     &pageID,
		SectionId:  &sectionID,
		GroupId:    &groupID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("deleting group %q: %s", model.Label.ValueString(), err))
	}
}

// ── ImportState ────────────────────────────────────────────────────────────────

// ImportState imports by "process_id/work_item_type_reference_name/page_id/section_id/group_id".
func (r *groupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 5)
	if len(parts) != 5 {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Expected process_id/work_item_type_reference_name/page_id/section_id/group_id, got %q", req.ID))
		return
	}
	if _, err := uuid.Parse(parts[0]); err != nil {
		resp.Diagnostics.AddError("Invalid process_id", fmt.Sprintf("process_id must be a UUID: %s", err))
		return
	}

	model := groupResourceModel{
		ProcessID:      types.StringValue(parts[0]),
		WorkItemTypeID: types.StringValue(parts[1]),
		PageID:         types.StringValue(parts[2]),
		SectionID:      types.StringValue(parts[3]),
		ID:             types.StringValue(parts[4]),
	}

	if err := r.readGroupIntoModel(ctx, &model); err != nil {
		resp.Diagnostics.AddError("Import error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ────────────────────────────────────────────────────────────────────

func (r *groupResource) readGroupIntoModel(ctx context.Context, model *groupResourceModel) error {
	processID := converter.UUID(model.ProcessID.ValueString())
	witRefName := model.WorkItemTypeID.ValueString()
	groupID := model.ID.ValueString()

	workItemType, err := r.client.WorkItemTrackingProcessClient.GetProcessWorkItemType(ctx, workitemtrackingprocess.GetProcessWorkItemTypeArgs{
		ProcessId:  processID,
		WitRefName: &witRefName,
		Expand:     &workitemtrackingprocess.GetWorkItemTypeExpandValues.Layout,
	})
	if err != nil {
		return fmt.Errorf("getting work item type %q: %w", witRefName, err)
	}

	foundGroup := findGroupById(workItemType.Layout, groupID)
	if foundGroup == nil {
		return fmt.Errorf("not found")
	}

	if foundGroup.Label != nil {
		model.Label = types.StringValue(*foundGroup.Label)
	}
	if foundGroup.Visible != nil {
		model.Visible = types.BoolValue(*foundGroup.Visible)
	}
	if foundGroup.Order != nil {
		model.Order = types.Int64Value(int64(*foundGroup.Order))
	}

	// Flatten controls
	if foundGroup.Controls != nil && len(*foundGroup.Controls) > 0 {
		groupControls := *foundGroup.Controls
		slices.SortStableFunc(groupControls, func(a, b workitemtrackingprocess.Control) int {
			return cmp.Compare(converter.ToInt(a.Order, math.MaxInt), converter.ToInt(b.Order, math.MaxInt))
		})

		controlObjs := make([]attr.Value, len(groupControls))
		for i, c := range groupControls {
			controlObj, err2 := flattenGroupControl(c)
			if err2 != nil {
				return fmt.Errorf("flattening control: %w", err2)
			}
			controlObjs[i] = controlObj
		}

		controlsList, diags := types.ListValue(groupControlObjectType, controlObjs)
		if diags.HasError() {
			return fmt.Errorf("building controls list: %v", diags)
		}
		model.Controls = controlsList
	} else {
		model.Controls = types.ListValueMust(groupControlObjectType, []attr.Value{})
	}

	return nil
}

func flattenGroupControl(c workitemtrackingprocess.Control) (attr.Value, error) {
	// Flatten contribution
	var contribList types.List
	if c.Contribution != nil {
		contrib := c.Contribution
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
		contribObj, diags := types.ObjectValue(groupContributionAttrTypes, contribAttrs)
		if diags.HasError() {
			return nil, fmt.Errorf("building contribution object")
		}
		contribList = types.ListValueMust(groupContributionObjectType, []attr.Value{contribObj})
	} else {
		contribList = types.ListValueMust(groupContributionObjectType, []attr.Value{})
	}

	controlAttrs := map[string]attr.Value{
		"id":              types.StringValue(converter.ToString(c.Id, "")),
		"label":           types.StringValue(converter.ToString(c.Label, "")),
		"control_type":    types.StringValue(converter.ToString(c.ControlType, "")),
		"visible":         types.BoolValue(converter.ToBool(c.Visible, true)),
		"read_only":       types.BoolValue(converter.ToBool(c.ReadOnly, false)),
		"order":           types.Int64Value(int64(converter.ToInt(c.Order, 0))),
		"metadata":        types.StringValue(converter.ToString(c.Metadata, "")),
		"watermark":       types.StringValue(converter.ToString(c.Watermark, "")),
		"inherited":       types.BoolValue(converter.ToBool(c.Inherited, false)),
		"overridden":      types.BoolValue(converter.ToBool(c.Overridden, false)),
		"is_contribution": types.BoolValue(converter.ToBool(c.IsContribution, false)),
		"contribution":    contribList,
	}

	obj, diags := types.ObjectValue(groupControlAttrTypes, controlAttrs)
	if diags.HasError() {
		return nil, fmt.Errorf("building control object")
	}
	return obj, nil
}

func expandGroupContribution(m groupContributionModel) *workitemtrackingprocess.WitContribution {
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
