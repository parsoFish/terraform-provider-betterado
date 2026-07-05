package core

import (
	"context"
	"fmt"
	"regexp"

	"github.com/ahmetb/go-linq"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	adocore "github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = (*TeamsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*TeamsDataSource)(nil)
)

// TeamsDataSource is the terraform-plugin-framework implementation of
// the betterado_teams data source.
type TeamsDataSource struct {
	client *client.AggregatedClient
}

// NewTeamsDataSource returns a new datasource.DataSource.
func NewTeamsDataSource() datasource.DataSource {
	return &TeamsDataSource{}
}

// teamsDataSourceModel is the Terraform state model for the betterado_teams data source.
type teamsDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Top       types.Int64  `tfsdk:"top"`
	Teams     types.List   `tfsdk:"teams"`
}

// teamItemAttrTypes defines the object attribute types for a single team entry.
var teamItemAttrTypes = map[string]attr.Type{
	"id":             types.StringType,
	"project_id":     types.StringType,
	"name":           types.StringType,
	"description":    types.StringType,
	"administrators": types.SetType{ElemType: types.StringType},
	"members":        types.SetType{ElemType: types.StringType},
}

// teamItemObject is the Go struct that maps to the team list element object.
type teamItemObject struct {
	ID             types.String `tfsdk:"id"`
	ProjectID      types.String `tfsdk:"project_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Administrators types.Set    `tfsdk:"administrators"`
	Members        types.Set    `tfsdk:"members"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *TeamsDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "betterado_teams"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *TeamsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to access information about Azure DevOps teams.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the project; if omitted all projects are queried.",
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
						"must be a valid UUID",
					),
				},
			},
			"top": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Maximum number of teams to return per project.",
			},
			"teams": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"project_id":  schema.StringAttribute{Computed: true},
						"name":        schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"administrators": schema.SetAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
						"members": schema.SetAttribute{
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *TeamsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TeamsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model teamsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	top := 100
	if !model.Top.IsNull() && !model.Top.IsUnknown() {
		top = int(model.Top.ValueInt64())
	}

	var projectIDList []string
	if !model.ProjectID.IsNull() && !model.ProjectID.IsUnknown() {
		projectIDList = []string{model.ProjectID.ValueString()}
	} else {
		projectList, err := getProjectsForStateAndName(d.client, string(adocore.ProjectStateValues.All), "")
		if err != nil {
			resp.Diagnostics.AddError("listing projects", err.Error())
			return
		}
		linq.From(projectList).
			Select(func(e interface{}) interface{} {
				return e.(adocore.TeamProjectReference).Id.String()
			}).
			ToSlice(&projectIDList)
	}

	teamObjType := types.ObjectType{AttrTypes: teamItemAttrTypes}
	teamObjs := make([]attr.Value, 0)

	for _, projectID := range projectIDList {
		teamList, err := d.client.CoreClient.GetTeams(d.client.Ctx, adocore.GetTeamsArgs{
			ProjectId:      converter.String(projectID),
			Mine:           converter.Bool(false),
			Top:            converter.Int(top),
			ExpandIdentity: converter.Bool(false),
		})
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("listing teams for project %s", projectID), err.Error())
			return
		}
		if teamList == nil || len(*teamList) == 0 {
			continue
		}

		for i := range *teamList {
			team := &(*teamList)[i]

			membersSet, err := getTeamMembers(d.client, team)
			if err != nil {
				resp.Diagnostics.AddError("reading team members", err.Error())
				return
			}
			memberList := membersSet.List()
			memberStrings := make([]string, len(memberList))
			for j, m := range memberList {
				memberStrings[j] = m.(string)
			}
			membersVal, _ := types.SetValueFrom(ctx, types.StringType, memberStrings)

			adminsSet, err := getTeamAdministrators(nil, d.client, team)
			if err != nil {
				resp.Diagnostics.AddError("reading team administrators", err.Error())
				return
			}
			adminList := adminsSet.List()
			adminStrings := make([]string, len(adminList))
			for j, a := range adminList {
				adminStrings[j] = a.(string)
			}
			adminsVal, _ := types.SetValueFrom(ctx, types.StringType, adminStrings)

			item := teamItemObject{
				Members:        membersVal,
				Administrators: adminsVal,
			}
			if team.Id != nil {
				item.ID = types.StringValue(team.Id.String())
			} else {
				item.ID = types.StringValue("")
			}
			if team.ProjectId != nil {
				item.ProjectID = types.StringValue(team.ProjectId.String())
			} else {
				item.ProjectID = types.StringValue("")
			}
			if team.Name != nil {
				item.Name = types.StringValue(*team.Name)
			} else {
				item.Name = types.StringValue("")
			}
			if team.Description != nil {
				item.Description = types.StringValue(*team.Description)
			} else {
				item.Description = types.StringValue("")
			}

			obj, diags := types.ObjectValueFrom(ctx, teamItemAttrTypes, item)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			teamObjs = append(teamObjs, obj)
		}
	}

	teamsVal, diags := types.ListValueFrom(ctx, teamObjType, teamObjs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	model.Teams = teamsVal
	model.Top = types.Int64Value(int64(top))
	model.ID = types.StringValue("teams")

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
