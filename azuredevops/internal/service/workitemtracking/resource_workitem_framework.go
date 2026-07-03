package workitemtracking

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtracking"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

const (
	customFieldsPrefix = "Custom."
)

var systemFieldMapping = map[string]string{
	"System.State":         "state",
	"System.Title":         "title",
	"System.WorkItemType":  "type",
	"System.AreaPath":      "area_path",
	"System.IterationPath": "iteration_path",
	"System.Parent":        "parent_id",
	"System.Description":   "description",
}

// Compile-time interface checks.
var (
	_ resource.Resource                     = (*WorkItemResource)(nil)
	_ resource.ResourceWithConfigure        = (*WorkItemResource)(nil)
	_ resource.ResourceWithImportState      = (*WorkItemResource)(nil)
	_ resource.ResourceWithConfigValidators = (*WorkItemResource)(nil)
)

// WorkItemResource is the terraform-plugin-framework implementation of betterado_workitem.
type WorkItemResource struct {
	client *client.AggregatedClient
}

// NewWorkItemResource returns a new resource.Resource for betterado_workitem.
func NewWorkItemResource() resource.Resource {
	return &WorkItemResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *WorkItemResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_workitem"
}

// ── planmodifier helpers ──────────────────────────────────────────────────────

// workItemStringRequiresReplace is a String plan modifier that forces replacement
// when the attribute value changes (equivalent to SDKv2 ForceNew).
type workItemStringRequiresReplace struct{}

func workItemRequiresReplace() planmodifier.String { return workItemStringRequiresReplace{} }

func (m workItemStringRequiresReplace) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m workItemStringRequiresReplace) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m workItemStringRequiresReplace) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// workItemStringUseStateForUnknown copies the prior state value when the plan
// value is unknown (prevents perpetual diffs for computed-only attributes).
type workItemStringUseStateForUnknown struct{}

func workItemUseStateForUnknown() planmodifier.String { return workItemStringUseStateForUnknown{} }

func (m workItemStringUseStateForUnknown) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m workItemStringUseStateForUnknown) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m workItemStringUseStateForUnknown) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ── Validators ────────────────────────────────────────────────────────────────

// wiIsUUIDValidator validates that a string value is a valid UUID.
type wiIsUUIDValidator struct{}

func (v wiIsUUIDValidator) Description(_ context.Context) string {
	return "value must be a valid UUID"
}

func (v wiIsUUIDValidator) MarkdownDescription(_ context.Context) string {
	return "value must be a valid UUID"
}

func (v wiIsUUIDValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if _, err := uuid.Parse(req.ConfigValue.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid UUID",
			fmt.Sprintf("%q is not a valid UUID", req.ConfigValue.ValueString()))
	}
}

// wiNotWhitespaceValidator validates that a string value is not empty or whitespace.
type wiNotWhitespaceValidator struct{}

func (v wiNotWhitespaceValidator) Description(_ context.Context) string {
	return "value must not be empty or whitespace"
}

func (v wiNotWhitespaceValidator) MarkdownDescription(_ context.Context) string {
	return "value must not be empty or whitespace"
}

func (v wiNotWhitespaceValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if strings.TrimSpace(req.ConfigValue.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid value",
			"value must not be empty or whitespace")
	}
}

// wiTagsSizeFloorValidator validates that a set attribute has at least 1 element.
// Mirrors SDKv2 MinItems: 1 behaviour on tags.
type wiTagsSizeFloorValidator struct{}

func (v wiTagsSizeFloorValidator) Description(_ context.Context) string {
	return "set must contain at least 1 element"
}

func (v wiTagsSizeFloorValidator) MarkdownDescription(_ context.Context) string {
	return "set must contain at least 1 element"
}

func (v wiTagsSizeFloorValidator) ValidateSet(_ context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if len(req.ConfigValue.Elements()) < 1 {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid tags",
			"tags must contain at least 1 element when specified")
	}
}

// wiParentIDAtLeastValidator validates that parent_id is >= 1 when specified.
// Mirrors SDKv2 ValidateFunc: validation.IntAtLeast(1).
type wiParentIDAtLeastValidator struct{}

func (v wiParentIDAtLeastValidator) Description(_ context.Context) string {
	return "value must be at least 1"
}

func (v wiParentIDAtLeastValidator) MarkdownDescription(_ context.Context) string {
	return "value must be at least 1"
}

func (v wiParentIDAtLeastValidator) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if req.ConfigValue.ValueInt64() < 1 {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid parent_id",
			fmt.Sprintf("parent_id must be at least 1, got %d", req.ConfigValue.ValueInt64()))
	}
}

// wiConflictingFieldsValidator implements resource.ConfigValidator to enforce
// that custom_fields and additional_fields_json are mutually exclusive.
// This mirrors the SDKv2 ConflictsWith constraint (resourcevalidator.Conflicting).
type wiConflictingFieldsValidator struct{}

func (v wiConflictingFieldsValidator) Description(_ context.Context) string {
	return "custom_fields and additional_fields_json are mutually exclusive"
}

func (v wiConflictingFieldsValidator) MarkdownDescription(_ context.Context) string {
	return "custom_fields and additional_fields_json are mutually exclusive"
}

func (v wiConflictingFieldsValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var customFields types.Map
	var additionalFieldsJSON types.String

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("custom_fields"), &customFields)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("additional_fields_json"), &additionalFieldsJSON)...)
	if resp.Diagnostics.HasError() {
		return
	}

	customFieldsSet := !customFields.IsNull() && !customFields.IsUnknown()
	additionalFieldsSet := !additionalFieldsJSON.IsNull() && !additionalFieldsJSON.IsUnknown()

	if customFieldsSet && additionalFieldsSet {
		resp.Diagnostics.AddError(
			"Conflicting attributes",
			"'custom_fields' and 'additional_fields_json' are mutually exclusive; only one may be set",
		)
	}
}

// ConfigValidators returns resource-level config validators.
func (r *WorkItemResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		wiConflictingFieldsValidator{},
	}
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *WorkItemResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					workItemUseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					wiNotWhitespaceValidator{},
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				// TODO: Remove Computed in a major release — matches SDKv2 behaviour.
				PlanModifiers: []planmodifier.String{
					workItemUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					workItemRequiresReplace(),
				},
				Validators: []validator.String{
					wiIsUUIDValidator{},
				},
			},
			"type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					workItemRequiresReplace(),
				},
				Validators: []validator.String{
					wiNotWhitespaceValidator{},
				},
			},
			"state": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					workItemUseStateForUnknown(),
				},
				Validators: []validator.String{
					wiNotWhitespaceValidator{},
				},
			},
			"custom_fields": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				DeprecationMessage: "This property is deprecated and will be removed in a future release. " +
					"Please use \"additional_fields_json\" argument instead.",
			},
			"additional_fields_json": schema.StringAttribute{
				Optional: true,
			},
			"tags": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					wiTagsSizeFloorValidator{},
				},
			},
			"area_path": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					workItemUseStateForUnknown(),
				},
				Validators: []validator.String{
					wiNotWhitespaceValidator{},
				},
			},
			"iteration_path": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					workItemUseStateForUnknown(),
				},
				Validators: []validator.String{
					wiNotWhitespaceValidator{},
				},
			},
			"parent_id": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					wiParentIDAtLeastValidator{},
				},
			},
			"url": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					workItemUseStateForUnknown(),
				},
			},
			"relations": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"rel": schema.StringAttribute{
							Computed: true,
						},
						"url": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *WorkItemResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ── workItemModel ─────────────────────────────────────────────────────────────

// workItemModel is the Terraform state model for betterado_workitem.
type workItemModel struct {
	ID                   types.String `tfsdk:"id"`
	Title                types.String `tfsdk:"title"`
	Description          types.String `tfsdk:"description"`
	ProjectID            types.String `tfsdk:"project_id"`
	Type                 types.String `tfsdk:"type"`
	State                types.String `tfsdk:"state"`
	CustomFields         types.Map    `tfsdk:"custom_fields"`
	AdditionalFieldsJSON types.String `tfsdk:"additional_fields_json"`
	Tags                 types.Set    `tfsdk:"tags"`
	AreaPath             types.String `tfsdk:"area_path"`
	IterationPath        types.String `tfsdk:"iteration_path"`
	ParentID             types.Int64  `tfsdk:"parent_id"`
	URL                  types.String `tfsdk:"url"`
	Relations            types.List   `tfsdk:"relations"`
}

// relationAttrTypes returns the attr.Type map used for nested relation objects.
var relationAttrTypes = map[string]attr.Type{
	"rel": types.StringType,
	"url": types.StringType,
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *WorkItemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workItemModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgName := strings.Split(r.client.OrganizationURL, "/")[3]

	var ops []webapi.JsonPatchOperation
	ops = fwExpandSystemFields(plan, ops, orgName, nil)
	ops = fwExpandCustomFields(plan, ops)
	ops = fwExpandAdditionalFieldsCreate(plan, ops)
	ops = fwExpandTags(plan, ops, webapi.OperationValues.Add)

	workItem, err := r.client.WorkItemTrackingClient.CreateWorkItem(r.client.Ctx, workitemtracking.CreateWorkItemArgs{
		Project:  converter.String(plan.ProjectID.ValueString()),
		Type:     converter.String(plan.Type.ValueString()),
		Document: &ops,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating work item", err.Error())
		return
	}

	plan.ID = types.StringValue(strconv.Itoa(*workItem.Id))
	resp.Diagnostics.Append(fwReadWorkItemInto(ctx, r.client, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *WorkItemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workItemModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := fwReadWorkItemInto(ctx, r.client, &state)
	if diags.HasError() {
		// Check if it was a not-found — remove from state.
		for _, d := range diags {
			if strings.Contains(d.Detail(), "not found") || strings.Contains(d.Summary(), "not found") {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.Append(diags...)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *WorkItemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state workItemModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid work item ID", err.Error())
		return
	}
	project := plan.ProjectID.ValueString()
	orgName := strings.Split(r.client.OrganizationURL, "/")[3]

	var ops []webapi.JsonPatchOperation
	ops = fwExpandSystemFields(plan, ops, orgName, &state)
	ops = fwExpandCustomFields(plan, ops)
	ops = fwExpandAdditionalFieldsUpdate(ctx, r.client, plan, state, id, project, ops, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	ops = fwExpandTags(plan, ops, webapi.OperationValues.Replace)

	_, err = r.client.WorkItemTrackingClient.UpdateWorkItem(r.client.Ctx, workitemtracking.UpdateWorkItemArgs{
		Id:       &id,
		Project:  &project,
		Document: &ops,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating work item",
			fmt.Sprintf("Project ID: %s, Work Item: %d, Error: %v", project, id, err),
		)
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(fwReadWorkItemInto(ctx, r.client, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *WorkItemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workItemModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid work item ID", err.Error())
		return
	}

	_, err = r.client.WorkItemTrackingClient.DeleteWorkItem(r.client.Ctx, workitemtracking.DeleteWorkItemArgs{
		Id: &id,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting work item", err.Error())
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

func (r *WorkItemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: <project_id>/<work_item_id>
	// Mirrors SDKv2 ImportProjectQualifiedResource.
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format <project_id>/<work_item_id>, got: %q", req.ID),
		)
		return
	}

	var state workItemModel
	state.ProjectID = types.StringValue(parts[0])
	state.ID = types.StringValue(parts[1])
	// Zero out optional fields so Read can populate them cleanly.
	state.CustomFields = types.MapValueMust(types.StringType, map[string]attr.Value{})
	state.AdditionalFieldsJSON = types.StringNull()
	state.Tags = types.SetValueMust(types.StringType, []attr.Value{})
	state.Relations = types.ListValueMust(
		types.ObjectType{AttrTypes: relationAttrTypes},
		[]attr.Value{},
	)

	resp.Diagnostics.Append(fwReadWorkItemInto(ctx, r.client, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// fwReadWorkItemInto performs a live API GET and populates model fields in-place.
func fwReadWorkItemInto(ctx context.Context, clients *client.AggregatedClient, m *workItemModel) diag.Diagnostics {
	var diags diag.Diagnostics

	id, err := strconv.Atoi(m.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid work item ID", err.Error())
		return diags
	}

	workItem, err := clients.WorkItemTrackingClient.GetWorkItem(clients.Ctx, workitemtracking.GetWorkItemArgs{
		Project: converter.String(m.ProjectID.ValueString()),
		Id:      &id,
		Expand:  converter.ToPtr(workitemtracking.WorkItemExpandValues.All),
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			diags.AddError("work item not found", fmt.Sprintf("work item %d not found: %v", id, err))
			return diags
		}
		diags.AddError("Error reading work item", err.Error())
		return diags
	}

	if workItem == nil {
		return diags
	}

	if workItem.Url != nil {
		m.URL = types.StringValue(*workItem.Url)
	}

	if workItem.Fields != nil {
		diags.Append(fwFlattenFields(ctx, m, workItem.Fields)...)
	}

	// Relations
	var relVals []attr.Value
	if workItem.Relations != nil {
		for _, rel := range *workItem.Relations {
			relObj, d := types.ObjectValue(relationAttrTypes, map[string]attr.Value{
				"rel": types.StringValue(converter.ToString(rel.Rel, "")),
				"url": types.StringValue(converter.ToString(rel.Url, "")),
			})
			diags.Append(d...)
			if !d.HasError() {
				relVals = append(relVals, relObj)
			}
		}
	}
	if relVals == nil {
		relVals = []attr.Value{}
	}
	listVal, d := types.ListValue(
		types.ObjectType{AttrTypes: relationAttrTypes},
		relVals,
	)
	diags.Append(d...)
	if !d.HasError() {
		m.Relations = listVal
	}

	return diags
}

// fwFlattenFields populates scalar + map fields from the ADO API fields map.
func fwFlattenFields(ctx context.Context, m *workItemModel, fields *map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Parse config maps for additional_fields_json tracking.
	configMap := make(map[string]interface{})
	stateMap := make(map[string]interface{})

	if !m.AdditionalFieldsJSON.IsNull() && !m.AdditionalFieldsJSON.IsUnknown() {
		v := m.AdditionalFieldsJSON.ValueString()
		if v != "" {
			if err := json.Unmarshal([]byte(v), &configMap); err != nil {
				diags.AddError("Error parsing additional_fields_json (config)", err.Error())
				return diags
			}
			// stateMap mirrors configMap for tracking purposes.
			if err := json.Unmarshal([]byte(v), &stateMap); err != nil {
				diags.AddError("Error parsing additional_fields_json (state)", err.Error())
				return diags
			}
		}
	}

	customFieldsMap := make(map[string]attr.Value)
	additionalFields := make(map[string]interface{})

	// Pre-initialise all Computed optional fields to null/zero so that they are
	// always known values even when the ADO API omits the corresponding field
	// from its response. Without this the framework sees the field as
	// (known after apply) and rejects the apply with "provider still indicated
	// an unknown value".
	if m.Description.IsUnknown() {
		m.Description = types.StringNull()
	}
	if m.State.IsUnknown() {
		m.State = types.StringNull()
	}
	if m.AreaPath.IsUnknown() {
		m.AreaPath = types.StringNull()
	}
	if m.IterationPath.IsUnknown() {
		m.IterationPath = types.StringNull()
	}
	if m.URL.IsUnknown() {
		m.URL = types.StringNull()
	}

	for key, value := range *fields {
		if tfName, ok := systemFieldMapping[key]; ok {
			switch tfName {
			case "title":
				m.Title = types.StringValue(fmt.Sprintf("%v", value))
			case "type":
				m.Type = types.StringValue(fmt.Sprintf("%v", value))
			case "state":
				m.State = types.StringValue(fmt.Sprintf("%v", value))
			case "area_path":
				m.AreaPath = types.StringValue(fmt.Sprintf("%v", value))
			case "iteration_path":
				m.IterationPath = types.StringValue(fmt.Sprintf("%v", value))
			case "description":
				if value != nil {
					m.Description = types.StringValue(fmt.Sprintf("%v", value))
				} else {
					m.Description = types.StringNull()
				}
			case "parent_id":
				// parent_id comes back as a float64 from JSON
				if value != nil {
					switch v := value.(type) {
					case float64:
						m.ParentID = types.Int64Value(int64(v))
					case int:
						m.ParentID = types.Int64Value(int64(v))
					case int64:
						m.ParentID = types.Int64Value(v)
					}
				}
			}
		} else if _, ok := configMap[key]; ok {
			additionalFields[key] = value
		} else if _, ok := stateMap[key]; ok {
			additionalFields[key] = value
		} else if strings.HasPrefix(key, customFieldsPrefix) {
			shortName := strings.ReplaceAll(key, customFieldsPrefix, "")
			customFieldsMap[shortName] = types.StringValue(fmt.Sprintf("%v", value))
		} else if key == "System.Tags" {
			// Tags come back as a semicolon-separated string.
			// Only populate m.Tags when the config/state already has a non-null
			// value (meaning the user manages tags) OR when the API returned
			// actual tags. This prevents a perpetual diff where config omits
			// `tags` (null) but state would be set to an empty set [].
			tagStr, _ := value.(string)
			var tagVals []attr.Value
			if tagStr != "" {
				for _, t := range strings.Split(tagStr, "; ") {
					tagVals = append(tagVals, types.StringValue(t))
				}
			}
			if tagVals != nil {
				// Real tags returned by ADO — always persist them.
				setVal, d := types.SetValue(types.StringType, tagVals)
				diags.Append(d...)
				if !d.HasError() {
					m.Tags = setVal
				}
			} else if !m.Tags.IsNull() {
				// No tags from ADO, but the user has a non-null tags attribute
				// (e.g. empty set [] or previously populated set) — persist
				// an empty set so the diff remains zero.
				setVal, d := types.SetValue(types.StringType, []attr.Value{})
				diags.Append(d...)
				if !d.HasError() {
					m.Tags = setVal
				}
			}
			// If m.Tags.IsNull() and no tags from ADO → leave it null (no diff).
		}
	}

	// Persist custom_fields map — but only when the user has a non-null value
	// (i.e. is managing custom_fields) OR when the API returned custom fields.
	// When the config omits custom_fields (null) and ADO returned no Custom.*
	// fields, leave m.CustomFields null to avoid a perpetual null vs {} diff.
	if len(customFieldsMap) > 0 {
		cfMap, d := types.MapValue(types.StringType, customFieldsMap)
		diags.Append(d...)
		if !d.HasError() {
			m.CustomFields = cfMap
		}
	} else if !m.CustomFields.IsNull() {
		cfMap, d := types.MapValue(types.StringType, map[string]attr.Value{})
		diags.Append(d...)
		if !d.HasError() {
			m.CustomFields = cfMap
		}
	}
	// If m.CustomFields.IsNull() and no Custom.* fields → leave it null (no diff).

	// Persist additional_fields_json only when there are tracked fields.
	if len(additionalFields) > 0 {
		b, err := json.Marshal(additionalFields)
		if err != nil {
			diags.AddError("Error serialising additional_fields_json", err.Error())
		} else {
			m.AdditionalFieldsJSON = types.StringValue(string(b))
		}
	}

	return diags
}

// fwExpandSystemFields builds JsonPatchOperations from model scalar fields.
// prevState is non-nil on Update; nil on Create.
func fwExpandSystemFields(plan workItemModel, ops []webapi.JsonPatchOperation, orgName string, prevState *workItemModel) []webapi.JsonPatchOperation {
	// Scalar system fields (excluding parent_id and description).
	scalarFields := map[string]string{
		"title":          "System.Title",
		"type":           "System.WorkItemType",
		"state":          "System.State",
		"area_path":      "System.AreaPath",
		"iteration_path": "System.IterationPath",
	}
	modelVals := map[string]string{
		"title":          plan.Title.ValueString(),
		"type":           plan.Type.ValueString(),
		"state":          plan.State.ValueString(),
		"area_path":      plan.AreaPath.ValueString(),
		"iteration_path": plan.IterationPath.ValueString(),
	}
	for tf, api := range scalarFields {
		v := modelVals[tf]
		if v != "" {
			ops = append(ops, webapi.JsonPatchOperation{
				Op:    &webapi.OperationValues.Add,
				From:  nil,
				Path:  converter.String("/fields/" + api),
				Value: v,
			})
		}
	}

	// description — always patch on change (even when empty).
	descChanged := prevState == nil || !plan.Description.Equal(prevState.Description)
	if descChanged {
		desc := plan.Description.ValueString()
		ops = append(ops, webapi.JsonPatchOperation{
			Op:    &webapi.OperationValues.Add,
			From:  nil,
			Path:  converter.String("/fields/System.Description"),
			Value: desc,
		})
	}

	// parent_id — requires link operations.
	var oldParentID, newParentID int64
	if prevState != nil && !prevState.ParentID.IsNull() {
		oldParentID = prevState.ParentID.ValueInt64()
	}
	if !plan.ParentID.IsNull() {
		newParentID = plan.ParentID.ValueInt64()
	}
	parentChanged := prevState == nil || oldParentID != newParentID

	if parentChanged {
		// Remove existing parent link when updating from a non-zero value.
		if prevState != nil && oldParentID > 0 {
			// We need the current relations from state to find the index.
			// This matches the SDKv2 implementation which iterates d.Get("relations").
			// On Update we have prevState.Relations already in model form; we rely on
			// the relations list populated by the prior Read.
			//
			// Because the framework Read ran before Update, prevState.Relations holds
			// the live relations. We iterate them to find the hierarchy-reverse link.
			if !prevState.Relations.IsNull() && !prevState.Relations.IsUnknown() {
				relElems := prevState.Relations.Elements()
				for idx, elem := range relElems {
					obj, ok := elem.(types.Object)
					if !ok {
						continue
					}
					relAttr := obj.Attributes()
					relType, _ := relAttr["rel"].(types.String)
					if strings.EqualFold(relType.ValueString(), "System.LinkTypes.Hierarchy-Reverse") {
						idxInt := idx
						ops = append(ops, webapi.JsonPatchOperation{
							Op:   &webapi.OperationValues.Remove,
							From: nil,
							Path: converter.String(fmt.Sprintf("/relations/%d", idxInt)),
						})
					}
				}
			}
		}

		if newParentID > 0 {
			ops = append(ops, webapi.JsonPatchOperation{
				Op:   &webapi.OperationValues.Add,
				From: nil,
				Path: converter.String("/relations/-"),
				Value: &map[string]string{
					"rel": "System.LinkTypes.Hierarchy-Reverse",
					"url": fmt.Sprintf("https://dev.azure.com/%s/_apis/wit/workItems/%d", orgName, newParentID),
				},
			})
		}
	}

	return ops
}

// fwExpandCustomFields builds JsonPatchOperations from custom_fields map.
func fwExpandCustomFields(plan workItemModel, ops []webapi.JsonPatchOperation) []webapi.JsonPatchOperation {
	if plan.CustomFields.IsNull() || plan.CustomFields.IsUnknown() {
		return ops
	}
	cfElems := plan.CustomFields.Elements()
	for name, val := range cfElems {
		strVal, _ := val.(types.String)
		ops = append(ops, webapi.JsonPatchOperation{
			Op:    &webapi.OperationValues.Add,
			From:  nil,
			Path:  converter.String("/fields/" + customFieldsPrefix + name),
			Value: strVal.ValueString(),
		})
	}
	return ops
}

// fwExpandAdditionalFieldsCreate builds operations for additional_fields_json on Create.
func fwExpandAdditionalFieldsCreate(plan workItemModel, ops []webapi.JsonPatchOperation) []webapi.JsonPatchOperation {
	if plan.AdditionalFieldsJSON.IsNull() || plan.AdditionalFieldsJSON.IsUnknown() {
		return ops
	}
	v := plan.AdditionalFieldsJSON.ValueString()
	if v == "" {
		return ops
	}
	var fields map[string]interface{}
	if err := json.Unmarshal([]byte(v), &fields); err != nil {
		return ops // parse error will surface at validate; silently skip here
	}
	for name, value := range fields {
		ops = append(ops, webapi.JsonPatchOperation{
			Op:    &webapi.OperationValues.Add,
			From:  nil,
			Path:  converter.String("/fields/" + name),
			Value: value,
		})
	}
	return ops
}

// fwExpandAdditionalFieldsUpdate performs the three-way diff (remove/update/add) for
// additional_fields_json on Update, matching the SDKv2 logic exactly.
func fwExpandAdditionalFieldsUpdate(
	ctx context.Context,
	clients *client.AggregatedClient,
	plan, state workItemModel,
	id int, project string,
	ops []webapi.JsonPatchOperation,
	diags *diag.Diagnostics,
) []webapi.JsonPatchOperation {
	if plan.AdditionalFieldsJSON.Equal(state.AdditionalFieldsJSON) {
		return ops
	}

	oldStr := ""
	newStr := ""
	if !state.AdditionalFieldsJSON.IsNull() {
		oldStr = state.AdditionalFieldsJSON.ValueString()
	}
	if !plan.AdditionalFieldsJSON.IsNull() {
		newStr = plan.AdditionalFieldsJSON.ValueString()
	}

	oldFieldsMap := make(map[string]interface{})
	newFieldsMap := make(map[string]interface{})
	apiFields := make(map[string]interface{})
	removeFields := make(map[string]interface{})
	updateFields := make(map[string]interface{})
	addFields := make(map[string]interface{})

	// Fetch live API fields to know which require update vs add.
	workItem, err := clients.WorkItemTrackingClient.GetWorkItem(clients.Ctx, workitemtracking.GetWorkItemArgs{
		Project: converter.String(project),
		Id:      &id,
		Expand:  converter.ToPtr(workitemtracking.WorkItemExpandValues.All),
	})
	if err != nil {
		diags.AddError("Error fetching work item during update", err.Error())
		return ops
	}
	if workItem.Fields != nil {
		apiFields = *workItem.Fields
	}

	if oldStr != "" {
		if err := json.Unmarshal([]byte(oldStr), &oldFieldsMap); err != nil {
			diags.AddError("Error parsing old additional_fields_json", err.Error())
			return ops
		}
	}
	if newStr != "" {
		if err := json.Unmarshal([]byte(newStr), &newFieldsMap); err != nil {
			diags.AddError("Error parsing new additional_fields_json", err.Error())
			return ops
		}
	}

	for k := range oldFieldsMap {
		if _, ok := newFieldsMap[k]; !ok {
			removeFields[k] = ""
		}
	}
	for k := range apiFields {
		if _, ok := newFieldsMap[k]; ok {
			updateFields[k] = newFieldsMap[k]
		}
	}
	for k, v := range newFieldsMap {
		if _, ok := updateFields[k]; !ok {
			addFields[k] = v
		}
	}

	for name := range removeFields {
		ops = append(ops, webapi.JsonPatchOperation{
			Op:    &webapi.OperationValues.Remove,
			From:  nil,
			Path:  converter.String("/fields/" + name),
			Value: nil,
		})
	}
	for name, value := range updateFields {
		ops = append(ops, webapi.JsonPatchOperation{
			Op:    &webapi.OperationValues.Replace,
			From:  nil,
			Path:  converter.String("/fields/" + name),
			Value: value,
		})
	}
	for name, value := range addFields {
		ops = append(ops, webapi.JsonPatchOperation{
			Op:    &webapi.OperationValues.Add,
			From:  nil,
			Path:  converter.String("/fields/" + name),
			Value: value,
		})
	}

	return ops
}

// fwExpandTags builds JsonPatchOperations for the tags set attribute.
func fwExpandTags(plan workItemModel, ops []webapi.JsonPatchOperation, op webapi.Operation) []webapi.JsonPatchOperation {
	var tagStr string
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		elems := plan.Tags.Elements()
		tags := make([]string, 0, len(elems))
		for _, e := range elems {
			s, _ := e.(types.String)
			tags = append(tags, s.ValueString())
		}
		tagStr = strings.Join(tags, "; ")
	}
	ops = append(ops, webapi.JsonPatchOperation{
		Op:    &op,
		From:  nil,
		Path:  converter.String("/fields/System.Tags"),
		Value: tagStr,
	})
	return ops
}
