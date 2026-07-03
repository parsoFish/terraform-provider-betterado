package testplan

import (
	"context"
	"fmt"
	"strconv"
	"time"

	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/testplan"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// ── Plan modifier helpers ─────────────────────────────────────────────────────

// tpRequiresReplace marks an attribute for replacement when its value changes.
type tpRequiresReplace struct{}

func (m tpRequiresReplace) Description(_ context.Context) string {
	return "forces replacement when value changes"
}
func (m tpRequiresReplace) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}
func (m tpRequiresReplace) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// tpUseStateForUnknown carries the prior state value when the plan is unknown.
type tpUseStateForUnknown struct{}

func (m tpUseStateForUnknown) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}
func (m tpUseStateForUnknown) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}
func (m tpUseStateForUnknown) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// Compile-time interface checks.
var (
	_ resource.Resource              = (*TestPlanResource)(nil)
	_ resource.ResourceWithConfigure = (*TestPlanResource)(nil)
)

// TestPlanResource is the terraform-plugin-framework implementation of
// betterado_test_plan.
type TestPlanResource struct {
	client *client.AggregatedClient
}

// NewTestPlanResource returns a new resource.Resource for betterado_test_plan.
func NewTestPlanResource() resource.Resource {
	return &TestPlanResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *TestPlanResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_test_plan"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *TestPlanResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					tpUseStateForUnknown{},
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					tpRequiresReplace{},
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"area_path": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"iteration_path": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"start_date": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"end_date": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *TestPlanResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	agg, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *client.AggregatedClient, got: %T", req.ProviderData),
		)
		return
	}
	r.client = agg
}

// ── State model ───────────────────────────────────────────────────────────────

// testPlanModel is the Terraform state model for betterado_test_plan.
type testPlanModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	Name          types.String `tfsdk:"name"`
	AreaPath      types.String `tfsdk:"area_path"`
	IterationPath types.String `tfsdk:"iteration_path"`
	StartDate     types.String `tfsdk:"start_date"`
	EndDate       types.String `tfsdk:"end_date"`
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *TestPlanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model testPlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	createParams := expandTestPlan(&model)

	created, err := r.client.TestPlanClient.CreateTestPlan(r.client.Ctx, testplan.CreateTestPlanArgs{
		TestPlanCreateParams: createParams,
		Project:              &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating test plan", err.Error())
		return
	}

	flattenTestPlan(&model, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TestPlanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model testPlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	planID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid test plan ID", err.Error())
		return
	}

	plan, err := r.client.TestPlanClient.GetTestPlanById(r.client.Ctx, testplan.GetTestPlanByIdArgs{
		Project: &projectID,
		PlanId:  &planID,
	})
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading test plan", err.Error())
		return
	}
	if plan == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	flattenTestPlan(&model, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TestPlanResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan testPlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state testPlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ProjectID.ValueString()
	planID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid test plan ID", err.Error())
		return
	}

	updateParams := expandTestPlanUpdate(&plan)

	updated, err := r.client.TestPlanClient.UpdateTestPlan(r.client.Ctx, testplan.UpdateTestPlanArgs{
		TestPlanUpdateParams: updateParams,
		Project:              &projectID,
		PlanId:               &planID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating test plan", err.Error())
		return
	}

	plan.ID = state.ID
	flattenTestPlan(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TestPlanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model testPlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	planID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid test plan ID", err.Error())
		return
	}

	err = r.client.TestPlanClient.DeleteTestPlan(r.client.Ctx, testplan.DeleteTestPlanArgs{
		Project: &projectID,
		PlanId:  &planID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting test plan", err.Error())
	}
}

// ── Expand / Flatten ──────────────────────────────────────────────────────────

// expandTestPlan converts state model → API create params.
func expandTestPlan(model *testPlanModel) *testplan.TestPlanCreateParams {
	params := &testplan.TestPlanCreateParams{
		Name: converter.String(model.Name.ValueString()),
	}
	if !model.AreaPath.IsNull() && !model.AreaPath.IsUnknown() {
		params.AreaPath = converter.String(model.AreaPath.ValueString())
	}
	if !model.IterationPath.IsNull() && !model.IterationPath.IsUnknown() {
		params.Iteration = converter.String(model.IterationPath.ValueString())
	}
	if !model.StartDate.IsNull() && !model.StartDate.IsUnknown() && model.StartDate.ValueString() != "" {
		if t, err := time.Parse(time.RFC3339, model.StartDate.ValueString()); err == nil {
			params.StartDate = &azuredevops.Time{Time: t}
		}
	}
	if !model.EndDate.IsNull() && !model.EndDate.IsUnknown() && model.EndDate.ValueString() != "" {
		if t, err := time.Parse(time.RFC3339, model.EndDate.ValueString()); err == nil {
			params.EndDate = &azuredevops.Time{Time: t}
		}
	}
	return params
}

// expandTestPlanUpdate converts state model → API update params.
func expandTestPlanUpdate(model *testPlanModel) *testplan.TestPlanUpdateParams {
	params := &testplan.TestPlanUpdateParams{
		Name: converter.String(model.Name.ValueString()),
	}
	if !model.AreaPath.IsNull() && !model.AreaPath.IsUnknown() {
		params.AreaPath = converter.String(model.AreaPath.ValueString())
	}
	if !model.IterationPath.IsNull() && !model.IterationPath.IsUnknown() {
		params.Iteration = converter.String(model.IterationPath.ValueString())
	}
	if !model.StartDate.IsNull() && !model.StartDate.IsUnknown() && model.StartDate.ValueString() != "" {
		if t, err := time.Parse(time.RFC3339, model.StartDate.ValueString()); err == nil {
			params.StartDate = &azuredevops.Time{Time: t}
		}
	}
	if !model.EndDate.IsNull() && !model.EndDate.IsUnknown() && model.EndDate.ValueString() != "" {
		if t, err := time.Parse(time.RFC3339, model.EndDate.ValueString()); err == nil {
			params.EndDate = &azuredevops.Time{Time: t}
		}
	}
	return params
}

// flattenTestPlan maps API response → state model.
func flattenTestPlan(model *testPlanModel, plan *testplan.TestPlan) {
	if plan.Id != nil {
		model.ID = types.StringValue(strconv.Itoa(*plan.Id))
	}
	model.Name = types.StringValue(converter.ToString(plan.Name, ""))
	model.AreaPath = types.StringValue(converter.ToString(plan.AreaPath, ""))
	model.IterationPath = types.StringValue(converter.ToString(plan.Iteration, ""))

	if plan.StartDate != nil && !plan.StartDate.Time.IsZero() {
		model.StartDate = types.StringValue(plan.StartDate.Time.UTC().Format(time.RFC3339))
	} else {
		model.StartDate = types.StringValue("")
	}
	if plan.EndDate != nil && !plan.EndDate.Time.IsZero() {
		model.EndDate = types.StringValue(plan.EndDate.Time.UTC().Format(time.RFC3339))
	} else {
		model.EndDate = types.StringValue("")
	}
}

// isNotFound returns true for 404 HTTP errors from the ADO API.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	if wrappedErr, ok := err.(azuredevops.WrappedError); ok {
		if wrappedErr.StatusCode != nil && *wrappedErr.StatusCode == 404 {
			return true
		}
	}
	return false
}
