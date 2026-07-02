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

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &BuildDefinitionResource{}
	_ resource.ResourceWithImportState = &BuildDefinitionResource{}
	_ resource.ResourceWithConfigure   = &BuildDefinitionResource{}
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

type buildDefinitionCITriggerModel struct {
	UseYaml types.Bool `tfsdk:"use_yaml"`
}

type buildDefinitionPRTriggerModel struct {
	UseYaml         types.Bool   `tfsdk:"use_yaml"`
	InitialBranch   types.String `tfsdk:"initial_branch"`
	CommentRequired types.String `tfsdk:"comment_required"`
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
			},
			"queue_status": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The queue status of the build definition. Valid values: `enabled`, `paused`, `disabled`.",
				Default:     bdFwStaticString("enabled"),
			},
			"skip_first_run": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "If true, the first run of the pipeline is skipped after creation.",
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
							Description: "The URL of the GitHub Enterprise instance.",
							Default:     bdFwStaticString(""),
						},
						"url": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "The URL of the repository.",
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
							Description: "Whether to use the YAML-defined trigger.",
							Default:     bdFwStaticBool(false),
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
							Description: "Whether to use the YAML-defined PR trigger.",
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
							Description: "Whether a comment is required to trigger the PR build. Valid values: '', 'All', 'NonTeamMembers'.",
							Default:     bdFwStaticString(""),
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
	_ context.Context,
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

	return nil
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

	if repoType == "GitHub" && repoURL == "" {
		repoURL = fmt.Sprintf("https://github.com/%s.git", repoID)
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

	_ = projectID // captured in caller
	return &def, nil
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
