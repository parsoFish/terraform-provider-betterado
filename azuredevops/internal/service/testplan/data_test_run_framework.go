package testplan

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/test"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = (*TestRunDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*TestRunDataSource)(nil)
)

// TestRunDataSource is the terraform-plugin-framework implementation of the
// betterado_test_run data source.
type TestRunDataSource struct {
	client *client.AggregatedClient
}

// NewTestRunDataSource returns a new datasource.DataSource for betterado_test_run.
func NewTestRunDataSource() datasource.DataSource {
	return &TestRunDataSource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *TestRunDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "betterado_test_run"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *TestRunDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Required: true,
			},
			"run_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"total_tests": schema.Int64Attribute{
				Computed: true,
			},
			"passed_tests": schema.Int64Attribute{
				Computed: true,
			},
			"failed_tests": schema.Int64Attribute{
				Computed: true,
			},
			"started_date": schema.StringAttribute{
				Computed: true,
			},
			"completed_date": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *TestRunDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// testRunDataModel is the Terraform state model for the betterado_test_run data source.
type testRunDataModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	RunID         types.String `tfsdk:"run_id"`
	Name          types.String `tfsdk:"name"`
	State         types.String `tfsdk:"state"`
	TotalTests    types.Int64  `tfsdk:"total_tests"`
	PassedTests   types.Int64  `tfsdk:"passed_tests"`
	FailedTests   types.Int64  `tfsdk:"failed_tests"`
	StartedDate   types.String `tfsdk:"started_date"`
	CompletedDate types.String `tfsdk:"completed_date"`
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (d *TestRunDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model testRunDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	runID, err := strconv.Atoi(model.RunID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid run_id", fmt.Sprintf("run_id must be a numeric string: %s", err.Error()))
		return
	}

	run, err := d.client.TestClient.GetTestRunById(d.client.Ctx, test.GetTestRunByIdArgs{
		Project: &projectID,
		RunId:   &runID,
	})
	if err != nil {
		if isNotFound(err) {
			resp.Diagnostics.AddError(
				"Test run not found",
				fmt.Sprintf("No test run found with id %d in project %s", runID, projectID),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading test run", err.Error())
		return
	}
	if run == nil {
		resp.Diagnostics.AddError("Test run not found", fmt.Sprintf("No test run found with id %d in project %s", runID, projectID))
		return
	}

	flattenTestRunData(&model, run)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// flattenTestRunData maps API response → data source state model.
func flattenTestRunData(model *testRunDataModel, run *test.TestRun) {
	if run.Id != nil {
		model.ID = types.StringValue(strconv.Itoa(*run.Id))
		model.RunID = types.StringValue(strconv.Itoa(*run.Id))
	}
	if run.Name != nil {
		model.Name = types.StringValue(*run.Name)
	} else {
		model.Name = types.StringValue("")
	}
	if run.State != nil {
		model.State = types.StringValue(*run.State)
	} else {
		model.State = types.StringValue("")
	}
	if run.TotalTests != nil {
		model.TotalTests = types.Int64Value(int64(*run.TotalTests))
	} else {
		model.TotalTests = types.Int64Value(0)
	}
	if run.PassedTests != nil {
		model.PassedTests = types.Int64Value(int64(*run.PassedTests))
	} else {
		model.PassedTests = types.Int64Value(0)
	}
	// FailedTests is UnanalyzedTests in the API (failed = unanalyzed / failed tests count).
	// The API does not have a direct FailedTests field; use UnanalyzedTests as the closest proxy.
	if run.UnanalyzedTests != nil {
		model.FailedTests = types.Int64Value(int64(*run.UnanalyzedTests))
	} else {
		model.FailedTests = types.Int64Value(0)
	}
	if run.StartedDate != nil && !run.StartedDate.Time.IsZero() {
		model.StartedDate = types.StringValue(run.StartedDate.Time.UTC().Format(time.RFC3339))
	} else {
		model.StartedDate = types.StringValue("")
	}
	if run.CompletedDate != nil && !run.CompletedDate.Time.IsZero() {
		model.CompletedDate = types.StringValue(run.CompletedDate.Time.UTC().Format(time.RFC3339))
	} else {
		model.CompletedDate = types.StringValue("")
	}
}
