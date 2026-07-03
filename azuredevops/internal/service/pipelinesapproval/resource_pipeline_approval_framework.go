package pipelinesapproval

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	adoApproval "github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelinesapproval"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// ── Compile-time interface checks ─────────────────────────────────────────────

var (
	_ resource.Resource              = (*PipelineApprovalResource)(nil)
	_ resource.ResourceWithConfigure = (*PipelineApprovalResource)(nil)
)

// ── Resource struct ───────────────────────────────────────────────────────────

// PipelineApprovalResource is the terraform-plugin-framework implementation of
// betterado_pipeline_approval.
type PipelineApprovalResource struct {
	client *client.AggregatedClient
}

// NewPipelineApprovalResource returns a new resource.Resource for
// betterado_pipeline_approval.
func NewPipelineApprovalResource() resource.Resource {
	return &PipelineApprovalResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *PipelineApprovalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pipeline_approval"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *PipelineApprovalResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Records an approval decision for a pending Azure DevOps pipeline run approval. " +
			"Because approval decisions cannot be revoked via the API, deleting this resource is a no-op " +
			"(the resource is simply removed from Terraform state).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The approval ID (same as `approval_id`), stored as the resource identifier.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"project_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The ID or name of the Azure DevOps project. Changing this forces a new resource.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"approval_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "UUID of the pending approval to resolve. Changing this forces a new resource.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"status": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The decision to record. Must be `approved` or `rejected`.",
				Validators: []validator.String{
					stringvalidator.OneOf("approved", "rejected"),
				},
			},
			"comment": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Comment recorded with the approval decision.",
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *PipelineApprovalResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// pipelineApprovalModel is the Terraform state model for betterado_pipeline_approval.
type pipelineApprovalModel struct {
	ID         types.String `tfsdk:"id"`
	ProjectID  types.String `tfsdk:"project_id"`
	ApprovalID types.String `tfsdk:"approval_id"`
	Status     types.String `tfsdk:"status"`
	Comment    types.String `tfsdk:"comment"`
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

// Create records the approval decision via UpdateApprovals and sets the resource ID.
func (r *PipelineApprovalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model pipelineApprovalModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project := model.ProjectID.ValueString()
	approvalIDStr := model.ApprovalID.ValueString()

	approvalUUID, err := uuid.Parse(approvalIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid approval_id", fmt.Sprintf("approval_id must be a valid UUID: %s", err))
		return
	}

	statusEnum := stringToApprovalStatus(model.Status.ValueString())
	params := adoApproval.ApprovalUpdateParameters{
		ApprovalId: &approvalUUID,
		Status:     &statusEnum,
	}
	if !model.Comment.IsNull() && !model.Comment.IsUnknown() {
		comment := model.Comment.ValueString()
		params.Comment = &comment
	}

	_, err = r.client.PipelinesApprovalClient.UpdateApprovals(ctx, adoApproval.UpdateApprovalsArgs{
		Project:          &project,
		UpdateParameters: &[]adoApproval.ApprovalUpdateParameters{params},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error recording approval decision", err.Error())
		return
	}

	// Set the resource ID to approval_id, then refresh state via Read.
	model.ID = model.ApprovalID
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Refresh computed state from the API.
	readReq := resource.ReadRequest{State: resp.State}
	readResp := &resource.ReadResponse{State: resp.State}
	r.Read(ctx, readReq, readResp)
	resp.Diagnostics.Append(readResp.Diagnostics...)
	resp.State = readResp.State
}

// Read refreshes status and comment from the API.
func (r *PipelineApprovalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model pipelineApprovalModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project := model.ProjectID.ValueString()
	approvalIDStr := model.ApprovalID.ValueString()

	approvalUUID, err := uuid.Parse(approvalIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid approval_id", fmt.Sprintf("approval_id must be a valid UUID: %s", err))
		return
	}

	approval, err := r.client.PipelinesApprovalClient.GetApproval(ctx, adoApproval.GetApprovalArgs{
		Project:    &project,
		ApprovalId: &approvalUUID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading approval", err.Error())
		return
	}

	flattenApproval(approval, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update is intentionally not implemented — both key attributes have RequiresReplace,
// so any change forces a destroy+create cycle.
func (r *PipelineApprovalResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"betterado_pipeline_approval does not support in-place update. "+
			"Change approval_id or status to force replacement.",
	)
}

// Delete is a no-op. Approval decisions cannot be revoked via the API.
// The resource is simply removed from Terraform state.
func (r *PipelineApprovalResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No-op: approval decisions are immutable in the ADO API.
}

// ── Expand / Flatten helpers ──────────────────────────────────────────────────

// flattenApproval writes the API response into the Terraform state model.
func flattenApproval(a *adoApproval.Approval, model *pipelineApprovalModel) {
	if a.Status != nil {
		model.Status = types.StringValue(string(*a.Status))
	}
	// Approval.Steps carries per-step comments; the overall approval comment is
	// surfaced on the first step where actedOn matches our status.
	if a.Steps != nil && len(*a.Steps) > 0 {
		for _, step := range *a.Steps {
			if step.Comment != nil && *step.Comment != "" {
				model.Comment = types.StringValue(*step.Comment)
				break
			}
		}
	}
}

// stringToApprovalStatus converts the Terraform string value to the SDK enum constant.
func stringToApprovalStatus(s string) adoApproval.ApprovalStatus {
	switch s {
	case "approved":
		return adoApproval.ApprovalStatusValues.Approved
	case "rejected":
		return adoApproval.ApprovalStatusValues.Rejected
	default:
		return adoApproval.ApprovalStatusValues.Pending
	}
}
