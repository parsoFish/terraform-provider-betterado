package taskagent

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = (*AgentPoolResource)(nil)
	_ resource.ResourceWithConfigure   = (*AgentPoolResource)(nil)
	_ resource.ResourceWithImportState = (*AgentPoolResource)(nil)
)

// AgentPoolResource is the terraform-plugin-framework implementation of betterado_agent_pool.
type AgentPoolResource struct {
	client *client.AggregatedClient
}

// NewAgentPoolResource returns a new resource.Resource for betterado_agent_pool.
func NewAgentPoolResource() resource.Resource {
	return &AgentPoolResource{}
}

// agentPoolModel is the Terraform state model for betterado_agent_pool.
type agentPoolModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	PoolType      types.String `tfsdk:"pool_type"`
	AutoProvision types.Bool   `tfsdk:"auto_provision"`
	AutoUpdate    types.Bool   `tfsdk:"auto_update"`
}

func (r *AgentPoolResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_agent_pool"
}

func (r *AgentPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Azure DevOps Agent Pool.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					useStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the agent pool.",
			},
			"pool_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     defaultString(string(taskagent.TaskAgentPoolTypeValues.Automation)),
				Description: "Specifies whether the agent pool type is Automation or Deployment.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"auto_provision": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     defaultBool(false),
				Description: "Setting this value to true will automatically provision Azure Pipelines agent pools.",
			},
			"auto_update": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     defaultBool(true),
				Description: "Setting this value to true will automatically update Azure Pipelines agents.",
			},
		},
	}
}

func (r *AgentPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AgentPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model agentPoolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	args := taskagent.AddAgentPoolArgs{
		Pool: &taskagent.TaskAgentPool{
			Name:          converter.String(model.Name.ValueString()),
			PoolType:      converter.ToPtr(taskagent.TaskAgentPoolType(model.PoolType.ValueString())),
			AutoProvision: converter.Bool(model.AutoProvision.ValueBool()),
			AutoUpdate:    converter.Bool(model.AutoUpdate.ValueBool()),
		},
	}

	agentPool, err := r.client.TaskAgentClient.AddAgentPool(r.client.Ctx, args)
	if err != nil {
		resp.Diagnostics.AddError("Error creating agent pool",
			fmt.Sprintf("creating agent pool in Azure DevOps: %+v", err))
		return
	}

	// auto_update can only be set to false after creation
	if args.Pool.AutoUpdate != nil && !*args.Pool.AutoUpdate {
		updateArgs := taskagent.UpdateAgentPoolArgs{
			PoolId: agentPool.Id,
			Pool: &taskagent.TaskAgentPool{
				Name:          args.Pool.Name,
				PoolType:      args.Pool.PoolType,
				AutoProvision: args.Pool.AutoProvision,
				AutoUpdate:    args.Pool.AutoUpdate,
			},
		}
		agentPool, err = r.client.TaskAgentClient.UpdateAgentPool(r.client.Ctx, updateArgs)
		if err != nil {
			resp.Diagnostics.AddError("Error updating agent pool after creation",
				fmt.Sprintf("updating agent pool in Azure DevOps: %+v", err))
			return
		}
		if err := r.syncPoolStatus(updateArgs); err != nil {
			resp.Diagnostics.AddError("Error syncing agent pool status", err.Error())
			return
		}
	}

	model.ID = types.StringValue(strconv.Itoa(*agentPool.Id))

	resp.Diagnostics.Append(r.readIntoModel(&model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AgentPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model agentPoolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.readIntoModel(&model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AgentPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan agentPoolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state agentPoolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	poolID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid agent pool ID",
			fmt.Sprintf("parse agent pool ID: %+v", err))
		return
	}

	updateArgs := taskagent.UpdateAgentPoolArgs{
		PoolId: &poolID,
		Pool: &taskagent.TaskAgentPool{
			Name:          converter.String(plan.Name.ValueString()),
			PoolType:      converter.ToPtr(taskagent.TaskAgentPoolType(plan.PoolType.ValueString())),
			AutoProvision: converter.Bool(plan.AutoProvision.ValueBool()),
			AutoUpdate:    converter.Bool(plan.AutoUpdate.ValueBool()),
		},
	}

	if _, err = r.client.TaskAgentClient.UpdateAgentPool(r.client.Ctx, updateArgs); err != nil {
		resp.Diagnostics.AddError("Error updating agent pool",
			fmt.Sprintf("updating agent pool in Azure DevOps: %+v", err))
		return
	}

	if err := r.syncPoolStatus(updateArgs); err != nil {
		resp.Diagnostics.AddError("Error syncing agent pool status", err.Error())
		return
	}

	resp.Diagnostics.Append(r.readIntoModel(&plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AgentPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model agentPoolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	poolID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid agent pool ID",
			fmt.Sprintf("parse agent pool ID: %+v", err))
		return
	}

	if err := r.client.TaskAgentClient.DeleteAgentPool(r.client.Ctx, taskagent.DeleteAgentPoolArgs{PoolId: &poolID}); err != nil {
		resp.Diagnostics.AddError("Error deleting agent pool", err.Error())
		return
	}

	// Wait for pool to be fully deleted
	stateConf := &retry.StateChangeConf{
		Pending: []string{"Waiting"},
		Target:  []string{"Synched"},
		Refresh: func() (interface{}, string, error) {
			s := "Waiting"
			ap, err := r.client.TaskAgentClient.GetAgentPool(r.client.Ctx, taskagent.GetAgentPoolArgs{PoolId: &poolID})
			if err != nil {
				if utils.ResponseWasNotFound(err) {
					s = "Synched"
				} else {
					return nil, "", fmt.Errorf("looking up Agent Pool with ID: %+v", err)
				}
			}
			if ap == nil {
				s = "Synched"
			}
			return s, s, nil
		},
		Timeout:                   5 * time.Minute,
		MinTimeout:                3 * time.Second,
		Delay:                     1 * time.Second,
		ContinuousTargetOccurence: 2,
	}
	if _, err := stateConf.WaitForStateContext(r.client.Ctx); err != nil {
		resp.Diagnostics.AddError("Error waiting for agent pool deletion",
			fmt.Sprintf("waiting for Agent Pool deleted. %v", err))
	}
}

func (r *AgentPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	model := agentPoolModel{
		ID: types.StringValue(req.ID),
	}
	resp.Diagnostics.Append(r.readIntoModel(&model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// readIntoModel fetches the agent pool by the ID in model and populates model fields.
// If the pool is not found (404), model.ID is set to null.
func (r *AgentPoolResource) readIntoModel(model *agentPoolModel) diag.Diagnostics {
	var diags diag.Diagnostics

	poolID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid agent pool ID", fmt.Sprintf("parse agent pool ID: %+v", err))
		return diags
	}

	agentPool, err := r.client.TaskAgentClient.GetAgentPool(r.client.Ctx, taskagent.GetAgentPoolArgs{PoolId: &poolID})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			model.ID = types.StringNull()
			return diags
		}
		diags.AddError("Error reading agent pool",
			fmt.Sprintf("looking up Agent Pool with ID %d. Error: %v", poolID, err))
		return diags
	}

	model.Name = types.StringValue(converter.ToString(agentPool.Name, ""))
	if agentPool.PoolType != nil {
		model.PoolType = types.StringValue(string(*agentPool.PoolType))
	}
	if agentPool.AutoProvision != nil {
		model.AutoProvision = types.BoolValue(*agentPool.AutoProvision)
	}
	if agentPool.AutoUpdate != nil {
		model.AutoUpdate = types.BoolValue(*agentPool.AutoUpdate)
	}
	return diags
}

// syncPoolStatus waits for the agent pool to match the desired state.
func (r *AgentPoolResource) syncPoolStatus(params taskagent.UpdateAgentPoolArgs) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"Waiting"},
		Target:  []string{"Synched"},
		Refresh: func() (interface{}, string, error) {
			s := "Waiting"
			ap, err := r.client.TaskAgentClient.GetAgentPool(r.client.Ctx, taskagent.GetAgentPoolArgs{PoolId: params.PoolId})
			if err != nil {
				return nil, "", fmt.Errorf("looking up Agent Pool with ID: %+v", err)
			}
			if *ap.AutoUpdate == *params.Pool.AutoUpdate &&
				*ap.AutoProvision == *params.Pool.AutoProvision &&
				*ap.PoolType == *params.Pool.PoolType {
				s = "Synched"
			}
			return s, s, nil
		},
		Timeout:                   5 * time.Minute,
		MinTimeout:                3 * time.Second,
		Delay:                     1 * time.Second,
		ContinuousTargetOccurence: 2,
	}
	if _, err := stateConf.WaitForStateContext(r.client.Ctx); err != nil {
		return fmt.Errorf("waiting for Agent Pool ready. %v", err)
	}
	return nil
}
