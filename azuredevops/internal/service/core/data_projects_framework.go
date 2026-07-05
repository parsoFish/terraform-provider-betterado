package core

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
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = (*ProjectsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ProjectsDataSource)(nil)
)

// ProjectsDataSource is the terraform-plugin-framework implementation of
// data.betterado_projects.
type ProjectsDataSource struct {
	client *client.AggregatedClient
}

// NewProjectsDataSource returns a new datasource.DataSource for data.betterado_projects.
func NewProjectsDataSource() datasource.DataSource {
	return &ProjectsDataSource{}
}

// projectsDataModel is the Terraform state model for data.betterado_projects.
type projectsDataModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	State    types.String `tfsdk:"state"`
	Projects types.List   `tfsdk:"projects"`
}

// projectRefModel holds a single project reference within data.betterado_projects.
type projectRefModel struct {
	Name       types.String `tfsdk:"name"`
	ProjectID  types.String `tfsdk:"project_id"`
	ProjectURL types.String `tfsdk:"project_url"`
	State      types.String `tfsdk:"state"`
}

// projectRefAttrTypes returns the attribute type map for a projectRefModel.
var projectRefAttrTypes = map[string]attr.Type{
	"name":        types.StringType,
	"project_id":  types.StringType,
	"project_url": types.StringType,
	"state":       types.StringType,
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *ProjectsDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "betterado_projects"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *ProjectsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter projects by name (case-insensitive).",
			},
			"state": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Filter projects by state (e.g. wellFormed, all, deleting, etc.).",
			},
			"projects": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":        schema.StringAttribute{Computed: true},
						"project_id":  schema.StringAttribute{Computed: true},
						"project_url": schema.StringAttribute{Computed: true},
						"state":       schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *ProjectsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model projectsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	stateFilter := model.State.ValueString()
	if stateFilter == "" {
		stateFilter = string(core.ProjectStateValues.All)
	}
	nameFilter := model.Name.ValueString()

	projects, err := getProjectsForStateAndName(d.client, stateFilter, nameFilter)
	if err != nil {
		resp.Diagnostics.AddError("listing projects", fmt.Sprintf("Error listing projects: %v", err))
		return
	}

	// Collect project names for ID hash.
	var names []string
	projectObjs := make([]attr.Value, 0, len(projects))
	for _, p := range projects {
		m := projectRefModel{}
		if p.Name != nil {
			m.Name = types.StringValue(*p.Name)
			names = append(names, *p.Name)
		} else {
			m.Name = types.StringValue("")
		}
		if p.Id != nil {
			m.ProjectID = types.StringValue(p.Id.String())
		} else {
			m.ProjectID = types.StringValue("")
		}
		if p.Url != nil {
			m.ProjectURL = types.StringValue(*p.Url)
		} else {
			m.ProjectURL = types.StringValue("")
		}
		if p.State != nil {
			m.State = types.StringValue(string(*p.State))
		} else {
			m.State = types.StringValue("")
		}

		obj, diags := types.ObjectValueFrom(ctx, projectRefAttrTypes, m)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		projectObjs = append(projectObjs, obj)
	}

	if len(names) == 0 && nameFilter != "" {
		names = append(names, nameFilter)
	}
	h := sha1.New()
	_, _ = h.Write([]byte(stateFilter + strings.Join(names, "-")))
	model.ID = types.StringValue("projects#" + base64.URLEncoding.EncodeToString(h.Sum(nil)))
	model.State = types.StringValue(stateFilter)

	listVal, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: projectRefAttrTypes}, projectObjs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Projects = listVal

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
