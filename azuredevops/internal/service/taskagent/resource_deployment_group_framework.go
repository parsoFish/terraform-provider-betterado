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
	_ resource.Resource                = (*DeploymentGroupResource)(nil)
	_ resource.ResourceWithConfigure   = (*DeploymentGroupResource)(nil)
	_ resource.ResourceWithImportState = (*DeploymentGroupResource)(nil)
)

// DeploymentGroupResource is the terraform-plugin-framework implementation of betterado_deployment_group.
type DeploymentGroupResource struct {
	client *client.AggregatedClient
}

// NewDeploymentGroupResource returns a new resource.Resource for betterado_deployment_group.
func NewDeploymentGroupResource() resource.Resource {
	return &DeploymentGroupResource{}
}

// deploymentGroupModel is the Terraform state model for betterado_deployment_group.
type deploymentGroupModel struct {
	ID           types.String `tfsdk:"id"`
	ProjectID    types.String `tfsdk:"project_id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	PoolID       types.Int64  `tfsdk:"pool_id"`
	MachineCount types.Int64  `tfsdk:"machine_count"`
}

func (r *DeploymentGroupResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_deployment_group"
}

func (r *DeploymentGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Azure DevOps Deployment Group.",
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
				Required:    true,
				Description: "The name of the deployment group.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     defaultString(""),
				Description: "The description of the deployment group.",
			},
			"pool_id": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The ID of the deployment pool in which deployment agents are registered. If not specified, a new pool will be created.",
				PlanModifiers: []planmodifier.Int64{
					requiresReplaceInt64(),
				},
			},
			"machine_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of deployment targets in the deployment group.",
			},
		},
	}
}

func (r *DeploymentGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DeploymentGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model deploymentGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	name := model.Name.ValueString()
	description := model.Description.ValueString()

	createParams := &taskagent.DeploymentGroupCreateParameter{
		Name:        &name,
		Description: &description,
	}

	// If pool_id is specified, use it.
	if !model.PoolID.IsNull() && !model.PoolID.IsUnknown() {
		poolID := int(model.PoolID.ValueInt64())
		createParams.PoolId = &poolID
	}

	created, err := r.client.TaskAgentClient.AddDeploymentGroup(r.client.Ctx, taskagent.AddDeploymentGroupArgs{
		Project:         &projectID,
		DeploymentGroup: createParams,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating deployment group",
			fmt.Sprintf("creating deployment group in Azure DevOps: %+v", err))
		return
	}
	if created == nil || created.Id == nil {
		resp.Diagnostics.AddError("Error creating deployment group", "response or ID is nil")
		return
	}

	model.ID = types.StringValue(strconv.Itoa(*created.Id))

	resp.Diagnostics.Append(r.readIntoModel(&model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *DeploymentGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model deploymentGroupModel
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

func (r *DeploymentGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan deploymentGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state deploymentGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	deploymentGroupID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid deployment group ID",
			fmt.Sprintf("parse deployment group ID: %+v", err))
		return
	}

	projectID := plan.ProjectID.ValueString()
	name := plan.Name.ValueString()
	description := plan.Description.ValueString()

	_, err = r.client.TaskAgentClient.UpdateDeploymentGroup(r.client.Ctx, taskagent.UpdateDeploymentGroupArgs{
		Project:           &projectID,
		DeploymentGroupId: &deploymentGroupID,
		DeploymentGroup: &taskagent.DeploymentGroupUpdateParameter{
			Name:        &name,
			Description: &description,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating deployment group",
			fmt.Sprintf("updating deployment group in Azure DevOps: %+v", err))
		return
	}

	resp.Diagnostics.Append(r.readIntoModel(&plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DeploymentGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model deploymentGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deploymentGroupID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid deployment group ID",
			fmt.Sprintf("parse deployment group ID: %+v", err))
		return
	}

	projectID := model.ProjectID.ValueString()

	if err := r.client.TaskAgentClient.DeleteDeploymentGroup(r.client.Ctx, taskagent.DeleteDeploymentGroupArgs{
		Project:           &projectID,
		DeploymentGroupId: &deploymentGroupID,
	}); err != nil {
		resp.Diagnostics.AddError("Error deleting deployment group",
			fmt.Sprintf("deleting deployment group: %+v", err))
	}
}

func (r *DeploymentGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: <project_id>/<deployment_group_id>
	projectID, dgIDStr, err := splitProjectQualifiedID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("expected <project_id>/<deployment_group_id>, got %q: %v", req.ID, err))
		return
	}
	model := deploymentGroupModel{
		ID:        types.StringValue(dgIDStr),
		ProjectID: types.StringValue(projectID),
	}
	resp.Diagnostics.Append(r.readIntoModel(&model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// readIntoModel fetches the deployment group by ID and project_id in model and populates model fields.
// If the resource is not found (404), model.ID is set to null (triggers removal).
func (r *DeploymentGroupResource) readIntoModel(model *deploymentGroupModel) diag.Diagnostics {
	var diags diag.Diagnostics

	deploymentGroupID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid deployment group ID", fmt.Sprintf("parse deployment group ID: %+v", err))
		return diags
	}

	projectID := model.ProjectID.ValueString()

	dg, err := r.client.TaskAgentClient.GetDeploymentGroup(r.client.Ctx, taskagent.GetDeploymentGroupArgs{
		Project:           &projectID,
		DeploymentGroupId: &deploymentGroupID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			model.ID = types.StringNull()
			return diags
		}
		diags.AddError("Error reading deployment group",
			fmt.Sprintf("looking up deployment group with ID %d in project %s: %v",
				deploymentGroupID, projectID, err))
		return diags
	}

	model.Name = types.StringValue(converter.ToString(dg.Name, ""))
	model.Description = types.StringValue(converter.ToString(dg.Description, ""))
	model.MachineCount = types.Int64Value(int64(converter.ToInt(dg.MachineCount, 0)))

	if dg.Pool != nil && dg.Pool.Id != nil {
		model.PoolID = types.Int64Value(int64(*dg.Pool.Id))
	} else if model.PoolID.IsNull() || model.PoolID.IsUnknown() {
		model.PoolID = types.Int64Value(0)
	}

	return diags
}
