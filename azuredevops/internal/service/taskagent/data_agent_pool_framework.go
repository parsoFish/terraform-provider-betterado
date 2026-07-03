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
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Ensure interface compliance.
var (
	_ datasource.DataSource              = (*AgentPoolDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*AgentPoolDataSource)(nil)
)

// AgentPoolDataSource is the terraform-plugin-framework data source for
// betterado_agent_pool.
type AgentPoolDataSource struct {
	client *client.AggregatedClient
}

// NewAgentPoolDataSource returns a new framework data source for betterado_agent_pool.
func NewAgentPoolDataSource() datasource.DataSource {
	return &AgentPoolDataSource{}
}

// agentPoolDataModel is the tfsdk model for the data source.
type agentPoolDataModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	PoolType      types.String `tfsdk:"pool_type"`
	AutoProvision types.Bool   `tfsdk:"auto_provision"`
	AutoUpdate    types.Bool   `tfsdk:"auto_update"`
}

func (d *AgentPoolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_pool"
}

func (d *AgentPoolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to access information about an existing Azure DevOps Agent Pool.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the agent pool.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the agent pool.",
			},
			"pool_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the agent pool (Automation or Deployment).",
			},
			"auto_provision": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the agent pool is automatically provisioned.",
			},
			"auto_update": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether agents in the pool are automatically updated.",
			},
		},
	}
}

func (d *AgentPoolDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AgentPoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model agentPoolDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	poolName := model.Name.ValueString()
	agentPools, err := d.client.TaskAgentClient.GetAgentPools(d.client.Ctx, taskagent.GetAgentPoolsArgs{
		PoolName: converter.String(poolName),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading agent pools", err.Error())
		return
	}

	if agentPools == nil || len(*agentPools) == 0 {
		resp.Diagnostics.AddError("Agent pool not found",
			fmt.Sprintf("Unable to find agent pool with name: %s", poolName))
		return
	}

	if len(*agentPools) > 1 {
		resp.Diagnostics.AddError("Multiple agent pools found",
			fmt.Sprintf("Found multiple agent pools for name: %s. Agent pools found: %+v", poolName, agentPools))
		return
	}

	pool := (*agentPools)[0]
	model.ID = types.StringValue(strconv.Itoa(*pool.Id))
	model.Name = types.StringValue(converter.ToString(pool.Name, ""))
	if pool.PoolType != nil {
		model.PoolType = types.StringValue(string(*pool.PoolType))
	}
	if pool.AutoProvision != nil {
		model.AutoProvision = types.BoolValue(*pool.AutoProvision)
	}
	if pool.AutoUpdate != nil {
		model.AutoUpdate = types.BoolValue(*pool.AutoUpdate)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
