package git

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = &gitRepositoriesDataSource{}
	_ datasource.DataSourceWithConfigure = &gitRepositoriesDataSource{}
)

// NewGitRepositoriesDataSource returns a new framework data source.
func NewGitRepositoriesDataSource() datasource.DataSource {
	return &gitRepositoriesDataSource{}
}

// gitRepositoriesDataSource is the framework implementation of betterado_git_repositories (data source).
type gitRepositoriesDataSource struct {
	client *client.AggregatedClient
}

// gitRepositoriesDataModel is the tfsdk model for the data source.
type gitRepositoriesDataModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	Name          types.String `tfsdk:"name"`
	IncludeHidden types.Bool   `tfsdk:"include_hidden"`
	Repositories  types.List   `tfsdk:"repositories"`
}

// repositoryAttrTypes returns the attr.Type map for a single repository object.
func repositoryAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":             types.StringType,
		"name":           types.StringType,
		"url":            types.StringType,
		"ssh_url":        types.StringType,
		"web_url":        types.StringType,
		"remote_url":     types.StringType,
		"project_id":     types.StringType,
		"size":           types.Int64Type,
		"default_branch": types.StringType,
		"disabled":       types.BoolType,
	}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *gitRepositoriesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_repositories"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *gitRepositoriesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads Git repositories in an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "A hash-based ID derived from the repository names.",
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the project to list repositories for.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter repositories by name (case-insensitive).",
			},
			"include_hidden": schema.BoolAttribute{
				Optional:    true,
				Description: "Include hidden repositories in the result. Defaults to false.",
			},
			"repositories": schema.ListNestedAttribute{
				Computed:    true,
				Description: "A list of repositories matching the filter criteria.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the repository.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the repository.",
						},
						"url": schema.StringAttribute{
							Computed:    true,
							Description: "REST API URL for the repository.",
						},
						"ssh_url": schema.StringAttribute{
							Computed:    true,
							Description: "Git SSH clone URL.",
						},
						"web_url": schema.StringAttribute{
							Computed:    true,
							Description: "Web URL for the repository.",
						},
						"remote_url": schema.StringAttribute{
							Computed:    true,
							Description: "Git HTTPS clone URL.",
						},
						"project_id": schema.StringAttribute{
							Computed:    true,
							Description: "The project ID the repository belongs to.",
						},
						"size": schema.Int64Attribute{
							Computed:    true,
							Description: "Size of the repository in bytes.",
						},
						"default_branch": schema.StringAttribute{
							Computed:    true,
							Description: "The ref of the default branch.",
						},
						"disabled": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the repository is disabled.",
						},
					},
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *gitRepositoriesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *gitRepositoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been configured with Azure DevOps credentials.")
		return
	}

	var config gitRepositoriesDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	projectID := config.ProjectID.ValueString()
	includeHidden := false
	if !config.IncludeHidden.IsNull() && !config.IncludeHidden.IsUnknown() {
		includeHidden = config.IncludeHidden.ValueBool()
	}

	projectRepos, err := getGitRepositoriesByNameAndProject(d.client, name, projectID, includeHidden)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			// Return empty list on not-found; the data source is a list query.
			state := gitRepositoriesDataModel{
				ID:            types.StringValue("gitRepos#empty"),
				ProjectID:     config.ProjectID,
				Name:          config.Name,
				IncludeHidden: config.IncludeHidden,
				Repositories:  types.ListValueMust(types.ObjectType{AttrTypes: repositoryAttrTypes()}, []attr.Value{}),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
		resp.Diagnostics.AddError("Failed to find repositories", fmt.Sprintf("Error: %v", err))
		return
	}

	// Build repository list.
	repoItems := make([]attr.Value, 0)
	var repoNames []string

	if projectRepos != nil {
		for _, repo := range *projectRepos {
			repoName := converter.ToString(repo.Name, "")
			repoNames = append(repoNames, repoName)

			repoIDStr := ""
			if repo.Id != nil {
				repoIDStr = repo.Id.String()
			}

			projIDStr := ""
			if repo.Project != nil && repo.Project.Id != nil {
				projIDStr = repo.Project.Id.String()
			}

			sizeVal := int64(0)
			if repo.Size != nil {
				sizeVal = int64(*repo.Size)
			}

			obj, diags := types.ObjectValue(repositoryAttrTypes(), map[string]attr.Value{
				"id":             types.StringValue(repoIDStr),
				"name":           types.StringValue(repoName),
				"url":            types.StringValue(converter.ToString(repo.Url, "")),
				"ssh_url":        types.StringValue(converter.ToString(repo.SshUrl, "")),
				"web_url":        types.StringValue(converter.ToString(repo.WebUrl, "")),
				"remote_url":     types.StringValue(converter.ToString(repo.RemoteUrl, "")),
				"project_id":     types.StringValue(projIDStr),
				"size":           types.Int64Value(sizeVal),
				"default_branch": types.StringValue(converter.ToString(repo.DefaultBranch, "")),
				"disabled":       types.BoolValue(converter.ToBool(repo.IsDisabled, false)),
			})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			repoItems = append(repoItems, obj)
		}
	}

	// Compute a hash-based ID.
	id := computeGitRepositoriesID(projectID, repoNames)

	reposList, diags := types.ListValue(types.ObjectType{AttrTypes: repositoryAttrTypes()}, repoItems)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := gitRepositoriesDataModel{
		ID:            types.StringValue(id),
		ProjectID:     config.ProjectID,
		Name:          config.Name,
		IncludeHidden: config.IncludeHidden,
		Repositories:  reposList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// computeGitRepositoriesID mirrors the SDKv2 ID logic: sha1 over joined repo names.
func computeGitRepositoriesID(projectID string, repoNames []string) string {
	names := make([]string, len(repoNames))
	copy(names, repoNames)
	if len(names) == 0 {
		names = append(names, "empty")
	}
	if projectID != "" {
		names = append([]string{projectID}, names...)
	}
	h := sha1.New()
	_, _ = h.Write([]byte(strings.Join(names, "-")))
	return "gitRepos#" + base64.URLEncoding.EncodeToString(h.Sum(nil))
}
