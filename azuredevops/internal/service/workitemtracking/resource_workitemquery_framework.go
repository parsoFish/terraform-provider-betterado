package workitemtracking

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.Resource                     = (*WorkItemQueryResource)(nil)
	_ resource.ResourceWithConfigure        = (*WorkItemQueryResource)(nil)
	_ resource.ResourceWithImportState      = (*WorkItemQueryResource)(nil)
	_ resource.ResourceWithConfigValidators = (*WorkItemQueryResource)(nil)
)

// WorkItemQueryResource is the terraform-plugin-framework implementation of betterado_workitemquery.
type WorkItemQueryResource struct {
	client *client.AggregatedClient
}

// NewWorkItemQueryResource returns a new resource.Resource for betterado_workitemquery.
func NewWorkItemQueryResource() resource.Resource {
	return &WorkItemQueryResource{}
}

// workItemQueryModel is the Terraform state model for betterado_workitemquery.
type workItemQueryModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	ParentID  types.String `tfsdk:"parent_id"`
	Area      types.String `tfsdk:"area"`
	Wiql      types.String `tfsdk:"wiql"`
	IsPublic  types.Bool   `tfsdk:"is_public"`
	Path      types.String `tfsdk:"path"`
}

// ── Plan modifiers ─────────────────────────────────────────────────────────────

// wiqStringRequiresReplace forces replacement when value changes (ForceNew equivalent).
type wiqStringRequiresReplace struct{}

func (m wiqStringRequiresReplace) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m wiqStringRequiresReplace) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m wiqStringRequiresReplace) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// wiqStringUseStateForUnknown copies the prior state value for computed attributes.
type wiqStringUseStateForUnknown struct{}

func (m wiqStringUseStateForUnknown) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m wiqStringUseStateForUnknown) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m wiqStringUseStateForUnknown) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
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

// wiqBoolUseStateForUnknown copies the prior state bool value for computed attributes.
type wiqBoolUseStateForUnknown struct{}

func (m wiqBoolUseStateForUnknown) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m wiqBoolUseStateForUnknown) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m wiqBoolUseStateForUnknown) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
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

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *WorkItemQueryResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_workitemquery"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *WorkItemQueryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a work item query in Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the work item query.",
				PlanModifiers: []planmodifier.String{
					wiqStringUseStateForUnknown{},
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					wiqStringRequiresReplace{},
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegexp, "value must be a valid UUID"),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the work item query.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(nonWhitespaceRegexp, "value must not be empty or whitespace"),
				},
			},
			"parent_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the parent folder. Mutually exclusive with 'area'.",
				PlanModifiers: []planmodifier.String{
					wiqStringRequiresReplace{},
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidRegexp, "value must be a valid UUID"),
				},
			},
			"area": schema.StringAttribute{
				Optional:    true,
				Description: "The root area for the query ('Shared Queries' or 'My Queries'). Mutually exclusive with 'parent_id'.",
				PlanModifiers: []planmodifier.String{
					wiqStringRequiresReplace{},
				},
				Validators: []validator.String{
					stringvalidator.OneOf("Shared Queries", "My Queries"),
				},
			},
			"wiql": schema.StringAttribute{
				Required:    true,
				Description: "The WIQL query string.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32000),
				},
			},
			"is_public": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the query is public.",
				PlanModifiers: []planmodifier.Bool{
					wiqBoolUseStateForUnknown{},
				},
			},
			"path": schema.StringAttribute{
				Computed:    true,
				Description: "The full path of the query.",
				PlanModifiers: []planmodifier.String{
					wiqStringUseStateForUnknown{},
				},
			},
		},
	}
}

// ── ConfigValidators ──────────────────────────────────────────────────────────

// ConfigValidators uses resourcevalidator.ExactlyOneOf from terraform-plugin-framework-validators
// to enforce that exactly one of parent_id or area is set.
func (r *WorkItemQueryResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("parent_id"),
			path.MatchRoot("area"),
		),
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *WorkItemQueryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *WorkItemQueryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workItemQueryModel
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
			Wiql:     converter.String(plan.Wiql.ValueString()),
			IsFolder: converter.Bool(false),
		},
	}

	q, err := r.client.WorkItemTrackingClient.CreateQuery(r.client.Ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating work item query", err.Error())
		return
	}
	if q == nil || q.Id == nil {
		resp.Diagnostics.AddError("Error creating work item query", "API returned nil response or nil ID")
		return
	}

	plan.ID = types.StringValue(q.Id.String())
	r.flattenQuery(q, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WorkItemQueryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workItemQueryModel
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
			"Error reading work item query",
			fmt.Sprintf("GetQuery(%s): %v", id, err),
		)
		return
	}

	if q == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Validate that the returned item is not a folder.
	if q.IsFolder != nil && *q.IsFolder {
		resp.Diagnostics.AddError(
			"Unexpected query type",
			fmt.Sprintf("query %s is a folder, expected a query", id),
		)
		return
	}

	r.flattenQuery(q, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WorkItemQueryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan workItemQueryModel
	var state workItemQueryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	projectID := plan.ProjectID.ValueString()

	// Fetch the existing item to preserve server-side fields.
	existing, err := r.client.WorkItemTrackingClient.GetQuery(r.client.Ctx, workitemtracking.GetQueryArgs{
		Project: &projectID,
		Query:   &id,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error fetching work item query for update", err.Error())
		return
	}

	existing.Name = converter.String(plan.Name.ValueString())
	existing.Wiql = converter.String(plan.Wiql.ValueString())

	updated, err := r.client.WorkItemTrackingClient.UpdateQuery(r.client.Ctx, workitemtracking.UpdateQueryArgs{
		Project:     &projectID,
		Query:       &id,
		QueryUpdate: existing,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating work item query", err.Error())
		return
	}

	plan.ID = state.ID
	r.flattenQuery(updated, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WorkItemQueryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workItemQueryModel
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
		resp.Diagnostics.AddError("Error deleting work item query", err.Error())
	}
}

func (r *WorkItemQueryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// flattenQuery copies API response fields into the state model.
func (r *WorkItemQueryResource) flattenQuery(q *workitemtracking.QueryHierarchyItem, m *workItemQueryModel) {
	if q == nil {
		return
	}
	if q.Name != nil {
		m.Name = types.StringValue(*q.Name)
	}
	if q.IsPublic != nil {
		m.IsPublic = types.BoolValue(*q.IsPublic)
	} else if m.IsPublic.IsNull() || m.IsPublic.IsUnknown() {
		m.IsPublic = types.BoolValue(false)
	}
	if q.Path != nil {
		m.Path = types.StringValue(*q.Path)
	}
}
