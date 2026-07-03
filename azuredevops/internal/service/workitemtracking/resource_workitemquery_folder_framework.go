package workitemtracking

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
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

// Compile-time interface checks.
var (
	_ resource.Resource                     = (*WorkItemQueryFolderResource)(nil)
	_ resource.ResourceWithConfigure        = (*WorkItemQueryFolderResource)(nil)
	_ resource.ResourceWithImportState      = (*WorkItemQueryFolderResource)(nil)
	_ resource.ResourceWithConfigValidators = (*WorkItemQueryFolderResource)(nil)
)

// WorkItemQueryFolderResource is the terraform-plugin-framework implementation of betterado_workitemquery_folder.
type WorkItemQueryFolderResource struct {
	client *client.AggregatedClient
}

// NewWorkItemQueryFolderResource returns a new resource.Resource for betterado_workitemquery_folder.
func NewWorkItemQueryFolderResource() resource.Resource {
	return &WorkItemQueryFolderResource{}
}

// workItemQueryFolderModel is the Terraform state model for betterado_workitemquery_folder.
type workItemQueryFolderModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	ParentID  types.String `tfsdk:"parent_id"`
	Area      types.String `tfsdk:"area"`
}

// ── Plan modifiers ─────────────────────────────────────────────────────────────
// Re-use same pattern as workitemquery — all folder fields are ForceNew.

// wiqfStringRequiresReplace forces replacement when value changes (ForceNew equivalent).
type wiqfStringRequiresReplace struct{}

func (m wiqfStringRequiresReplace) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m wiqfStringRequiresReplace) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m wiqfStringRequiresReplace) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// wiqfStringUseStateForUnknown copies the prior state value for computed attributes.
type wiqfStringUseStateForUnknown struct{}

func (m wiqfStringUseStateForUnknown) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m wiqfStringUseStateForUnknown) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m wiqfStringUseStateForUnknown) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	// Only copy state when there is a known prior state value.
	// On first apply state is null; leaving the plan as unknown allows the
	// provider to set the real value after creation without a consistency error.
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ── ConfigValidator: ExactlyOneOf(parent_id, area) for folder ─────────────────

// wiqfExactlyOneOfValidator enforces that exactly one of parent_id or area is set.
type wiqfExactlyOneOfValidator struct{}

func (v wiqfExactlyOneOfValidator) Description(_ context.Context) string {
	return "exactly one of parent_id or area must be set"
}

func (v wiqfExactlyOneOfValidator) MarkdownDescription(_ context.Context) string {
	return "exactly one of parent_id or area must be set"
}

func (v wiqfExactlyOneOfValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var parentID types.String
	var area types.String

	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("parent_id"), &parentID)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("area"), &area)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentSet := !parentID.IsNull() && !parentID.IsUnknown()
	areaSet := !area.IsNull() && !area.IsUnknown()

	if parentSet && areaSet {
		resp.Diagnostics.AddError("Conflicting attributes",
			"exactly one of 'parent_id' or 'area' must be set; both are currently set")
	} else if !parentSet && !areaSet {
		resp.Diagnostics.AddError("Missing required attribute",
			"exactly one of 'parent_id' or 'area' must be set; neither is currently set")
	}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *WorkItemQueryFolderResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_workitemquery_folder"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *WorkItemQueryFolderResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a work item query folder in Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the work item query folder.",
				PlanModifiers: []planmodifier.String{
					wiqfStringUseStateForUnknown{},
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					wiqfStringRequiresReplace{},
				},
				Validators: []validator.String{
					wiqIsUUIDValidator{},
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the work item query folder.",
				PlanModifiers: []planmodifier.String{
					wiqfStringRequiresReplace{},
				},
				Validators: []validator.String{
					wiqNotEmptyValidator{},
				},
			},
			"parent_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the parent folder. Mutually exclusive with 'area'.",
				PlanModifiers: []planmodifier.String{
					wiqfStringRequiresReplace{},
				},
				Validators: []validator.String{
					wiqIsUUIDValidator{},
				},
			},
			"area": schema.StringAttribute{
				Optional:    true,
				Description: "The root area for the folder ('Shared Queries' or 'My Queries'). Mutually exclusive with 'parent_id'.",
				PlanModifiers: []planmodifier.String{
					wiqfStringRequiresReplace{},
				},
				Validators: []validator.String{
					wiqAreaValidator{},
				},
			},
		},
	}
}

// ── ConfigValidators ──────────────────────────────────────────────────────────

func (r *WorkItemQueryFolderResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		wiqfExactlyOneOfValidator{},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *WorkItemQueryFolderResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	agg, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.AggregatedClient, got: %T", req.ProviderData),
		)
		return
	}
	r.client = agg
}

// ── CRUD (Create + Read + Delete only; folder is immutable) ───────────────────

func (r *WorkItemQueryFolderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workItemQueryFolderModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ProjectID.ValueString()
	parent := plan.Area.ValueString()
	if parent == "" {
		parent = plan.ParentID.ValueString()
	}

	params := workitemtracking.CreateQueryArgs{
		Project: &projectID,
		Query:   &parent,
		PostedQuery: &workitemtracking.QueryHierarchyItem{
			Name:     converter.String(plan.Name.ValueString()),
			IsFolder: converter.Bool(true),
		},
	}

	q, err := r.client.WorkItemTrackingClient.CreateQuery(r.client.Ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating work item query folder", err.Error())
		return
	}
	if q == nil || q.Id == nil {
		resp.Diagnostics.AddError("Error creating work item query folder", "API returned nil response or nil ID")
		return
	}

	plan.ID = types.StringValue(q.Id.String())
	r.flattenFolder(q, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WorkItemQueryFolderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workItemQueryFolderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	projectID := state.ProjectID.ValueString()

	q, err := r.client.WorkItemTrackingClient.GetQuery(r.client.Ctx, workitemtracking.GetQueryArgs{
		Project: &projectID,
		Query:   &id,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading work item query folder",
			fmt.Sprintf("GetQuery(%s): %v", id, err),
		)
		return
	}

	if q == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Validate that this is actually a folder.
	if q.IsFolder != nil && !*q.IsFolder {
		resp.Diagnostics.AddError(
			"Unexpected item type",
			fmt.Sprintf("item %s is not a folder", id),
		)
		return
	}

	r.flattenFolder(q, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is not implemented — folder is immutable (no Update CRUD).
// All fields are ForceNew so terraform will destroy+create instead of calling Update.
func (r *WorkItemQueryFolderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported",
		"betterado_workitemquery_folder is immutable; all fields trigger replacement")
}

func (r *WorkItemQueryFolderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workItemQueryFolderModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	projectID := state.ProjectID.ValueString()

	err := r.client.WorkItemTrackingClient.DeleteQuery(r.client.Ctx, workitemtracking.DeleteQueryArgs{
		Project: &projectID,
		Query:   &id,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting work item query folder", err.Error())
	}
}

func (r *WorkItemQueryFolderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// flattenFolder copies API response fields into the state model.
func (r *WorkItemQueryFolderResource) flattenFolder(q *workitemtracking.QueryHierarchyItem, m *workItemQueryFolderModel) {
	if q == nil {
		return
	}
	if q.Name != nil {
		m.Name = types.StringValue(*q.Name)
	}
}
