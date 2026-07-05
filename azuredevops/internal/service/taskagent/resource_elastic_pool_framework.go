package taskagent

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/elastic"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = (*ElasticPoolResource)(nil)
	_ resource.ResourceWithConfigure   = (*ElasticPoolResource)(nil)
	_ resource.ResourceWithImportState = (*ElasticPoolResource)(nil)
)

// ElasticPoolResource is the terraform-plugin-framework implementation of betterado_elastic_pool.
type ElasticPoolResource struct {
	client *client.AggregatedClient
}

// NewElasticPoolResource returns a new resource.Resource for betterado_elastic_pool.
func NewElasticPoolResource() resource.Resource {
	return &ElasticPoolResource{}
}

// elasticPoolModel is the Terraform state model for betterado_elastic_pool.
type elasticPoolModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	AzureResourceID      types.String `tfsdk:"azure_resource_id"`
	ServiceEndpointID    types.String `tfsdk:"service_endpoint_id"`
	ServiceEndpointScope types.String `tfsdk:"service_endpoint_scope"`
	DesiredIdle          types.Int64  `tfsdk:"desired_idle"`
	MaxCapacity          types.Int64  `tfsdk:"max_capacity"`
	RecycleAfterEachUse  types.Bool   `tfsdk:"recycle_after_each_use"`
	AgentInteractiveUI   types.Bool   `tfsdk:"agent_interactive_ui"`
	TimeToLiveMinutes    types.Int64  `tfsdk:"time_to_live_minutes"`
	AutoProvision        types.Bool   `tfsdk:"auto_provision"`
	AutoUpdate           types.Bool   `tfsdk:"auto_update"`
	ProjectID            types.String `tfsdk:"project_id"`
}

func (r *ElasticPoolResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_elastic_pool"
}

func (r *ElasticPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Azure DevOps Elastic Pool (VMSS agent pool).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					useStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the elastic pool.",
			},
			"azure_resource_id": schema.StringAttribute{
				Required:    true,
				Description: "The Azure resource ID of the virtual machine scale set.",
			},
			"service_endpoint_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the service endpoint used to connect to the Azure subscription containing the scale set.",
			},
			"service_endpoint_scope": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project where the service endpoint is defined.",
			},
			"desired_idle": schema.Int64Attribute{
				Required:    true,
				Description: "Number of agents to have ready waiting for jobs.",
			},
			"max_capacity": schema.Int64Attribute{
				Required:    true,
				Description: "Maximum number of machines in the elastic pool.",
			},
			"recycle_after_each_use": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     defaultBool(false),
				Description: "Setting this value to true will recycle the agent after each use.",
			},
			"agent_interactive_ui": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     defaultBool(false),
				Description: "Set whether agents should be configured to run with interactive UI.",
			},
			"time_to_live_minutes": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     defaultInt64(30),
				Description: "The time in minutes to keep idle agents alive.",
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
			"project_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     defaultString(""),
				Description: "The ID of the project (optional).",
			},
		},
	}
}

func (r *ElasticPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ElasticPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model elasticPoolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	desiredIdle := int(model.DesiredIdle.ValueInt64())
	maxCapacity := int(model.MaxCapacity.ValueInt64())
	if desiredIdle > maxCapacity {
		resp.Diagnostics.AddError(
			"Invalid elastic pool configuration",
			"`desired_idle` cannot be greater than `max_capacity`. Valid range is from 0 to the maximum number of virtual machines in the scale set.",
		)
		return
	}

	seID, err := uuid.Parse(model.ServiceEndpointID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid service_endpoint_id", fmt.Sprintf("parse service_endpoint_id: %+v", err))
		return
	}

	seScope, err := uuid.Parse(model.ServiceEndpointScope.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid service_endpoint_scope", fmt.Sprintf("parse service_endpoint_scope: %+v", err))
		return
	}

	ttl := int(model.TimeToLiveMinutes.ValueInt64())
	args := elastic.CreateElasticPoolArgs{
		ElasticPool: &elastic.ElasticPool{
			AgentInteractiveUI:   converter.ToPtr(model.AgentInteractiveUI.ValueBool()),
			AzureId:              converter.ToPtr(model.AzureResourceID.ValueString()),
			DesiredIdle:          &desiredIdle,
			MaxCapacity:          &maxCapacity,
			TimeToLiveMinutes:    &ttl,
			RecycleAfterEachUse:  converter.ToPtr(model.RecycleAfterEachUse.ValueBool()),
			ServiceEndpointId:    &seID,
			ServiceEndpointScope: &seScope,
		},
		PoolName: converter.String(model.Name.ValueString()),
	}

	if !model.ProjectID.IsNull() && !model.ProjectID.IsUnknown() && model.ProjectID.ValueString() != "" {
		projectID, parseErr := uuid.Parse(model.ProjectID.ValueString())
		if parseErr != nil {
			resp.Diagnostics.AddError("Invalid project_id", fmt.Sprintf("parse project_id: %+v", parseErr))
			return
		}
		args.ProjectId = &projectID
	}

	elasticPool, err := r.client.ElasticClient.CreateElasticPool(r.client.Ctx, args)
	if err != nil {
		resp.Diagnostics.AddError("Error creating elastic pool",
			fmt.Sprintf("creating Elastic Pool: %+v", err))
		return
	}

	updateArgs := taskagent.UpdateAgentPoolArgs{
		PoolId: elasticPool.ElasticPool.PoolId,
		Pool: &taskagent.TaskAgentPool{
			Name:          args.PoolName,
			AutoProvision: converter.ToPtr(model.AutoProvision.ValueBool()),
			AutoUpdate:    converter.ToPtr(model.AutoUpdate.ValueBool()),
		},
	}
	if _, err = r.client.TaskAgentClient.UpdateAgentPool(r.client.Ctx, updateArgs); err != nil {
		resp.Diagnostics.AddError("Error updating agent pool after elastic pool creation",
			fmt.Sprintf("updating agent pool in Azure DevOps: %+v", err))
		return
	}
	if err := r.syncElasticPoolStatus(updateArgs); err != nil {
		resp.Diagnostics.AddError("Error syncing elastic pool status", err.Error())
		return
	}

	model.ID = types.StringValue(strconv.Itoa(*elasticPool.ElasticPool.PoolId))

	resp.Diagnostics.Append(r.readIntoModel(&model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ElasticPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model elasticPoolModel
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

func (r *ElasticPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan elasticPoolModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state elasticPoolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	poolID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid elastic pool ID",
			fmt.Sprintf("getting Elastic Pool Id: %+v", err))
		return
	}

	desiredIdle := int(plan.DesiredIdle.ValueInt64())
	maxCapacity := int(plan.MaxCapacity.ValueInt64())
	if desiredIdle > maxCapacity {
		resp.Diagnostics.AddError(
			"Invalid elastic pool configuration",
			"`desired_idle` cannot be greater than `max_capacity`. Valid range is from 0 to the maximum number of virtual machines in the scale set.",
		)
		return
	}

	seID, err := uuid.Parse(plan.ServiceEndpointID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid service_endpoint_id", fmt.Sprintf("parse service_endpoint_id: %+v", err))
		return
	}

	seScope, err := uuid.Parse(plan.ServiceEndpointScope.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid service_endpoint_scope", fmt.Sprintf("parse service_endpoint_scope: %+v", err))
		return
	}

	ttl := int(plan.TimeToLiveMinutes.ValueInt64())
	elasticPoolArgs := elastic.UpdateElasticPoolArgs{
		PoolId: &poolID,
		ElasticPoolSettings: &elastic.ElasticPoolSettings{
			AgentInteractiveUI:   converter.ToPtr(plan.AgentInteractiveUI.ValueBool()),
			AzureId:              converter.ToPtr(plan.AzureResourceID.ValueString()),
			TimeToLiveMinutes:    &ttl,
			RecycleAfterEachUse:  converter.ToPtr(plan.RecycleAfterEachUse.ValueBool()),
			ServiceEndpointId:    &seID,
			ServiceEndpointScope: &seScope,
			DesiredIdle:          &desiredIdle,
			MaxCapacity:          &maxCapacity,
		},
	}

	if _, err := r.client.ElasticClient.UpdateElasticPool(r.client.Ctx, elasticPoolArgs); err != nil {
		resp.Diagnostics.AddError("Error updating elastic pool",
			fmt.Sprintf("updating Elastic Pool in Azure DevOps: %+v", err))
		return
	}

	agentPoolArgs := taskagent.UpdateAgentPoolArgs{
		PoolId: &poolID,
		Pool: &taskagent.TaskAgentPool{
			Name:          converter.String(plan.Name.ValueString()),
			AutoProvision: converter.Bool(plan.AutoProvision.ValueBool()),
			AutoUpdate:    converter.Bool(plan.AutoUpdate.ValueBool()),
		},
	}
	if _, err := r.client.TaskAgentClient.UpdateAgentPool(r.client.Ctx, agentPoolArgs); err != nil {
		resp.Diagnostics.AddError("Error updating agent pool settings",
			fmt.Sprintf("updating Elastic Pool in Azure DevOps: %+v", err))
		return
	}
	if err := r.syncElasticPoolStatus(agentPoolArgs); err != nil {
		resp.Diagnostics.AddError("Error syncing elastic pool status", err.Error())
		return
	}

	resp.Diagnostics.Append(r.readIntoModel(&plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ElasticPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model elasticPoolModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	poolID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid elastic pool ID",
			fmt.Sprintf("parse agent pool ID: %+v", err))
		return
	}

	if err := r.client.TaskAgentClient.DeleteAgentPool(r.client.Ctx, taskagent.DeleteAgentPoolArgs{PoolId: &poolID}); err != nil {
		resp.Diagnostics.AddError("Error deleting elastic pool", err.Error())
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
		resp.Diagnostics.AddError("Error waiting for elastic pool deletion",
			fmt.Sprintf("waiting for Elastic Pool deleted. %v", err))
	}
}

func (r *ElasticPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	model := elasticPoolModel{
		ID: types.StringValue(req.ID),
	}
	resp.Diagnostics.Append(r.readIntoModel(&model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// readIntoModel fetches the elastic pool and agent pool by ID and populates model fields.
// If the resource is not found (404), model.ID is set to null (triggers removal).
func (r *ElasticPoolResource) readIntoModel(model *elasticPoolModel) diag.Diagnostics {
	var diags diag.Diagnostics

	poolID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid elastic pool ID", fmt.Sprintf("parse Elastic Pool ID: %+v", err))
		return diags
	}

	elasticPool, err := r.client.ElasticClient.GetElasticPool(r.client.Ctx, elastic.GetElasticPoolArgs{
		PoolId: &poolID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			model.ID = types.StringNull()
			return diags
		}
		diags.AddError("Error reading elastic pool",
			fmt.Sprintf("looking up Elastic Pool with ID %d. Error: %v", poolID, err))
		return diags
	}

	agentPool, err := r.client.TaskAgentClient.GetAgentPool(r.client.Ctx, taskagent.GetAgentPoolArgs{
		PoolId: &poolID,
	})
	if err != nil {
		diags.AddError("Error reading agent pool",
			fmt.Sprintf("looking up Agent Pool with ID %d. Error: %v", poolID, err))
		return diags
	}

	model.Name = types.StringValue(converter.ToString(agentPool.Name, ""))
	model.AzureResourceID = types.StringValue(converter.ToString(elasticPool.AzureId, ""))
	if elasticPool.ServiceEndpointId != nil {
		model.ServiceEndpointID = types.StringValue(elasticPool.ServiceEndpointId.String())
	}
	if elasticPool.ServiceEndpointScope != nil {
		model.ServiceEndpointScope = types.StringValue(elasticPool.ServiceEndpointScope.String())
	}
	model.DesiredIdle = types.Int64Value(int64(converter.ToInt(elasticPool.DesiredIdle, 0)))
	model.MaxCapacity = types.Int64Value(int64(converter.ToInt(elasticPool.MaxCapacity, 0)))
	model.RecycleAfterEachUse = types.BoolValue(converter.ToBool(elasticPool.RecycleAfterEachUse, false))
	model.AgentInteractiveUI = types.BoolValue(converter.ToBool(elasticPool.AgentInteractiveUI, false))
	model.TimeToLiveMinutes = types.Int64Value(int64(converter.ToInt(elasticPool.TimeToLiveMinutes, 30)))
	if agentPool.AutoProvision != nil {
		model.AutoProvision = types.BoolValue(*agentPool.AutoProvision)
	}
	if agentPool.AutoUpdate != nil {
		model.AutoUpdate = types.BoolValue(*agentPool.AutoUpdate)
	}

	// project_id is not returned by the API; preserve existing state value.
	// (The SDKv2 implementation did the same: d.Set("project_id", d.Get("project_id").(string)))

	return diags
}

// syncElasticPoolStatus waits for the agent pool to match the desired auto_provision/auto_update state.
func (r *ElasticPoolResource) syncElasticPoolStatus(params taskagent.UpdateAgentPoolArgs) error {
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
				*ap.AutoProvision == *params.Pool.AutoProvision {
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
