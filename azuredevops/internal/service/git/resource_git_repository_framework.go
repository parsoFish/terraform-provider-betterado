package git

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/serviceendpoint"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &gitRepositoryResource{}
	_ resource.ResourceWithConfigure   = &gitRepositoryResource{}
	_ resource.ResourceWithImportState = &gitRepositoryResource{}
)

// ---- plan-modifier / default helpers ----------------------------------------

type gitStringDefault struct{ value string }

func gitDefaultString(v string) defaults.String { return gitStringDefault{v} }
func (d gitStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d gitStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", d.value)
}

func (d gitStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(d.value)
}

type gitBoolDefault struct{ value bool }

func gitDefaultBool(v bool) defaults.Bool { return gitBoolDefault{v} }
func (d gitBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", d.value)
}

func (d gitBoolDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to %v", d.value)
}

func (d gitBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.value)
}

type gitUseStateString struct{}

func gitStateString() planmodifier.String { return gitUseStateString{} }
func (m gitUseStateString) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m gitUseStateString) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m gitUseStateString) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() || req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type gitUseStateBoolMod struct{}

func gitStateBool() planmodifier.Bool { return gitUseStateBoolMod{} }
func (m gitUseStateBoolMod) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m gitUseStateBoolMod) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m gitUseStateBoolMod) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if !req.PlanValue.IsUnknown() || req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type gitUseStateInt64Mod struct{}

func gitStateInt64() planmodifier.Int64 { return gitUseStateInt64Mod{} }
func (m gitUseStateInt64Mod) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m gitUseStateInt64Mod) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m gitUseStateInt64Mod) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	if !req.PlanValue.IsUnknown() || req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

type gitRequiresReplaceMod struct{}

func gitRequiresReplace() planmodifier.String { return gitRequiresReplaceMod{} }
func (m gitRequiresReplaceMod) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m gitRequiresReplaceMod) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m gitRequiresReplaceMod) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() || req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ---- gitRepositoryResource --------------------------------------------------

type gitRepositoryResource struct {
	client *client.AggregatedClient
}

func NewGitRepositoryResource() resource.Resource {
	return &gitRepositoryResource{}
}

// ---- Model types ------------------------------------------------------------

type initializationModel struct {
	InitType            types.String `tfsdk:"init_type"`
	SourceType          types.String `tfsdk:"source_type"`
	SourceURL           types.String `tfsdk:"source_url"`
	ServiceConnectionID types.String `tfsdk:"service_connection_id"`
	Username            types.String `tfsdk:"username"`
	Password            types.String `tfsdk:"password"`
}

var initializationAttrTypes = map[string]attr.Type{
	"init_type":             types.StringType,
	"source_type":           types.StringType,
	"source_url":            types.StringType,
	"service_connection_id": types.StringType,
	"username":              types.StringType,
	"password":              types.StringType,
}

type gitRepositoryModel struct {
	ID                 types.String `tfsdk:"id"`
	ProjectID          types.String `tfsdk:"project_id"`
	Name               types.String `tfsdk:"name"`
	Initialization     types.List   `tfsdk:"initialization"`
	ParentRepositoryID types.String `tfsdk:"parent_repository_id"`
	DefaultBranch      types.String `tfsdk:"default_branch"`
	Disabled           types.Bool   `tfsdk:"disabled"`
	IsFork             types.Bool   `tfsdk:"is_fork"`
	RemoteURL          types.String `tfsdk:"remote_url"`
	Size               types.Int64  `tfsdk:"size"`
	SSHURL             types.String `tfsdk:"ssh_url"`
	URL                types.String `tfsdk:"url"`
	WebURL             types.String `tfsdk:"web_url"`
}

// ---- Metadata ---------------------------------------------------------------

func (r *gitRepositoryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_repository"
}

// ---- Schema -----------------------------------------------------------------

func (r *gitRepositoryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Git repository in Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{gitStateString()},
			},
			"project_id": schema.StringAttribute{
				Required:      true,
				Description:   "The ID of the project the repository belongs to.",
				PlanModifiers: []planmodifier.String{gitRequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the git repository.",
			},
			"parent_repository_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of a Git repository from which a fork is to be created.",
			},
			"default_branch": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "The ref of the default branch.",
				PlanModifiers: []planmodifier.String{gitStateString()},
			},
			"disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     gitDefaultBool(false),
				Description: "Whether the repository is disabled.",
			},
			"is_fork": schema.BoolAttribute{
				Computed:      true,
				Description:   "True if the repository was created as a fork.",
				PlanModifiers: []planmodifier.Bool{gitStateBool()},
			},
			"remote_url": schema.StringAttribute{
				Computed:      true,
				Description:   "Git HTTPS clone URL.",
				PlanModifiers: []planmodifier.String{gitStateString()},
			},
			"size": schema.Int64Attribute{
				Computed:      true,
				Description:   "Size of the repository in bytes.",
				PlanModifiers: []planmodifier.Int64{gitStateInt64()},
			},
			"ssh_url": schema.StringAttribute{
				Computed:      true,
				Description:   "Git SSH clone URL.",
				PlanModifiers: []planmodifier.String{gitStateString()},
			},
			"url": schema.StringAttribute{
				Computed:      true,
				Description:   "REST API URL for the repository.",
				PlanModifiers: []planmodifier.String{gitStateString()},
			},
			"web_url": schema.StringAttribute{
				Computed:      true,
				Description:   "Web URL for the repository.",
				PlanModifiers: []planmodifier.String{gitStateString()},
			},
		},
		Blocks: map[string]schema.Block{
			"initialization": schema.ListNestedBlock{
				Description: "How the repository is initialized.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"init_type": schema.StringAttribute{
							Required:    true,
							Description: "The type of repository initialization. Valid values are `Clean`, `Fork`, `Import`, `Uninitialized`.",
						},
						"source_type": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     gitDefaultString(""),
							Description: "Type of the source repository. Only `Git` is supported.",
						},
						"source_url": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     gitDefaultString(""),
							Description: "The URL of the source repository.",
						},
						"service_connection_id": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     gitDefaultString(""),
							Description: "The ID of a service connection for authentication when importing.",
						},
						"username": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     gitDefaultString(""),
							Description: "The username for importing a private repository.",
						},
						"password": schema.StringAttribute{
							Optional:    true,
							Sensitive:   true,
							Description: "The password for importing a private repository.",
						},
					},
				},
			},
		},
	}
}

// ---- Configure --------------------------------------------------------------

func (r *gitRepositoryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *client.AggregatedClient, got %T", req.ProviderData))
		return
	}
	r.client = c
}

// ---- ImportState ------------------------------------------------------------

func (r *gitRepositoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	var state gitRepositoryModel
	if len(parts) == 2 {
		state.ProjectID = types.StringValue(parts[0])
		state.ID = types.StringValue(parts[1])
	} else {
		state.ID = types.StringValue(req.ID)
		state.ProjectID = types.StringValue("")
	}
	state.Name = types.StringUnknown()
	state.DefaultBranch = types.StringUnknown()
	state.Disabled = types.BoolUnknown()
	state.IsFork = types.BoolUnknown()
	state.RemoteURL = types.StringUnknown()
	state.Size = types.Int64Unknown()
	state.SSHURL = types.StringUnknown()
	state.URL = types.StringUnknown()
	state.WebURL = types.StringUnknown()
	state.ParentRepositoryID = types.StringNull()
	state.Initialization = types.ListValueMust(types.ObjectType{AttrTypes: initializationAttrTypes}, []attr.Value{})
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ---- Create -----------------------------------------------------------------

func (r *gitRepositoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured with Azure DevOps credentials.")
		return
	}
	var plan gitRepositoryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	initMeta, diags := expandInitialization(ctx, plan.Initialization)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !plan.DefaultBranch.IsNull() && !plan.DefaultBranch.IsUnknown() && plan.DefaultBranch.ValueString() != "" {
		if strings.EqualFold(initMeta.initType, string(RepoInitTypeValues.Uninitialized)) {
			resp.Diagnostics.AddError("Invalid configuration",
				"Repository 'initialization.init_type = Uninitialized', there will be no branches, 'default_branch' cannot be set.")
			return
		}
	}
	projectID, err := uuid.Parse(plan.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid project_id", err.Error())
		return
	}
	createArgs := git.CreateRepositoryArgs{
		GitRepositoryToCreate: &git.GitRepositoryCreateOptions{
			Name:    converter.String(plan.Name.ValueString()),
			Project: &core.TeamProjectReference{Id: &projectID},
		},
	}
	if strings.EqualFold(initMeta.initType, string(RepoInitTypeValues.Fork)) {
		if !plan.ParentRepositoryID.IsNull() && !plan.ParentRepositoryID.IsUnknown() {
			pID := plan.ParentRepositoryID.ValueString()
			parentRepo, err := gitRepositoryRead(r.client, pID, "", "")
			if err != nil {
				resp.Diagnostics.AddError("Failed to read parent repository", fmt.Sprintf("ID: %s, Error: %+v", pID, err))
				return
			}
			createArgs.GitRepositoryToCreate.ParentRepository = &git.GitRepositoryRef{
				Id:      parentRepo.Id,
				Name:    parentRepo.Name,
				Project: parentRepo.Project,
			}
		}
	}
	createdRepo, err := r.client.GitReposClient.CreateRepository(r.client.Ctx, createArgs)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create repository", err.Error())
		return
	}
	if !plan.DefaultBranch.IsNull() && !plan.DefaultBranch.IsUnknown() && plan.DefaultBranch.ValueString() != "" {
		createdRepo.DefaultBranch = converter.String(plan.DefaultBranch.ValueString())
	}
	if err = r.initializeRepository(initMeta, createdRepo, projectID.String()); err != nil {
		resp.Diagnostics.AddError("Failed to initialize repository", err.Error())
		return
	}
	if !strings.EqualFold(initMeta.initType, string(RepoInitTypeValues.Uninitialized)) {
		if err = waitForBranch(r.client, createdRepo.Name, &projectID, 10*time.Minute); err != nil {
			resp.Diagnostics.AddError("Waiting for repository branch", err.Error())
			return
		}
	}
	if !plan.DefaultBranch.IsNull() && !plan.DefaultBranch.IsUnknown() && plan.DefaultBranch.ValueString() != "" {
		createdRepo.DefaultBranch = converter.String(plan.DefaultBranch.ValueString())
		if _, err = updateGitRepository(r.client, createdRepo, &projectID); err != nil {
			resp.Diagnostics.AddError("Failed to set default branch", err.Error())
			return
		}
	}
	if !plan.Disabled.IsNull() && !plan.Disabled.IsUnknown() && plan.Disabled.ValueBool() {
		if _, err = updateIsDisabledGitRepository(r.client, createdRepo.Id.String(), projectID.String(), true); err != nil {
			resp.Diagnostics.AddError("Failed to disable repository", err.Error())
			return
		}
	}
	repo, err := gitRepositoryRead(r.client, createdRepo.Id.String(), plan.Name.ValueString(), projectID.String())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read repository after create", err.Error())
		return
	}
	state := flattenGitRepositoryFramework(repo, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ---- Read -------------------------------------------------------------------

func (r *gitRepositoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured with Azure DevOps credentials.")
		return
	}
	var state gitRepositoryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	repo, err := gitRepositoryRead(r.client, state.ID.ValueString(), state.Name.ValueString(), state.ProjectID.ValueString())
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			tflog.Warn(ctx, fmt.Sprintf("Git repository %s not found, removing from state", state.ID.ValueString()))
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read repository",
			fmt.Sprintf("ID: %s, Error: %v", state.ID.ValueString(), err))
		return
	}
	if repo == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	newState := flattenGitRepositoryFramework(repo, state)
	newState.ID = types.StringValue(repo.Id.String())
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// ---- Update -----------------------------------------------------------------

func (r *gitRepositoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured with Azure DevOps credentials.")
		return
	}
	var plan, state gitRepositoryModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	initMeta, diags := expandInitialization(ctx, plan.Initialization)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	projectID, err := uuid.Parse(plan.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid project_id", err.Error())
		return
	}
	repoID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid repository id", err.Error())
		return
	}
	disabled := plan.Disabled.ValueBool()
	repoExist, err := gitRepositoryRead(r.client, repoID.String(), plan.Name.ValueString(), projectID.String())
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read repository", err.Error())
		return
	}
	// Enable before update: Disabled -> Enabled
	if converter.ToBool(repoExist.IsDisabled, false) && !disabled {
		repoExist, err = updateIsDisabledGitRepository(r.client, repoID.String(), projectID.String(), false)
		if err != nil {
			resp.Diagnostics.AddError("Failed to enable repository", err.Error())
			return
		}
	}
	if converter.ToBool(repoExist.IsDisabled, false) {
		resp.Diagnostics.AddError("Cannot update disabled repository",
			fmt.Sprintf("Repository %s is disabled. Enable it before updating.", repoID))
		return
	}
	// Initialize if not yet initialized
	if (repoExist.DefaultBranch == nil || *repoExist.DefaultBranch == "") &&
		(repoExist.Size == nil || *repoExist.Size == 0) {
		repo := &git.GitRepository{
			Id:            &repoID,
			Name:          converter.String(plan.Name.ValueString()),
			DefaultBranch: converter.String(plan.DefaultBranch.ValueString()),
			Project:       repoExist.Project,
		}
		if err = r.initializeRepository(initMeta, repo, projectID.String()); err != nil {
			resp.Diagnostics.AddError("Failed to initialize repository", err.Error())
			return
		}
	}
	updateRepo := &git.GitRepository{
		Id:            &repoID,
		Name:          converter.String(plan.Name.ValueString()),
		DefaultBranch: converter.String(plan.DefaultBranch.ValueString()),
	}
	if _, err = updateGitRepository(r.client, updateRepo, &projectID); err != nil {
		resp.Diagnostics.AddError("Failed to update repository", err.Error())
		return
	}
	// Disable after update: Enabled -> Disabled
	if !converter.ToBool(repoExist.IsDisabled, false) && disabled {
		if _, err = updateIsDisabledGitRepository(r.client, repoID.String(), projectID.String(), true); err != nil {
			resp.Diagnostics.AddError("Failed to disable repository", err.Error())
			return
		}
	}
	repo, err := gitRepositoryRead(r.client, repoID.String(), plan.Name.ValueString(), projectID.String())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read repository after update", err.Error())
		return
	}
	newState := flattenGitRepositoryFramework(repo, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

// ---- Delete -----------------------------------------------------------------

func (r *gitRepositoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured with Azure DevOps credentials.")
		return
	}
	var state gitRepositoryModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	repoActual, err := gitRepositoryRead(r.client, state.ID.ValueString(), state.Name.ValueString(), state.ProjectID.ValueString())
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to read repository before delete", err.Error())
		return
	}
	if repoActual == nil {
		return
	}
	if converter.ToBool(repoActual.IsDisabled, false) {
		resp.Diagnostics.AddError("Cannot delete disabled repository",
			fmt.Sprintf("Repository %s is disabled. Enable it before deleting.", state.ID.ValueString()))
		return
	}
	repoUUID, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid repository ID",
			fmt.Sprintf("Cannot parse %s as UUID: %+v", state.ID.ValueString(), err))
		return
	}
	if err = r.client.GitReposClient.DeleteRepository(r.client.Ctx, git.DeleteRepositoryArgs{
		RepositoryId: &repoUUID,
	}); err != nil {
		resp.Diagnostics.AddError("Failed to delete repository", err.Error())
	}
}

// ---- Helpers ----------------------------------------------------------------

func expandInitialization(ctx context.Context, initList types.List) (*repoInitializationMeta, diag.Diagnostics) {
	var diags diag.Diagnostics
	if initList.IsNull() || initList.IsUnknown() || len(initList.Elements()) == 0 {
		return &repoInitializationMeta{}, diags
	}
	var blocks []initializationModel
	diags.Append(initList.ElementsAs(ctx, &blocks, false)...)
	if diags.HasError() {
		return nil, diags
	}
	b := blocks[0]
	meta := &repoInitializationMeta{
		initType:            b.InitType.ValueString(),
		sourceType:          b.SourceType.ValueString(),
		sourceURL:           b.SourceURL.ValueString(),
		serviceConnectionID: b.ServiceConnectionID.ValueString(),
		userName:            b.Username.ValueString(),
		password:            b.Password.ValueString(),
	}
	if strings.EqualFold(meta.initType, "clean") {
		meta.sourceType = ""
		meta.sourceURL = ""
		meta.serviceConnectionID = ""
		meta.userName = ""
		meta.password = ""
	}
	return meta, diags
}

func flattenGitRepositoryFramework(repo *git.GitRepository, prev gitRepositoryModel) gitRepositoryModel {
	m := gitRepositoryModel{
		ID:                 types.StringValue(repo.Id.String()),
		Name:               types.StringValue(converter.ToString(repo.Name, "")),
		DefaultBranch:      types.StringValue(converter.ToString(repo.DefaultBranch, "")),
		Disabled:           types.BoolValue(converter.ToBool(repo.IsDisabled, false)),
		IsFork:             types.BoolValue(converter.ToBool(repo.IsFork, false)),
		RemoteURL:          types.StringValue(converter.ToString(repo.RemoteUrl, "")),
		SSHURL:             types.StringValue(converter.ToString(repo.SshUrl, "")),
		URL:                types.StringValue(converter.ToString(repo.Url, "")),
		WebURL:             types.StringValue(converter.ToString(repo.WebUrl, "")),
		Initialization:     prev.Initialization,
		ParentRepositoryID: prev.ParentRepositoryID,
	}
	if repo.Project != nil && repo.Project.Id != nil {
		m.ProjectID = types.StringValue(repo.Project.Id.String())
	} else {
		m.ProjectID = prev.ProjectID
	}
	if repo.Size != nil {
		m.Size = types.Int64Value(int64(*repo.Size))
	} else {
		m.Size = types.Int64Value(0)
	}
	return m
}

func (r *gitRepositoryResource) initializeRepository(initialization *repoInitializationMeta, repository *git.GitRepository, projectID string) error {
	if initialization == nil {
		return nil
	}
	if strings.EqualFold(initialization.initType, string(RepoInitTypeValues.Import)) && strings.EqualFold(initialization.sourceType, "Git") {
		importRequest := git.GitImportRequest{
			Parameters: &git.GitImportRequestParameters{
				GitSource: &git.GitImportGitSource{Url: &initialization.sourceURL},
			},
			Repository: repository,
		}
		if initialization.serviceConnectionID != "" {
			importRequest.Parameters.ServiceEndpointId = converter.UUID(initialization.serviceConnectionID)
			importRequest.Parameters.DeleteServiceEndpointAfterImportIsDone = converter.Bool(false)
		} else if initialization.userName != "" || initialization.password != "" {
			seName := fmt.Sprintf("Repository Import (%s)", uuid.New().String())
			se, err := r.client.ServiceEndpointClient.CreateServiceEndpoint(
				r.client.Ctx,
				serviceendpoint.CreateServiceEndpointArgs{
					Endpoint: &serviceendpoint.ServiceEndpoint{
						Authorization: &serviceendpoint.EndpointAuthorization{
							Parameters: &map[string]string{
								"username": initialization.userName,
								"password": initialization.password,
							},
							Scheme: converter.String("UsernamePassword"),
						},
						Name:  &seName,
						Type:  converter.String("git"),
						Url:   &initialization.sourceURL,
						Owner: converter.String("library"),
						ServiceEndpointProjectReferences: &[]serviceendpoint.ServiceEndpointProjectReference{
							{
								ProjectReference: &serviceendpoint.ProjectReference{
									Id: converter.ToPtr(uuid.MustParse(projectID)),
								},
								Name: &seName,
							},
						},
					},
				},
			)
			if err != nil {
				return err
			}
			importRequest.Parameters.ServiceEndpointId = se.Id
			importRequest.Parameters.DeleteServiceEndpointAfterImportIsDone = converter.Bool(true)
		}
		_, importErr := r.client.GitReposClient.CreateImportRequest(r.client.Ctx, git.CreateImportRequestArgs{
			ImportRequest: &importRequest,
			Project:       &projectID,
			RepositoryId:  repository.Name,
		})
		if importErr != nil {
			var wrapperError *azuredevops.WrappedError
			if errors.As(importErr, &wrapperError) {
				if wrapperError.StatusCode != nil && *wrapperError.StatusCode == http.StatusBadRequest {
					return fmt.Errorf("Import repository in Azure DevOps: %+v \nImport request cannot be processed due to one of the following reasons:\n\n\tClone URL is incorrect.\n\tClone URL requires authorization.\n", importErr)
				}
			}
			cleanupErr := r.client.GitReposClient.DeleteRepository(r.client.Ctx, git.DeleteRepositoryArgs{
				RepositoryId: repository.Id,
			})
			if cleanupErr != nil {
				return fmt.Errorf("Creating repository in Azure DevOps: %+v", cleanupErr)
			}
			return fmt.Errorf("Import repository in Azure DevOps: %+v ", importErr)
		}
	}
	if strings.EqualFold(initialization.initType, string(RepoInitTypeValues.Clean)) {
		if err := initializeGitRepository(r.client, repository, repository.DefaultBranch); err != nil {
			return fmt.Errorf("initializing repository in Azure DevOps: %+v", err)
		}
	}
	return nil
}
