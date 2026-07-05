package workitemtracking

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// referenceNameValidChars matches strings that do NOT contain forbidden characters.
// Allowed: any character that is NOT in the forbidden set: ,;~:/\*|?"&%$!+=()[]{}<>-
var referenceNameValidChars = regexp.MustCompile(`^[^,;~:/\\*|?"&%$!+=()[\]{}<>\-]+$`)

// Compile-time interface checks.
var (
	_ resource.Resource                = (*FieldResource)(nil)
	_ resource.ResourceWithConfigure   = (*FieldResource)(nil)
	_ resource.ResourceWithImportState = (*FieldResource)(nil)
)

// FieldResource is the terraform-plugin-framework implementation of betterado_workitemtracking_field.
type FieldResource struct {
	client *client.AggregatedClient
}

// NewFieldResource returns a new resource.Resource for betterado_workitemtracking_field.
func NewFieldResource() resource.Resource {
	return &FieldResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *FieldResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_workitemtracking_field"
}

// ── Plan modifiers ─────────────────────────────────────────────────────────────

// fieldRequiresReplace is a bool plan modifier that forces replacement when the value changes.
type fieldBoolRequiresReplace struct{}

func (m fieldBoolRequiresReplace) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m fieldBoolRequiresReplace) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m fieldBoolRequiresReplace) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// fieldStringRequiresReplace is a String plan modifier that forces replacement when value changes.
type fieldStringRequiresReplace struct{}

func (m fieldStringRequiresReplace) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m fieldStringRequiresReplace) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m fieldStringRequiresReplace) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// fieldBoolUseStateForUnknown copies the prior state value when the plan value is unknown.
type fieldBoolUseStateForUnknown struct{}

func (m fieldBoolUseStateForUnknown) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m fieldBoolUseStateForUnknown) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m fieldBoolUseStateForUnknown) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// fieldStringUseStateForUnknown copies the prior state value when the plan value is unknown.
type fieldStringUseStateForUnknown struct{}

func (m fieldStringUseStateForUnknown) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m fieldStringUseStateForUnknown) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m fieldStringUseStateForUnknown) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// fieldListUseStateForUnknown copies the prior state list value when the plan value is unknown.
type fieldListUseStateForUnknown struct{}

func (m fieldListUseStateForUnknown) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m fieldListUseStateForUnknown) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m fieldListUseStateForUnknown) PlanModifyList(_ context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ── Schema ────────────────────────────────────────────────────────────────────

// operationAttrTypes returns the attr.Type map used for nested supported_operations objects.
var operationAttrTypes = map[string]attr.Type{
	"name":           types.StringType,
	"reference_name": types.StringType,
}

func (r *FieldResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a work item tracking field in Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			// https://learn.microsoft.com/en-us/azure/devops/boards/work-items/work-item-fields?view=azure-devops#field-attributes
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					fieldStringUseStateForUnknown{},
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The friendly name of the field.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(nonWhitespaceRegexp, "value must not be empty or whitespace"),
				},
				PlanModifiers: []planmodifier.String{
					fieldStringRequiresReplace{},
				},
			},
			"reference_name": schema.StringAttribute{
				Required:    true,
				Description: "The reference name of the field (e.g., Custom.MyField).",
				Validators: []validator.String{
					stringvalidator.RegexMatches(nonWhitespaceRegexp, "value must not be empty or whitespace"),
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(referenceNameValidChars, `cannot contain the following characters: ',;~:/\*|?"&%$!+=()[]{}<>-'`),
				},
				PlanModifiers: []planmodifier.String{
					fieldStringRequiresReplace{},
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of the field. Possible values: `string`, `integer`, `dateTime`, `plainText`, `html`, `treePath`, `history`, `double`, `guid`, `boolean`, `identity`.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						"string", "integer", "dateTime", "plainText", "html",
						"treePath", "history", "double", "guid", "boolean", "identity",
					),
				},
				PlanModifiers: []planmodifier.String{
					fieldStringRequiresReplace{},
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The description of the field.",
				PlanModifiers: []planmodifier.String{
					fieldStringUseStateForUnknown{},
					fieldStringRequiresReplace{},
				},
			},
			"usage": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The usage of the field. Possible values: `none`, `workItem`, `workItemLink`, `tree`, `workItemTypeExtension`. Default: `workItem`.",
				Validators: []validator.String{
					stringvalidator.OneOf("none", "workItem", "workItemLink", "tree", "workItemTypeExtension"),
				},
				PlanModifiers: []planmodifier.String{
					fieldStringUseStateForUnknown{},
					fieldStringRequiresReplace{},
				},
			},
			"read_only": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Indicates whether the field is read-only. Default: `false`.",
				PlanModifiers: []planmodifier.Bool{
					fieldBoolUseStateForUnknown{},
					fieldBoolRequiresReplace{},
				},
			},
			"can_sort_by": schema.BoolAttribute{
				Computed:    true,
				Description: "Indicates whether the field can be sorted in server queries.",
				PlanModifiers: []planmodifier.Bool{
					fieldBoolUseStateForUnknown{},
				},
			},
			"is_queryable": schema.BoolAttribute{
				Computed:    true,
				Description: "Indicates whether the field can be queried in the server.",
				PlanModifiers: []planmodifier.Bool{
					fieldBoolUseStateForUnknown{},
				},
			},
			"is_identity": schema.BoolAttribute{
				Computed:    true,
				Description: "Indicates whether this field is an identity field.",
				PlanModifiers: []planmodifier.Bool{
					fieldBoolUseStateForUnknown{},
				},
			},
			"is_picklist": schema.BoolAttribute{
				Computed:    true,
				Description: "Indicates whether this field is a picklist.",
				PlanModifiers: []planmodifier.Bool{
					fieldBoolUseStateForUnknown{},
				},
			},
			"is_picklist_suggested": schema.BoolAttribute{
				// NOTE! This should be computed only but is kept as optional for backward compatibility.
				Optional:    true,
				Computed:    true,
				Description: "Indicates whether this field is a suggested picklist.",
				PlanModifiers: []planmodifier.Bool{
					fieldBoolUseStateForUnknown{},
				},
			},
			"picklist_id": schema.StringAttribute{
				Optional:    true,
				Description: "The identifier of the picklist associated with this field, if applicable.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegexp, "value must be a valid UUID"),
				},
				PlanModifiers: []planmodifier.String{
					fieldStringRequiresReplace{},
				},
			},
			"is_locked": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Indicates whether this field is locked for editing. Default: `false`.",
				PlanModifiers: []planmodifier.Bool{
					fieldBoolUseStateForUnknown{},
				},
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The URL of the field resource.",
				PlanModifiers: []planmodifier.String{
					fieldStringUseStateForUnknown{},
				},
			},
			"restore": schema.BoolAttribute{
				Optional:    true,
				WriteOnly:   true,
				Description: "Set to `true` to restore a previously deleted field.",
			},
			"supported_operations": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The supported operations on this field.",
				PlanModifiers: []planmodifier.List{
					fieldListUseStateForUnknown{},
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The friendly name of the operation.",
						},
						"reference_name": schema.StringAttribute{
							Computed:    true,
							Description: "The reference name of the operation.",
						},
					},
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *FieldResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	agg, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.AggregatedClient, got: %T", req.ProviderData),
		)
		return
	}
	r.client = agg
}

// ── fieldModel ────────────────────────────────────────────────────────────────

type fieldModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	ReferenceName       types.String `tfsdk:"reference_name"`
	Type                types.String `tfsdk:"type"`
	Description         types.String `tfsdk:"description"`
	Usage               types.String `tfsdk:"usage"`
	ReadOnly            types.Bool   `tfsdk:"read_only"`
	CanSortBy           types.Bool   `tfsdk:"can_sort_by"`
	IsQueryable         types.Bool   `tfsdk:"is_queryable"`
	IsIdentity          types.Bool   `tfsdk:"is_identity"`
	IsPicklist          types.Bool   `tfsdk:"is_picklist"`
	IsPicklistSuggested types.Bool   `tfsdk:"is_picklist_suggested"`
	PicklistID          types.String `tfsdk:"picklist_id"`
	IsLocked            types.Bool   `tfsdk:"is_locked"`
	URL                 types.String `tfsdk:"url"`
	Restore             types.Bool   `tfsdk:"restore"`
	SupportedOperations types.List   `tfsdk:"supported_operations"`
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *FieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan fieldModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if restore was requested.
	restore := !plan.Restore.IsNull() && !plan.Restore.IsUnknown() && plan.Restore.ValueBool()

	if restore {
		// Restore a previously deleted field by setting is_deleted = false
		referenceName := plan.ReferenceName.ValueString()
		isDeleted := false
		args := workitemtracking.UpdateWorkItemFieldArgs{
			FieldNameOrRefName: &referenceName,
			Payload: &workitemtracking.FieldUpdate{
				IsDeleted: &isDeleted,
			},
		}
		restoredField, err := r.client.WorkItemTrackingClient.UpdateWorkItemField(r.client.Ctx, args)
		if err != nil {
			resp.Diagnostics.AddError("Error restoring field", err.Error())
			return
		}
		if restoredField == nil || restoredField.ReferenceName == nil {
			resp.Diagnostics.AddError("Error restoring field", "restored field is nil or has no reference name")
			return
		}
		plan.ID = types.StringValue(*restoredField.ReferenceName)
	} else {
		field := &workitemtracking.WorkItemField2{
			Name:          converter.String(plan.Name.ValueString()),
			ReferenceName: converter.String(plan.ReferenceName.ValueString()),
			Type:          converter.ToPtr(workitemtracking.FieldType(plan.Type.ValueString())),
		}

		if !plan.ReadOnly.IsNull() && !plan.ReadOnly.IsUnknown() {
			field.ReadOnly = converter.Bool(plan.ReadOnly.ValueBool())
		} else {
			field.ReadOnly = converter.Bool(false)
		}
		if !plan.IsPicklistSuggested.IsNull() && !plan.IsPicklistSuggested.IsUnknown() {
			field.IsPicklistSuggested = converter.Bool(plan.IsPicklistSuggested.ValueBool())
		} else {
			field.IsPicklistSuggested = converter.Bool(false)
		}
		if !plan.IsLocked.IsNull() && !plan.IsLocked.IsUnknown() {
			field.IsLocked = converter.Bool(plan.IsLocked.ValueBool())
		} else {
			field.IsLocked = converter.Bool(false)
		}
		if !plan.Usage.IsNull() && !plan.Usage.IsUnknown() && plan.Usage.ValueString() != "" {
			fieldUsage := workitemtracking.FieldUsage(plan.Usage.ValueString())
			field.Usage = &fieldUsage
		}
		if !plan.Description.IsNull() && !plan.Description.IsUnknown() && plan.Description.ValueString() != "" {
			field.Description = converter.String(plan.Description.ValueString())
		}
		if !plan.PicklistID.IsNull() && !plan.PicklistID.IsUnknown() && plan.PicklistID.ValueString() != "" {
			picklistID, err := uuid.Parse(plan.PicklistID.ValueString())
			if err != nil {
				resp.Diagnostics.AddError("Invalid picklist_id", fmt.Sprintf("parsing picklist_id %s: %+v", plan.PicklistID.ValueString(), err))
				return
			}
			field.PicklistId = &picklistID
		}

		args := workitemtracking.CreateWorkItemFieldArgs{
			WorkItemField: field,
		}
		createdField, err := r.client.WorkItemTrackingClient.CreateWorkItemField(r.client.Ctx, args)
		if err != nil {
			resp.Diagnostics.AddError("Error creating field", err.Error())
			return
		}
		if createdField == nil || createdField.ReferenceName == nil {
			resp.Diagnostics.AddError("Error creating field", "created field is nil or has no reference name")
			return
		}
		plan.ID = types.StringValue(*createdField.ReferenceName)
	}

	// Read back the full state.
	d := r.readFieldInto(ctx, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *FieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state fieldModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	d := r.readFieldInto(ctx, &state)
	if d.HasError() {
		for _, de := range d {
			if strings.Contains(de.Detail(), "not found") || strings.Contains(de.Summary(), "not found") ||
				strings.Contains(de.Detail(), "404") || strings.Contains(de.Summary(), "404") {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.Append(d...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *FieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan fieldModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state fieldModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	referenceName := state.ID.ValueString()

	// Only is_locked can change in-place (all ForceNew fields replaced via Create).
	if !plan.IsLocked.Equal(state.IsLocked) {
		isLocked := plan.IsLocked.ValueBool()
		args := workitemtracking.UpdateWorkItemFieldArgs{
			FieldNameOrRefName: &referenceName,
			Payload: &workitemtracking.FieldUpdate{
				IsLocked: &isLocked,
			},
		}
		_, err := r.client.WorkItemTrackingClient.UpdateWorkItemField(r.client.Ctx, args)
		if err != nil {
			resp.Diagnostics.AddError("Error updating field", fmt.Sprintf("updating field %s: %+v", referenceName, err))
			return
		}
	}

	plan.ID = state.ID

	d := r.readFieldInto(ctx, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *FieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state fieldModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	referenceName := state.ID.ValueString()
	args := workitemtracking.DeleteWorkItemFieldArgs{
		FieldNameOrRefName: &referenceName,
	}

	err := utils.RetryOnUnexpectedException(ctx, 10*time.Minute, func() error {
		return r.client.WorkItemTrackingClient.DeleteWorkItemField(r.client.Ctx, args)
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting field", fmt.Sprintf("deleting field %s: %+v", referenceName, err))
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

func (r *FieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The import ID is the reference name (e.g. Custom.MyField).
	plan := fieldModel{
		ID:                  types.StringValue(req.ID),
		SupportedOperations: types.ListValueMust(types.ObjectType{AttrTypes: operationAttrTypes}, []attr.Value{}),
	}

	d := r.readFieldInto(ctx, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── readFieldInto ─────────────────────────────────────────────────────────────

// readFieldInto reads the remote field into m. It uses m.ID as the reference name.
func (r *FieldResource) readFieldInto(_ context.Context, m *fieldModel) diag.Diagnostics {
	var d diag.Diagnostics

	referenceName := m.ID.ValueString()
	args := workitemtracking.GetWorkItemFieldArgs{
		FieldNameOrRefName: &referenceName,
	}

	field, err := r.client.WorkItemTrackingClient.GetWorkItemField(r.client.Ctx, args)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			d.AddError("Field not found", fmt.Sprintf("field %s not found (404)", referenceName))
			return d
		}
		d.AddError("Error reading field", fmt.Sprintf("reading field %s: %+v", referenceName, err))
		return d
	}
	if field == nil {
		d.AddError("Error reading field", "read field is nil")
		return d
	}

	if field.Name != nil {
		m.Name = types.StringValue(*field.Name)
	}
	if field.ReferenceName != nil {
		m.ReferenceName = types.StringValue(*field.ReferenceName)
		m.ID = types.StringValue(*field.ReferenceName)
	}
	if field.Type != nil {
		m.Type = types.StringValue(string(*field.Type))
	}
	if field.Description != nil {
		m.Description = types.StringValue(*field.Description)
	} else {
		m.Description = types.StringValue("")
	}
	if field.Usage != nil {
		m.Usage = types.StringValue(string(*field.Usage))
	}
	if field.ReadOnly != nil {
		m.ReadOnly = types.BoolValue(*field.ReadOnly)
	}
	if field.CanSortBy != nil {
		m.CanSortBy = types.BoolValue(*field.CanSortBy)
	}
	if field.IsQueryable != nil {
		m.IsQueryable = types.BoolValue(*field.IsQueryable)
	}
	if field.IsIdentity != nil {
		m.IsIdentity = types.BoolValue(*field.IsIdentity)
		// NOTE! Azure DevOps API incorrectly returns type=string for identity fields.
		if *field.IsIdentity {
			m.Type = types.StringValue(string(workitemtracking.FieldTypeValues.Identity))
		}
	}
	if field.IsPicklist != nil {
		m.IsPicklist = types.BoolValue(*field.IsPicklist)
	}
	if field.IsPicklistSuggested != nil {
		m.IsPicklistSuggested = types.BoolValue(*field.IsPicklistSuggested)
	}
	if field.PicklistId != nil {
		m.PicklistID = types.StringValue(field.PicklistId.String())
	}
	if field.IsLocked != nil {
		m.IsLocked = types.BoolValue(*field.IsLocked)
	}
	if field.Url != nil {
		m.URL = types.StringValue(*field.Url)
	}

	// supported_operations
	if field.SupportedOperations != nil {
		operationVals := make([]attr.Value, len(*field.SupportedOperations))
		for i, op := range *field.SupportedOperations {
			opName := ""
			if op.Name != nil {
				opName = *op.Name
			}
			opRef := ""
			if op.ReferenceName != nil {
				opRef = *op.ReferenceName
			}
			obj, objDiags := types.ObjectValue(operationAttrTypes, map[string]attr.Value{
				"name":           types.StringValue(opName),
				"reference_name": types.StringValue(opRef),
			})
			d.Append(objDiags...)
			operationVals[i] = obj
		}
		if d.HasError() {
			return d
		}
		list, listDiags := types.ListValue(types.ObjectType{AttrTypes: operationAttrTypes}, operationVals)
		d.Append(listDiags...)
		m.SupportedOperations = list
	} else {
		m.SupportedOperations = types.ListValueMust(types.ObjectType{AttrTypes: operationAttrTypes}, []attr.Value{})
	}

	return d
}
