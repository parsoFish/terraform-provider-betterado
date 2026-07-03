package workitemtrackingprocess

// resource_list_framework.go — terraform-plugin-framework implementation of
// betterado_workitemtrackingprocess_list.

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/workitemtrackingprocess"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &listResource{}
	_ resource.ResourceWithConfigure   = &listResource{}
	_ resource.ResourceWithImportState = &listResource{}
)

// ── inline plan modifiers ─────────────────────────────────────────────────────

type listUseStateForUnknownString struct{}

func (listUseStateForUnknownString) Description(_ context.Context) string { return "use prior state" }

func (listUseStateForUnknownString) MarkdownDescription(_ context.Context) string {
	return "use prior state"
}

func (listUseStateForUnknownString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type listRequiresReplaceString struct{}

func (listRequiresReplaceString) Description(_ context.Context) string {
	return "requires replacement if changed"
}

func (listRequiresReplaceString) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}

func (listRequiresReplaceString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ── inline defaults ───────────────────────────────────────────────────────────

type listStaticStringDefault struct{ value string }

func (s listStaticStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("default value %q", s.value)
}

func (s listStaticStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("default value `%s`", s.value)
}

func (s listStaticStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(s.value)
}

type listStaticBoolDefault struct{ value bool }

func (s listStaticBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("default value %v", s.value)
}

func (s listStaticBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("default value `%v`", s.value)
}

func (s listStaticBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(s.value)
}

// ── validators ─────────────────────────────────────────────────────────────────

type listTypeValidator struct{}

func (v listTypeValidator) Description(_ context.Context) string {
	return "Valid values: string, integer"
}

func (v listTypeValidator) MarkdownDescription(_ context.Context) string {
	return "Valid values: `string`, `integer`"
}

func (v listTypeValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	valid := map[string]bool{"string": true, "integer": true}
	if !valid[req.ConfigValue.ValueString()] {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid list type", "Valid values: string, integer")
	}
}

// ── resource struct ────────────────────────────────────────────────────────────

type listResource struct {
	client *client.AggregatedClient
}

// NewListResource returns a new resource.Resource for betterado_workitemtrackingprocess_list.
func NewListResource() resource.Resource {
	return &listResource{}
}

// ── Model ──────────────────────────────────────────────────────────────────────

type listResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Items       types.List   `tfsdk:"items"`
	IsSuggested types.Bool   `tfsdk:"is_suggested"`
	URL         types.String `tfsdk:"url"`
}

// ── Metadata ───────────────────────────────────────────────────────────────────

func (r *listResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workitemtrackingprocess_list"
}

// ── Schema ─────────────────────────────────────────────────────────────────────

func (r *listResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a picklist (allowed values list) in an Azure DevOps process.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the list.",
				PlanModifiers: []planmodifier.String{
					listUseStateForUnknownString{},
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the list.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     listStaticStringDefault{value: "string"},
				Description: "Data type of the list. Valid values: string, integer.",
				Validators: []validator.String{
					listTypeValidator{},
				},
				PlanModifiers: []planmodifier.String{
					listRequiresReplaceString{},
				},
			},
			"items": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "A list of items.",
			},
			"is_suggested": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     listStaticBoolDefault{value: false},
				Description: "Indicates whether items outside of the suggested list are allowed.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "URL of the list.",
				PlanModifiers: []planmodifier.String{
					listUseStateForUnknownString{},
				},
			},
		},
	}
}

// ── Configure ──────────────────────────────────────────────────────────────────

func (r *listResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *listResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_list Create: provider client not configured")
		return
	}

	var model listResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	items, d := expandListItemsFramework(ctx, model.Items)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := model.Name.ValueString()
	listType := model.Type.ValueString()
	isSuggested := model.IsSuggested.ValueBool()

	picklist := &workitemtrackingprocess.PickList{
		Name:        &name,
		Type:        &listType,
		IsSuggested: &isSuggested,
		Items:       &items,
	}

	created, err := r.client.WorkItemTrackingProcessClient.CreateList(ctx, workitemtrackingprocess.CreateListArgs{
		Picklist: picklist,
	})
	if err != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("creating list: %s", err))
		return
	}
	if created == nil {
		resp.Diagnostics.AddError("Create error", "created list is nil")
		return
	}
	if created.Id == nil {
		resp.Diagnostics.AddError("Create error", "created list has no ID")
		return
	}

	model.ID = types.StringValue(created.Id.String())
	resp.Diagnostics.Append(flattenListFramework(ctx, &model, created)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ───────────────────────────────────────────────────────────────────────

func (r *listResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_list Read: provider client not configured")
		return
	}

	var model listResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("list ID %q is not a valid UUID: %s", model.ID.ValueString(), err))
		return
	}

	list, err := r.client.WorkItemTrackingProcessClient.GetList(ctx, workitemtrackingprocess.GetListArgs{
		ListId: &listID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading list %s: %s", model.ID.ValueString(), err))
		return
	}
	if list == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(flattenListFramework(ctx, &model, list)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ─────────────────────────────────────────────────────────────────────

func (r *listResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_list Update: provider client not configured")
		return
	}

	var plan, state listResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("list ID %q is not a valid UUID: %s", state.ID.ValueString(), err))
		return
	}

	items, d := expandListItemsFramework(ctx, plan.Items)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	isSuggested := plan.IsSuggested.ValueBool()

	picklist := &workitemtrackingprocess.PickList{
		Id:          &listID,
		Name:        &name,
		IsSuggested: &isSuggested,
		Items:       &items,
	}

	_, err = r.client.WorkItemTrackingProcessClient.UpdateList(ctx, workitemtrackingprocess.UpdateListArgs{
		ListId:   &listID,
		Picklist: picklist,
	})
	if err != nil {
		resp.Diagnostics.AddError("Update error", fmt.Sprintf("updating list %s: %s", state.ID.ValueString(), err))
		return
	}

	// Poll GetList until it reflects the desired plan values (eventual consistency).
	// We compare against the desired values rather than the UpdateList response,
	// because the Azure DevOps API can return stale data in the UpdateList response
	// which would cause perpetual drift if we used it as the polling target.
	desiredIsSuggested := isSuggested
	desiredItems := items
	desired := &workitemtrackingprocess.PickList{
		Id:          &listID,
		Name:        &name,
		IsSuggested: &desiredIsSuggested,
		Items:       &desiredItems,
	}

	stateConf := &sdkretry.StateChangeConf{
		Pending:                   []string{"inconsistent"},
		Target:                    []string{"consistent"},
		ContinuousTargetOccurence: 3,
		Refresh: func() (interface{}, string, error) {
			readList, err := r.client.WorkItemTrackingProcessClient.GetList(ctx, workitemtrackingprocess.GetListArgs{
				ListId: &listID,
			})
			if err != nil {
				return nil, "", err
			}
			if !listPickListsEqual(desired, readList) {
				return nil, "inconsistent", nil
			}
			return readList, "consistent", nil
		},
		Timeout: 10 * time.Minute,
	}

	result, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Update error", fmt.Sprintf("waiting for list %s to be consistent: %s", state.ID.ValueString(), err))
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(flattenListFramework(ctx, &plan, result.(*workitemtrackingprocess.PickList))...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ─────────────────────────────────────────────────────────────────────

func (r *listResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_workitemtrackingprocess_list Delete: provider client not configured")
		return
	}

	var model listResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listID, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("list ID %q is not a valid UUID: %s", model.ID.ValueString(), err))
		return
	}

	err = r.client.WorkItemTrackingProcessClient.DeleteList(ctx, workitemtrackingprocess.DeleteListArgs{
		ListId: &listID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("deleting list %s: %s", model.ID.ValueString(), err))
	}
}

// ── ImportState ────────────────────────────────────────────────────────────────

func (r *listResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if _, err := uuid.Parse(req.ID); err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("list ID must be a UUID, got %q: %s", req.ID, err))
		return
	}

	var model listResourceModel
	model.ID = types.StringValue(req.ID)
	// Set empty non-null list; Read will populate items after ImportState.
	model.Items = types.ListValueMust(types.StringType, []attr.Value{})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ────────────────────────────────────────────────────────────────────

func flattenListFramework(ctx context.Context, model *listResourceModel, list *workitemtrackingprocess.PickList) diag.Diagnostics {
	var d diag.Diagnostics
	if list == nil {
		d.AddError("flattenList error", "list is nil")
		return d
	}
	if list.Id != nil {
		model.ID = types.StringValue(list.Id.String())
	}
	if list.Name != nil {
		model.Name = types.StringValue(*list.Name)
	}
	if list.Type != nil {
		model.Type = types.StringValue(strings.ToLower(*list.Type))
	}
	if list.IsSuggested != nil {
		model.IsSuggested = types.BoolValue(*list.IsSuggested)
	}
	if list.Items != nil {
		elems := make([]attr.Value, len(*list.Items))
		for i, item := range *list.Items {
			elems[i] = types.StringValue(item)
		}
		var listD diag.Diagnostics
		model.Items, listD = types.ListValue(types.StringType, elems)
		d.Append(listD...)
	}
	if list.Url != nil {
		model.URL = types.StringValue(*list.Url)
	}
	return d
}

func expandListItemsFramework(_ context.Context, list types.List) ([]string, diag.Diagnostics) {
	var d diag.Diagnostics
	elems := list.Elements()
	items := make([]string, 0, len(elems))
	for _, elem := range elems {
		sv, ok := elem.(types.String)
		if !ok {
			d.AddError("expandListItems error", fmt.Sprintf("expected string element, got %T", elem))
			return nil, d
		}
		items = append(items, sv.ValueString())
	}
	return items, d
}

// listPickListsEqual reports whether the two picklists are equal for the fields
// that are set (non-nil) in a. Fields that are nil in a are treated as wildcards
// (not compared). This allows callers to build a partial "desired" struct with
// only the fields they updated and compare it against the full API response.
// Type comparison is case-insensitive to handle "String" vs "string" from the API.
func listPickListsEqual(a, b *workitemtrackingprocess.PickList) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	if a.Name != nil {
		if b.Name == nil || *a.Name != *b.Name {
			return false
		}
	}
	if a.Type != nil {
		if b.Type == nil || !strings.EqualFold(*a.Type, *b.Type) {
			return false
		}
	}
	if a.IsSuggested != nil {
		if b.IsSuggested == nil || *a.IsSuggested != *b.IsSuggested {
			return false
		}
	}
	if a.Items != nil {
		if b.Items == nil || !slices.Equal(*a.Items, *b.Items) {
			return false
		}
	}
	return true
}
