package taskagent

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var (
	_ datasource.DataSource              = (*AgentQueueDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*AgentQueueDataSource)(nil)
)

// AgentQueueDataSource is the terraform-plugin-framework data source for betterado_agent_queue.
type AgentQueueDataSource struct {
	client *client.AggregatedClient
}

// NewAgentQueueDataSource returns a new framework data source for betterado_agent_queue.
func NewAgentQueueDataSource() datasource.DataSource {
	return &AgentQueueDataSource{}
}

// agentQueueDataModel is the tfsdk model for the data source.
type agentQueueDataModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	AgentPoolID types.Int64  `tfsdk:"agent_pool_id"`
}

func (d *AgentQueueDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_queue"
}

func (d *AgentQueueDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to access information about an existing Azure DevOps Agent Queue.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the agent queue.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the agent queue to look up.",
			},
			"agent_pool_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the underlying agent pool.",
			},
		},
	}
}

func (d *AgentQueueDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AgentQueueDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model agentQueueDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := model.Name.ValueString()
	projectID := model.ProjectID.ValueString()

	agentQueues, err := d.client.TaskAgentClient.GetAgentQueues(d.client.Ctx, taskagent.GetAgentQueuesArgs{
		Project:   &projectID,
		QueueName: &name,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading agent queues", err.Error())
		return
	}

	if agentQueues == nil || len(*agentQueues) == 0 {
		resp.Diagnostics.AddError("Agent queue not found",
			fmt.Sprintf("Unable to find agent queue with name: %s in project: %s", name, projectID))
		return
	}

	if len(*agentQueues) > 1 {
		resp.Diagnostics.AddError("Multiple agent queues found",
			fmt.Sprintf("Found multiple agent queues for name: %s in project: %s", name, projectID))
		return
	}

	queue := (*agentQueues)[0]
	model.ID = types.StringValue(strconv.Itoa(*queue.Id))
	if queue.Name != nil {
		model.Name = types.StringValue(*queue.Name)
	}
	if queue.Pool != nil && queue.Pool.Id != nil {
		model.AgentPoolID = types.Int64Value(int64(*queue.Pool.Id))
	}
	if queue.ProjectId != nil {
		model.ProjectID = types.StringValue(queue.ProjectId.String())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
