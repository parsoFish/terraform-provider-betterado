package testplan

import (
	"context"
	"fmt"
	"strconv"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/testplan"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource              = (*TestSuiteResource)(nil)
	_ resource.ResourceWithConfigure = (*TestSuiteResource)(nil)
)

// TestSuiteResource is the terraform-plugin-framework implementation of
// betterado_test_suite.
type TestSuiteResource struct {
	client *client.AggregatedClient
}

// NewTestSuiteResource returns a new resource.Resource for betterado_test_suite.
func NewTestSuiteResource() resource.Resource {
	return &TestSuiteResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *TestSuiteResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_test_suite"
}

// ── Schema ────────────────────────────────────────────────────────────────────

// Suite type constants (exposed values in TF schema).
const (
	suiteTypeStatic      = "staticTestSuite"
	suiteTypeDynamic     = "dynamicTestSuite"
	suiteTypeRequirement = "requirementTestSuite"
)

func (r *TestSuiteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"plan_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					tpRequiresReplace{},
				},
			},
			"parent_suite_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					tpRequiresReplace{},
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			// suite_type: "staticTestSuite", "dynamicTestSuite", "requirementTestSuite"
			"suite_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					tpRequiresReplace{},
				},
			},
			// query_string is required for dynamicTestSuite; ignored for other types.
			"query_string": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *TestSuiteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// testSuiteModel is the Terraform state model for betterado_test_suite.
type testSuiteModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	PlanID        types.String `tfsdk:"plan_id"`
	ParentSuiteID types.String `tfsdk:"parent_suite_id"`
	Name          types.String `tfsdk:"name"`
	SuiteType     types.String `tfsdk:"suite_type"`
	QueryString   types.String `tfsdk:"query_string"`
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *TestSuiteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model testSuiteModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	planID, err := strconv.Atoi(model.PlanID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid plan_id", err.Error())
		return
	}

	createParams := expandTestSuite(&model)
	created, err := r.client.TestPlanClient.CreateTestSuite(r.client.Ctx, testplan.CreateTestSuiteArgs{
		TestSuiteCreateParams: createParams,
		Project:               &projectID,
		PlanId:                &planID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating test suite", err.Error())
		return
	}

	flattenTestSuite(&model, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TestSuiteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model testSuiteModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	planID, err := strconv.Atoi(model.PlanID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid plan_id", err.Error())
		return
	}
	suiteID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid suite id", err.Error())
		return
	}

	suite, err := r.client.TestPlanClient.GetTestSuiteById(r.client.Ctx, testplan.GetTestSuiteByIdArgs{
		Project: &projectID,
		PlanId:  &planID,
		SuiteId: &suiteID,
	})
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading test suite", err.Error())
		return
	}
	if suite == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	flattenTestSuite(&model, suite)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TestSuiteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan testSuiteModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state testSuiteModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ProjectID.ValueString()
	planID, err := strconv.Atoi(plan.PlanID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid plan_id", err.Error())
		return
	}
	suiteID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid suite id", err.Error())
		return
	}

	updateParams := expandTestSuiteUpdate(&plan)
	updated, err := r.client.TestPlanClient.UpdateTestSuite(r.client.Ctx, testplan.UpdateTestSuiteArgs{
		TestSuiteUpdateParams: updateParams,
		Project:               &projectID,
		PlanId:                &planID,
		SuiteId:               &suiteID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating test suite", err.Error())
		return
	}

	plan.ID = state.ID
	flattenTestSuite(&plan, updated)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TestSuiteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model testSuiteModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	planID, err := strconv.Atoi(model.PlanID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid plan_id", err.Error())
		return
	}
	suiteID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid suite id", err.Error())
		return
	}

	err = r.client.TestPlanClient.DeleteTestSuite(r.client.Ctx, testplan.DeleteTestSuiteArgs{
		Project: &projectID,
		PlanId:  &planID,
		SuiteId: &suiteID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting test suite", err.Error())
	}
}

// ── Expand / Flatten ──────────────────────────────────────────────────────────

// expandTestSuite converts state model → API create params.
func expandTestSuite(model *testSuiteModel) *testplan.TestSuiteCreateParams {
	suiteTyp := testplan.TestSuiteType(model.SuiteType.ValueString())
	params := &testplan.TestSuiteCreateParams{
		Name:      converter.String(model.Name.ValueString()),
		SuiteType: &suiteTyp,
	}

	// ParentSuite reference
	if !model.ParentSuiteID.IsNull() && !model.ParentSuiteID.IsUnknown() {
		if pid, err := strconv.Atoi(model.ParentSuiteID.ValueString()); err == nil {
			params.ParentSuite = &testplan.TestSuiteReference{Id: &pid}
		}
	}

	// query_string only applies to dynamic suites
	if model.SuiteType.ValueString() == suiteTypeDynamic &&
		!model.QueryString.IsNull() && !model.QueryString.IsUnknown() {
		params.QueryString = converter.String(model.QueryString.ValueString())
	}

	return params
}

// expandTestSuiteUpdate converts state model → API update params.
func expandTestSuiteUpdate(model *testSuiteModel) *testplan.TestSuiteUpdateParams {
	params := &testplan.TestSuiteUpdateParams{
		Name: converter.String(model.Name.ValueString()),
	}

	// query_string only applies to dynamic suites
	if model.SuiteType.ValueString() == suiteTypeDynamic &&
		!model.QueryString.IsNull() && !model.QueryString.IsUnknown() {
		params.QueryString = converter.String(model.QueryString.ValueString())
	}

	return params
}

// flattenTestSuite maps API response → state model.
func flattenTestSuite(model *testSuiteModel, suite *testplan.TestSuite) {
	if suite.Id != nil {
		model.ID = types.StringValue(strconv.Itoa(*suite.Id))
	}
	model.Name = types.StringValue(converter.ToString(suite.Name, ""))

	if suite.SuiteType != nil {
		model.SuiteType = types.StringValue(string(*suite.SuiteType))
	}

	if suite.QueryString != nil {
		model.QueryString = types.StringValue(*suite.QueryString)
	} else {
		model.QueryString = types.StringValue("")
	}

	if suite.Plan != nil && suite.Plan.Id != nil {
		model.PlanID = types.StringValue(strconv.Itoa(*suite.Plan.Id))
	}

	if suite.ParentSuite != nil && suite.ParentSuite.Id != nil {
		model.ParentSuiteID = types.StringValue(strconv.Itoa(*suite.ParentSuite.Id))
	}
}
