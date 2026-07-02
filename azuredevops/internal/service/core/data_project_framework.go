package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = (*ProjectDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ProjectDataSource)(nil)
)

// ProjectDataSource is the terraform-plugin-framework implementation of
// data.betterado_project.
type ProjectDataSource struct {
	client *client.AggregatedClient
}

// NewProjectDataSource returns a new datasource.DataSource for data.betterado_project.
func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

// projectDataModel is the Terraform state model for data.betterado_project.
type projectDataModel struct {
	ID                types.String `tfsdk:"id"`
	ProjectID         types.String `tfsdk:"project_id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	Visibility        types.String `tfsdk:"visibility"`
	VersionControl    types.String `tfsdk:"version_control"`
	WorkItemTemplate  types.String `tfsdk:"work_item_template"`
	ProcessTemplateID types.String `tfsdk:"process_template_id"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *ProjectDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "betterado_project"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *ProjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The project ID to look up. Conflicts with name.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The project name to look up. Conflicts with project_id.",
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"visibility": schema.StringAttribute{
				Computed: true,
			},
			"version_control": schema.StringAttribute{
				Computed: true,
			},
			"work_item_template": schema.StringAttribute{
				Computed: true,
			},
			"process_template_id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *ProjectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	agg, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data",
			fmt.Sprintf("Expected *client.AggregatedClient, got: %T", req.ProviderData),
		)
		return
	}
	d.client = agg
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model projectDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	identifier := model.ProjectID.ValueString()
	if identifier == "" {
		identifier = model.Name.ValueString()
	}
	if identifier == "" {
		resp.Diagnostics.AddError("missing required attribute", "either name or project_id must be set")
		return
	}

	project, err := d.client.CoreClient.GetProject(d.client.Ctx, core.GetProjectArgs{
		ProjectId:           &identifier,
		IncludeCapabilities: converter.Bool(true),
		IncludeHistory:      converter.Bool(false),
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.Diagnostics.AddError("project not found",
				fmt.Sprintf("Project with name or ID %q does not exist", identifier))
			return
		}
		resp.Diagnostics.AddError("reading project",
			fmt.Sprintf("Error looking up project %q: %v", identifier, err))
		return
	}

	model.ID = types.StringValue(project.Id.String())
	model.ProjectID = types.StringValue(project.Id.String())
	if project.Name != nil {
		model.Name = types.StringValue(*project.Name)
	}
	if project.Description != nil {
		model.Description = types.StringValue(*project.Description)
	}
	if project.Visibility != nil {
		model.Visibility = types.StringValue(strings.ToLower(string(*project.Visibility)))
	}

	if project.Capabilities != nil {
		caps := *project.Capabilities
		if vc, ok := caps["versioncontrol"]; ok {
			model.VersionControl = types.StringValue(vc["sourceControlType"])
		}
		if pt, ok := caps["processTemplate"]; ok {
			ptID := pt["templateTypeId"]
			model.ProcessTemplateID = types.StringValue(ptID)
			if ptID != "" {
				name, err := lookupProcessTemplateName(d.client, ptID)
				if err != nil {
					resp.Diagnostics.AddWarning("resolving work_item_template", err.Error())
				} else {
					model.WorkItemTemplate = types.StringValue(name)
				}
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
