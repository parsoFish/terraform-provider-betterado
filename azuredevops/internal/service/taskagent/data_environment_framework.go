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
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Ensure interface compliance.
var (
	_ datasource.DataSource              = (*EnvironmentDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*EnvironmentDataSource)(nil)
)

// EnvironmentDataSource is the terraform-plugin-framework data source for
// betterado_environment.
type EnvironmentDataSource struct {
	client *client.AggregatedClient
}

// NewEnvironmentDataSource returns a new framework data source for betterado_environment.
func NewEnvironmentDataSource() datasource.DataSource {
	return &EnvironmentDataSource{}
}

// environmentDataModel is the tfsdk model for the betterado_environment data source.
type environmentDataModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	EnvironmentID types.Int64  `tfsdk:"environment_id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
}

func (d *EnvironmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

func (d *EnvironmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to access information about an existing Azure DevOps Environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the environment.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
			},
			"environment_id": schema.Int64Attribute{
				Optional:    true,
				Description: "The ID of the environment to look up. Conflicts with name.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the environment. Conflicts with environment_id.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the environment.",
			},
		},
	}
}

func (d *EnvironmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EnvironmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model environmentDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()

	var env *taskagent.EnvironmentInstance

	if !model.EnvironmentID.IsNull() && !model.EnvironmentID.IsUnknown() {
		envID := int(model.EnvironmentID.ValueInt64())
		result, err := d.client.TaskAgentClient.GetEnvironmentById(d.client.Ctx, taskagent.GetEnvironmentByIdArgs{
			EnvironmentId: &envID,
			Project:       &projectID,
		})
		if err != nil {
			if utils.ResponseWasNotFound(err) {
				resp.Diagnostics.AddError("Environment not found",
					fmt.Sprintf("Environment with ID %d not found in project %s", envID, projectID))
				return
			}
			resp.Diagnostics.AddError("Error reading environment",
				fmt.Sprintf("reading environment by ID: %+v", err))
			return
		}
		env = result
	} else {
		name := model.Name.ValueString()
		response, err := d.client.TaskAgentClient.GetEnvironments(d.client.Ctx, taskagent.GetEnvironmentsArgs{
			Name:    converter.String(name),
			Project: &projectID,
			Top:     converter.Int(1),
		})
		if err != nil {
			resp.Diagnostics.AddError("Error reading environments",
				fmt.Sprintf("listing environments: %+v", err))
			return
		}
		if len(response.Value) == 0 {
			resp.Diagnostics.AddError("Environment not found",
				fmt.Sprintf("Unable to find environment with name: %s", name))
			return
		}
		env = &response.Value[0]
	}

	model.ID = types.StringValue(strconv.Itoa(*env.Id))
	model.EnvironmentID = types.Int64Value(int64(*env.Id))
	if env.Name != nil {
		model.Name = types.StringValue(*env.Name)
	}
	model.Description = types.StringValue(converter.ToString(env.Description, ""))
	if env.Project != nil && env.Project.Id != nil {
		model.ProjectID = types.StringValue(env.Project.Id.String())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
