package pipelinesapproval

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	adoApproval "github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelinesapproval"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// ── Compile-time interface check ───────────────────────────────────────────────

var _ datasource.DataSource = (*PipelineApprovalsDataSource)(nil)

// ── Data source struct ────────────────────────────────────────────────────────

// PipelineApprovalsDataSource is the terraform-plugin-framework data source for
// betterado_pipeline_approvals.
type PipelineApprovalsDataSource struct {
	client *client.AggregatedClient
}

// NewPipelineApprovalsDataSource returns a new datasource.DataSource for
// betterado_pipeline_approvals.
func NewPipelineApprovalsDataSource() datasource.DataSource {
	return &PipelineApprovalsDataSource{}
}

// ── State models ──────────────────────────────────────────────────────────────

// pipelineApprovalsDataModel is the tfsdk state model for betterado_pipeline_approvals.
type pipelineApprovalsDataModel struct {
	ID            types.String        `tfsdk:"id"`
	ProjectID     types.String        `tfsdk:"project_id"`
	PipelineRunID types.String        `tfsdk:"pipeline_run_id"`
	Approvals     []approvalItemModel `tfsdk:"approvals"`
}

// approvalItemModel is the nested model for each approval in the list.
type approvalItemModel struct {
	ID           types.String `tfsdk:"id"`
	Status       types.String `tfsdk:"status"`
	Comment      types.String `tfsdk:"comment"`
	Instructions types.String `tfsdk:"instructions"`
	ApprovedByID types.String `tfsdk:"approved_by_id"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *PipelineApprovalsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_approvals"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *PipelineApprovalsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists pipeline approvals for a given pipeline run in Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Computed ID set to `project_id/pipeline_run_id`.",
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID or name of the Azure DevOps project.",
			},
			"pipeline_run_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Integer run ID as a string. The data source lists all pending approvals for this run.",
			},
			"approvals": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of approvals associated with the pipeline run.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The UUID of the approval.",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The current status of the approval.",
						},
						"comment": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Comment associated with the approval decision.",
						},
						"instructions": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Instructions for the approver.",
						},
						"approved_by_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Identity descriptor of the approver (if resolved).",
						},
					},
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *PipelineApprovalsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// rawApprovalListResponse is used to unmarshal the response from the raw HTTP
// approvals endpoint when filtering by runId.
type rawApprovalListResponse struct {
	Value []rawApprovalItem `json:"value"`
	Count int               `json:"count"`
}

type rawApprovalItem struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	Comment      string `json:"comment"`
	Instructions string `json:"instructions"`
	Steps        []struct {
		ActualApprover *struct {
			ID string `json:"id"`
		} `json:"actualApprover,omitempty"`
		Comment string `json:"comment,omitempty"`
		Status  string `json:"status,omitempty"`
	} `json:"steps,omitempty"`
}

func (d *PipelineApprovalsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_pipeline_approvals data source Read: provider client not configured")
		return
	}

	var model pipelineApprovalsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project := model.ProjectID.ValueString()
	runID := model.PipelineRunID.ValueString()

	// The SDK's QueryApprovals only supports filtering by explicit ApprovalIds.
	// We need to filter by runId, so we use the raw HTTP client instead.
	// Endpoint: {org}/{project}/_apis/pipelines/approvals?runId={runId}&api-version=7.1-preview.1
	impl, ok := d.client.PipelinesApprovalClient.(*adoApproval.ClientImpl)
	if !ok {
		resp.Diagnostics.AddError(
			"Client type assertion failed",
			"betterado_pipeline_approvals: cannot access underlying azuredevops.Client for raw HTTP",
		)
		return
	}

	routeValues := map[string]string{
		"project": project,
	}
	queryParams := url.Values{}
	queryParams.Set("runId", runID)
	queryParams.Set("$expand", "steps")

	// Use the same locationId as the approvals endpoint.
	// approvals location ID: 37794717-f36f-4d78-b2bf-4dc30d0cfbcd
	approvalLocationID := [16]byte{
		0x37, 0x79, 0x47, 0x17, 0xf3, 0x6f, 0x4d, 0x78,
		0xb2, 0xbf, 0x4d, 0xc3, 0x0d, 0x0c, 0xfb, 0xcd,
	}

	httpResp, err := impl.Client.Send(ctx, http.MethodGet, approvalLocationID, "7.1-preview.1", routeValues, queryParams, nil, "", "application/json", nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error listing pipeline approvals",
			fmt.Sprintf("listing approvals for project %q run %q: %v", project, runID, err),
		)
		return
	}
	defer httpResp.Body.Close()

	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		resp.Diagnostics.AddError("Error reading response body", err.Error())
		return
	}

	var result rawApprovalListResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		resp.Diagnostics.AddError("Error parsing approvals response", fmt.Sprintf("body: %s\nerror: %v", string(bodyBytes), err))
		return
	}

	approvals := make([]approvalItemModel, 0, len(result.Value))
	for _, a := range result.Value {
		item := approvalItemModel{
			ID:           types.StringValue(a.ID),
			Status:       types.StringValue(a.Status),
			Comment:      types.StringValue(a.Comment),
			Instructions: types.StringValue(a.Instructions),
			ApprovedByID: types.StringValue(""),
		}
		// Pull approved_by_id from the first step that has an actual approver.
		for _, step := range a.Steps {
			if step.ActualApprover != nil && step.ActualApprover.ID != "" {
				item.ApprovedByID = types.StringValue(step.ActualApprover.ID)
				break
			}
		}
		approvals = append(approvals, item)
	}

	model.ID = types.StringValue(project + "/" + runID)
	model.Approvals = approvals

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
