package core

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	adocore "github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = (*TeamResource)(nil)
	_ resource.ResourceWithConfigure   = (*TeamResource)(nil)
	_ resource.ResourceWithImportState = (*TeamResource)(nil)
)

// TeamResource is the terraform-plugin-framework implementation of betterado_team.
type TeamResource struct {
	client *client.AggregatedClient
}

// NewTeamResource returns a new resource.Resource for betterado_team.
func NewTeamResource() resource.Resource {
	return &TeamResource{}
}

// teamModel is the Terraform state model for betterado_team.
type teamModel struct {
	ID             types.String `tfsdk:"id"`
	ProjectID      types.String `tfsdk:"project_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Administrators types.Set    `tfsdk:"administrators"`
	Members        types.Set    `tfsdk:"members"`
	Descriptor     types.String `tfsdk:"descriptor"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *TeamResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_team"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *TeamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a team in an Azure DevOps project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the team.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`\S`),
						"must not be whitespace-only",
					),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The description of the team.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"administrators": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Set of subject descriptors for team administrators.",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
					),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"members": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Set of subject descriptors for team members.",
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
					),
				},
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"descriptor": schema.StringAttribute{
				Computed:    true,
				Description: "The descriptor of the team (for use in permission references).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *TeamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = agg
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *TeamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model teamModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	teamData := adocore.WebApiTeam{
		Name: converter.ToPtr(model.Name.ValueString()),
	}
	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		desc := model.Description.ValueString()
		teamData.Description = &desc
	}

	team, err := r.client.CoreClient.CreateTeam(r.client.Ctx, adocore.CreateTeamArgs{
		ProjectId: &projectID,
		Team:      &teamData,
	})
	if err != nil {
		resp.Diagnostics.AddError("creating team", err.Error())
		return
	}

	model.ID = types.StringValue(team.Id.String())

	// Apply administrators if set.
	if !model.Administrators.IsNull() && !model.Administrators.IsUnknown() {
		var admins []string
		model.Administrators.ElementsAs(ctx, &admins, false)
		if err := updateTeamAdministrators(nil, r.client, team, &admins); err != nil {
			resp.Diagnostics.AddError("setting team administrators", err.Error())
			return
		}
	}

	// Apply members if set.
	if !model.Members.IsNull() && !model.Members.IsUnknown() {
		var members []string
		model.Members.ElementsAs(ctx, &members, false)
		if err := setTeamMembers(r.client, team, &members); err != nil {
			resp.Diagnostics.AddError("setting team members", err.Error())
			return
		}
	}

	if err := r.readTeamIntoModel(ctx, &model); err != nil {
		resp.Diagnostics.AddError("reading team after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TeamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model teamModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readTeamIntoModel(ctx, &model); err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("reading team", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TeamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan teamModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID
	projectID := state.ProjectID.ValueString()
	teamID := state.ID.ValueString()

	teamData := adocore.WebApiTeam{}
	needsUpdate := false

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		teamData.Name = &name
		needsUpdate = true
	}
	if !plan.Description.Equal(state.Description) {
		desc := plan.Description.ValueString()
		teamData.Description = &desc
		needsUpdate = true
	}

	var team *adocore.WebApiTeam
	var err error
	if needsUpdate {
		team, err = r.client.CoreClient.UpdateTeam(r.client.Ctx, adocore.UpdateTeamArgs{
			ProjectId: &projectID,
			TeamId:    &teamID,
			TeamData:  &teamData,
		})
		if err != nil {
			resp.Diagnostics.AddError("updating team", err.Error())
			return
		}
	} else {
		team, err = r.client.CoreClient.GetTeam(r.client.Ctx, adocore.GetTeamArgs{
			ProjectId:      &projectID,
			TeamId:         &teamID,
			ExpandIdentity: converter.Bool(false),
		})
		if err != nil {
			resp.Diagnostics.AddError("reading team during update", err.Error())
			return
		}
	}

	if !plan.Administrators.Equal(state.Administrators) {
		var admins []string
		plan.Administrators.ElementsAs(ctx, &admins, false)
		if err := updateTeamAdministrators(nil, r.client, team, &admins); err != nil {
			resp.Diagnostics.AddError("updating team administrators", err.Error())
			return
		}
	}

	if !plan.Members.Equal(state.Members) {
		var members []string
		plan.Members.ElementsAs(ctx, &members, false)
		if err := setTeamMembers(r.client, team, &members); err != nil {
			resp.Diagnostics.AddError("updating team members", err.Error())
			return
		}
	}

	if err := r.readTeamIntoModel(ctx, &plan); err != nil {
		resp.Diagnostics.AddError("reading team after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TeamResource) Delete(_ context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model teamModel
	resp.Diagnostics.Append(req.State.Get(context.Background(), &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	teamID := model.ID.ValueString()

	if err := r.client.CoreClient.DeleteTeam(r.client.Ctx, adocore.DeleteTeamArgs{
		ProjectId: &projectID,
		TeamId:    &teamID,
	}); err != nil {
		resp.Diagnostics.AddError("deleting team", err.Error())
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

// ImportState supports import by "projectID/teamID".
func (r *TeamResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Accept "projectID/teamID" format.
	var projectID, teamID string
	fmt.Sscanf(req.ID, "%36s/%36s", &projectID, &teamID)
	if _, err := uuid.Parse(projectID); err != nil || teamID == "" {
		// Fall back: treat req.ID as teamID; project_id will need to be set.
		resp.Diagnostics.AddError(
			"Invalid import ID",
			`Expected format: "<project_uuid>/<team_uuid>", got: `+req.ID,
		)
		return
	}

	model := teamModel{
		ID:        types.StringValue(teamID),
		ProjectID: types.StringValue(projectID),
	}
	if err := r.readTeamIntoModel(ctx, &model); err != nil {
		resp.Diagnostics.AddError("importing team", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (r *TeamResource) readTeamIntoModel(ctx context.Context, model *teamModel) error {
	projectID := model.ProjectID.ValueString()
	teamID := model.ID.ValueString()

	team, err := r.client.CoreClient.GetTeam(r.client.Ctx, adocore.GetTeamArgs{
		ProjectId:      &projectID,
		TeamId:         &teamID,
		ExpandIdentity: converter.Bool(false),
	})
	if err != nil {
		return err
	}
	if team == nil {
		return fmt.Errorf("team %s not found in project %s", teamID, projectID)
	}

	model.Name = types.StringValue(*team.Name)
	if team.Description != nil {
		model.Description = types.StringValue(*team.Description)
	} else {
		model.Description = types.StringValue("")
	}

	// Members.
	membersSet, err := getTeamMembers(r.client, team)
	if err != nil {
		return fmt.Errorf("reading team members: %w", err)
	}
	membersList := membersSet.List()
	memberStrings := make([]string, len(membersList))
	for i, m := range membersList {
		memberStrings[i] = m.(string)
	}
	membersVal, _ := types.SetValueFrom(ctx, types.StringType, memberStrings)
	model.Members = membersVal

	// Administrators — pass nil for schema.ResourceData (not needed by helper for reads).
	adminsSet, err := getTeamAdministrators(nil, r.client, team)
	if err != nil {
		return fmt.Errorf("reading team administrators: %w", err)
	}
	adminsList := adminsSet.List()
	adminStrings := make([]string, len(adminsList))
	for i, a := range adminsList {
		adminStrings[i] = a.(string)
	}
	adminsVal, _ := types.SetValueFrom(ctx, types.StringType, adminStrings)
	model.Administrators = adminsVal

	// Descriptor.
	descriptor, err := r.client.GraphClient.GetDescriptor(r.client.Ctx, graph.GetDescriptorArgs{
		StorageKey: team.Id,
	})
	if err != nil {
		return fmt.Errorf("reading team descriptor: %w", err)
	}
	model.Descriptor = types.StringValue(*descriptor.Value)

	return nil
}
