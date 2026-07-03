package workitemtrackingprocess

// resource_field_framework.go — terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_field.

import (
	"context"
	"fmt"
	"strings"

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
	_ resource.Resource                = &fieldResource{}
	_ resource.ResourceWithConfigure   = &fieldResource{}
	_ resource.ResourceWithImportState = &fieldResource{}
)

// ── inline plan modifiers ─────────────────────────────────────────────────────

type fieldUseStateForUnknownString struct{}

func (fieldUseStateForUnknownString) Description(_ context.Context) string { return "use prior state" }

func (fieldUseStateForUnknownString) MarkdownDescription(_ context.Context) string {
	return "use prior state"
}

func (fieldUseStateForUnknownString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type fieldRequiresReplaceString struct{}

func (fieldRequiresReplaceString) Description(_ context.Context) string {
	return "requires replacement if changed"
}

func (fieldRequiresReplaceString) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}

func (fieldRequiresReplaceString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ── inline defaults ───────────────────────────────────────────────────────────

type fieldStaticBoolDefault struct{ value bool }

func (s fieldStaticBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("default value %v", s.value)
}

func (s fieldStaticBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("default value `%v`", s.value)
}

func (s fieldStaticBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(s.value)
}

// ── resource struct ────────────────────────────────────────────────────────────

type fieldResource struct {
	client *client.AggregatedClient
}

// NewFieldResource returns a new resource.Resource for betterado_workitemtrackingprocess_field.
func NewFieldResource() resource.Resource {
	return &fieldResource{}
}

// ── Model ──────────────────────────────────────────────────────────────────────

type fieldResourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProcessID      types.String `tfsdk:"process_id"`
	WorkItemTypeID types.String `tfsdk:"work_item_type_id"`
	FieldID        types.String `tfsdk:"field_id"`
	DefaultValue   types.String `tfsdk:"default_value"`
	ReadOnly       types.Bool   `tfsdk:"read_only"`
	Required       types.Bool   `tfsdk:"required"`
	AllowGroups    types.Bool   `tfsdk:"allow_groups"`
	Customization  types.String `tfsdk:"customization"`
	IsLocked       types.Bool   `tfsdk:"is_locked"`
	URL            types.String `tfsdk:"url"`
}

// ── Metadata ───────────────────────────────────────────────────────────────────

func (r *fieldResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_field"
}

// ── Schema ─────────────────────────────────────────────────────────────────────

func (r *fieldResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a field on a work item type in an Azure DevOps process.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the field assignment (process_id/work_item_type_id/field_id).",
				PlanModifiers: []planmodifier.String{
					fieldUseStateForUnknownString{},
				},
			},
			"process_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the process.",
				PlanModifiers: []planmodifier.String{
					fieldRequiresReplaceString{},
				},
			},
			"work_item_type_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID (reference name) of the work item type.",
				PlanModifiers: []planmodifier.String{
					fieldRequiresReplaceString{},
				},
			},
			"field_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID (reference name) of the field.",
				PlanModifiers: []planmodifier.String{
					fieldRequiresReplaceString{},
				},
			},
			"default_value": schema.StringAttribute{
				Optional:    true,
				Description: "The default value of the field.",
			},
			"read_only": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     fieldStaticBoolDefault{value: false},
				Description: "If true, the field cannot be edited.",
			},
			"required": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     fieldStaticBoolDefault{value: false},
				Description: "If true, the field cannot be empty.",
			},
			// allow_groups is write-only to circumvent the Azure DevOps API bug
			// https://developercommunity.visualstudio.com/t/Custom-field-APIs-are-missing-attributes/11032086
			// The framework does not have a built-in WriteOnly, so we mark it as Optional
			// and simply never read it back from the API. Import will set it to null.
			"allow_groups": schema.BoolAttribute{
				Optional:    true,
				Description: "Allow setting field value to a group identity. Only applies to identity fields.",
			},
			"customization": schema.StringAttribute{
				Computed:    true,
				Description: "Indicates the type of customization on this work item. Possible values are `system`, `inherited`, or `custom`.",
				PlanModifiers: []planmodifier.String{
					fieldUseStateForUnknownString{},
				},
			},
			"is_locked": schema.BoolAttribute{
				Computed:    true,
				Description: "Indicates whether the field definition is locked for editing.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL of the field resource.",
				PlanModifiers: []planmodifier.String{
					fieldUseStateForUnknownString{},
				},
			},
		},
	}
}

// ── Configure ──────────────────────────────────────────────────────────────────

func (r *fieldResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *fieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_field Create: provider client not configured")
		return
	}

	var model fieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	processID := model.ProcessID.ValueString()
	witRefName := model.WorkItemTypeID.ValueString()
	referenceName := model.FieldID.ValueString()

	fieldRequest := &workitemtrackingprocess.AddProcessWorkItemTypeFieldRequest{
		ReferenceName: &referenceName,
		ReadOnly:      converter.Bool(model.ReadOnly.ValueBool()),
		Required:      converter.Bool(model.Required.ValueBool()),
	}

	if !model.DefaultValue.IsNull() && !model.DefaultValue.IsUnknown() {
		fieldRequest.DefaultValue = model.DefaultValue.ValueString()
	}
	if !model.AllowGroups.IsNull() && !model.AllowGroups.IsUnknown() {
		fieldRequest.AllowGroups = converter.Bool(model.AllowGroups.ValueBool())
	}

	args := workitemtrackingprocess.AddFieldToWorkItemTypeArgs{
		ProcessId:  converter.UUID(processID),
		WitRefName: &witRefName,
		Field:      fieldRequest,
	}

	createdField, err := r.client.WorkItemTrackingProcessClient.AddFieldToWorkItemType(ctx, args)
	if err != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("adding field %s to work item type %s: %s", referenceName, witRefName, err))
		return
	}
	if createdField == nil {
		resp.Diagnostics.AddError("Create error", "created field is nil")
		return
	}
	if createdField.ReferenceName == nil {
		resp.Diagnostics.AddError("Create error", "created field has no reference name")
		return
	}

	model.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", processID, witRefName, *createdField.ReferenceName))
	r.flattenField(&model, createdField)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ───────────────────────────────────────────────────────────────────────

func (r *fieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_field Read: provider client not configured")
		return
	}

	var model fieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	processID := model.ProcessID.ValueString()
	witRefName := model.WorkItemTypeID.ValueString()
	fieldRefName := model.FieldID.ValueString()

	args := workitemtrackingprocess.GetWorkItemTypeFieldArgs{
		ProcessId:    converter.UUID(processID),
		WitRefName:   &witRefName,
		FieldRefName: &fieldRefName,
		Expand:       &workitemtrackingprocess.ProcessWorkItemTypeFieldsExpandLevelValues.All,
	}

	field, err := r.client.WorkItemTrackingProcessClient.GetWorkItemTypeField(ctx, args)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading field %s: %s", fieldRefName, err))
		return
	}
	if field == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	r.flattenField(&model, field)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ─────────────────────────────────────────────────────────────────────

func (r *fieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_field Update: provider client not configured")
		return
	}

	var plan, state fieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	processID := plan.ProcessID.ValueString()
	witRefName := plan.WorkItemTypeID.ValueString()
	fieldRefName := plan.FieldID.ValueString()

	fieldUpdate := &workitemtrackingprocess.UpdateProcessWorkItemTypeFieldRequest{
		ReadOnly: converter.Bool(plan.ReadOnly.ValueBool()),
		Required: converter.Bool(plan.Required.ValueBool()),
	}

	if !plan.DefaultValue.IsNull() && !plan.DefaultValue.IsUnknown() {
		fieldUpdate.DefaultValue = plan.DefaultValue.ValueString()
	}
	if !plan.AllowGroups.IsNull() && !plan.AllowGroups.IsUnknown() {
		fieldUpdate.AllowGroups = converter.Bool(plan.AllowGroups.ValueBool())
	}

	args := workitemtrackingprocess.UpdateWorkItemTypeFieldArgs{
		ProcessId:    converter.UUID(processID),
		WitRefName:   &witRefName,
		FieldRefName: &fieldRefName,
		Field:        fieldUpdate,
	}

	updatedField, err := r.client.WorkItemTrackingProcessClient.UpdateWorkItemTypeField(ctx, args)
	if err != nil {
		resp.Diagnostics.AddError("Update error", fmt.Sprintf("updating field %s: %s", fieldRefName, err))
		return
	}

	plan.ID = state.ID
	if updatedField != nil {
		r.flattenField(&plan, updatedField)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ─────────────────────────────────────────────────────────────────────

func (r *fieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_field Delete: provider client not configured")
		return
	}

	var model fieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	processID := model.ProcessID.ValueString()
	witRefName := model.WorkItemTypeID.ValueString()
	fieldRefName := model.FieldID.ValueString()

	args := workitemtrackingprocess.RemoveWorkItemTypeFieldArgs{
		ProcessId:    converter.UUID(processID),
		WitRefName:   &witRefName,
		FieldRefName: &fieldRefName,
	}

	err := r.client.WorkItemTrackingProcessClient.RemoveWorkItemTypeField(ctx, args)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("removing field %s: %s", fieldRefName, err))
	}
}

// ── ImportState ────────────────────────────────────────────────────────────────

// ImportState imports an existing field by "process_id/work_item_type_id/field_id".
func (r *fieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 3)
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected process_id/work_item_type_id/field_id, got %q", req.ID))
		return
	}

	var model fieldResourceModel
	model.ID = types.StringValue(req.ID)
	model.ProcessID = types.StringValue(parts[0])
	model.WorkItemTypeID = types.StringValue(parts[1])
	model.FieldID = types.StringValue(parts[2])

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ────────────────────────────────────────────────────────────────────

func (r *fieldResource) flattenField(model *fieldResourceModel, field *workitemtrackingprocess.ProcessWorkItemTypeField) {
	if field == nil {
		return
	}
	if field.ReferenceName != nil {
		model.FieldID = types.StringValue(*field.ReferenceName)
	}
	if field.DefaultValue != nil {
		model.DefaultValue = types.StringValue(fmt.Sprintf("%v", field.DefaultValue))
	}
	if field.ReadOnly != nil {
		model.ReadOnly = types.BoolValue(*field.ReadOnly)
	} else {
		model.ReadOnly = types.BoolValue(false)
	}
	if field.Required != nil {
		model.Required = types.BoolValue(*field.Required)
	} else {
		model.Required = types.BoolValue(false)
	}
	if field.Customization != nil {
		model.Customization = types.StringValue(string(*field.Customization))
	}
	if field.IsLocked != nil {
		model.IsLocked = types.BoolValue(*field.IsLocked)
	}
	if field.Url != nil {
		model.URL = types.StringValue(*field.Url)
	}
	// allow_groups is write-only; never read back from API.
}
