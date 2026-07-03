package git

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = &gitRepositoryDataSource{}
	_ datasource.DataSourceWithConfigure = &gitRepositoryDataSource{}
)

// gitRepositoryDataSource is the framework implementation of betterado_git_repository (data source).
type gitRepositoryDataSource struct {
	client *client.AggregatedClient
}

// NewGitRepositoryDataSource returns a new framework data source.
func NewGitRepositoryDataSource() datasource.DataSource {
	return &gitRepositoryDataSource{}
}

// gitRepositoryDataModel is the tfsdk model for the data source.
type gitRepositoryDataModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	Name          types.String `tfsdk:"name"`
	DefaultBranch types.String `tfsdk:"default_branch"`
	IsFork        types.Bool   `tfsdk:"is_fork"`
	RemoteURL     types.String `tfsdk:"remote_url"`
	Size          types.Int64  `tfsdk:"size"`
	SSHURL        types.String `tfsdk:"ssh_url"`
	URL           types.String `tfsdk:"url"`
	WebURL        types.String `tfsdk:"web_url"`
	Disabled      types.Bool   `tfsdk:"disabled"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *gitRepositoryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_repository"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *gitRepositoryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a Git repository in Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the repository.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project the repository belongs to.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the git repository.",
			},
			"default_branch": schema.StringAttribute{
				Computed:    true,
				Description: "The ref of the default branch.",
			},
			"is_fork": schema.BoolAttribute{
				Computed:    true,
				Description: "True if the repository was created as a fork.",
			},
			"remote_url": schema.StringAttribute{
				Computed:    true,
				Description: "Git HTTPS clone URL.",
			},
			"size": schema.Int64Attribute{
				Computed:    true,
				Description: "Size of the repository in bytes.",
			},
			"ssh_url": schema.StringAttribute{
				Computed:    true,
				Description: "Git SSH clone URL.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "REST API URL for the repository.",
			},
			"web_url": schema.StringAttribute{
				Computed:    true,
				Description: "Web URL for the repository.",
			},
			"disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the repository is disabled.",
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *gitRepositoryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *gitRepositoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured with Azure DevOps credentials.")
		return
	}

	var config gitRepositoryDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	projectID := config.ProjectID.ValueString()

	projectRepos, err := getGitRepositoriesByNameAndProject(d.client, name, projectID, true)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			tflog.Warn(ctx, fmt.Sprintf("Git repository %s not found in project %s", name, projectID))
			resp.Diagnostics.AddError(
				"Repository not found",
				fmt.Sprintf("Repository with name %s does not exist in project %s", name, projectID),
			)
			return
		}
		resp.Diagnostics.AddError("Failed to find repositories", fmt.Sprintf("Error: %v", err))
		return
	}

	if projectRepos == nil || len(*projectRepos) == 0 {
		resp.Diagnostics.AddError(
			"Repository not found",
			fmt.Sprintf("Repository with name %s does not exist in project %s", name, projectID),
		)
		return
	}

	if len(*projectRepos) > 1 {
		resp.Diagnostics.AddError(
			"Multiple repositories found",
			fmt.Sprintf("Multiple repositories with name %s found in project %s", name, projectID),
		)
		return
	}

	repo := (*projectRepos)[0]

	state := gitRepositoryDataModel{
		ID:            types.StringValue(repo.Id.String()),
		ProjectID:     types.StringValue(projectID),
		Name:          types.StringValue(converter.ToString(repo.Name, "")),
		DefaultBranch: types.StringValue(converter.ToString(repo.DefaultBranch, "")),
		IsFork:        types.BoolValue(converter.ToBool(repo.IsFork, false)),
		RemoteURL:     types.StringValue(converter.ToString(repo.RemoteUrl, "")),
		SSHURL:        types.StringValue(converter.ToString(repo.SshUrl, "")),
		URL:           types.StringValue(converter.ToString(repo.Url, "")),
		WebURL:        types.StringValue(converter.ToString(repo.WebUrl, "")),
		Disabled:      types.BoolValue(converter.ToBool(repo.IsDisabled, false)),
	}

	if repo.Size != nil {
		state.Size = types.Int64Value(int64(*repo.Size))
	} else {
		state.Size = types.Int64Value(0)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
