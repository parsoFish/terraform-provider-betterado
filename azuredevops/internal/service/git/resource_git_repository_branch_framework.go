package git

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/tfhelper"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = &gitRepositoryBranchResource{}
	_ resource.ResourceWithConfigure   = &gitRepositoryBranchResource{}
	_ resource.ResourceWithImportState = &gitRepositoryBranchResource{}
)

// NewGitRepositoryBranchResource is the factory function registered with the framework provider.
func NewGitRepositoryBranchResource() resource.Resource {
	return &gitRepositoryBranchResource{}
}

// gitRepositoryBranchResource implements resource.Resource for betterado_git_repository_branch.
type gitRepositoryBranchResource struct {
	clients *client.AggregatedClient
}

// gitRepositoryBranchModel is the Terraform state model.
type gitRepositoryBranchModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	RepositoryID types.String `tfsdk:"repository_id"`
	RefBranch    types.String `tfsdk:"ref_branch"`
	RefTag       types.String `tfsdk:"ref_tag"`
	RefCommitID  types.String `tfsdk:"ref_commit_id"`
	LastCommitID types.String `tfsdk:"last_commit_id"`
}

// ---------- Metadata / Schema ------------------------------------------------

func (r *gitRepositoryBranchResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_repository_branch"
}

func (r *gitRepositoryBranchResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Git repository branch in Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					gitStateString(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the branch (short format, without refs/heads/ prefix).",
				PlanModifiers: []planmodifier.String{
					gitRequiresReplace(),
				},
			},
			"repository_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Git repository.",
				PlanModifiers: []planmodifier.String{
					gitRequiresReplace(),
				},
			},
			"ref_branch": schema.StringAttribute{
				Optional:    true,
				Description: "The branch name to branch from (mutually exclusive with ref_tag and ref_commit_id).",
				PlanModifiers: []planmodifier.String{
					gitRequiresReplace(),
				},
			},
			"ref_tag": schema.StringAttribute{
				Optional:    true,
				Description: "The tag name to branch from (mutually exclusive with ref_branch and ref_commit_id).",
				PlanModifiers: []planmodifier.String{
					gitRequiresReplace(),
				},
			},
			"ref_commit_id": schema.StringAttribute{
				Optional:    true,
				Description: "The commit ID to branch from (mutually exclusive with ref_branch and ref_tag).",
				PlanModifiers: []planmodifier.String{
					gitRequiresReplace(),
				},
			},
			"last_commit_id": schema.StringAttribute{
				Computed:    true,
				Description: "The commit ID of the last commit on this branch.",
				PlanModifiers: []planmodifier.String{
					gitStateString(),
				},
			},
		},
	}
}

// ---------- Configure --------------------------------------------------------

func (r *gitRepositoryBranchResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gitRepositoryBranchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan gitRepositoryBranchModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate mutual exclusion of ref_* attributes.
	if err := validateRefMutualExclusion(plan); err != nil {
		resp.Diagnostics.AddError("Invalid configuration", err.Error())
		return
	}

	branchName := plan.Name.ValueString()
	if strings.HasPrefix(branchName, REF_BRANCH_PREFIX) {
		resp.Diagnostics.AddError("Invalid branch name",
			fmt.Sprintf("Branch name must be in short format without refs/heads/ prefix, got: %q", branchName))
		return
	}

	repoID := plan.RepositoryID.ValueString()

	newObjectID, err := r.resolveRefToCommitID(repoID, plan)
	if err != nil {
		resp.Diagnostics.AddError("Error resolving ref", err.Error())
		return
	}

	if err := updateRefsFramework(r.clients, git.UpdateRefsArgs{
		RefUpdates: &[]git.GitRefUpdate{{
			Name:        converter.String(REF_BRANCH_PREFIX + branchName),
			NewObjectId: &newObjectID,
			OldObjectId: converter.String("0000000000000000000000000000000000000000"),
		}},
		RepositoryId: converter.String(repoID),
	}); err != nil {
		resp.Diagnostics.AddError("Error creating branch", fmt.Sprintf("Creating branch %q: %v", branchName, err))
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s:%s", repoID, branchName))

	// Read back to populate computed attributes.
	gotBranch, err := r.clients.GitReposClient.GetBranch(r.clients.Ctx, git.GetBranchArgs{
		RepositoryId: &repoID,
		Name:         &branchName,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading branch after create", err.Error())
		return
	}
	plan.LastCommitID = types.StringValue(*gotBranch.Commit.CommitId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ---------- Read -------------------------------------------------------------

func (r *gitRepositoryBranchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state gitRepositoryBranchModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repoID, branchName, err := tfhelper.ParseGitRepoBranchID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error parsing resource ID", err.Error())
		return
	}

	if strings.HasPrefix(branchName, REF_BRANCH_PREFIX) {
		resp.Diagnostics.AddError("Invalid branch name in state",
			fmt.Sprintf("Branch name must be in short format without refs/heads/ prefix, got: %q", branchName))
		return
	}

	gotBranch, err := r.clients.GitReposClient.GetBranch(r.clients.Ctx, git.GetBranchArgs{
		RepositoryId: &repoID,
		Name:         &branchName,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		var branchErr azuredevops.WrappedError
		if errors.As(err, &branchErr) {
			regx := regexp.MustCompile(fmt.Sprintf("Branch \"%[1]s\" does not exist in the %[2]s repository.", branchName, repoID))
			if branchErr.Message != nil && regx.MatchString(*branchErr.Message) {
				resp.State.RemoveResource(ctx)
				return
			}
		}
		resp.Diagnostics.AddError("Error reading branch", fmt.Sprintf("Reading branch %q: %v", branchName, err))
		return
	}

	state.Name = types.StringValue(branchName)
	state.RepositoryID = types.StringValue(repoID)
	state.LastCommitID = types.StringValue(*gotBranch.Commit.CommitId)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ---------- Update -----------------------------------------------------------

// Update is not supported — all fields are ForceNew (RequiresReplace).
// Terraform will destroy and recreate rather than call Update.
func (r *gitRepositoryBranchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported",
		"betterado_git_repository_branch does not support in-place updates; all attributes require replacement.")
}

// ---------- Delete -----------------------------------------------------------

func (r *gitRepositoryBranchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state gitRepositoryBranchModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repoID, branchName, err := tfhelper.ParseGitRepoBranchID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error parsing resource ID", err.Error())
		return
	}

	if strings.HasPrefix(branchName, REF_BRANCH_PREFIX) {
		resp.Diagnostics.AddError("Invalid branch name in state",
			fmt.Sprintf("Branch name must be in short format without refs/heads/ prefix, got: %q", branchName))
		return
	}

	gotBranch, err := r.clients.GitReposClient.GetBranch(r.clients.Ctx, git.GetBranchArgs{
		RepositoryId: &repoID,
		Name:         &branchName,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error getting branch for delete", fmt.Sprintf("Getting latest commit of %q: %v", branchName, err))
		return
	}

	if err := updateRefsFramework(r.clients, git.UpdateRefsArgs{
		RefUpdates: &[]git.GitRefUpdate{{
			Name:        converter.String(REF_BRANCH_PREFIX + branchName),
			OldObjectId: gotBranch.Commit.CommitId,
			NewObjectId: converter.String("0000000000000000000000000000000000000000"),
		}},
		RepositoryId: converter.String(repoID),
	}); err != nil {
		resp.Diagnostics.AddError("Error deleting branch", fmt.Sprintf("Deleting branch %q: %v", branchName, err))
		return
	}
}

// ---------- ImportState ------------------------------------------------------

func (r *gitRepositoryBranchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: "<repository_id>:<branch_name>"
	repoID, branchName, err := tfhelper.ParseGitRepoBranchID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Expected format '<repository_id>:<branch_name>', got: %q. Error: %v", req.ID, err))
		return
	}

	state := gitRepositoryBranchModel{
		ID:           types.StringValue(req.ID),
		Name:         types.StringValue(branchName),
		RepositoryID: types.StringValue(repoID),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ---------- Helpers ----------------------------------------------------------

// validateRefMutualExclusion ensures at most one of ref_branch, ref_tag, ref_commit_id is set.
func validateRefMutualExclusion(m gitRepositoryBranchModel) error {
	count := 0
	if !m.RefBranch.IsNull() && !m.RefBranch.IsUnknown() && m.RefBranch.ValueString() != "" {
		count++
	}
	if !m.RefTag.IsNull() && !m.RefTag.IsUnknown() && m.RefTag.ValueString() != "" {
		count++
	}
	if !m.RefCommitID.IsNull() && !m.RefCommitID.IsUnknown() && m.RefCommitID.ValueString() != "" {
		count++
	}
	if count > 1 {
		return fmt.Errorf("only one of ref_branch, ref_tag, or ref_commit_id may be specified")
	}
	return nil
}

// resolveRefToCommitID resolves whichever ref_* attribute is set to a commit ID.
func (r *gitRepositoryBranchResource) resolveRefToCommitID(repoID string, m gitRepositoryBranchModel) (string, error) {
	// ref_commit_id is already a commit ID — use directly.
	if !m.RefCommitID.IsNull() && !m.RefCommitID.IsUnknown() && m.RefCommitID.ValueString() != "" {
		return m.RefCommitID.ValueString(), nil
	}

	var refSpec string
	switch {
	case !m.RefBranch.IsNull() && !m.RefBranch.IsUnknown() && m.RefBranch.ValueString() != "":
		refSpec = withPrefix(REF_BRANCH_PREFIX, m.RefBranch.ValueString())
	case !m.RefTag.IsNull() && !m.RefTag.IsUnknown() && m.RefTag.ValueString() != "":
		refSpec = withPrefix(REF_TAG_PREFIX, m.RefTag.ValueString())
	default:
		return "", fmt.Errorf("one of ref_branch, ref_tag, or ref_commit_id must be specified")
	}

	filter := strings.TrimPrefix(refSpec, "refs/")
	gotRefs, err := r.clients.GitReposClient.GetRefs(r.clients.Ctx, git.GetRefsArgs{
		RepositoryId: converter.String(repoID),
		Filter:       converter.String(filter),
		Top:          converter.Int(1),
		PeelTags:     converter.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("getting refs matching %q: %w", filter, err)
	}
	if len(gotRefs.Value) == 0 {
		return "", fmt.Errorf("No refs found that match ref %q.", refSpec)
	}

	gotRef := gotRefs.Value[0]
	if gotRef.Name == nil {
		return "", fmt.Errorf("got unexpected GetRefs response, a ref without a name was returned")
	}
	// Require complete match to avoid prefix collisions.
	if *gotRef.Name != refSpec {
		return "", fmt.Errorf("ref %q not found, closest match is %q", filter, *gotRef.Name)
	}

	switch {
	case gotRef.PeeledObjectId != nil:
		return *gotRef.PeeledObjectId, nil
	case gotRef.ObjectId != nil:
		return *gotRef.ObjectId, nil
	default:
		return "", fmt.Errorf("GetRefs response doesn't have a valid commit id")
	}
}

// updateRefsFramework is the framework-layer equivalent of updateRefs (SDKv2).
func updateRefsFramework(clients *client.AggregatedClient, args git.UpdateRefsArgs) error {
	updateRefResults, err := clients.GitReposClient.UpdateRefs(clients.Ctx, args)
	if err != nil {
		return fmt.Errorf("updating refs: %w", err)
	}
	for _, refUpdate := range *updateRefResults {
		if !*refUpdate.Success {
			return fmt.Errorf("Update refs failed. Update Status: %s", *refUpdate.UpdateStatus)
		}
	}
	return nil
}
