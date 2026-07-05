package pipelines

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	adoPipelines "github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// ── Compile-time interface check ───────────────────────────────────────────────

var _ datasource.DataSource = (*PipelineRunDataSource)(nil)

// ── Data source struct ────────────────────────────────────────────────────────

// PipelineRunDataSource is the terraform-plugin-framework data source for betterado_pipeline_run.
type PipelineRunDataSource struct {
	client *client.AggregatedClient
}

// NewPipelineRunDataSource returns a new datasource.DataSource for betterado_pipeline_run.
func NewPipelineRunDataSource() datasource.DataSource {
	return &PipelineRunDataSource{}
}

// ── State model ───────────────────────────────────────────────────────────────

// pipelineRunDataModel is the tfsdk state model for the betterado_pipeline_run data source.
type pipelineRunDataModel struct {
	ID           types.String `tfsdk:"id"`
	ProjectID    types.String `tfsdk:"project_id"`
	PipelineID   types.Int64  `tfsdk:"pipeline_id"`
	RunID        types.Int64  `tfsdk:"run_id"`
	Name         types.String `tfsdk:"name"`
	State        types.String `tfsdk:"state"`
	Result       types.String `tfsdk:"result"`
	CreatedDate  types.String `tfsdk:"created_date"`
	FinishedDate types.String `tfsdk:"finished_date"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *PipelineRunDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_run"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *PipelineRunDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Pipelines v2 run (`_apis/pipelines/{id}/runs/{runId}`) from Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The run ID as a string (set as the data source ID).",
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the project the pipeline run belongs to.",
			},
			"pipeline_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The integer ID of the pipeline this run belongs to.",
			},
			"run_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The integer ID of the pipeline run.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name (run number) of the pipeline run.",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The state of the run. One of: `unknown`, `inProgress`, `canceling`, `completed`.",
			},
			"result": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The result of the run. One of: `unknown`, `succeeded`, `failed`, `canceled`. Empty if the run has not yet completed.",
			},
			"created_date": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The creation date/time of the run in RFC3339 format.",
			},
			"finished_date": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The finish date/time of the run in RFC3339 format. Empty if the run has not yet finished.",
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *PipelineRunDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// ── Read ──────────────────────────────────────────────────────────────────────

func (d *PipelineRunDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_pipeline_run data source Read: provider client not configured")
		return
	}

	var model pipelineRunDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project := model.ProjectID.ValueString()
	pipelineID := int(model.PipelineID.ValueInt64())
	runID := int(model.RunID.ValueInt64())

	run, err := d.client.PipelinesClient.GetRun(ctx, adoPipelines.GetRunArgs{
		Project:    &project,
		PipelineId: &pipelineID,
		RunId:      &runID,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading pipeline run",
			fmt.Sprintf("reading pipeline run (pipeline ID: %d, run ID: %d): %+v", pipelineID, runID, err),
		)
		return
	}

	flattenPipelineRunDataModel(run, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Flatten helper ────────────────────────────────────────────────────────────

func flattenPipelineRunDataModel(r *adoPipelines.Run, model *pipelineRunDataModel) {
	if r.Id != nil {
		model.ID = types.StringValue(strconv.Itoa(*r.Id))
	}
	if r.Name != nil {
		model.Name = types.StringValue(*r.Name)
	} else {
		model.Name = types.StringValue("")
	}
	if r.State != nil {
		model.State = types.StringValue(string(*r.State))
	} else {
		model.State = types.StringValue("")
	}
	if r.Result != nil {
		model.Result = types.StringValue(string(*r.Result))
	} else {
		model.Result = types.StringValue("")
	}
	if r.CreatedDate != nil && !r.CreatedDate.Time.IsZero() {
		model.CreatedDate = types.StringValue(r.CreatedDate.Time.UTC().Format(time.RFC3339))
	} else {
		model.CreatedDate = types.StringValue("")
	}
	if r.FinishedDate != nil && !r.FinishedDate.Time.IsZero() {
		model.FinishedDate = types.StringValue(r.FinishedDate.Time.UTC().Format(time.RFC3339))
	} else {
		model.FinishedDate = types.StringValue("")
	}
	// pipeline_id is already set from the config; keep it consistent with the run's pipeline reference.
	if r.Pipeline != nil && r.Pipeline.Id != nil {
		model.PipelineID = types.Int64Value(int64(*r.Pipeline.Id))
	}
}
