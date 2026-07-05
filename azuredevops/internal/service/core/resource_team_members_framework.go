package core

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/ahmetb/go-linq"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	adocore "github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource              = (*TeamMembersResource)(nil)
	_ resource.ResourceWithConfigure = (*TeamMembersResource)(nil)
)

// TeamMembersResource is the terraform-plugin-framework implementation of
// betterado_team_members.
type TeamMembersResource struct {
	client *client.AggregatedClient
}

// NewTeamMembersResource returns a new resource.Resource.
func NewTeamMembersResource() resource.Resource {
	return &TeamMembersResource{}
}

// teamMembersModel is the Terraform state model for betterado_team_members.
type teamMembersModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	TeamID    types.String `tfsdk:"team_id"`
	Mode      types.String `tfsdk:"mode"`
	Members   types.Set    `tfsdk:"members"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *TeamMembersResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_team_members"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *TeamMembersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages members of a team in Azure DevOps.",
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
			"team_id": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the team.",
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
			"mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: `"add" (default) preserves unmanaged members; "overwrite" removes all not in the list.`,
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive("add", "overwrite"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"members": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Set of subject descriptors to add as team members.",
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						stringvalidator.LengthAtLeast(1),
					),
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *TeamMembersResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TeamMembersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model teamMembersModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if model.Mode.IsNull() || model.Mode.IsUnknown() {
		model.Mode = types.StringValue("add")
	}

	team, err := r.getTeam(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("reading team", err.Error())
		return
	}

	var members []string
	model.Members.ElementsAs(ctx, &members, false)

	if strings.EqualFold(model.Mode.ValueString(), "overwrite") {
		if err := setTeamMembers(r.client, team, &members); err != nil {
			resp.Diagnostics.AddError("setting team members (overwrite)", err.Error())
			return
		}
	} else {
		if err := addTeamMembers(r.client, team, linq.From(r.toIface(members)), true); err != nil {
			resp.Diagnostics.AddError("adding team members", err.Error())
			return
		}
	}

	model.ID = types.StringValue(team.ProjectId.String() + "/" + team.Id.String())
	// For Create, use the plan values for non-computed attributes so the
	// framework's plan-vs-state consistency check passes. The Azure DevOps API
	// may not immediately reflect membership changes, which would cause
	// readIntoModel to return a null/empty members set and trigger "Provider
	// produced inconsistent result after apply". Subsequent Read calls refresh
	// state from the API.
	model.ProjectID = types.StringValue(team.ProjectId.String())
	model.TeamID = types.StringValue(team.Id.String())
	// members is already set from the plan via req.Plan.Get above
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TeamMembersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model teamMembersModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := r.getTeam(ctx, model)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("reading team", err.Error())
		return
	}

	if err := r.readIntoModel(ctx, &model, team); err != nil {
		resp.Diagnostics.AddError("reading members", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TeamMembersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan teamMembersModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID
	team, err := r.getTeam(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError("reading team during update", err.Error())
		return
	}

	var newMembers []string
	plan.Members.ElementsAs(ctx, &newMembers, false)

	if strings.EqualFold(plan.Mode.ValueString(), "overwrite") {
		if err := setTeamMembers(r.client, team, &newMembers); err != nil {
			resp.Diagnostics.AddError("updating team members (overwrite)", err.Error())
			return
		}
	} else {
		var oldMembers []string
		state.Members.ElementsAs(ctx, &oldMembers, false)
		toAdd := subtract(newMembers, oldMembers)
		toRemove := subtract(oldMembers, newMembers)
		if len(toAdd) > 0 {
			if err := addTeamMembers(r.client, team, linq.From(r.toIface(toAdd)), true); err != nil {
				resp.Diagnostics.AddError("adding team members", err.Error())
				return
			}
		}
		if len(toRemove) > 0 {
			if err := removeTeamMembers(r.client, team, linq.From(r.toIface(toRemove))); err != nil {
				resp.Diagnostics.AddError("removing team members", err.Error())
				return
			}
		}
	}

	// Azure DevOps membership changes propagate asynchronously. Use the plan
	// values as the new state (same as Create) so the post-apply plan-
	// consistency check passes even if the API has not yet reflected the
	// new members. A subsequent Read (e.g. terraform refresh) will re-read
	// from the API once propagation has completed.
	plan.ProjectID = types.StringValue(team.ProjectId.String())
	plan.TeamID = types.StringValue(team.Id.String())
	// plan.Members already holds the desired set from req.Plan.Get
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TeamMembersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model teamMembersModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := r.getTeam(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("reading team during delete", err.Error())
		return
	}

	if strings.EqualFold("overwrite", model.Mode.ValueString()) {
		if err := setTeamMembers(r.client, team, nil); err != nil {
			resp.Diagnostics.AddError("removing all team members", err.Error())
		}
	} else {
		var membersToRemove []string
		model.Members.ElementsAs(ctx, &membersToRemove, false)
		if len(membersToRemove) > 0 {
			if err := removeTeamMembers(r.client, team, linq.From(r.toIface(membersToRemove))); err != nil {
				resp.Diagnostics.AddError("removing team members on delete", err.Error())
			}
		}
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (r *TeamMembersResource) getTeam(_ context.Context, model teamMembersModel) (*adocore.WebApiTeam, error) {
	projectID := model.ProjectID.ValueString()
	teamID := model.TeamID.ValueString()
	return r.client.CoreClient.GetTeam(r.client.Ctx, adocore.GetTeamArgs{
		ProjectId:      &projectID,
		TeamId:         &teamID,
		ExpandIdentity: converter.Bool(true),
	})
}

func (r *TeamMembersResource) readIntoModel(ctx context.Context, model *teamMembersModel, team *adocore.WebApiTeam) error {
	mode := model.Mode.ValueString()
	allMembers, err := getTeamMembers(r.client, team)
	if err != nil {
		return err
	}

	var currentMembers []string
	model.Members.ElementsAs(ctx, &currentMembers, false)
	stateSet := make(map[string]bool, len(currentMembers))
	for _, m := range currentMembers {
		stateSet[m] = true
	}

	// Use a non-nil slice so that types.SetValueFrom produces an empty Set
	// (not a null Set) when no members match.
	result := make([]string, 0)
	for _, m := range allMembers.List() {
		s := m.(string)
		if strings.EqualFold("overwrite", mode) || stateSet[s] {
			result = append(result, s)
		}
	}

	// Azure DevOps membership changes propagate asynchronously. The API may
	// return a partial (stale) member list immediately after adding members.
	// In "add" mode, fall back to state values whenever the API returns fewer
	// members than the state knows about — a subset result means propagation is
	// still in progress. In "overwrite" mode the API is authoritative so we
	// accept whatever it returns. A subsequent refresh will re-read the
	// membership once propagation has completed.
	if !strings.EqualFold("overwrite", mode) && len(result) < len(currentMembers) {
		result = currentMembers
	}

	membersVal, diags := types.SetValueFrom(ctx, types.StringType, result)
	if diags.HasError() {
		return fmt.Errorf("building members set: %s", diags)
	}
	model.Members = membersVal
	model.ProjectID = types.StringValue(team.ProjectId.String())
	model.TeamID = types.StringValue(team.Id.String())
	return nil
}

func (r *TeamMembersResource) toIface(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
}
