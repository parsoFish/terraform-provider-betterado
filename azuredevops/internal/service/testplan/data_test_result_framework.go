package testplan

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/test"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = (*TestResultDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*TestResultDataSource)(nil)
)

// TestResultDataSource is the terraform-plugin-framework implementation of the
// betterado_test_result data source.
type TestResultDataSource struct {
	client *client.AggregatedClient
}

// NewTestResultDataSource returns a new datasource.DataSource for betterado_test_result.
func NewTestResultDataSource() datasource.DataSource {
	return &TestResultDataSource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *TestResultDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "betterado_test_result"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *TestResultDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
			"result_id": schema.StringAttribute{
				Required: true,
			},
			"test_case_title": schema.StringAttribute{
				Computed: true,
			},
			"outcome": schema.StringAttribute{
				Computed: true,
			},
			"duration_in_ms": schema.Int64Attribute{
				Computed: true,
			},
			"failure_type": schema.StringAttribute{
				Computed: true,
			},
			"error_message": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *TestResultDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// testResultDataModel is the Terraform state model for the betterado_test_result data source.
type testResultDataModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	RunID         types.String `tfsdk:"run_id"`
	ResultID      types.String `tfsdk:"result_id"`
	TestCaseTitle types.String `tfsdk:"test_case_title"`
	Outcome       types.String `tfsdk:"outcome"`
	DurationInMs  types.Int64  `tfsdk:"duration_in_ms"`
	FailureType   types.String `tfsdk:"failure_type"`
	ErrorMessage  types.String `tfsdk:"error_message"`
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (d *TestResultDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model testResultDataModel
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
	resultID, err := strconv.Atoi(model.ResultID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid result_id", fmt.Sprintf("result_id must be a numeric string: %s", err.Error()))
		return
	}

	result, err := d.client.TestClient.GetTestResultById(d.client.Ctx, test.GetTestResultByIdArgs{
		Project:          &projectID,
		RunId:            &runID,
		TestCaseResultId: &resultID,
	})
	if err != nil {
		if isNotFound(err) {
			resp.Diagnostics.AddError(
				"Test result not found",
				fmt.Sprintf("No test result found with id %d in run %d, project %s", resultID, runID, projectID),
			)
			return
		}
		resp.Diagnostics.AddError("Error reading test result", err.Error())
		return
	}
	if result == nil {
		resp.Diagnostics.AddError("Test result not found", fmt.Sprintf("No test result found with id %d in run %d, project %s", resultID, runID, projectID))
		return
	}

	flattenTestResultData(&model, result)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// flattenTestResultData maps API response → data source state model.
func flattenTestResultData(model *testResultDataModel, result *test.TestCaseResult) {
	if result.Id != nil {
		model.ID = types.StringValue(strconv.Itoa(*result.Id))
		model.ResultID = types.StringValue(strconv.Itoa(*result.Id))
	}
	if result.TestCaseTitle != nil {
		model.TestCaseTitle = types.StringValue(*result.TestCaseTitle)
	} else {
		model.TestCaseTitle = types.StringValue("")
	}
	if result.Outcome != nil {
		model.Outcome = types.StringValue(*result.Outcome)
	} else {
		model.Outcome = types.StringValue("")
	}
	if result.DurationInMs != nil {
		model.DurationInMs = types.Int64Value(int64(*result.DurationInMs))
	} else {
		model.DurationInMs = types.Int64Value(0)
	}
	if result.FailureType != nil {
		model.FailureType = types.StringValue(*result.FailureType)
	} else {
		model.FailureType = types.StringValue("")
	}
	if result.ErrorMessage != nil {
		model.ErrorMessage = types.StringValue(*result.ErrorMessage)
	} else {
		model.ErrorMessage = types.StringValue("")
	}
}
