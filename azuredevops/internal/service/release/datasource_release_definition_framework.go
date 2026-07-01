package release

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	releaseapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/release"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Ensure interface compliance.
var _ datasource.DataSource = &releaseDefinitionDataSource{}

// releaseDefinitionDataSource is the terraform-plugin-framework data source for
// betterado_release_definition.
type releaseDefinitionDataSource struct {
	client *client.AggregatedClient
}

// NewReleaseDefinitionDataSource returns a new framework data source for betterado_release_definition.
func NewReleaseDefinitionDataSource() datasource.DataSource {
	return &releaseDefinitionDataSource{}
}

// releaseDefinitionDataModel is the tfsdk model for the data source.
type releaseDefinitionDataModel struct {
	ID                  types.String `tfsdk:"id"`
	ProjectID           types.String `tfsdk:"project_id"`
	ReleaseDefinitionID types.Int64  `tfsdk:"release_definition_id"`
	Name                types.String `tfsdk:"name"`
	Path                types.String `tfsdk:"path"`
	Description         types.String `tfsdk:"description"`
	ReleaseNameFormat   types.String `tfsdk:"release_name_format"`
}

func (d *releaseDefinitionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_release_definition"
}

func (d *releaseDefinitionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads a release definition from an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the release definition.",
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project the release definition belongs to.",
			},
			"release_definition_id": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "The numeric ID of the release definition. One of release_definition_id or name must be set.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the release definition. One of release_definition_id or name must be set.",
			},
			"path": schema.StringAttribute{
				Computed:    true,
				Description: "The folder path of the release definition.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the release definition.",
			},
			"release_name_format": schema.StringAttribute{
				Computed:    true,
				Description: "The release name format of the release definition.",
			},
		},
	}
}

func (d *releaseDefinitionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *releaseDefinitionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_release_definition data source Read: provider client not configured")
		return
	}

	var model releaseDefinitionDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()

	// Look up by ID if provided.
	if !model.ReleaseDefinitionID.IsNull() && !model.ReleaseDefinitionID.IsUnknown() {
		defID := int(model.ReleaseDefinitionID.ValueInt64())
		def, err := d.client.ReleaseClient.GetReleaseDefinition(d.client.Ctx, releaseapi.GetReleaseDefinitionArgs{
			Project:      &projectID,
			DefinitionId: &defID,
		})
		if err != nil {
			if utils.ResponseWasNotFound(err) {
				resp.Diagnostics.AddError("Not found", fmt.Sprintf("release definition with ID %d not found in project %s", defID, projectID))
				return
			}
			resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading release definition (ID: %d): %+v", defID, err))
			return
		}
		flattenReleaseDefinitionDataModel(&model, def, projectID)
		resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
		return
	}

	// Look up by name.
	name := model.Name.ValueString()
	listResp, err := d.client.ReleaseClient.GetReleaseDefinitions(d.client.Ctx, releaseapi.GetReleaseDefinitionsArgs{
		Project:    &projectID,
		SearchText: converter.String(name),
	})
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("listing release definitions with name %q in project %s: %+v", name, projectID, err))
		return
	}

	var matches []releaseapi.ReleaseDefinition
	if listResp != nil {
		for _, rd := range listResp.Value {
			rdCopy := rd
			if rd.Name != nil && *rd.Name == name {
				matches = append(matches, rdCopy)
			}
		}
	}

	if len(matches) == 0 {
		resp.Diagnostics.AddError("Not found", fmt.Sprintf("no release definition with name %q found in project %s", name, projectID))
		return
	}
	if len(matches) > 1 {
		resp.Diagnostics.AddError("Ambiguous", fmt.Sprintf("multiple release definitions with name %q found in project %s; use release_definition_id to disambiguate", name, projectID))
		return
	}

	flattenReleaseDefinitionDataModel(&model, &matches[0], projectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func flattenReleaseDefinitionDataModel(model *releaseDefinitionDataModel, def *releaseapi.ReleaseDefinition, projectID string) {
	model.ID = types.StringValue(strconv.Itoa(*def.Id))
	model.ProjectID = types.StringValue(projectID)
	model.ReleaseDefinitionID = types.Int64Value(int64(*def.Id))
	if def.Name != nil {
		model.Name = types.StringValue(*def.Name)
	} else {
		model.Name = types.StringValue("")
	}
	if def.Path != nil {
		model.Path = types.StringValue(*def.Path)
	} else {
		model.Path = types.StringValue("")
	}
	if def.Description != nil {
		model.Description = types.StringValue(*def.Description)
	} else {
		model.Description = types.StringValue("")
	}
	if def.ReleaseNameFormat != nil {
		model.ReleaseNameFormat = types.StringValue(*def.ReleaseNameFormat)
	} else {
		model.ReleaseNameFormat = types.StringValue("")
	}
}
