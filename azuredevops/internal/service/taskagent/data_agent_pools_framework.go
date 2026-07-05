package taskagent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var (
	_ datasource.DataSource              = (*AgentPoolsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*AgentPoolsDataSource)(nil)
)

// AgentPoolsDataSource is the terraform-plugin-framework data source for
// betterado_agent_pools.
type AgentPoolsDataSource struct {
	client *client.AggregatedClient
}

// NewAgentPoolsDataSource returns a new framework data source for betterado_agent_pools.
func NewAgentPoolsDataSource() datasource.DataSource {
	return &AgentPoolsDataSource{}
}

// agentPoolsDataModel is the tfsdk model for the data source.
type agentPoolsDataModel struct {
	ID         types.String         `tfsdk:"id"`
	AgentPools []agentPoolItemModel `tfsdk:"agent_pools"`
}

// agentPoolItemModel is the model for each agent pool entry in the list.
type agentPoolItemModel struct {
	ID            types.Int64  `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	PoolType      types.String `tfsdk:"pool_type"`
	AutoProvision types.Bool   `tfsdk:"auto_provision"`
	AutoUpdate    types.Bool   `tfsdk:"auto_update"`
}

func (d *AgentPoolsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent_pools"
}

func (d *AgentPoolsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to access information about all Azure DevOps Agent Pools.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the data source (timestamp).",
			},
			"agent_pools": schema.ListNestedAttribute{
				Computed:    true,
				Description: "A list of existing agent pools.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:    true,
							Description: "The ID of the agent pool.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
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
				},
			},
		},
	}
}

func (d *AgentPoolsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AgentPoolsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model agentPoolsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	agentPools, err := d.client.TaskAgentClient.GetAgentPools(d.client.Ctx, taskagent.GetAgentPoolsArgs{})
	if err != nil {
		resp.Diagnostics.AddError("Error reading agent pools",
			fmt.Sprintf("finding agent pools. Error: %v", err))
		return
	}
	log.Printf("[TRACE] plugin.terraform-provider-betterado: Read [%d] agent pools from current organization", len(*agentPools))

	model.ID = types.StringValue(time.Now().UTC().String())
	model.AgentPools = flattenAgentPoolReferencesFramework(agentPools)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func flattenAgentPoolReferencesFramework(input *[]taskagent.TaskAgentPool) []agentPoolItemModel {
	if input == nil {
		return []agentPoolItemModel{}
	}

	results := make([]agentPoolItemModel, 0, len(*input))
	for _, element := range *input {
		item := agentPoolItemModel{}

		if element.Id != nil {
			item.ID = types.Int64Value(int64(*element.Id))
		}
		if element.Name != nil {
			item.Name = types.StringValue(*element.Name)
		}
		if element.PoolType != nil {
			item.PoolType = types.StringValue(string(*element.PoolType))
		}
		if element.AutoProvision != nil {
			item.AutoProvision = types.BoolValue(*element.AutoProvision)
		}
		if element.AutoUpdate != nil {
			item.AutoUpdate = types.BoolValue(*element.AutoUpdate)
		}

		results = append(results, item)
	}
	return results
}
