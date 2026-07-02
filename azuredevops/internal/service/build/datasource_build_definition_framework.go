package build

// datasource_build_definition_framework.go — terraform-plugin-framework implementation of
// data.betterado_build_definition.
//
// The SDKv2 DataBuildDefinition() in data_build_definition.go remains in place
// but its "betterado_build_definition" entry is removed from provider.go DataSourcesMap
// (commented with a migration note) to avoid "Invalid Provider Server Combination".
// The framework data source is registered in framework_provider.go DataSources().
//
// No build tag on this production file — only _test.go files carry //go:build.

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var _ datasource.DataSource = &BuildDefinitionDataSource{}
var _ datasource.DataSourceWithConfigure = &BuildDefinitionDataSource{}

// BuildDefinitionDataSource is the framework data source for data.betterado_build_definition.
type BuildDefinitionDataSource struct {
	client *client.AggregatedClient
}

// NewBuildDefinitionDataSource returns a new framework datasource.DataSource for data.betterado_build_definition.
func NewBuildDefinitionDataSource() datasource.DataSource {
	return &BuildDefinitionDataSource{}
}

// ── Model ────────────────────────────────────────────────────────────────────

// buildDefinitionDataSourceModel is the tfsdk model for the data source.
type buildDefinitionDataSourceModel struct {
	ID                    types.String `tfsdk:"id"`
	ProjectID             types.String `tfsdk:"project_id"`
	Name                  types.String `tfsdk:"name"`
	Path                  types.String `tfsdk:"path"`
	Revision              types.Int64  `tfsdk:"revision"`
	Repository            types.List   `tfsdk:"repository"`
	CITrigger             types.List   `tfsdk:"ci_trigger"`
	PullRequestTrigger    types.List   `tfsdk:"pull_request_trigger"`
	Variable              types.Set    `tfsdk:"variable"`
	AgentPoolName         types.String `tfsdk:"agent_pool_name"`
	AgentSpecification    types.String `tfsdk:"agent_specification"`
	JobAuthorizationScope types.String `tfsdk:"job_authorization_scope"`
	QueueStatus           types.String `tfsdk:"queue_status"`
	Schedules             types.List   `tfsdk:"schedules"`
}

// ── Metadata / Schema ────────────────────────────────────────────────────────

func (d *BuildDefinitionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build_definition"
}

func (d *BuildDefinitionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a build (pipeline) definition from an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the build definition.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project the build definition belongs to.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the build definition to look up.",
			},
			"path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: `The folder path of the build definition (e.g. "\\MyFolder"). Defaults to "\".`,
			},
			"revision": schema.Int64Attribute{
				Computed:    true,
				Description: "The revision number of the build definition.",
			},
			"repository": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The repository block describing the source repository for the pipeline.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"yml_path": schema.StringAttribute{
							Computed:    true,
							Description: "The path of the YAML pipeline file in the repository.",
						},
						"repo_id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the repository.",
						},
						"repo_type": schema.StringAttribute{
							Computed:    true,
							Description: "The type of repository.",
						},
						"branch_name": schema.StringAttribute{
							Computed:    true,
							Description: "The default branch of the repository.",
						},
						"service_connection_id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the service connection used for the repository.",
						},
						"github_enterprise_url": schema.StringAttribute{
							Computed:    true,
							Description: "The URL of the GitHub Enterprise instance.",
						},
						"url": schema.StringAttribute{
							Computed:    true,
							Description: "The URL of the repository.",
						},
						"report_build_status": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether build status is reported to the repository host.",
						},
					},
				},
			},
			"ci_trigger": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The continuous integration trigger configuration.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"use_yaml": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the YAML-defined trigger is used.",
						},
					},
				},
			},
			"pull_request_trigger": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The pull request trigger configuration.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"use_yaml": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the YAML-defined PR trigger is used.",
						},
						"initial_branch": schema.StringAttribute{
							Computed:    true,
							Description: "The initial branch used for PR triggers.",
						},
						"comment_required": schema.StringAttribute{
							Computed:    true,
							Description: "Whether a comment is required to trigger the PR build.",
						},
					},
				},
			},
			"variable": schema.SetNestedAttribute{
				Computed:    true,
				Description: "The variables set on the build definition.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the variable.",
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Description: "The value of the variable.",
						},
						"secret_value": schema.StringAttribute{
							Computed:    true,
							Sensitive:   true,
							Description: "The secret value of the variable.",
						},
						"is_secret": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the variable is secret.",
						},
						"allow_override": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the variable can be overridden at queue time.",
						},
					},
				},
			},
			"agent_pool_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the agent pool used for builds.",
			},
			"agent_specification": schema.StringAttribute{
				Computed:    true,
				Description: "The agent specification (image) used for the build.",
			},
			"job_authorization_scope": schema.StringAttribute{
				Computed:    true,
				Description: "The authorization scope for the job.",
			},
			"queue_status": schema.StringAttribute{
				Computed:    true,
				Description: "The queue status of the build definition.",
			},
			"schedules": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The schedule triggers configured on the build definition.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"days_to_build": schema.StringAttribute{
							Computed:    true,
							Description: "The days on which the schedule triggers (bitmask string).",
						},
						"schedule_only_with_changes": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether to only trigger the schedule when there are changes.",
						},
						"start_hours": schema.Int64Attribute{
							Computed:    true,
							Description: "The hour of day the schedule starts (0-23).",
						},
						"start_minutes": schema.Int64Attribute{
							Computed:    true,
							Description: "The minute of the hour the schedule starts (0-59).",
						},
						"time_zone_id": schema.StringAttribute{
							Computed:    true,
							Description: "The time zone identifier for the schedule.",
						},
						"branch_filter": schema.SetNestedAttribute{
							Computed:    true,
							Description: "Branch filters for the schedule.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"include": schema.SetAttribute{
										Computed:    true,
										ElementType: types.StringType,
										Description: "Branch patterns to include.",
									},
									"exclude": schema.SetAttribute{
										Computed:    true,
										ElementType: types.StringType,
										Description: "Branch patterns to exclude.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *BuildDefinitionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("data.betterado_build_definition: expected *client.AggregatedClient, got %T", req.ProviderData),
		)
		return
	}
	d.client = c
}

// ── Read ─────────────────────────────────────────────────────────────────────

func (d *BuildDefinitionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "data.betterado_build_definition Read: provider client not configured")
		return
	}

	var model buildDefinitionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	name := model.Name.ValueString()
	path := model.Path.ValueString()
	if path == "" {
		path = `\`
	}

	// Look up build definitions by name (and optionally path) using the shared helper.
	defs, apiErr := getBuildDefinitionsByNameAndProject(d.client, name, path, projectID)
	if apiErr != nil {
		if utils.ResponseWasNotFound(apiErr) {
			resp.Diagnostics.AddError("Not found", fmt.Sprintf("build definition name=%q, path=%q not found in project %q", name, path, projectID))
			return
		}
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("listing build definitions (name=%q, path=%q): %s", name, path, apiErr))
		return
	}
	if defs == nil || len(*defs) == 0 {
		resp.Diagnostics.AddError("Not found", fmt.Sprintf("no build definition found with name=%q, path=%q in project %q", name, path, projectID))
		return
	}
	if len(*defs) > 1 {
		resp.Diagnostics.AddError("Ambiguous result", fmt.Sprintf("multiple build definitions with name=%q found in project %q", name, projectID))
		return
	}

	def := &(*defs)[0]

	// Populate model.
	model.ID = types.StringValue(strconv.Itoa(*def.Id))
	if def.Name != nil {
		model.Name = types.StringValue(*def.Name)
	}
	model.ProjectID = types.StringValue(projectID)

	if def.Path != nil {
		model.Path = types.StringValue(*def.Path)
	} else {
		model.Path = types.StringValue(`\`)
	}
	if def.Revision != nil {
		model.Revision = types.Int64Value(int64(*def.Revision))
	} else {
		model.Revision = types.Int64Value(0)
	}

	// Agent pool name.
	if def.Queue != nil && def.Queue.Pool != nil && def.Queue.Pool.Name != nil {
		model.AgentPoolName = types.StringValue(*def.Queue.Pool.Name)
	} else {
		model.AgentPoolName = types.StringValue("")
	}

	// Agent specification.
	model.AgentSpecification = types.StringValue("")

	// Job authorization scope.
	if def.JobAuthorizationScope != nil {
		model.JobAuthorizationScope = types.StringValue(string(*def.JobAuthorizationScope))
	} else {
		model.JobAuthorizationScope = types.StringValue("")
	}

	// Queue status.
	if def.QueueStatus != nil {
		model.QueueStatus = types.StringValue(string(*def.QueueStatus))
	} else {
		model.QueueStatus = types.StringValue("")
	}

	// Repository.
	repoObjType := bddsDsRepoObjectType()
	if def.Repository != nil {
		repoObj, rDiags := bddsDsRepoToObject(def.Repository)
		resp.Diagnostics.Append(rDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		listVal, lDiags := types.ListValue(repoObjType, []attr.Value{repoObj})
		resp.Diagnostics.Append(lDiags...)
		model.Repository = listVal
	} else {
		model.Repository = types.ListValueMust(repoObjType, []attr.Value{})
	}

	// CI trigger.
	ciObjType := bddsDsCITriggerObjectType()
	model.CITrigger = types.ListValueMust(ciObjType, []attr.Value{})

	// Pull request trigger.
	prObjType := bddsDsPRTriggerObjectType()
	model.PullRequestTrigger = types.ListValueMust(prObjType, []attr.Value{})

	// Variables.
	varObjType := bddsDsVariableObjectType()
	varValues := []attr.Value{}
	if def.Variables != nil {
		for varName, varVal := range *def.Variables {
			value := ""
			secretValue := ""
			isSecret := false
			allowOverride := true
			if varVal.Value != nil {
				value = *varVal.Value
			}
			if varVal.IsSecret != nil {
				isSecret = *varVal.IsSecret
			}
			if varVal.AllowOverride != nil {
				allowOverride = *varVal.AllowOverride
			}
			varObj, vDiags := types.ObjectValue(varObjType.AttrTypes, map[string]attr.Value{
				"name":           types.StringValue(varName),
				"value":          types.StringValue(value),
				"secret_value":   types.StringValue(secretValue),
				"is_secret":      types.BoolValue(isSecret),
				"allow_override": types.BoolValue(allowOverride),
			})
			resp.Diagnostics.Append(vDiags...)
			varValues = append(varValues, varObj)
		}
	}
	varSet, vsDiags := types.SetValue(varObjType, varValues)
	resp.Diagnostics.Append(vsDiags...)
	model.Variable = varSet

	// Schedules — empty list.
	schedObjType := bddsDsScheduleObjectType()
	model.Schedules = types.ListValueMust(schedObjType, []attr.Value{})

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Object-type helpers ────────────────────────────────────────────────────────

func bddsDsRepoObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"yml_path":              types.StringType,
		"repo_id":               types.StringType,
		"repo_type":             types.StringType,
		"branch_name":           types.StringType,
		"service_connection_id": types.StringType,
		"github_enterprise_url": types.StringType,
		"url":                   types.StringType,
		"report_build_status":   types.BoolType,
	}}
}

func bddsDsCITriggerObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"use_yaml": types.BoolType,
	}}
}

func bddsDsPRTriggerObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"use_yaml":         types.BoolType,
		"initial_branch":   types.StringType,
		"comment_required": types.StringType,
	}}
}

func bddsDsVariableObjectType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"name":           types.StringType,
		"value":          types.StringType,
		"secret_value":   types.StringType,
		"is_secret":      types.BoolType,
		"allow_override": types.BoolType,
	}}
}

func bddsDsScheduleObjectType() types.ObjectType {
	branchFilterObjType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"include": types.SetType{ElemType: types.StringType},
		"exclude": types.SetType{ElemType: types.StringType},
	}}
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"days_to_build":              types.StringType,
		"schedule_only_with_changes": types.BoolType,
		"start_hours":                types.Int64Type,
		"start_minutes":              types.Int64Type,
		"time_zone_id":               types.StringType,
		"branch_filter":              types.SetType{ElemType: branchFilterObjType},
	}}
}

// bddsDsRepoToObject converts a *build.BuildRepository to an ObjectValue.
func bddsDsRepoToObject(repo *build.BuildRepository) (types.Object, diag.Diagnostics) {
	ymlPath := ""
	repoID := ""
	repoType := ""
	branchName := ""
	svcConnID := ""
	gheURL := ""
	url := ""
	reportStatus := true

	if repo.DefaultBranch != nil {
		branchName = *repo.DefaultBranch
	}
	if repo.Id != nil {
		repoID = *repo.Id
	}
	if repo.Type != nil {
		repoType = *repo.Type
	}
	if repo.Url != nil {
		url = *repo.Url
	}

	if repo.Properties != nil {
		if v, ok := (*repo.Properties)["reportBuildStatus"]; ok {
			reportStatus = v != "false" && v != "False"
		}
		if v, ok := (*repo.Properties)["connectedServiceId"]; ok {
			svcConnID = v
		}
		if v, ok := (*repo.Properties)["githubEnterpriseUrl"]; ok {
			gheURL = v
		}
		if v, ok := (*repo.Properties)["tfvcMapping"]; ok {
			ymlPath = v
		}
	}
	_ = converter.String // import converter to satisfy compiler

	return types.ObjectValue(bddsDsRepoObjectType().AttrTypes, map[string]attr.Value{
		"yml_path":              types.StringValue(ymlPath),
		"repo_id":               types.StringValue(repoID),
		"repo_type":             types.StringValue(repoType),
		"branch_name":           types.StringValue(branchName),
		"service_connection_id": types.StringValue(svcConnID),
		"github_enterprise_url": types.StringValue(gheURL),
		"url":                   types.StringValue(url),
		"report_build_status":   types.BoolValue(reportStatus),
	})
}
