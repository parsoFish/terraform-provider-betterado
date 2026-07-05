package core

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	adocore "github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = (*TeamDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*TeamDataSource)(nil)
)

// TeamDataSource is the terraform-plugin-framework implementation of
// the betterado_team data source.
type TeamDataSource struct {
	client *client.AggregatedClient
}

// NewTeamDataSource returns a new datasource.DataSource.
func NewTeamDataSource() datasource.DataSource {
	return &TeamDataSource{}
}

// teamDataSourceModel is the Terraform state model for the betterado_team data source.
type teamDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	ProjectID      types.String `tfsdk:"project_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Administrators types.Set    `tfsdk:"administrators"`
	Members        types.Set    `tfsdk:"members"`
	Descriptor     types.String `tfsdk:"descriptor"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *TeamDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "betterado_team"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *TeamDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to access information about an Azure DevOps team.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the project.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
						"must be a valid UUID",
					),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the team.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the team.",
			},
			"administrators": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Subject descriptors of the team's administrators.",
			},
			"members": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Subject descriptors of the team's members.",
			},
			"descriptor": schema.StringAttribute{
				Computed:    true,
				Description: "The descriptor of the team.",
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *TeamDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TeamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model teamDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	teamName := model.Name.ValueString()

	team, err := d.client.CoreClient.GetTeam(d.client.Ctx, adocore.GetTeamArgs{
		ProjectId: converter.String(projectID),
		TeamId:    converter.String(teamName),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"reading team",
			fmt.Sprintf("Get team %q: %v", teamName, err),
		)
		return
	}

	model.ID = types.StringValue(team.Id.String())
	model.Name = types.StringValue(*team.Name)
	if team.Description != nil {
		model.Description = types.StringValue(*team.Description)
	} else {
		model.Description = types.StringValue("")
	}

	// Members.
	membersSet, err := getTeamMembers(d.client, team)
	if err != nil {
		resp.Diagnostics.AddError("reading team members", err.Error())
		return
	}
	memberList := membersSet.List()
	memberStrings := make([]string, len(memberList))
	for i, m := range memberList {
		memberStrings[i] = m.(string)
	}
	membersVal, _ := types.SetValueFrom(ctx, types.StringType, memberStrings)
	model.Members = membersVal

	// Administrators.
	adminsSet, err := getTeamAdministrators(nil, d.client, team)
	if err != nil {
		resp.Diagnostics.AddError("reading team administrators", err.Error())
		return
	}
	adminList := adminsSet.List()
	adminStrings := make([]string, len(adminList))
	for i, a := range adminList {
		adminStrings[i] = a.(string)
	}
	adminsVal, _ := types.SetValueFrom(ctx, types.StringType, adminStrings)
	model.Administrators = adminsVal

	// Descriptor.
	descriptor, err := d.client.GraphClient.GetDescriptor(d.client.Ctx, graph.GetDescriptorArgs{
		StorageKey: team.Id,
	})
	if err != nil {
		resp.Diagnostics.AddError("reading team descriptor", err.Error())
		return
	}
	model.Descriptor = types.StringValue(*descriptor.Value)

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
