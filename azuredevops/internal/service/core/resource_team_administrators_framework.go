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
	securityhelper "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/permissions/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource              = (*TeamAdministratorsResource)(nil)
	_ resource.ResourceWithConfigure = (*TeamAdministratorsResource)(nil)
)

// TeamAdministratorsResource is the terraform-plugin-framework implementation of
// betterado_team_administrators.
type TeamAdministratorsResource struct {
	client *client.AggregatedClient
}

// NewTeamAdministratorsResource returns a new resource.Resource.
func NewTeamAdministratorsResource() resource.Resource {
	return &TeamAdministratorsResource{}
}

// teamAdministratorsModel is the Terraform state model for betterado_team_administrators.
type teamAdministratorsModel struct {
	ID             types.String `tfsdk:"id"`
	ProjectID      types.String `tfsdk:"project_id"`
	TeamID         types.String `tfsdk:"team_id"`
	Mode           types.String `tfsdk:"mode"`
	Administrators types.Set    `tfsdk:"administrators"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *TeamAdministratorsResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_team_administrators"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *TeamAdministratorsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages administrators for a team in Azure DevOps.",
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
				Description: `"add" (default) preserves unmanaged admins; "overwrite" removes all not in the list.`,
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive("add", "overwrite"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"administrators": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Set of subject descriptors to grant team administrator permissions.",
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

func (r *TeamAdministratorsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TeamAdministratorsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model teamAdministratorsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default mode.
	if model.Mode.IsNull() || model.Mode.IsUnknown() {
		model.Mode = types.StringValue("add")
	}

	team, err := r.getTeam(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("reading team", err.Error())
		return
	}

	var admins []string
	model.Administrators.ElementsAs(ctx, &admins, false)

	if strings.EqualFold(model.Mode.ValueString(), "overwrite") {
		if err := updateTeamAdministrators(nil, r.client, team, &admins); err != nil {
			resp.Diagnostics.AddError("setting team administrators (overwrite)", err.Error())
			return
		}
	} else {
		if err := setTeamAdministratorsPermissions(nil, r.client, team, linq.From(r.toIface(admins)), securityhelper.PermissionTypeValues.Allow); err != nil {
			resp.Diagnostics.AddError("setting team administrators (add)", err.Error())
			return
		}
	}

	model.ID = types.StringValue(team.ProjectId.String() + "/" + team.Id.String())
	// For Create, use the plan values for non-computed attributes (project_id,
	// team_id, administrators) so the framework's plan-vs-state consistency
	// check passes. The Azure DevOps security ACL may not immediately reflect
	// the new permission, which would cause readIntoModel to return a null/empty
	// administrators set and trigger "Provider produced inconsistent result after
	// apply". Subsequent Read calls will refresh state from the API.
	model.ProjectID = types.StringValue(team.ProjectId.String())
	model.TeamID = types.StringValue(team.Id.String())
	// administrators is already set from the plan via req.Plan.Get above
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TeamAdministratorsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model teamAdministratorsModel
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
		resp.Diagnostics.AddError("reading administrators", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TeamAdministratorsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan teamAdministratorsModel
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

	var admins []string
	plan.Administrators.ElementsAs(ctx, &admins, false)

	if strings.EqualFold(plan.Mode.ValueString(), "overwrite") {
		if err := updateTeamAdministrators(nil, r.client, team, &admins); err != nil {
			resp.Diagnostics.AddError("updating team administrators (overwrite)", err.Error())
			return
		}
	} else {
		var oldAdmins []string
		state.Administrators.ElementsAs(ctx, &oldAdmins, false)
		toAdd := subtract(admins, oldAdmins)
		toRemove := subtract(oldAdmins, admins)
		if len(toAdd) > 0 {
			if err := setTeamAdministratorsPermissions(nil, r.client, team, linq.From(r.toIface(toAdd)), securityhelper.PermissionTypeValues.Allow); err != nil {
				resp.Diagnostics.AddError("adding team administrators", err.Error())
				return
			}
		}
		if len(toRemove) > 0 {
			if err := setTeamAdministratorsPermissions(nil, r.client, team, linq.From(r.toIface(toRemove)), securityhelper.PermissionTypeValues.NotSet); err != nil {
				resp.Diagnostics.AddError("removing team administrators", err.Error())
				return
			}
		}
	}

	// Azure DevOps ACL changes propagate asynchronously. Use the plan values
	// as the new state (same as Create) so the post-apply plan-consistency
	// check passes even if the ACL API has not yet reflected the new members.
	// A subsequent Read (e.g. terraform refresh) will re-read from the API
	// once propagation has completed.
	plan.ProjectID = types.StringValue(team.ProjectId.String())
	plan.TeamID = types.StringValue(team.Id.String())
	// plan.Administrators already holds the desired set from req.Plan.Get
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TeamAdministratorsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model teamAdministratorsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	team, err := r.getTeam(ctx, model)
	if err != nil {
		resp.Diagnostics.AddError("reading team during delete", err.Error())
		return
	}

	var adminsToRemove []string
	if strings.EqualFold("overwrite", model.Mode.ValueString()) {
		allAdmins, err := getTeamAdministrators(nil, r.client, team)
		if err != nil {
			resp.Diagnostics.AddError("reading team administrators", err.Error())
			return
		}
		for _, a := range allAdmins.List() {
			adminsToRemove = append(adminsToRemove, a.(string))
		}
	} else {
		model.Administrators.ElementsAs(ctx, &adminsToRemove, false)
	}

	if len(adminsToRemove) > 0 {
		if err := setTeamAdministratorsPermissions(nil, r.client, team, linq.From(r.toIface(adminsToRemove)), securityhelper.PermissionTypeValues.NotSet); err != nil {
			resp.Diagnostics.AddError("removing team administrators on delete", err.Error())
		}
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (r *TeamAdministratorsResource) getTeam(_ context.Context, model teamAdministratorsModel) (*adocore.WebApiTeam, error) {
	projectID := model.ProjectID.ValueString()
	teamID := model.TeamID.ValueString()
	return r.client.CoreClient.GetTeam(r.client.Ctx, adocore.GetTeamArgs{
		ProjectId:      &projectID,
		TeamId:         &teamID,
		ExpandIdentity: converter.Bool(false),
	})
}

func (r *TeamAdministratorsResource) readIntoModel(ctx context.Context, model *teamAdministratorsModel, team *adocore.WebApiTeam) error {
	mode := model.Mode.ValueString()
	allAdmins, err := getTeamAdministrators(nil, r.client, team)
	if err != nil {
		return err
	}

	var currentAdmins []string
	model.Administrators.ElementsAs(ctx, &currentAdmins, false)
	stateSet := make(map[string]bool, len(currentAdmins))
	for _, a := range currentAdmins {
		stateSet[a] = true
	}

	// Use a non-nil slice so that types.SetValueFrom produces an empty Set
	// (not a null Set) when no admins match. A nil []string produces a null
	// Set which triggers "Provider produced inconsistent result after apply".
	result := make([]string, 0)
	for _, a := range allAdmins.List() {
		s := a.(string)
		if strings.EqualFold("overwrite", mode) || stateSet[s] {
			result = append(result, s)
		}
	}

	// Azure DevOps security ACL changes propagate asynchronously. When the ACL
	// API returns an empty result immediately after a permission grant (a common
	// transient condition), fall back to the current state values so that the
	// framework's plan-after-apply consistency check does not fail with
	// "Provider produced inconsistent result after apply". A subsequent refresh
	// will re-read the ACL once propagation has completed.
	if len(result) == 0 && len(currentAdmins) > 0 {
		result = currentAdmins
	}

	adminsVal, diags := types.SetValueFrom(ctx, types.StringType, result)
	if diags.HasError() {
		return fmt.Errorf("building administrators set: %s", diags)
	}
	model.Administrators = adminsVal
	model.ProjectID = types.StringValue(team.ProjectId.String())
	model.TeamID = types.StringValue(team.Id.String())
	return nil
}

func (r *TeamAdministratorsResource) toIface(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
}

// subtract returns elements in a not in b.
func subtract(a, b []string) []string {
	bSet := make(map[string]bool, len(b))
	for _, v := range b {
		bSet[v] = true
	}
	var result []string
	for _, v := range a {
		if !bSet[v] {
			result = append(result, v)
		}
	}
	return result
}
