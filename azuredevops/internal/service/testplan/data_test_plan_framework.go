package testplan

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/testplan"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = (*TestPlanDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*TestPlanDataSource)(nil)
)

// TestPlanDataSource is the terraform-plugin-framework implementation of the
// betterado_test_plan data source.
type TestPlanDataSource struct {
	client *client.AggregatedClient
}

// NewTestPlanDataSource returns a new datasource.DataSource for betterado_test_plan.
func NewTestPlanDataSource() datasource.DataSource {
	return &TestPlanDataSource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *TestPlanDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "betterado_test_plan"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *TestPlanDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Required: true,
			},
			"plan_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"area_path": schema.StringAttribute{
				Computed: true,
			},
			"iteration_path": schema.StringAttribute{
				Computed: true,
			},
			"start_date": schema.StringAttribute{
				Computed: true,
			},
			"end_date": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *TestPlanDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = agg
}

// ── State model ───────────────────────────────────────────────────────────────

// testPlanDataModel is the Terraform state model for the betterado_test_plan data source.
type testPlanDataModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	PlanID        types.String `tfsdk:"plan_id"`
	Name          types.String `tfsdk:"name"`
	AreaPath      types.String `tfsdk:"area_path"`
	IterationPath types.String `tfsdk:"iteration_path"`
	StartDate     types.String `tfsdk:"start_date"`
	EndDate       types.String `tfsdk:"end_date"`
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (d *TestPlanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model testPlanDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	planID, err := strconv.Atoi(model.PlanID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid plan_id", fmt.Sprintf("plan_id must be a numeric string: %s", err.Error()))
		return
	}

	plan, err := d.client.TestPlanClient.GetTestPlanById(d.client.Ctx, testplan.GetTestPlanByIdArgs{
		Project: &projectID,
		PlanId:  &planID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading test plan", err.Error())
		return
	}
	if plan == nil {
		resp.Diagnostics.AddError("Test plan not found", fmt.Sprintf("No test plan found with id %d in project %s", planID, projectID))
		return
	}

	flattenTestPlanData(&model, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// flattenTestPlanData maps API response → data source state model.
func flattenTestPlanData(model *testPlanDataModel, plan *testplan.TestPlan) {
	if plan.Id != nil {
		model.ID = types.StringValue(strconv.Itoa(*plan.Id))
		model.PlanID = types.StringValue(strconv.Itoa(*plan.Id))
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
