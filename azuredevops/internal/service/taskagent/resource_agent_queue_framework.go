package taskagent

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = (*AgentQueueResource)(nil)
	_ resource.ResourceWithConfigure   = (*AgentQueueResource)(nil)
	_ resource.ResourceWithImportState = (*AgentQueueResource)(nil)
)

// ── Int64 plan modifier helpers ───────────────────────────────────────────────

// requiresReplaceInt64Modifier forces replacement when an Int64 attribute changes.
type requiresReplaceInt64Modifier struct{}

func requiresReplaceInt64() planmodifier.Int64 { return requiresReplaceInt64Modifier{} }

func (requiresReplaceInt64Modifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (requiresReplaceInt64Modifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (requiresReplaceInt64Modifier) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// useStateForUnknownInt64QueueModifier copies prior state value for computed Int64.
// Named with suffix to avoid conflict with release package's useStateForUnknownInt64.
type useStateForUnknownInt64QueueModifier struct{}

func useStateForUnknownInt64Queue() planmodifier.Int64 { return useStateForUnknownInt64QueueModifier{} }

func (useStateForUnknownInt64QueueModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (useStateForUnknownInt64QueueModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (useStateForUnknownInt64QueueModifier) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ── Resource ──────────────────────────────────────────────────────────────────

// AgentQueueResource is the terraform-plugin-framework implementation of betterado_agent_queue.
type AgentQueueResource struct {
	client *client.AggregatedClient
}

// NewAgentQueueResource returns a new resource.Resource for betterado_agent_queue.
func NewAgentQueueResource() resource.Resource {
	return &AgentQueueResource{}
}

// agentQueueModel is the Terraform state model for betterado_agent_queue.
type agentQueueModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	AgentPoolID types.Int64  `tfsdk:"agent_pool_id"`
}

func (r *AgentQueueResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_agent_queue"
}

func (r *AgentQueueResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Azure DevOps Agent Queue.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					useStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the agent queue. Populated from the pool when agent_pool_id is used.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
					useStateForUnknown(),
				},
			},
			"agent_pool_id": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The ID of the agent pool to use.",
				PlanModifiers: []planmodifier.Int64{
					requiresReplaceInt64(),
					useStateForUnknownInt64Queue(),
				},
			},
		},
	}
}

func (r *AgentQueueResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AgentQueueResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model agentQueueModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	queue := &taskagent.TaskAgentQueue{}

	if !model.AgentPoolID.IsNull() && !model.AgentPoolID.IsUnknown() {
		poolID := int(model.AgentPoolID.ValueInt64())
		queue.Pool = &taskagent.TaskAgentPoolReference{
			Id: &poolID,
		}
		// Resolve pool name from pool ID
		referencedPool, err := r.client.TaskAgentClient.GetAgentPool(r.client.Ctx, taskagent.GetAgentPoolArgs{
			PoolId: &poolID,
		})
		if err != nil {
			resp.Diagnostics.AddError("Error looking up agent pool",
				fmt.Sprintf("looking up referenced agent pool: %+v", err))
			return
		}
		queue.Name = referencedPool.Name
	} else {
		queue.Name = converter.String(model.Name.ValueString())
	}

	createdQueue, err := r.client.TaskAgentClient.AddAgentQueue(r.client.Ctx, taskagent.AddAgentQueueArgs{
		Queue:              queue,
		Project:            &projectID,
		AuthorizePipelines: converter.Bool(false),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating agent queue",
			fmt.Sprintf("creating agent queue in Azure DevOps: %+v", err))
		return
	}

	model.ID = types.StringValue(strconv.Itoa(*createdQueue.Id))

	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AgentQueueResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model agentQueueModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// Update is a no-op because all attributes use RequiresReplace.
func (r *AgentQueueResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model agentQueueModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AgentQueueResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model agentQueueModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	queueID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid agent queue ID",
			fmt.Sprintf("parse agent queue ID: %+v", err))
		return
	}
	projectID := model.ProjectID.ValueString()

	if err := r.client.TaskAgentClient.DeleteAgentQueue(r.client.Ctx, taskagent.DeleteAgentQueueArgs{
		QueueId: &queueID,
		Project: &projectID,
	}); err != nil {
		resp.Diagnostics.AddError("Error deleting agent queue", err.Error())
	}
}

func (r *AgentQueueResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: <project_id>/<queue_id>
	projectID, queueIDStr, err := splitProjectQualifiedID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("expected <project_id>/<queue_id>, got %q: %v", req.ID, err))
		return
	}
	model := agentQueueModel{
		ID:        types.StringValue(queueIDStr),
		ProjectID: types.StringValue(projectID),
	}
	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// readIntoModel fetches the agent queue by the ID and project_id in model and populates model fields.
// If the queue is not found (404), model.ID is set to null.
func (r *AgentQueueResource) readIntoModel(_ context.Context, model *agentQueueModel) diag.Diagnostics {
	var diags diag.Diagnostics

	queueID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid agent queue ID", fmt.Sprintf("parse agent queue ID: %+v", err))
		return diags
	}
	projectID := model.ProjectID.ValueString()

	queue, err := r.client.TaskAgentClient.GetAgentQueue(r.client.Ctx, taskagent.GetAgentQueueArgs{
		QueueId: &queueID,
		Project: &projectID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			model.ID = types.StringNull()
			return diags
		}
		diags.AddError("Error reading agent queue",
			fmt.Sprintf("looking up Agent Queue with ID %d in project %s. Error: %v", queueID, projectID, err))
		return diags
	}

	if queue.Name != nil {
		model.Name = types.StringValue(*queue.Name)
	}
	if queue.Pool != nil && queue.Pool.Id != nil {
		model.AgentPoolID = types.Int64Value(int64(*queue.Pool.Id))
	}
	return diags
}

// splitProjectQualifiedID splits "projectID/resourceID" into its two parts.
func splitProjectQualifiedID(id string) (string, string, error) {
	for i := len(id) - 1; i >= 0; i-- {
		if id[i] == '/' {
			return id[:i], id[i+1:], nil
		}
	}
	return "", "", fmt.Errorf("no '/' separator found in %q", id)
}
