package release

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var _ datasource.DataSource = &releaseDefinitionsDataSource{}

// releaseDefinitionsDataSource is the terraform-plugin-framework data source for
// betterado_release_definitions (list).
type releaseDefinitionsDataSource struct {
	client *client.AggregatedClient
}

// NewReleaseDefinitionsDataSource returns a new framework data source for betterado_release_definitions.
func NewReleaseDefinitionsDataSource() datasource.DataSource {
	return &releaseDefinitionsDataSource{}
}

// releaseDefinitionItem represents a single item in the release_definitions list.
type releaseDefinitionItem struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Path types.String `tfsdk:"path"`
}

// releaseDefinitionsDataModel is the tfsdk model for the data source.
type releaseDefinitionsDataModel struct {
	ID                 types.String            `tfsdk:"id"`
	ProjectID          types.String            `tfsdk:"project_id"`
	Path               types.String            `tfsdk:"path"`
	Name               types.String            `tfsdk:"name"`
	ReleaseDefinitions []releaseDefinitionItem `tfsdk:"release_definitions"`
}

func (d *releaseDefinitionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_release_definitions"
}

func (d *releaseDefinitionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads all release definitions in an Azure DevOps project, optionally filtered by path or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "A stable ID computed from the filter parameters.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project.",
			},
			"path": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Filter release definitions by folder path.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Filter release definitions by name search text.",
			},
			"release_definitions": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The list of release definitions matching the filter.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the release definition.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the release definition.",
						},
						"path": schema.StringAttribute{
							Computed:    true,
							Description: "The folder path of the release definition.",
						},
					},
				},
			},
		},
	}
}

func (d *releaseDefinitionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *releaseDefinitionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_release_definitions data source Read: provider client not configured")
		return
	}

	var model releaseDefinitionsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	path := model.Path.ValueString()
	name := model.Name.ValueString()

	args := releaseapi.GetReleaseDefinitionsArgs{
		Project: &projectID,
	}
	if path != "" {
		args.Path = &path
	}
	if name != "" {
		args.SearchText = &name
	}

	listResp, err := d.client.ReleaseClient.GetReleaseDefinitions(d.client.Ctx, args)
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading release definitions for project %s: %+v", projectID, err))
		return
	}

	var defs []releaseapi.ReleaseDefinition
	if listResp != nil {
		defs = listResp.Value
	}

	items := make([]releaseDefinitionItem, 0, len(defs))
	for _, def := range defs {
		item := releaseDefinitionItem{
			ID:   types.StringValue(""),
			Name: types.StringValue(""),
			Path: types.StringValue(""),
		}
		if def.Id != nil {
			item.ID = types.StringValue(strconv.Itoa(*def.Id))
		}
		if def.Name != nil {
			item.Name = types.StringValue(*def.Name)
		}
		if def.Path != nil {
			item.Path = types.StringValue(*def.Path)
		}
		items = append(items, item)
	}

	model.ReleaseDefinitions = items

	// Ensure optional computed fields are set.
	if model.Path.IsNull() || model.Path.IsUnknown() {
		model.Path = types.StringValue(path)
	}
	if model.Name.IsNull() || model.Name.IsUnknown() {
		model.Name = types.StringValue(name)
	}

	// Compute a stable ID based on project + path + name.
	h := sha1.New()
	h.Write([]byte(projectID + ":" + path + ":" + name)) //nolint:errcheck
	model.ID = types.StringValue("release-definitions#" + base64.URLEncoding.EncodeToString(h.Sum(nil)))

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
