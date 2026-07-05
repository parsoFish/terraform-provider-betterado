package taskagent

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// Ensure interface compliance.
var _ datasource.DataSource = &TaskGroupDataSource{}

// TaskGroupDataSource is the terraform-plugin-framework data source for
// betterado_task_group.
type TaskGroupDataSource struct {
	client *client.AggregatedClient
}

// NewTaskGroupDataSource returns a new framework data source for betterado_task_group.
func NewTaskGroupDataSource() datasource.DataSource {
	return &TaskGroupDataSource{}
}

// taskGroupDataModel is the tfsdk model for the data source.
// It reuses the nested models already defined in resource_task_group_framework.go
// (versionModel, inputModel, taskStepModel).
type taskGroupDataModel struct {
	ID                 types.String `tfsdk:"id"`
	ProjectID          types.String `tfsdk:"project_id"`
	Name               types.String `tfsdk:"name"`
	FriendlyName       types.String `tfsdk:"friendly_name"`
	Description        types.String `tfsdk:"description"`
	Category           types.String `tfsdk:"category"`
	Author             types.String `tfsdk:"author"`
	IconURL            types.String `tfsdk:"icon_url"`
	InstanceNameFormat types.String `tfsdk:"instance_name_format"`
	RunsOn             types.List   `tfsdk:"runs_on"`
	Revision           types.Int64  `tfsdk:"revision"`
	DefinitionType     types.String `tfsdk:"definition_type"`
	Version            types.List   `tfsdk:"version"`
	Input              types.List   `tfsdk:"input"`
	Task               types.List   `tfsdk:"task"`
}

func (d *TaskGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task_group"
}

func (d *TaskGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a task group from an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID (UUID) of the task group.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project the task group belongs to.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the task group.",
			},
			"friendly_name": schema.StringAttribute{
				Computed:    true,
				Description: "The friendly name of the task group.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the task group.",
			},
			"category": schema.StringAttribute{
				Computed:    true,
				Description: "The category of the task group.",
			},
			"author": schema.StringAttribute{
				Computed:    true,
				Description: "The author of the task group.",
			},
			"icon_url": schema.StringAttribute{
				Computed:    true,
				Description: "The icon URL of the task group.",
			},
			"instance_name_format": schema.StringAttribute{
				Computed:    true,
				Description: "The instance name format of the task group.",
			},
			"runs_on": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "The list of agent specifications on which the task group runs.",
			},
			"revision": schema.Int64Attribute{
				Computed:    true,
				Description: "The revision number of the task group.",
			},
			"definition_type": schema.StringAttribute{
				Computed:    true,
				Description: "The definition type of the task group.",
			},
			"version": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The version of the task group.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"major":   schema.Int64Attribute{Computed: true},
						"minor":   schema.Int64Attribute{Computed: true},
						"patch":   schema.Int64Attribute{Computed: true},
						"is_test": schema.BoolAttribute{Computed: true},
					},
				},
			},
			"input": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of inputs for the task group.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":          schema.StringAttribute{Computed: true},
						"label":         schema.StringAttribute{Computed: true},
						"type":          schema.StringAttribute{Computed: true},
						"default_value": schema.StringAttribute{Computed: true},
						"required":      schema.BoolAttribute{Computed: true},
						"help_markdown": schema.StringAttribute{Computed: true},
						"group_name":    schema.StringAttribute{Computed: true},
						"options": schema.MapAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"visible_rule": schema.StringAttribute{Computed: true},
						"properties": schema.MapAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"aliases": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
			},
			"task": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of tasks in the task group.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"display_name":                schema.StringAttribute{Computed: true},
						"task_id":                     schema.StringAttribute{Computed: true},
						"task_version":                schema.StringAttribute{Computed: true},
						"task_definition_type":        schema.StringAttribute{Computed: true},
						"enabled":                     schema.BoolAttribute{Computed: true},
						"always_run":                  schema.BoolAttribute{Computed: true},
						"continue_on_error":           schema.BoolAttribute{Computed: true},
						"condition":                   schema.StringAttribute{Computed: true},
						"timeout_in_minutes":          schema.Int64Attribute{Computed: true},
						"retry_count_on_task_failure": schema.Int64Attribute{Computed: true},
						"inputs": schema.MapAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"environment": schema.MapAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *TaskGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TaskGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model taskGroupDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	tgIDStr := model.ID.ValueString()

	tgID, err := uuid.Parse(tgIDStr)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid task group ID",
			fmt.Sprintf("The task group ID %q is not a valid UUID: %s", tgIDStr, err.Error()),
		)
		return
	}

	taskGroups, err := d.client.TaskAgentClient.GetTaskGroups(d.client.Ctx, taskagent.GetTaskGroupsArgs{
		Project:     &projectID,
		TaskGroupId: &tgID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.Diagnostics.AddError(
				"Task group not found",
				fmt.Sprintf("Task group (ID: %s) not found in project %s", tgID, projectID),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading task group",
			fmt.Sprintf("Reading task group (ID: %s): %s", tgID, err.Error()),
		)
		return
	}

	if taskGroups == nil || len(*taskGroups) == 0 {
		resp.Diagnostics.AddError(
			"Task group not found",
			fmt.Sprintf("Task group (ID: %s) not found in project %s", tgID, projectID),
		)
		return
	}

	tg := (*taskGroups)[0]

	// Reuse the resource flatten helper — it populates all fields identically.
	// We bridge from taskGroupDataModel to taskGroupModel for the helper call,
	// then copy back.
	bridge := &taskGroupModel{
		ID:        model.ID,
		ProjectID: model.ProjectID,
	}
	diags := flattenTaskGroupFramework(ctx, bridge, &tg)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Copy the populated fields from bridge into our data model.
	model.ID = bridge.ID
	model.ProjectID = bridge.ProjectID
	model.Name = bridge.Name
	model.FriendlyName = bridge.FriendlyName
	model.Description = bridge.Description
	model.Category = bridge.Category
	model.Author = bridge.Author
	model.IconURL = bridge.IconURL
	model.InstanceNameFormat = bridge.InstanceNameFormat
	model.RunsOn = bridge.RunsOn
	model.Revision = bridge.Revision
	model.DefinitionType = bridge.DefinitionType
	model.Version = bridge.Version
	model.Input = bridge.Input
	model.Task = bridge.Task

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
