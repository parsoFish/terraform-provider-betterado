package pipelines

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	adoPipelines "github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// ── Compile-time interface check ───────────────────────────────────────────────

var _ datasource.DataSource = (*PipelineDataSource)(nil)

// ── Data source struct ────────────────────────────────────────────────────────

// PipelineDataSource is the terraform-plugin-framework data source for betterado_pipeline.
type PipelineDataSource struct {
	client *client.AggregatedClient
}

// NewPipelineDataSource returns a new datasource.DataSource for betterado_pipeline.
func NewPipelineDataSource() datasource.DataSource {
	return &PipelineDataSource{}
}

// ── State model ───────────────────────────────────────────────────────────────

// pipelineDataModel is the tfsdk state model for the betterado_pipeline data source.
type pipelineDataModel struct {
	ID                types.String `tfsdk:"id"`
	ProjectID         types.String `tfsdk:"project_id"`
	PipelineID        types.Int64  `tfsdk:"pipeline_id"`
	Name              types.String `tfsdk:"name"`
	Folder            types.String `tfsdk:"folder"`
	ConfigurationType types.String `tfsdk:"configuration_type"`
	Revision          types.Int64  `tfsdk:"revision"`
	URL               types.String `tfsdk:"url"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *PipelineDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *PipelineDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads a Pipelines v2 pipeline (`_apis/pipelines`) from Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The pipeline integer ID as a string (set as the data source ID).",
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID of the project the pipeline belongs to.",
			},
			"pipeline_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The integer ID of the pipeline.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the pipeline.",
			},
			"folder": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The folder path of the pipeline.",
			},
			"configuration_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The configuration type of the pipeline.",
			},
			"revision": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The pipeline revision number.",
			},
			"url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URL of the pipeline.",
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *PipelineDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PipelineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_pipeline data source Read: provider client not configured")
		return
	}

	var model pipelineDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project := model.ProjectID.ValueString()
	pipelineID := int(model.PipelineID.ValueInt64())

	pipeline, err := d.client.PipelinesClient.GetPipeline(ctx, adoPipelines.GetPipelineArgs{
		Project:    &project,
		PipelineId: &pipelineID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			model.ID = types.StringValue("")
			resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
			return
		}
		resp.Diagnostics.AddError("Error reading pipeline", fmt.Sprintf("reading pipeline (ID: %d): %+v", pipelineID, err))
		return
	}

	flattenPipelineDataModel(pipeline, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Flatten helper ────────────────────────────────────────────────────────────

func flattenPipelineDataModel(p *adoPipelines.Pipeline, model *pipelineDataModel) {
	if p.Id != nil {
		model.ID = types.StringValue(strconv.Itoa(*p.Id))
		model.PipelineID = types.Int64Value(int64(*p.Id))
	}
	if p.Name != nil {
		model.Name = types.StringValue(*p.Name)
	} else {
		model.Name = types.StringValue("")
	}
	if p.Folder != nil {
		model.Folder = types.StringValue(*p.Folder)
	} else {
		model.Folder = types.StringValue("")
	}
	if p.Revision != nil {
		model.Revision = types.Int64Value(int64(*p.Revision))
	}
	if p.Url != nil {
		model.URL = types.StringValue(*p.Url)
	} else {
		model.URL = types.StringValue("")
	}
	if p.Configuration != nil && p.Configuration.Type != nil {
		model.ConfigurationType = types.StringValue(string(*p.Configuration.Type))
	} else {
		model.ConfigurationType = types.StringValue("")
	}
}
