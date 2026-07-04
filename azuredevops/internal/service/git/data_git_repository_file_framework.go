package git

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = &gitRepositoryFileDataSource{}
	_ datasource.DataSourceWithConfigure = &gitRepositoryFileDataSource{}
)

// NewGitRepositoryFileDataSource returns a new framework data source.
func NewGitRepositoryFileDataSource() datasource.DataSource {
	return &gitRepositoryFileDataSource{}
}

// gitRepositoryFileDataSource is the framework implementation of betterado_git_repository_file (data source).
type gitRepositoryFileDataSource struct {
	client *client.AggregatedClient
}

// gitRepositoryFileDataModel is the tfsdk model for the data source.
type gitRepositoryFileDataModel struct {
	ID                types.String `tfsdk:"id"`
	RepositoryID      types.String `tfsdk:"repository_id"`
	File              types.String `tfsdk:"file"`
	Branch            types.String `tfsdk:"branch"`
	Tag               types.String `tfsdk:"tag"`
	Content           types.String `tfsdk:"content"`
	LastCommitMessage types.String `tfsdk:"last_commit_message"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *gitRepositoryFileDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_repository_file"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *gitRepositoryFileDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a file from a Git repository in Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the data source.",
			},
			"repository_id": schema.StringAttribute{
				Required:    true,
				Description: "The repository ID.",
			},
			"file": schema.StringAttribute{
				Required:    true,
				Description: "The file path.",
			},
			"branch": schema.StringAttribute{
				Optional:    true,
				Description: "The branch name, no default.",
			},
			"tag": schema.StringAttribute{
				Optional:    true,
				Description: "Optional tag name, no default.",
			},
			"content": schema.StringAttribute{
				Computed:    true,
				Description: "The file's content.",
			},
			"last_commit_message": schema.StringAttribute{
				Computed:    true,
				Description: "The last commit message.",
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *gitRepositoryFileDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.AggregatedClient, got %T", req.ProviderData),
		)
		return
	}
	d.client = c
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (d *gitRepositoryFileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured with Azure DevOps credentials.")
		return
	}

	var config gitRepositoryFileDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	repoID := config.RepositoryID.ValueString()
	file := config.File.ValueString()

	// Validate exactly one of branch/tag is set.
	hasBranch := !config.Branch.IsNull() && !config.Branch.IsUnknown() && config.Branch.ValueString() != ""
	hasTag := !config.Tag.IsNull() && !config.Tag.IsUnknown() && config.Tag.ValueString() != ""

	if !hasBranch && !hasTag {
		resp.Diagnostics.AddError("Invalid configuration",
			"Exactly one of `branch` or `tag` must be specified.")
		return
	}
	if hasBranch && hasTag {
		resp.Diagnostics.AddError("Invalid configuration",
			"Only one of `branch` or `tag` may be specified, not both.")
		return
	}

	var vDescriptor git.GitVersionDescriptor
	var versionKey, versionValue string
	if hasBranch {
		branchVal := shortBranchName(config.Branch.ValueString())
		vDescriptor = git.GitVersionDescriptor{
			VersionType: &git.GitVersionTypeValues.Branch,
			Version:     converter.String(branchVal),
		}
		versionKey = "branch"
		versionValue = branchVal
	} else {
		tagVal := shortTagName(config.Tag.ValueString())
		vDescriptor = git.GitVersionDescriptor{
			VersionType: &git.GitVersionTypeValues.Tag,
			Version:     converter.String(tagVal),
		}
		versionKey = "tag"
		versionValue = tagVal
	}

	repoItem, err := d.client.GitReposClient.GetItem(ctx, git.GetItemArgs{
		RepositoryId:      &repoID,
		Path:              &file,
		IncludeContent:    converter.Bool(true),
		VersionDescriptor: &vDescriptor,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.Diagnostics.AddError("Item not found",
				fmt.Sprintf("Item not found, repositoryID: %s, %s: %s, file: %s. Error: %+v", repoID, versionKey, versionValue, file, err))
			return
		}
		resp.Diagnostics.AddError("Error reading file",
			fmt.Sprintf("Get item failed, repositoryID: %s, %s: %s, file: %s. Error: %+v", repoID, versionKey, versionValue, file, err))
		return
	}

	commit, err := d.client.GitReposClient.GetCommit(ctx, git.GetCommitArgs{
		RepositoryId: &repoID,
		CommitId:     repoItem.CommitId,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error reading commit",
			fmt.Sprintf("Get commit failed, repositoryID: %s, commitID: %s. Error: %+v", repoID, *repoItem.CommitId, err))
		return
	}

	state := gitRepositoryFileDataModel{
		ID:                types.StringValue(fmt.Sprintf("%s/%s:%s:%s", repoID, file, string(*vDescriptor.VersionType), *vDescriptor.Version)),
		RepositoryID:      types.StringValue(repoID),
		File:              types.StringValue(file),
		Content:           types.StringValue(converter.ToString(repoItem.Content, "")),
		LastCommitMessage: types.StringValue(converter.ToString(commit.Comment, "")),
	}

	if hasBranch {
		state.Branch = config.Branch
		state.Tag = types.StringNull()
	} else {
		state.Branch = types.StringNull()
		state.Tag = config.Tag
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
