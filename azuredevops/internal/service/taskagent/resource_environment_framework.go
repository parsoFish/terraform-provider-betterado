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
	_ resource.Resource                = (*EnvironmentResource)(nil)
	_ resource.ResourceWithConfigure   = (*EnvironmentResource)(nil)
	_ resource.ResourceWithImportState = (*EnvironmentResource)(nil)
)

// EnvironmentResource is the terraform-plugin-framework implementation of betterado_environment.
type EnvironmentResource struct {
	client *client.AggregatedClient
}

// NewEnvironmentResource returns a new resource.Resource for betterado_environment.
func NewEnvironmentResource() resource.Resource {
	return &EnvironmentResource{}
}

// environmentModel is the Terraform state model for betterado_environment.
type environmentModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (r *EnvironmentResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_environment"
}

func (r *EnvironmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Azure DevOps Environment.",
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
				Description: "The name of the environment.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     defaultString(""),
				Description: "A description of the environment.",
			},
		},
	}
}

func (r *EnvironmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *EnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model environmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	created, err := r.client.TaskAgentClient.AddEnvironment(r.client.Ctx, taskagent.AddEnvironmentArgs{
		Project: &projectID,
		EnvironmentCreateParameter: &taskagent.EnvironmentCreateParameter{
			Name:        converter.String(model.Name.ValueString()),
			Description: converter.String(model.Description.ValueString()),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating environment",
			fmt.Sprintf("creating environment in Azure DevOps: %+v", err))
		return
	}

	model.ID = types.StringValue(strconv.Itoa(*created.Id))

	resp.Diagnostics.Append(r.readIntoModel(&model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *EnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model environmentModel
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

func (r *EnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan environmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state environmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	environmentID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid environment ID",
			fmt.Sprintf("parse environment ID: %+v", err))
		return
	}

	projectID := plan.ProjectID.ValueString()
	_, err = r.client.TaskAgentClient.UpdateEnvironment(r.client.Ctx, taskagent.UpdateEnvironmentArgs{
		Project:       &projectID,
		EnvironmentId: &environmentID,
		EnvironmentUpdateParameter: &taskagent.EnvironmentUpdateParameter{
			Name:        converter.String(plan.Name.ValueString()),
			Description: converter.String(plan.Description.ValueString()),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating environment",
			fmt.Sprintf("updating environment in Azure DevOps: %+v", err))
		return
	}

	resp.Diagnostics.Append(r.readIntoModel(&plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *EnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model environmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	environmentID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid environment ID",
			fmt.Sprintf("parse environment ID: %+v", err))
		return
	}
	projectID := model.ProjectID.ValueString()

	if err := r.client.TaskAgentClient.DeleteEnvironment(r.client.Ctx, taskagent.DeleteEnvironmentArgs{
		Project:       &projectID,
		EnvironmentId: &environmentID,
	}); err != nil {
		resp.Diagnostics.AddError("Error deleting environment", err.Error())
	}
}

func (r *EnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: <project_id>/<environment_id>
	projectID, envIDStr, err := splitProjectQualifiedID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("expected <project_id>/<environment_id>, got %q: %v", req.ID, err))
		return
	}
	model := environmentModel{
		ID:        types.StringValue(envIDStr),
		ProjectID: types.StringValue(projectID),
	}
	resp.Diagnostics.Append(r.readIntoModel(&model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// readIntoModel fetches the environment by the ID and project_id in model and populates model fields.
// If the environment is not found (404), model.ID is set to null.
func (r *EnvironmentResource) readIntoModel(model *environmentModel) diag.Diagnostics {
	var diags diag.Diagnostics

	environmentID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid environment ID", fmt.Sprintf("parse environment ID: %+v", err))
		return diags
	}
	projectID := model.ProjectID.ValueString()

	env, err := r.client.TaskAgentClient.GetEnvironmentById(r.client.Ctx, taskagent.GetEnvironmentByIdArgs{
		Project:       &projectID,
		EnvironmentId: &environmentID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			model.ID = types.StringNull()
			return diags
		}
		diags.AddError("Error reading environment",
			fmt.Sprintf("looking up Environment with ID %d in project %s. Error: %v", environmentID, projectID, err))
		return diags
	}

	if env.Name != nil {
		model.Name = types.StringValue(*env.Name)
	}
	model.Description = types.StringValue(converter.ToString(env.Description, ""))
	if env.Project != nil && env.Project.Id != nil {
		model.ProjectID = types.StringValue(env.Project.Id.String())
	}
	return diags
}
