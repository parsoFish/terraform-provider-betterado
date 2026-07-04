package git

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &gitRepositoryFileResource{}
	_ resource.ResourceWithConfigure   = &gitRepositoryFileResource{}
	_ resource.ResourceWithImportState = &gitRepositoryFileResource{}
)

// NewGitRepositoryFileResource is the factory registered with the framework provider.
func NewGitRepositoryFileResource() resource.Resource {
	return &gitRepositoryFileResource{}
}

// gitRepositoryFileResource implements resource.Resource for betterado_git_repository_file.
type gitRepositoryFileResource struct {
	clients *client.AggregatedClient
}

// gitRepositoryFileModel is the Terraform state model.
type gitRepositoryFileModel struct {
	ID                types.String `tfsdk:"id"`
	RepositoryID      types.String `tfsdk:"repository_id"`
	File              types.String `tfsdk:"file"`
	Content           types.String `tfsdk:"content"`
	Branch            types.String `tfsdk:"branch"`
	CommitMessage     types.String `tfsdk:"commit_message"`
	CommitterName     types.String `tfsdk:"committer_name"`
	CommitterEmail    types.String `tfsdk:"committer_email"`
	AuthorName        types.String `tfsdk:"author_name"`
	AuthorEmail       types.String `tfsdk:"author_email"`
	OverwriteOnCreate types.Bool   `tfsdk:"overwrite_on_create"`
}

// ---------- Metadata / Schema ------------------------------------------------

func (r *gitRepositoryFileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_repository_file"
}

func (r *gitRepositoryFileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a file within a Git repository in Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					gitStateString(),
				},
			},
			"repository_id": schema.StringAttribute{
				Required:    true,
				Description: "The repository ID.",
				PlanModifiers: []planmodifier.String{
					gitRequiresReplace(),
				},
			},
			"file": schema.StringAttribute{
				Required:    true,
				Description: "The file path to manage.",
				PlanModifiers: []planmodifier.String{
					gitRequiresReplace(),
				},
			},
			"content": schema.StringAttribute{
				Required:    true,
				Description: "The file's content.",
			},
			"branch": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: `The branch name, defaults to "refs/heads/master".`,
				Default:     gitDefaultString("refs/heads/master"),
				PlanModifiers: []planmodifier.String{
					gitRequiresReplace(),
				},
			},
			"commit_message": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The commit message when creating or updating the file.",
				PlanModifiers: []planmodifier.String{
					gitStateString(),
				},
			},
			"committer_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The committer name.",
				PlanModifiers: []planmodifier.String{
					gitStateString(),
				},
			},
			"committer_email": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The committer email.",
				PlanModifiers: []planmodifier.String{
					gitStateString(),
				},
			},
			"author_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The author name.",
				PlanModifiers: []planmodifier.String{
					gitStateString(),
				},
			},
			"author_email": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The author email.",
				PlanModifiers: []planmodifier.String{
					gitStateString(),
				},
			},
			"overwrite_on_create": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: `Enable overwriting existing files, defaults to "false".`,
				Default:     gitDefaultBool(false),
			},
		},
	}
}

// ---------- Configure --------------------------------------------------------

func (r *gitRepositoryFileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	clients, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.AggregatedClient, got: %T", req.ProviderData),
		)
		return
	}
	r.clients = clients
}

// ---------- Create -----------------------------------------------------------

func (r *gitRepositoryFileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.clients == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured with Azure DevOps credentials.")
		return
	}

	var plan gitRepositoryFileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repoID := plan.RepositoryID.ValueString()
	file := plan.File.ValueString()
	branch := plan.Branch.ValueString()
	overwriteOnCreate := plan.OverwriteOnCreate.ValueBool()

	ref, err := checkRepositoryBranchExists(r.clients, repoID, branch)
	if err != nil {
		resp.Diagnostics.AddError("Error checking branch", err.Error())
		return
	}
	if ref == nil {
		resp.Diagnostics.AddError("Branch not found", fmt.Sprintf("Creating Git file. Branch not found. Name: %s.", branch))
		return
	}

	version := shortBranchName(branch)
	repoItem, err := r.clients.GitReposClient.GetItem(ctx, git.GetItemArgs{
		RepositoryId: &repoID,
		Path:         &file,
		VersionDescriptor: &git.GitVersionDescriptor{
			Version:     &version,
			VersionType: &git.GitVersionTypeValues.Branch,
		},
	})
	if err != nil && !utils.ResponseWasNotFound(err) {
		resp.Diagnostics.AddError("Error checking file existence",
			fmt.Sprintf("Repository branch not found, repositoryID: %s, branch: %s. Error: %+v", repoID, branch, err))
		return
	}

	changeType := git.VersionControlChangeTypeValues.Add
	if repoItem != nil {
		if !overwriteOnCreate {
			resp.Diagnostics.AddError("File already exists",
				"Refusing to overwrite existing file. Configure `overwrite_on_create` to `true` to override.")
			return
		}
		changeType = git.VersionControlChangeTypeValues.Edit
	}

	timeout := 10 * time.Minute
	deadline := time.Now().Add(timeout)

	var createErr error
	for time.Now().Before(deadline) {
		objectID, err := getLastCommitId(r.clients, repoID, branch)
		if err != nil {
			resp.Diagnostics.AddError("Error getting last commit ID", err.Error())
			return
		}

		pushArgs := r.buildPushArgs(plan, objectID, changeType)
		commits := *pushArgs.Push.Commits
		if commits[0].Comment == nil {
			m := fmt.Sprintf("Add %s", file)
			commits[0].Comment = &m
			pushArgs.Push.Commits = &commits
		}

		_, createErr = r.clients.GitReposClient.CreatePush(ctx, *pushArgs)
		if createErr == nil {
			break
		}
		if !utils.ResponseContainsStatusMessage(createErr, "has already been updated by another client") {
			resp.Diagnostics.AddError("Error creating repository file",
				fmt.Sprintf("Create repository file failed, repositoryID: %s, branch: %s, file: %s. Error: %+v", repoID, branch, file, createErr))
			return
		}
		time.Sleep(2 * time.Second)
	}
	if createErr != nil {
		resp.Diagnostics.AddError("Error creating repository file (timeout)",
			fmt.Sprintf("Create repository file failed after retries, repositoryID: %s, branch: %s, file: %s. Error: %+v", repoID, branch, file, createErr))
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", repoID, file))

	diags := &resp.Diagnostics
	r.readIntoModel(ctx, &plan, diags)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ---------- Read -------------------------------------------------------------

func (r *gitRepositoryFileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.clients == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured with Azure DevOps credentials.")
		return
	}

	var state gitRepositoryFileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	gone := false
	diags := &resp.Diagnostics
	r.readIntoModelCheckGone(ctx, &state, diags, &gone)
	if resp.Diagnostics.HasError() {
		return
	}
	if gone {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ---------- Update -----------------------------------------------------------

func (r *gitRepositoryFileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.clients == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured with Azure DevOps credentials.")
		return
	}

	var plan gitRepositoryFileModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Carry over the stable ID from prior state.
	var state gitRepositoryFileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	repoID := plan.RepositoryID.ValueString()
	file := plan.File.ValueString()
	branch := plan.Branch.ValueString()

	_, err := checkRepositoryBranchExists(r.clients, repoID, branch)
	if err != nil {
		resp.Diagnostics.AddError("Error checking branch",
			fmt.Sprintf("Updating Git file. Failed to get repository branch. Repository ID: %s. Branch Name: %s. Error: %+v", repoID, branch, err))
		return
	}

	timeout := 10 * time.Minute
	deadline := time.Now().Add(timeout)

	var updateErr error
	for time.Now().Before(deadline) {
		objectID, err := getLastCommitId(r.clients, repoID, branch)
		if err != nil {
			resp.Diagnostics.AddError("Error getting last commit ID", err.Error())
			return
		}

		pushArgs := r.buildPushArgs(plan, objectID, git.VersionControlChangeTypeValues.Edit)
		commits := *pushArgs.Push.Commits
		if commits[0].Comment != nil && *commits[0].Comment == fmt.Sprintf("Add %s", file) {
			m := fmt.Sprintf("Update %s", file)
			commits[0].Comment = &m
			pushArgs.Push.Commits = &commits
		}

		_, updateErr = r.clients.GitReposClient.CreatePush(ctx, *pushArgs)
		if updateErr == nil {
			break
		}
		if !utils.ResponseContainsStatusMessage(updateErr, "has already been updated by another client") {
			resp.Diagnostics.AddError("Error updating repository file",
				fmt.Sprintf("Update repository file failed, repositoryID: %s, branch: %s, file: %s. Error: %+v", repoID, branch, file, updateErr))
			return
		}
		time.Sleep(2 * time.Second)
	}
	if updateErr != nil {
		resp.Diagnostics.AddError("Error updating repository file (timeout)",
			fmt.Sprintf("Update repository file failed after retries, repositoryID: %s, branch: %s, file: %s. Error: %+v", repoID, branch, file, updateErr))
		return
	}

	diags := &resp.Diagnostics
	r.readIntoModel(ctx, &plan, diags)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ---------- Delete -----------------------------------------------------------

func (r *gitRepositoryFileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.clients == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured with Azure DevOps credentials.")
		return
	}

	var state gitRepositoryFileModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repoID := state.RepositoryID.ValueString()
	file := state.File.ValueString()
	branch := state.Branch.ValueString()
	message := fmt.Sprintf("Delete %s", file)

	timeout := 10 * time.Minute
	deadline := time.Now().Add(timeout)

	var deleteErr error
	for time.Now().Before(deadline) {
		objectID, err := getLastCommitId(r.clients, repoID, branch)
		if err != nil {
			resp.Diagnostics.AddError("Error getting last commit ID", err.Error())
			return
		}

		change := &git.GitChange{
			ChangeType: &git.VersionControlChangeTypeValues.Delete,
			Item: git.GitItem{
				Path: &file,
			},
		}

		_, deleteErr = r.clients.GitReposClient.CreatePush(ctx, git.CreatePushArgs{
			RepositoryId: &repoID,
			Push: &git.GitPush{
				RefUpdates: &[]git.GitRefUpdate{
					{
						Name:        &branch,
						OldObjectId: &objectID,
					},
				},
				Commits: &[]git.GitCommitRef{
					{
						Author: &git.GitUserDate{
							Name:  converter.String(state.AuthorName.ValueString()),
							Email: converter.String(state.AuthorEmail.ValueString()),
						},
						Comment: &message,
						Changes: &[]interface{}{change},
					},
				},
			},
		})
		if deleteErr == nil {
			break
		}
		if !utils.ResponseContainsStatusMessage(deleteErr, "has already been updated by another client") {
			resp.Diagnostics.AddError("Error deleting repository file",
				fmt.Sprintf("Failed to destroy the repository file, repository ID: %s, branch: %s. file %s. Error %+v", repoID, branch, file, deleteErr))
			return
		}
		time.Sleep(2 * time.Second)
	}
	if deleteErr != nil {
		resp.Diagnostics.AddError("Error deleting repository file (timeout)",
			fmt.Sprintf("Failed to destroy the repository file after retries, repository ID: %s, branch: %s. file %s. Error %+v", repoID, branch, file, deleteErr))
	}
}

// ---------- ImportState ------------------------------------------------------

func (r *gitRepositoryFileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: "<repository_id>/<file_path>" or "<repository_id>/<file_path>:<branch>"
	parts := strings.Split(req.ID, ":")
	branch := "refs/heads/master"

	if len(parts) > 2 {
		resp.Diagnostics.AddError("Invalid import ID",
			`Invalid ID specified. Supplied ID must be written as <repository>/<file path> (when branch is "master") or <repository>/<file path>:<branch>`)
		return
	}
	if len(parts) == 2 {
		branch = parts[1]
	}

	repoID, file := splitRepoFilePath(parts[0])

	if err := checkRepositoryFileExists(r.clients, repoID, file, branch); err != nil {
		resp.Diagnostics.AddError("File not found",
			fmt.Sprintf("Repository not found, repository ID: %s, branch: %s, file: %s. Error: %+v", repoID, branch, file, err))
		return
	}

	state := gitRepositoryFileModel{
		ID:                types.StringValue(fmt.Sprintf("%s/%s", repoID, file)),
		RepositoryID:      types.StringValue(repoID),
		File:              types.StringValue(file),
		Branch:            types.StringValue(branch),
		OverwriteOnCreate: types.BoolValue(false),
		// Computed fields default to empty strings; Read will populate them
		Content:        types.StringValue(""),
		CommitMessage:  types.StringValue(""),
		CommitterName:  types.StringValue(""),
		CommitterEmail: types.StringValue(""),
		AuthorName:     types.StringValue(""),
		AuthorEmail:    types.StringValue(""),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ---------- Helpers ----------------------------------------------------------

// readIntoModel reads the current state of the file from Azure DevOps into m.
// Errors are appended to diags.
func (r *gitRepositoryFileResource) readIntoModel(ctx context.Context, m *gitRepositoryFileModel, diags *diag.Diagnostics) {
	gone := false
	r.readIntoModelCheckGone(ctx, m, diags, &gone)
}

// readIntoModelCheckGone is the internal implementation; sets *gone=true when
// the resource no longer exists (repository/branch/file deleted externally).
func (r *gitRepositoryFileResource) readIntoModelCheckGone(ctx context.Context, m *gitRepositoryFileModel, diags *diag.Diagnostics, gone *bool) {
	repoID, file := splitRepoFilePath(m.ID.ValueString())
	branch := m.Branch.ValueString()

	_, err := r.clients.GitReposClient.GetRepository(ctx, git.GetRepositoryArgs{
		RepositoryId: &repoID,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			*gone = true
			return
		}
		diags.AddError("Error reading repository",
			fmt.Sprintf("Get Git file. Repository not found, repositoryID: %s. Error: %+v", repoID, err))
		return
	}

	ref, err := checkRepositoryBranchExists(r.clients, repoID, branch)
	if err != nil {
		diags.AddError("Error checking branch",
			fmt.Sprintf("Get Git file. Failed to get repository branch. Repository ID: %s. Branch Name: %s. Error: %+v", repoID, branch, err))
		return
	}
	if ref == nil {
		*gone = true
		return
	}

	repoItem, err := r.clients.GitReposClient.GetItem(ctx, git.GetItemArgs{
		RepositoryId:   &repoID,
		Path:           &file,
		IncludeContent: converter.Bool(true),
		VersionDescriptor: &git.GitVersionDescriptor{
			Version:     converter.String(shortBranchName(branch)),
			VersionType: &git.GitVersionTypeValues.Branch,
		},
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			*gone = true
			return
		}
		diags.AddError("Error reading file",
			fmt.Sprintf("Query repository item failed, repositoryID: %s, branch: %s, file: %s. Error: %+v", repoID, branch, file, err))
		return
	}

	m.Content = types.StringValue(converter.ToString(repoItem.Content, ""))
	m.RepositoryID = types.StringValue(repoID)
	m.File = types.StringValue(file)

	commit, err := r.clients.GitReposClient.GetCommit(ctx, git.GetCommitArgs{
		RepositoryId: &repoID,
		CommitId:     repoItem.CommitId,
	})
	if err != nil {
		diags.AddError("Error reading commit",
			fmt.Sprintf("Get repository file commit failed, repositoryID: %s, branch: %s, file: %s. Error: %+v", repoID, branch, file, err))
		return
	}

	if commit.Committer != nil {
		if commit.Committer.Name != nil {
			m.CommitterName = types.StringValue(*commit.Committer.Name)
		}
		if commit.Committer.Email != nil {
			m.CommitterEmail = types.StringValue(*commit.Committer.Email)
		}
	}
	if commit.Author != nil {
		if commit.Author.Name != nil {
			m.AuthorName = types.StringValue(*commit.Author.Name)
		}
		if commit.Author.Email != nil {
			m.AuthorEmail = types.StringValue(*commit.Author.Email)
		}
	}
	if commit.Comment != nil {
		m.CommitMessage = types.StringValue(*commit.Comment)
	}
}

// buildPushArgs constructs the CreatePushArgs from the model.
func (r *gitRepositoryFileResource) buildPushArgs(m gitRepositoryFileModel, objectID string, changeType git.VersionControlChangeType) *git.CreatePushArgs {
	repoID := m.RepositoryID.ValueString()
	content := m.Content.ValueString()
	file := m.File.ValueString()
	branch := m.Branch.ValueString()

	var message *string
	if !m.CommitMessage.IsNull() && !m.CommitMessage.IsUnknown() && m.CommitMessage.ValueString() != "" {
		s := m.CommitMessage.ValueString()
		message = &s
	}

	change := git.GitChange{
		ChangeType: &changeType,
		Item: git.GitItem{
			Path: &file,
		},
		NewContent: &git.ItemContent{
			Content:     &content,
			ContentType: &git.ItemContentTypeValues.RawText,
		},
	}

	return &git.CreatePushArgs{
		RepositoryId: &repoID,
		Push: &git.GitPush{
			RefUpdates: &[]git.GitRefUpdate{
				{
					Name:        &branch,
					OldObjectId: &objectID,
				},
			},
			Commits: &[]git.GitCommitRef{
				{
					Author: &git.GitUserDate{
						Name:  converter.String(m.AuthorName.ValueString()),
						Email: converter.String(m.AuthorEmail.ValueString()),
					},
					Committer: &git.GitUserDate{
						Name:  converter.String(m.CommitterName.ValueString()),
						Email: converter.String(m.CommitterEmail.ValueString()),
					},
					Comment: message,
					Changes: &[]interface{}{change},
				},
			},
		},
	}
}
