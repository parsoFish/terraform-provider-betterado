package build

// resource_build_definition_framework.go — terraform-plugin-framework implementation of
// betterado_build_definition.
//
// The SDKv2 ResourceBuildDefinition() in resource_build_definition.go remains in place
// but its "betterado_build_definition" entry is removed from provider.go ResourcesMap
// (commented with a migration note) to avoid "Invalid Provider Server Combination".
// The framework resource is registered in framework_provider.go Resources().
//
// No build tag on this production file — only _test.go files carry //go:build.
//
// Deliberately deferred subtrees (NOT migrated to this framework schema):
//   - variable_groups  — complex int-set; documented in docs/build-gap-matrix.md
//   - build_completion_trigger — complex nested; documented in docs/build-gap-matrix.md
//   - jobs             — large nested block (OtherGit only); documented in docs/build-gap-matrix.md
//   - schedules        — complex nested with timezone list; documented in docs/build-gap-matrix.md
//   - features (top-level list) — skip_first_run is promoted to a top-level bool attribute here

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/pipelines"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                   = &BuildDefinitionResource{}
	_ resource.ResourceWithImportState    = &BuildDefinitionResource{}
	_ resource.ResourceWithConfigure      = &BuildDefinitionResource{}
	_ resource.ResourceWithValidateConfig = &BuildDefinitionResource{}
)

// BuildDefinitionResource is the framework resource for betterado_build_definition.
type BuildDefinitionResource struct {
	client *client.AggregatedClient
}

// NewBuildDefinitionResource returns a new framework resource.Resource for betterado_build_definition.
func NewBuildDefinitionResource() resource.Resource {
	return &BuildDefinitionResource{}
}

// ── Models ───────────────────────────────────────────────────────────────────

type buildDefinitionRepositoryModel struct {
	YmlPath             types.String `tfsdk:"yml_path"`
	RepoID              types.String `tfsdk:"repo_id"`
	RepoType            types.String `tfsdk:"repo_type"`
	BranchName          types.String `tfsdk:"branch_name"`
	ServiceConnectionID types.String `tfsdk:"service_connection_id"`
	GithubEnterpriseURL types.String `tfsdk:"github_enterprise_url"`
	URL                 types.String `tfsdk:"url"`
	ReportBuildStatus   types.Bool   `tfsdk:"report_build_status"`
}

type buildDefinitionVariableModel struct {
	Name          types.String `tfsdk:"name"`
	Value         types.String `tfsdk:"value"`
	SecretValue   types.String `tfsdk:"secret_value"`
	IsSecret      types.Bool   `tfsdk:"is_secret"`
	AllowOverride types.Bool   `tfsdk:"allow_override"`
}

type buildDefinitionCITriggerModel struct {
	UseYAML  types.Bool `tfsdk:"use_yaml"`
	Override types.List `tfsdk:"override"`
}

type buildDefinitionCIOverrideModel struct {
	Batch                        types.Bool   `tfsdk:"batch"`
	MaxConcurrentBuildsPerBranch types.Int64  `tfsdk:"max_concurrent_builds_per_branch"`
	PollingInterval              types.Int64  `tfsdk:"polling_interval"`
	PollingJobID                 types.String `tfsdk:"polling_job_id"`
	BranchFilter                 types.List   `tfsdk:"branch_filter"`
	PathFilter                   types.List   `tfsdk:"path_filter"`
}

type buildDefinitionFilterModel struct {
	Include types.List `tfsdk:"include"`
	Exclude types.List `tfsdk:"exclude"`
}

type buildDefinitionPRTriggerModel struct {
	UseYAML         types.Bool   `tfsdk:"use_yaml"`
	InitialBranch   types.String `tfsdk:"initial_branch"`
	CommentRequired types.String `tfsdk:"comment_required"`
	Override        types.List   `tfsdk:"override"`
	Forks           types.List   `tfsdk:"forks"`
}

type buildDefinitionPROverrideModel struct {
	AutoCancel   types.Bool `tfsdk:"auto_cancel"`
	BranchFilter types.List `tfsdk:"branch_filter"`
	PathFilter   types.List `tfsdk:"path_filter"`
}

type buildDefinitionForksModel struct {
	Enabled      types.Bool `tfsdk:"enabled"`
	ShareSecrets types.Bool `tfsdk:"share_secrets"`
}

type buildDefinitionModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	ProjectID             types.String `tfsdk:"project_id"`
	Revision              types.Int64  `tfsdk:"revision"`
	Path                  types.String `tfsdk:"path"`
	AgentPoolName         types.String `tfsdk:"agent_pool_name"`
	AgentSpecification    types.String `tfsdk:"agent_specification"`
	JobAuthorizationScope types.String `tfsdk:"job_authorization_scope"`
	QueueStatus           types.String `tfsdk:"queue_status"`
	SkipFirstRun          types.Bool   `tfsdk:"skip_first_run"`
	Variable              types.Set    `tfsdk:"variable"`
	Repository            types.List   `tfsdk:"repository"`
	CITrigger             types.List   `tfsdk:"ci_trigger"`
	PullRequestTrigger    types.List   `tfsdk:"pull_request_trigger"`
}

// ── Metadata / Schema ────────────────────────────────────────────────────────

func (r *BuildDefinitionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build_definition"
}

func (r *BuildDefinitionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// Shared filter sub-block used by ci_trigger.override and pull_request_trigger.override.
	filterNestedObj := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"include": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of branch/path patterns to include.",
			},
			"exclude": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "List of branch/path patterns to exclude.",
			},
		},
	}

	ciOverrideNestedObj := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"batch": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to batch changes while a build is in progress.",
				Default:     bdFwStaticBool(true),
			},
			"max_concurrent_builds_per_branch": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The maximum number of concurrent builds per branch.",
				Default:     bdFwStaticInt64(1),
			},
			"polling_interval": schema.Int64Attribute{
				Optional:    true,
				Description: "The polling interval in seconds.",
			},
			"polling_job_id": schema.StringAttribute{
				Computed:    true,
				Description: "The polling job ID (computed).",
				PlanModifiers: []planmodifier.String{
					bdFwUseStateForUnknown(),
				},
			},
			"branch_filter": schema.ListNestedAttribute{
				Required:     true,
				Description:  "Branch filter blocks.",
				NestedObject: filterNestedObj,
			},
			"path_filter": schema.ListNestedAttribute{
				Optional:     true,
				Description:  "Path filter blocks.",
				NestedObject: filterNestedObj,
			},
		},
	}

	prOverrideNestedObj := schema.NestedAttributeObject{
		Attributes: map[string]schema.Attribute{
			"auto_cancel": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to auto-cancel pending builds when new commits are pushed.",
				Default:     bdFwStaticBool(true),
			},
			"branch_filter": schema.ListNestedAttribute{
				Required:     true,
				Description:  "Branch filter blocks.",
				NestedObject: filterNestedObj,
			},
			"path_filter": schema.ListNestedAttribute{
				Optional:     true,
				Description:  "Path filter blocks.",
				NestedObject: filterNestedObj,
			},
		},
	}

	resp.Schema = schema.Schema{
		Description: "Manages a build (pipeline) definition within an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the build definition.",
				PlanModifiers: []planmodifier.String{
					bdFwUseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the build definition.",
				Validators: []validator.String{
					bdFwStringNotWhitespace{},
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project the build definition belongs to.",
				PlanModifiers: []planmodifier.String{
					bdFwRequiresReplace(),
				},
			},
			"revision": schema.Int64Attribute{
				Computed:    true,
				Description: "The revision number of the build definition.",
				PlanModifiers: []planmodifier.Int64{
					bdFwUseStateForUnknownInt64(),
				},
			},
			"path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: `The folder path of the build definition (e.g. "\\MyFolder"). Defaults to "\".`,
				Default:     bdFwStaticString(`\`),
				Validators: []validator.String{
					bdFwPathValidator{},
				},
			},
			"agent_pool_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the agent pool to use for builds.",
				Default:     bdFwStaticString("Azure Pipelines"),
			},
			"agent_specification": schema.StringAttribute{
				Optional:    true,
				Description: "The agent specification (image) to use for the build.",
			},
			"job_authorization_scope": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The authorization scope for the job. Valid values: `projectCollection`, `project`.",
				Default:     bdFwStaticString("projectCollection"),
				Validators: []validator.String{
					bdFwStringOneOf{values: []string{
						string(build.BuildAuthorizationScopeValues.ProjectCollection),
						string(build.BuildAuthorizationScopeValues.Project),
					}},
				},
			},
			"queue_status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The queue status of the build definition. Valid values: `enabled`, `paused`, `disabled`.",
				Default:     bdFwStaticString("enabled"),
				Validators: []validator.String{
					bdFwStringOneOf{values: []string{
						string(build.DefinitionQueueStatusValues.Enabled),
						string(build.DefinitionQueueStatusValues.Paused),
						string(build.DefinitionQueueStatusValues.Disabled),
					}},
				},
			},
			"skip_first_run": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, the first run of the pipeline is skipped after creation. (Promoted from features block — see docs/build-gap-matrix.md.)",
				Default:     bdFwStaticBool(false),
			},
			"variable": schema.SetNestedAttribute{
				Optional:    true,
				Description: "A set of variable blocks to set on the build definition.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the variable.",
							Validators: []validator.String{
								bdFwStringNotWhitespace{},
							},
						},
						"value": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The value of the variable.",
							Default:     bdFwStaticString(""),
						},
						"secret_value": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Sensitive:   true,
							Description: "The secret value of the variable (when is_secret is true).",
							Default:     bdFwStaticString(""),
						},
						"is_secret": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether the variable is secret.",
							Default:     bdFwStaticBool(false),
						},
						"allow_override": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether the variable can be overridden at queue time.",
							Default:     bdFwStaticBool(true),
						},
					},
				},
			},
			"repository": schema.ListNestedAttribute{
				Required:    true,
				Description: "The repository block describing the source repository for the pipeline.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"yml_path": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The path of the YAML pipeline file in the repository.",
							Default:     bdFwStaticString(""),
						},
						"repo_id": schema.StringAttribute{
							Required:    true,
							Description: "The ID of the repository.",
						},
						"repo_type": schema.StringAttribute{
							Required:    true,
							Description: "The type of repository (GitHub, TfsGit, Bitbucket, GitHubEnterprise, Git).",
							Validators: []validator.String{
								bdFwStringOneOf{values: []string{
									"GitHub", "TfsGit", "Bitbucket", "GitHubEnterprise", "Git",
								}},
							},
						},
						"branch_name": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The default branch of the repository.",
							Default:     bdFwStaticString("master"),
						},
						"service_connection_id": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The ID of the service connection to use for the repository.",
							Default:     bdFwStaticString(""),
						},
						"github_enterprise_url": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The URL of the GitHub Enterprise instance. Conflicts with url.",
							Default:     bdFwStaticString(""),
						},
						"url": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The URL of the repository. Conflicts with github_enterprise_url.",
							PlanModifiers: []planmodifier.String{
								bdFwUseStateForUnknown(),
							},
						},
						"report_build_status": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether to report the build status to the repository host.",
							Default:     bdFwStaticBool(true),
						},
					},
				},
			},
			"ci_trigger": schema.ListNestedAttribute{
				Optional:    true,
				Description: "A block for configuring the continuous integration trigger.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"use_yaml": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether to use the YAML-defined trigger. Conflicts with override.",
							Default:     bdFwStaticBool(false),
						},
						"override": schema.ListNestedAttribute{
							Optional:     true,
							Description:  "Override block for CI trigger settings (conflicts with use_yaml).",
							NestedObject: ciOverrideNestedObj,
						},
					},
				},
			},
			"pull_request_trigger": schema.ListNestedAttribute{
				Optional:    true,
				Description: "A block for configuring the pull request trigger.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"use_yaml": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether to use the YAML-defined PR trigger. Conflicts with override.",
							Default:     bdFwStaticBool(false),
						},
						"initial_branch": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The initial branch to use for PR triggers.",
							Default:     bdFwStaticString("Managed by Terraform"),
						},
						"comment_required": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Whether a comment is required to trigger the PR build. Valid values: ``, `All`, `NonTeamMembers`.",
							Default:     bdFwStaticString(""),
							Validators: []validator.String{
								bdFwStringOneOf{values: []string{"", "All", "NonTeamMembers"}},
							},
						},
						"override": schema.ListNestedAttribute{
							Optional:     true,
							Description:  "Override block for PR trigger settings (conflicts with use_yaml).",
							NestedObject: prOverrideNestedObj,
						},
						"forks": schema.ListNestedAttribute{
							Required:    true,
							Description: "Forks settings for the pull request trigger.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										Required:    true,
										Description: "Whether fork-based pull request builds are enabled.",
									},
									"share_secrets": schema.BoolAttribute{
										Required:    true,
										Description: "Whether secrets are shared with fork builds.",
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

// ── Provider data injection ───────────────────────────────────────────────────

func (r *BuildDefinitionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("betterado_build_definition: expected *client.AggregatedClient, got %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

// ── ValidateConfig ────────────────────────────────────────────────────────────

// ValidateConfig enforces cross-attribute constraints that cannot be expressed
// as individual attribute validators (ConflictsWith-equivalent).
func (r *BuildDefinitionResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// AC3: repository.github_enterprise_url and repository.url are mutually exclusive.
	// Fetch both values from the nested repository block.
	var gheURL types.String
	gheURLDiag := req.Config.GetAttribute(ctx, path.Root("repository").AtListIndex(0).AtName("github_enterprise_url"), &gheURL)
	if gheURLDiag.HasError() {
		// Block may not be present yet (unknown/null); skip validation.
		return
	}

	var repoURL types.String
	repoURLDiag := req.Config.GetAttribute(ctx, path.Root("repository").AtListIndex(0).AtName("url"), &repoURL)
	if repoURLDiag.HasError() {
		return
	}

	// Only report conflict if both are non-null, non-unknown, and non-empty.
	gheSet := !gheURL.IsNull() && !gheURL.IsUnknown() && gheURL.ValueString() != ""
	urlSet := !repoURL.IsNull() && !repoURL.IsUnknown() && repoURL.ValueString() != ""
	if gheSet && urlSet {
		resp.Diagnostics.AddAttributeError(
			path.Root("repository").AtListIndex(0).AtName("github_enterprise_url"),
			"Conflicting repository attributes",
			"repository.github_enterprise_url and repository.url are mutually exclusive. "+
				"Set only one of these attributes.",
		)
	}
}

// ── Create ───────────────────────────────────────────────────────────────────

func (r *BuildDefinitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_build_definition Create: provider client not configured")
		return
	}

	var model buildDefinitionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()

	def, err := r.expandBuildDefinitionFw(ctx, model, "")
	if err != nil {
		resp.Diagnostics.AddError("Expand error", err.Error())
		return
	}

	created, apiErr := r.client.BuildClient.CreateDefinition(r.client.Ctx, build.CreateDefinitionArgs{
		Definition: def,
		Project:    &projectID,
	})
	if apiErr != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("creating build definition: %s", apiErr))
		return
	}

	model.ID = types.StringValue(strconv.Itoa(*created.Id))
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// AC2: When skip_first_run = false, trigger the first pipeline run (SDKv2 parity).
	if !model.SkipFirstRun.ValueBool() {
		// Extract branch name from repository block.
		branchName := "master"
		if !model.Repository.IsNull() && !model.Repository.IsUnknown() {
			var repos []buildDefinitionRepositoryModel
			if diags := model.Repository.ElementsAs(ctx, &repos, false); !diags.HasError() && len(repos) > 0 {
				branchName = repos[0].BranchName.ValueString()
			}
		}
		branchName = strings.TrimPrefix(branchName, "refs/heads/")

		_, runErr := r.client.PipelinesClient.RunPipeline(r.client.Ctx, pipelines.RunPipelineArgs{
			Project:    converter.String(projectID),
			PipelineId: created.Id,
			RunParameters: &pipelines.RunPipelineParameters{
				Resources: &pipelines.RunResourcesParameters{
					Repositories: &map[string]pipelines.RepositoryResourceParameters{
						"self": {
							RefName: converter.String("refs/heads/" + branchName),
						},
					},
				},
			},
		})
		if runErr != nil {
			// Warn but do not fail — pipeline may not have a valid YAML file yet.
			resp.Diagnostics.AddWarning(
				"First pipeline run failed",
				fmt.Sprintf("First run of build definition failed: %s. "+
					"Try initializing the repository with a valid build definition file.", runErr),
			)
		}
	}

	// Read back to populate computed fields.
	if readErr := r.readIntoModel(ctx, model.ID.ValueString(), projectID, &model, resp.State.RemoveResource); readErr != nil {
		resp.Diagnostics.AddError("Read-after-create error", readErr.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ─────────────────────────────────────────────────────────────────────

func (r *BuildDefinitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_build_definition Read: provider client not configured")
		return
	}

	var model buildDefinitionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	defIDStr := model.ID.ValueString()

	if readErr := r.readIntoModel(ctx, defIDStr, projectID, &model, resp.State.RemoveResource); readErr != nil {
		resp.Diagnostics.AddError("Read error", readErr.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ───────────────────────────────────────────────────────────────────

func (r *BuildDefinitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_build_definition Update: provider client not configured")
		return
	}

	var plan buildDefinitionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state buildDefinitionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ProjectID.ValueString()
	defIDStr := state.ID.ValueString()
	defID, err := strconv.Atoi(defIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("betterado_build_definition: invalid ID %q: %s", defIDStr, err))
		return
	}

	def, expandErr := r.expandBuildDefinitionFw(ctx, plan, defIDStr)
	if expandErr != nil {
		resp.Diagnostics.AddError("Expand error", expandErr.Error())
		return
	}

	_, updateErr := r.client.BuildClient.UpdateDefinition(r.client.Ctx, build.UpdateDefinitionArgs{
		Definition:   def,
		Project:      &projectID,
		DefinitionId: &defID,
	})
	if updateErr != nil {
		// Handle stale revision: re-read and retry once (matches SDKv2 pattern).
		if strings.Contains(updateErr.Error(), "400") || strings.Contains(updateErr.Error(), "revision") {
			reDef, readErr := r.client.BuildClient.GetDefinition(r.client.Ctx, build.GetDefinitionArgs{
				Project:      &projectID,
				DefinitionId: &defID,
			})
			if readErr != nil {
				resp.Diagnostics.AddError("Update error (re-read)", fmt.Sprintf("re-reading for stale-revision retry: %s", readErr))
				return
			}
			if reDef.Revision != nil {
				def.Revision = reDef.Revision
			}
			_, retryErr := r.client.BuildClient.UpdateDefinition(r.client.Ctx, build.UpdateDefinitionArgs{
				Definition:   def,
				Project:      &projectID,
				DefinitionId: &defID,
			})
			if retryErr != nil {
				resp.Diagnostics.AddError("Update error (retry)", fmt.Sprintf("updating build definition after revision refresh: %s", retryErr))
				return
			}
		} else {
			resp.Diagnostics.AddError("Update error", fmt.Sprintf("updating build definition: %s", updateErr))
			return
		}
	}

	plan.ID = state.ID
	if readErr := r.readIntoModel(ctx, defIDStr, projectID, &plan, resp.State.RemoveResource); readErr != nil {
		resp.Diagnostics.AddError("Read-after-update error", readErr.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ───────────────────────────────────────────────────────────────────

func (r *BuildDefinitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_build_definition Delete: provider client not configured")
		return
	}

	var model buildDefinitionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	defIDStr := model.ID.ValueString()
	defID, err := strconv.Atoi(defIDStr)
	if err != nil {
		resp.Diagnostics.AddError("Invalid ID", fmt.Sprintf("betterado_build_definition: invalid ID %q: %s", defIDStr, err))
		return
	}

	if apiErr := r.client.BuildClient.DeleteDefinition(r.client.Ctx, build.DeleteDefinitionArgs{
		Project:      &projectID,
		DefinitionId: &defID,
	}); apiErr != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("deleting build definition (id: %s): %s", defIDStr, apiErr))
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

// ImportState supports `terraform import betterado_build_definition.x "<project_id>/<definition_id>"`.
func (r *BuildDefinitionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := req.ID
	slashIdx := strings.Index(id, "/")
	if slashIdx < 0 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected format '<project_id>/<definition_id>', got %q", id),
		)
		return
	}
	projectID := id[:slashIdx]
	defIDStr := id[slashIdx+1:]

	model := buildDefinitionModel{
		ID:        types.StringValue(defIDStr),
		ProjectID: types.StringValue(projectID),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Private helpers ──────────────────────────────────────────────────────────

// readIntoModel fetches the build definition from ADO and populates model fields.
// removeResource is called when the resource no longer exists (404).
func (r *BuildDefinitionResource) readIntoModel(
	ctx context.Context,
	defIDStr, projectID string,
	model *buildDefinitionModel,
	removeResource func(context.Context),
) error {
	defID, err := strconv.Atoi(defIDStr)
	if err != nil {
		return fmt.Errorf("invalid build definition ID %q: %w", defIDStr, err)
	}

	def, apiErr := r.client.BuildClient.GetDefinition(r.client.Ctx, build.GetDefinitionArgs{
		Project:      &projectID,
		DefinitionId: &defID,
	})
	if apiErr != nil {
		if utils.ResponseWasNotFound(apiErr) {
			removeResource(r.client.Ctx)
			return nil
		}
		return fmt.Errorf("reading build definition (id: %s): %w", defIDStr, apiErr)
	}

	model.ID = types.StringValue(strconv.Itoa(*def.Id))
	if def.Name != nil {
		model.Name = types.StringValue(*def.Name)
	}
	model.ProjectID = types.StringValue(projectID)

	if def.Path != nil {
		model.Path = types.StringValue(*def.Path)
	}

	if def.Revision != nil {
		model.Revision = types.Int64Value(int64(*def.Revision))
	}

	if def.Queue != nil && def.Queue.Pool != nil && def.Queue.Pool.Name != nil {
		model.AgentPoolName = types.StringValue(*def.Queue.Pool.Name)
	}

	if def.JobAuthorizationScope != nil {
		model.JobAuthorizationScope = types.StringValue(string(*def.JobAuthorizationScope))
	}

	if def.QueueStatus != nil {
		model.QueueStatus = types.StringValue(string(*def.QueueStatus))
	}

	// Read repository fields back.
	if def.Repository != nil {
		repo := def.Repository
		repoModel := buildDefinitionRepositoryModel{}
		if repo.Id != nil {
			repoModel.RepoID = types.StringValue(*repo.Id)
		} else {
			repoModel.RepoID = types.StringValue("")
		}
		if repo.Type != nil {
			repoModel.RepoType = types.StringValue(*repo.Type)
		} else {
			repoModel.RepoType = types.StringValue("")
		}
		if repo.DefaultBranch != nil {
			repoModel.BranchName = types.StringValue(*repo.DefaultBranch)
		} else {
			repoModel.BranchName = types.StringValue("")
		}
		if repo.Url != nil {
			repoModel.URL = types.StringValue(*repo.Url)
		} else {
			repoModel.URL = types.StringValue("")
		}
		// Properties: connectedServiceId, reportBuildStatus.
		if repo.Properties != nil {
			props := *repo.Properties
			if sc, ok := props["connectedServiceId"]; ok {
				repoModel.ServiceConnectionID = types.StringValue(sc)
			} else {
				repoModel.ServiceConnectionID = types.StringValue("")
			}
			if rbs, ok := props["reportBuildStatus"]; ok {
				repoModel.ReportBuildStatus = types.BoolValue(rbs == "true")
			} else {
				repoModel.ReportBuildStatus = types.BoolValue(true)
			}
		} else {
			repoModel.ServiceConnectionID = types.StringValue("")
			repoModel.ReportBuildStatus = types.BoolValue(true)
		}
		// YmlPath from process.
		if def.Process != nil {
			if yamlProc, ok := def.Process.(*build.YamlProcess); ok && yamlProc.YamlFilename != nil {
				repoModel.YmlPath = types.StringValue(*yamlProc.YamlFilename)
			} else {
				repoModel.YmlPath = types.StringValue("")
			}
		} else {
			repoModel.YmlPath = types.StringValue("")
		}
		// github_enterprise_url: not returned by API — preserve from state.
		if !model.Repository.IsNull() && !model.Repository.IsUnknown() {
			var existingRepos []buildDefinitionRepositoryModel
			if diags := model.Repository.ElementsAs(ctx, &existingRepos, false); !diags.HasError() && len(existingRepos) > 0 {
				repoModel.GithubEnterpriseURL = existingRepos[0].GithubEnterpriseURL
			} else {
				repoModel.GithubEnterpriseURL = types.StringValue("")
			}
		} else {
			repoModel.GithubEnterpriseURL = types.StringValue("")
		}

		// Build the list value.
		repoObjType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"yml_path":              types.StringType,
				"repo_id":               types.StringType,
				"repo_type":             types.StringType,
				"branch_name":           types.StringType,
				"service_connection_id": types.StringType,
				"github_enterprise_url": types.StringType,
				"url":                   types.StringType,
				"report_build_status":   types.BoolType,
			},
		}
		repoListVal, diags := types.ListValueFrom(ctx, repoObjType, []buildDefinitionRepositoryModel{repoModel})
		if diags.HasError() {
			return fmt.Errorf("converting repository to list value")
		}
		model.Repository = repoListVal
	}

	// Read variable fields back.
	if def.Variables != nil && len(*def.Variables) > 0 {
		// Build a map from existing state for secret_value (API never returns secret values).
		existingSecrets := map[string]string{}
		if !model.Variable.IsNull() && !model.Variable.IsUnknown() {
			var existingVars []buildDefinitionVariableModel
			if diags := model.Variable.ElementsAs(ctx, &existingVars, false); !diags.HasError() {
				for _, ev := range existingVars {
					existingSecrets[ev.Name.ValueString()] = ev.SecretValue.ValueString()
				}
			}
		}
		vars := make([]buildDefinitionVariableModel, 0, len(*def.Variables))
		for name, v := range *def.Variables {
			vm := buildDefinitionVariableModel{
				Name:          types.StringValue(name),
				IsSecret:      types.BoolValue(v.IsSecret != nil && *v.IsSecret),
				AllowOverride: types.BoolValue(v.AllowOverride == nil || *v.AllowOverride),
			}
			if v.IsSecret != nil && *v.IsSecret {
				// Secret: API returns empty value; preserve from state.
				vm.Value = types.StringValue("")
				if sv, ok := existingSecrets[name]; ok {
					vm.SecretValue = types.StringValue(sv)
				} else {
					vm.SecretValue = types.StringValue("")
				}
			} else {
				if v.Value != nil {
					vm.Value = types.StringValue(*v.Value)
				} else {
					vm.Value = types.StringValue("")
				}
				vm.SecretValue = types.StringValue("")
			}
			vars = append(vars, vm)
		}
		varObjType := types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":           types.StringType,
				"value":          types.StringType,
				"secret_value":   types.StringType,
				"is_secret":      types.BoolType,
				"allow_override": types.BoolType,
			},
		}
		varSetVal, diags := types.SetValueFrom(ctx, varObjType, vars)
		if diags.HasError() {
			return fmt.Errorf("converting variables to set value")
		}
		model.Variable = varSetVal
	}
	// If API returns no variables but we had some, leave model.Variable as-is from plan
	// (empty set from defaults will be used).

	// agent_specification and skip_first_run: ADO doesn't reliably echo these back
	// on read, so we preserve the values already in the model (from plan/state).

	// AC1: Parse def.Triggers back into model.CITrigger / model.PullRequestTrigger
	// so that ADO-side trigger changes surface as drift on terraform plan.
	if def.Triggers != nil {
		if err := r.flattenTriggersIntoModel(ctx, *def.Triggers, model); err != nil {
			return fmt.Errorf("parsing triggers from API response: %w", err)
		}
	}

	return nil
}

// flattenTriggersIntoModel parses the raw API trigger list ([]interface{} where each
// element is a map[string]interface{}) back into model.CITrigger and
// model.PullRequestTrigger, mirroring the SDKv2
// flattenBuildDefinitionContinuousIntegrationTrigger /
// flattenBuildDefinitionPullRequestTrigger logic.
func (r *BuildDefinitionResource) flattenTriggersIntoModel(ctx context.Context, triggers []interface{}, model *buildDefinitionModel) error {
	// Shared attr type maps.
	filterObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"include": types.ListType{ElemType: types.StringType},
			"exclude": types.ListType{ElemType: types.StringType},
		},
	}
	filterListType := types.ListType{ElemType: filterObjType}

	ciOverrideObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"batch":                            types.BoolType,
			"max_concurrent_builds_per_branch": types.Int64Type,
			"polling_interval":                 types.Int64Type,
			"polling_job_id":                   types.StringType,
			"branch_filter":                    filterListType,
			"path_filter":                      filterListType,
		},
	}
	ciObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"use_yaml": types.BoolType,
			"override": types.ListType{ElemType: ciOverrideObjType},
		},
	}
	prOverrideObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"auto_cancel":   types.BoolType,
			"branch_filter": filterListType,
			"path_filter":   filterListType,
		},
	}
	forksObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"enabled":       types.BoolType,
			"share_secrets": types.BoolType,
		},
	}
	prObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"use_yaml":         types.BoolType,
			"initial_branch":   types.StringType,
			"comment_required": types.StringType,
			"override":         types.ListType{ElemType: prOverrideObjType},
			"forks":            types.ListType{ElemType: forksObjType},
		},
	}

	var ciTriggers []buildDefinitionCITriggerModel
	var prTriggers []buildDefinitionPRTriggerModel

	for _, raw := range triggers {
		ms, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		triggerType, _ := ms["triggerType"].(string)

		isYaml := false
		if sst, ok := ms["settingsSourceType"]; ok {
			switch v := sst.(type) {
			case float64:
				isYaml = int(v) == 2
			case int:
				isYaml = v == 2
			}
		}

		switch {
		case strings.EqualFold(triggerType, string(build.DefinitionTriggerTypeValues.ContinuousIntegration)):
			ct := buildDefinitionCITriggerModel{
				UseYAML: types.BoolValue(isYaml),
			}
			if !isYaml {
				// Build override sub-block.
				batchVal, _ := ms["batchChanges"].(bool)
				var maxConc int64
				if mc, ok := ms["maxConcurrentBuildsPerBranch"]; ok {
					switch v := mc.(type) {
					case float64:
						maxConc = int64(v)
					case int:
						maxConc = int64(v)
					}
				}
				var pollingInterval int64
				var pollingIntervalKnown bool
				if pi, ok := ms["pollingInterval"]; ok && pi != nil {
					switch v := pi.(type) {
					case float64:
						pollingInterval = int64(v)
						pollingIntervalKnown = true
					case int:
						pollingInterval = int64(v)
						pollingIntervalKnown = true
					}
				}
				pollingJobID := ""
				if pj, ok := ms["pollingJobId"]; ok && pj != nil {
					if s, ok := pj.(string); ok {
						pollingJobID = s
					}
				}

				branchFilterList, err := bdFwFlattenFilters(ctx, ms["branchFilters"], filterObjType, filterListType)
				if err != nil {
					return fmt.Errorf("ci_trigger branch_filter: %w", err)
				}
				pathFilterList, err := bdFwFlattenFilters(ctx, ms["pathFilters"], filterObjType, filterListType)
				if err != nil {
					return fmt.Errorf("ci_trigger path_filter: %w", err)
				}

				ovModel := buildDefinitionCIOverrideModel{
					Batch:                        types.BoolValue(batchVal),
					MaxConcurrentBuildsPerBranch: types.Int64Value(maxConc),
					BranchFilter:                 branchFilterList,
					PathFilter:                   pathFilterList,
					PollingJobID:                 types.StringValue(pollingJobID),
				}
				if pollingIntervalKnown {
					ovModel.PollingInterval = types.Int64Value(pollingInterval)
				} else {
					ovModel.PollingInterval = types.Int64Null()
				}

				overrideListVal, diags := types.ListValueFrom(ctx, ciOverrideObjType, []buildDefinitionCIOverrideModel{ovModel})
				if diags.HasError() {
					return fmt.Errorf("ci_trigger override list: diags error")
				}
				ct.Override = overrideListVal
			} else {
				ct.Override = types.ListValueMust(ciOverrideObjType, []attr.Value{})
			}
			ciTriggers = append(ciTriggers, ct)

		case strings.EqualFold(triggerType, string(build.DefinitionTriggerTypeValues.PullRequest)):
			pt := buildDefinitionPRTriggerModel{}
			pt.UseYAML = types.BoolValue(isYaml)

			// initial_branch: first branch filter entry (strip leading "+").
			initialBranch := "Managed by Terraform"
			if bfs, ok := ms["branchFilters"]; ok {
				if bfSlice, ok := bfs.([]interface{}); ok && len(bfSlice) > 0 {
					if s, ok := bfSlice[0].(string); ok {
						initialBranch = strings.TrimPrefix(s, "+")
					}
				}
			}
			pt.InitialBranch = types.StringValue(initialBranch)

			// comment_required.
			commentRequired := ""
			isCommentReq, _ := ms["isCommentRequiredForPullRequest"].(bool)
			isCommentReqNonTeam, _ := ms["requireCommentsForNonTeamMembersOnly"].(bool)
			if isCommentReq {
				commentRequired = "All"
			}
			if isCommentReq && isCommentReqNonTeam {
				commentRequired = "NonTeamMembers"
			}
			// Try requireCommentsStrategy (newer API field).
			if rcs, ok := ms["requireCommentsStrategy"]; ok && rcs != nil {
				if s, ok := rcs.(string); ok && s != "" {
					commentRequired = s
				}
			}
			pt.CommentRequired = types.StringValue(commentRequired)

			// forks.
			forksEnabled := false
			shareSecrets := false
			if forks, ok := ms["forks"]; ok {
				if forkMap, ok := forks.(map[string]interface{}); ok {
					forksEnabled, _ = forkMap["enabled"].(bool)
					shareSecrets, _ = forkMap["allowSecrets"].(bool)
				}
			}
			forksModel := buildDefinitionForksModel{
				Enabled:      types.BoolValue(forksEnabled),
				ShareSecrets: types.BoolValue(shareSecrets),
			}
			forksListVal, diags := types.ListValueFrom(ctx, forksObjType, []buildDefinitionForksModel{forksModel})
			if diags.HasError() {
				return fmt.Errorf("pull_request_trigger forks list: diags error")
			}
			pt.Forks = forksListVal

			// override (only when not use_yaml).
			if !isYaml {
				autoCancel := true
				if ac, ok := ms["autoCancel"]; ok {
					autoCancel, _ = ac.(bool)
				}
				branchFilterList, err := bdFwFlattenFilters(ctx, ms["branchFilters"], filterObjType, filterListType)
				if err != nil {
					return fmt.Errorf("pull_request_trigger branch_filter: %w", err)
				}
				pathFilterList, err := bdFwFlattenFilters(ctx, ms["pathFilters"], filterObjType, filterListType)
				if err != nil {
					return fmt.Errorf("pull_request_trigger path_filter: %w", err)
				}
				ovModel := buildDefinitionPROverrideModel{
					AutoCancel:   types.BoolValue(autoCancel),
					BranchFilter: branchFilterList,
					PathFilter:   pathFilterList,
				}
				overrideListVal, diags := types.ListValueFrom(ctx, prOverrideObjType, []buildDefinitionPROverrideModel{ovModel})
				if diags.HasError() {
					return fmt.Errorf("pull_request_trigger override list: diags error")
				}
				pt.Override = overrideListVal
			} else {
				pt.Override = types.ListValueMust(prOverrideObjType, []attr.Value{})
			}

			prTriggers = append(prTriggers, pt)
		}
	}

	// Write CI trigger back to model if found in API response.
	if len(ciTriggers) > 0 {
		ciListVal, diags := types.ListValueFrom(ctx, ciObjType, ciTriggers)
		if diags.HasError() {
			return fmt.Errorf("ci_trigger list value: diags error")
		}
		model.CITrigger = ciListVal
	} else if !model.CITrigger.IsNull() {
		// API returned no CI trigger — clear model to surface drift.
		model.CITrigger = types.ListValueMust(ciObjType, []attr.Value{})
	}

	// Write PR trigger back to model if found in API response.
	if len(prTriggers) > 0 {
		prListVal, diags := types.ListValueFrom(ctx, prObjType, prTriggers)
		if diags.HasError() {
			return fmt.Errorf("pull_request_trigger list value: diags error")
		}
		model.PullRequestTrigger = prListVal
	} else if !model.PullRequestTrigger.IsNull() {
		// API returned no PR trigger — clear model to surface drift.
		model.PullRequestTrigger = types.ListValueMust(prObjType, []attr.Value{})
	}

	return nil
}

// bdFwFlattenFilters converts an API branch/path filter slice ([]string with "+"/"-" prefix)
// into a types.List of filterModel objects.
func bdFwFlattenFilters(ctx context.Context, raw interface{}, filterObjType types.ObjectType, filterListType types.ListType) (types.List, error) {
	emptyList := types.ListValueMust(filterObjType, []attr.Value{})
	if raw == nil {
		return emptyList, nil
	}
	rawSlice, ok := raw.([]interface{})
	if !ok || len(rawSlice) == 0 {
		return emptyList, nil
	}

	var include []string
	var exclude []string
	for _, v := range rawSlice {
		s, ok := v.(string)
		if !ok {
			continue
		}
		if strings.HasPrefix(s, "-") {
			exclude = append(exclude, strings.TrimPrefix(s, "-"))
		} else {
			include = append(include, strings.TrimPrefix(s, "+"))
		}
	}

	inclListVal, diags := types.ListValueFrom(ctx, types.StringType, include)
	if diags.HasError() {
		return emptyList, fmt.Errorf("include filter list")
	}
	exclListVal, diags := types.ListValueFrom(ctx, types.StringType, exclude)
	if diags.HasError() {
		return emptyList, fmt.Errorf("exclude filter list")
	}

	fm := buildDefinitionFilterModel{
		Include: inclListVal,
		Exclude: exclListVal,
	}
	listVal, diags := types.ListValueFrom(ctx, filterObjType, []buildDefinitionFilterModel{fm})
	if diags.HasError() {
		return emptyList, fmt.Errorf("filter model list")
	}
	_ = filterListType
	return listVal, nil
}

// expandBuildDefinitionFw converts the model to an ADO BuildDefinition struct.
func (r *BuildDefinitionResource) expandBuildDefinitionFw(ctx context.Context, model buildDefinitionModel, idStr string) (*build.BuildDefinition, error) {
	projectID := model.ProjectID.ValueString()

	// Parse repository.
	var repos []buildDefinitionRepositoryModel
	if !model.Repository.IsNull() && !model.Repository.IsUnknown() {
		diags := model.Repository.ElementsAs(ctx, &repos, false)
		if diags.HasError() {
			return nil, fmt.Errorf("expanding repository attributes")
		}
	}
	if len(repos) != 1 {
		return nil, fmt.Errorf("exactly one repository block is required; got %d", len(repos))
	}
	repo := repos[0]

	repoID := repo.RepoID.ValueString()
	repoType := repo.RepoType.ValueString()
	repoURL := repo.URL.ValueString()
	branchName := repo.BranchName.ValueString()
	ymlPath := repo.YmlPath.ValueString()
	serviceConnID := repo.ServiceConnectionID.ValueString()
	reportStatus := repo.ReportBuildStatus.ValueBool()
	gheURL := repo.GithubEnterpriseURL.ValueString()

	if repoType == "GitHub" && repoURL == "" {
		repoURL = fmt.Sprintf("https://github.com/%s.git", repoID)
	}
	if gheURL != "" {
		repoURL = gheURL
	}

	// Compute reference ID for updates.
	var defIDRef *int
	if idStr != "" {
		n, err := strconv.Atoi(idStr)
		if err == nil {
			defIDRef = &n
		}
	}

	revision := int(model.Revision.ValueInt64())

	queueStatusStr := model.QueueStatus.ValueString()
	if queueStatusStr == "" {
		queueStatusStr = "enabled"
	}
	queueStatus := build.DefinitionQueueStatus(queueStatusStr)

	def := build.BuildDefinition{
		Id:       defIDRef,
		Name:     converter.String(model.Name.ValueString()),
		Path:     converter.String(model.Path.ValueString()),
		Revision: converter.Int(revision),
		Repository: &build.BuildRepository{
			Url:           &repoURL,
			Id:            &repoID,
			Name:          &repoID,
			DefaultBranch: &branchName,
			Type:          &repoType,
			Properties: &map[string]string{
				"connectedServiceId": serviceConnID,
				"reportBuildStatus":  strconv.FormatBool(reportStatus),
			},
		},
		Process: &build.YamlProcess{
			YamlFilename: &ymlPath,
		},
		QueueStatus: &queueStatus,
		Type:        &build.DefinitionTypeValues.Build,
		Quality:     &build.DefinitionQualityValues.Definition,
	}

	if agentPool := model.AgentPoolName.ValueString(); agentPool != "" {
		def.Queue = &build.AgentPoolQueue{
			Name: &agentPool,
			Pool: &build.TaskAgentPoolReference{
				Name: &agentPool,
			},
		}
	}

	if scope := model.JobAuthorizationScope.ValueString(); scope != "" {
		s := build.BuildAuthorizationScope(scope)
		def.JobAuthorizationScope = &s
	}

	// agent_specification: only relevant for OtherGit/jobs-based pipelines (deferred subtree).
	// For YAML pipelines it is stored inside the Process/target map. Since jobs are
	// deliberately deferred in this framework implementation, agent_specification is
	// accepted in schema (for future use) but not wired to the API object here.
	// See docs/build-gap-matrix.md for the NOT-migrated note.

	// Wire variables.
	if !model.Variable.IsNull() && !model.Variable.IsUnknown() {
		var vars []buildDefinitionVariableModel
		if diags := model.Variable.ElementsAs(ctx, &vars, false); !diags.HasError() && len(vars) > 0 {
			varMap := make(map[string]build.BuildDefinitionVariable, len(vars))
			for _, v := range vars {
				name := v.Name.ValueString()
				isSecret := v.IsSecret.ValueBool()
				allowOverride := v.AllowOverride.ValueBool()
				bdv := build.BuildDefinitionVariable{
					IsSecret:      &isSecret,
					AllowOverride: &allowOverride,
				}
				if isSecret {
					sv := v.SecretValue.ValueString()
					bdv.Value = &sv
				} else {
					val := v.Value.ValueString()
					bdv.Value = &val
				}
				varMap[name] = bdv
			}
			def.Variables = &varMap
		}
	}

	// Wire ci_trigger.
	var allTriggers []interface{}
	if !model.CITrigger.IsNull() && !model.CITrigger.IsUnknown() {
		var ciTriggers []buildDefinitionCITriggerModel
		if diags := model.CITrigger.ElementsAs(ctx, &ciTriggers, false); !diags.HasError() && len(ciTriggers) > 0 {
			ct := ciTriggers[0]
			switch {
			case ct.UseYAML.ValueBool():
				allTriggers = append(allTriggers, map[string]interface{}{
					"triggerType":        "continuousIntegration",
					"settingsSourceType": 2,
				})
			case !ct.Override.IsNull() && !ct.Override.IsUnknown():
				var overrides []buildDefinitionCIOverrideModel
				if diags2 := ct.Override.ElementsAs(ctx, &overrides, false); !diags2.HasError() && len(overrides) > 0 {
					ov := overrides[0]
					trigger := map[string]interface{}{
						"triggerType":                  "continuousIntegration",
						"batchChanges":                 ov.Batch.ValueBool(),
						"maxConcurrentBuildsPerBranch": int(ov.MaxConcurrentBuildsPerBranch.ValueInt64()),
						"branchFilters":                expandBdFwFilters(ctx, ov.BranchFilter),
						"pathFilters":                  expandBdFwFilters(ctx, ov.PathFilter),
					}
					if !ov.PollingInterval.IsNull() && !ov.PollingInterval.IsUnknown() {
						trigger["pollingInterval"] = int(ov.PollingInterval.ValueInt64())
					}
					allTriggers = append(allTriggers, trigger)
				}
			default:
				// Bare ci_trigger block (no override, no use_yaml) — use_yaml defaults false.
				allTriggers = append(allTriggers, map[string]interface{}{
					"triggerType":        "continuousIntegration",
					"settingsSourceType": 1,
				})
			}
		}
	}

	// Wire pull_request_trigger.
	if !model.PullRequestTrigger.IsNull() && !model.PullRequestTrigger.IsUnknown() {
		var prTriggers []buildDefinitionPRTriggerModel
		if diags := model.PullRequestTrigger.ElementsAs(ctx, &prTriggers, false); !diags.HasError() && len(prTriggers) > 0 {
			pt := prTriggers[0]
			trigger := map[string]interface{}{
				"triggerType": "pullRequest",
				"autoCancel":  true,
			}
			if pt.UseYAML.ValueBool() {
				trigger["settingsSourceType"] = 2
			} else {
				trigger["settingsSourceType"] = 1
			}
			if cr := pt.CommentRequired.ValueString(); cr != "" {
				trigger["requireCommentsStrategy"] = cr
			}
			// Forks (Required in SDKv2 parity).
			if !pt.Forks.IsNull() && !pt.Forks.IsUnknown() {
				var forks []buildDefinitionForksModel
				if diags2 := pt.Forks.ElementsAs(ctx, &forks, false); !diags2.HasError() && len(forks) > 0 {
					f := forks[0]
					trigger["forks"] = map[string]interface{}{
						"enabled":      f.Enabled.ValueBool(),
						"allowSecrets": f.ShareSecrets.ValueBool(),
					}
				}
			}
			// Override.
			if !pt.Override.IsNull() && !pt.Override.IsUnknown() {
				var overrides []buildDefinitionPROverrideModel
				if diags2 := pt.Override.ElementsAs(ctx, &overrides, false); !diags2.HasError() && len(overrides) > 0 {
					ov := overrides[0]
					trigger["autoCancel"] = ov.AutoCancel.ValueBool()
					trigger["branchFilters"] = expandBdFwFilters(ctx, ov.BranchFilter)
					trigger["pathFilters"] = expandBdFwFilters(ctx, ov.PathFilter)
				}
			}
			allTriggers = append(allTriggers, trigger)
		}
	}

	if len(allTriggers) > 0 {
		def.Triggers = &allTriggers
	}

	_ = projectID // captured in caller
	return &def, nil
}

// expandBdFwFilters converts a list of filter models to ADO branch/path filter strings.
func expandBdFwFilters(ctx context.Context, filterList types.List) []string {
	if filterList.IsNull() || filterList.IsUnknown() {
		return nil
	}
	var filters []buildDefinitionFilterModel
	if diags := filterList.ElementsAs(ctx, &filters, false); diags.HasError() {
		return nil
	}
	var result []string
	for _, f := range filters {
		if !f.Include.IsNull() && !f.Include.IsUnknown() {
			var incl []string
			if diags := f.Include.ElementsAs(ctx, &incl, false); !diags.HasError() {
				for _, v := range incl {
					result = append(result, "+"+v)
				}
			}
		}
		if !f.Exclude.IsNull() && !f.Exclude.IsUnknown() {
			var excl []string
			if diags := f.Exclude.ElementsAs(ctx, &excl, false); !diags.HasError() {
				for _, v := range excl {
					result = append(result, "-"+v)
				}
			}
		}
	}
	return result
}

// ── Inline defaults ───────────────────────────────────────────────────────────
// Mirrors the pattern in azuredevops/internal/service/build/resource_build_folder_framework.go.
// Prefixed bdFw* to avoid conflicts with the bf* prefixed modifiers in that file.

// bdFwStaticStringImpl is a minimal defaults.String that returns a constant.
type bdFwStaticStringImpl struct{ value string }

func bdFwStaticString(v string) defaults.String { return bdFwStaticStringImpl{value: v} }

func (d bdFwStaticStringImpl) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d bdFwStaticStringImpl) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%q`", d.value)
}

func (d bdFwStaticStringImpl) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

// bdFwStaticBoolImpl is a minimal defaults.Bool that returns a constant.
type bdFwStaticBoolImpl struct{ value bool }

func bdFwStaticBool(v bool) defaults.Bool { return bdFwStaticBoolImpl{value: v} }

func (d bdFwStaticBoolImpl) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", d.value)
}

func (d bdFwStaticBoolImpl) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%v`", d.value)
}

func (d bdFwStaticBoolImpl) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.value)
}

// bdFwStaticInt64Impl is a minimal defaults.Int64 that returns a constant.
type bdFwStaticInt64Impl struct{ value int64 }

func bdFwStaticInt64(v int64) defaults.Int64 { return bdFwStaticInt64Impl{value: v} }

func (d bdFwStaticInt64Impl) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %d", d.value)
}

func (d bdFwStaticInt64Impl) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%d`", d.value)
}

func (d bdFwStaticInt64Impl) DefaultInt64(_ context.Context, _ defaults.Int64Request, resp *defaults.Int64Response) {
	resp.PlanValue = types.Int64Value(d.value)
}

// ── Inline plan-modifiers ─────────────────────────────────────────────────────

// bdFwUseStateForUnknownImpl is equivalent to stringplanmodifier.UseStateForUnknown().
type bdFwUseStateForUnknownImpl struct{}

func bdFwUseStateForUnknown() planmodifier.String { return bdFwUseStateForUnknownImpl{} }

func (bdFwUseStateForUnknownImpl) Description(_ context.Context) string {
	return "use prior state value for unknown"
}

func (bdFwUseStateForUnknownImpl) MarkdownDescription(_ context.Context) string {
	return "use prior state value for unknown"
}

func (bdFwUseStateForUnknownImpl) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// bdFwRequiresReplaceImpl is equivalent to stringplanmodifier.RequiresReplace().
type bdFwRequiresReplaceImpl struct{}

func bdFwRequiresReplace() planmodifier.String { return bdFwRequiresReplaceImpl{} }

func (bdFwRequiresReplaceImpl) Description(_ context.Context) string {
	return "requires replacement if changed"
}

func (bdFwRequiresReplaceImpl) MarkdownDescription(_ context.Context) string {
	return "requires replacement if changed"
}

func (bdFwRequiresReplaceImpl) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// bdFwUseStateForUnknownInt64Impl is equivalent to int64planmodifier.UseStateForUnknown().
type bdFwUseStateForUnknownInt64Impl struct{}

func bdFwUseStateForUnknownInt64() planmodifier.Int64 { return bdFwUseStateForUnknownInt64Impl{} }

func (bdFwUseStateForUnknownInt64Impl) Description(_ context.Context) string {
	return "use prior state value for unknown"
}

func (bdFwUseStateForUnknownInt64Impl) MarkdownDescription(_ context.Context) string {
	return "use prior state value for unknown"
}

func (bdFwUseStateForUnknownInt64Impl) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ── Validators ────────────────────────────────────────────────────────────────

// bdFwStringNotWhitespace rejects empty/whitespace strings (SDKv2 StringIsNotWhiteSpace parity).
type bdFwStringNotWhitespace struct{}

func (v bdFwStringNotWhitespace) Description(_ context.Context) string {
	return "value must not be empty or whitespace"
}

func (v bdFwStringNotWhitespace) MarkdownDescription(_ context.Context) string {
	return "value must not be empty or whitespace"
}

func (v bdFwStringNotWhitespace) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if strings.TrimSpace(req.ConfigValue.ValueString()) == "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Value must not be whitespace",
			"The value must contain at least one non-whitespace character.")
	}
}

// bdFwStringOneOf rejects values not in the allowed set (SDKv2 StringInSlice parity).
type bdFwStringOneOf struct{ values []string }

func (v bdFwStringOneOf) Description(_ context.Context) string {
	return fmt.Sprintf("value must be one of: %s", strings.Join(v.values, ", "))
}

func (v bdFwStringOneOf) MarkdownDescription(_ context.Context) string {
	return v.Description(context.Background())
}

func (v bdFwStringOneOf) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	for _, allowed := range v.values {
		if val == allowed {
			return
		}
	}
	resp.Diagnostics.AddAttributeError(req.Path, "Invalid value",
		fmt.Sprintf("value must be one of %v, got: %q", v.values, val))
}

// bdFwPathValidator validates that a build definition folder path is correctly formatted.
// Rules (matching SDKv2 validate.Path):
//   - must not be empty
//   - must start with backslash
//   - must not end with backslash (unless it IS the root "\" path)
//   - must not contain any of: < > | : $ @ " / % + * ?
var bdFwInvalidPathRe = regexp.MustCompile(`[<>|:$@"/%+*?]`)

type bdFwPathValidator struct{}

func (v bdFwPathValidator) Description(_ context.Context) string {
	return `path must start with \ and must not contain <, >, |, :, $, @, ", /, %, +, *, or ?`
}

func (v bdFwPathValidator) MarkdownDescription(_ context.Context) string {
	return v.Description(context.Background())
}

func (v bdFwPathValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	val := req.ConfigValue.ValueString()
	if len(val) == 0 {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid path", "path must not be empty")
		return
	}
	if val[0] != '\\' {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid path", "path must start with a backslash")
		return
	}
	if len(val) >= 2 && val[len(val)-1] == '\\' {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid path", "path must not end with a backslash when not the root path")
		return
	}
	if bdFwInvalidPathRe.MatchString(val) {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid path",
			`path must not contain <, >, |, :, $, @, ", /, %, +, *, or ?`)
	}
}
