package branch

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/policy"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// Compile-time interface checks.
var (
	_ resource.Resource              = (*MinReviewersResource)(nil)
	_ resource.ResourceWithConfigure = (*MinReviewersResource)(nil)
	_ resource.ResourceWithImportState = (*MinReviewersResource)(nil)
)

// MinReviewersResource is the framework implementation of betterado_branch_policy_min_reviewers.
type MinReviewersResource struct {
	client *client.AggregatedClient
}

// NewMinReviewersResource returns a new resource.Resource for betterado_branch_policy_min_reviewers.
func NewMinReviewersResource() resource.Resource {
	return &MinReviewersResource{}
}

func (r *MinReviewersResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_branch_policy_min_reviewers"
}

func (r *MinReviewersResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	agg, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data",
			fmt.Sprintf("Expected *client.AggregatedClient, got: %T", req.ProviderData))
		return
	}
	r.client = agg
}

// ── Schema ──────────────────────────────────────────────────────────────────

func (r *MinReviewersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					policyUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					policyRequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  staticPolicyBool(true),
			},
			"blocking": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  staticPolicyBool(true),
			},
		},
		Blocks: map[string]schema.Block{
			"settings": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"reviewer_count": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyInt64(1),
						},
						"submitter_can_vote": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyBool(false),
						},
						"allow_completion_with_rejects_or_waits": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyBool(false),
						},
						"on_last_iteration_require_vote": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyBool(false),
						},
						"on_each_iteration_require_vote": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyBool(false),
						},
						"on_push_reset_approved_votes": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyBool(false),
						},
						"on_push_reset_all_votes": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyBool(false),
						},
						"last_pusher_cannot_approve": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyBool(false),
						},
					},
					Blocks: map[string]schema.Block{
						"scope": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"repository_id": schema.StringAttribute{
										Optional: true,
										Computed: true,
									},
									"repository_ref": schema.StringAttribute{
										Optional: true,
										Computed: true,
									},
									"match_type": schema.StringAttribute{
										Optional: true,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// ── State model ─────────────────────────────────────────────────────────────

type minReviewersModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Blocking  types.Bool   `tfsdk:"blocking"`
	Settings  types.List   `tfsdk:"settings"`
}

type minReviewersSettingsModel struct {
	ReviewerCount                     types.Int64 `tfsdk:"reviewer_count"`
	SubmitterCanVote                  types.Bool  `tfsdk:"submitter_can_vote"`
	AllowCompletionWithRejectsOrWaits types.Bool  `tfsdk:"allow_completion_with_rejects_or_waits"`
	OnLastIterationRequireVote        types.Bool  `tfsdk:"on_last_iteration_require_vote"`
	OnEachIterationRequireVote        types.Bool  `tfsdk:"on_each_iteration_require_vote"`
	OnPushResetApprovedVotes          types.Bool  `tfsdk:"on_push_reset_approved_votes"`
	OnPushResetAllVotes               types.Bool  `tfsdk:"on_push_reset_all_votes"`
	LastPusherCannotApprove           types.Bool  `tfsdk:"last_pusher_cannot_approve"`
	Scope                             types.List  `tfsdk:"scope"`
}

func minReviewersSettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"reviewer_count":                           types.Int64Type,
		"submitter_can_vote":                       types.BoolType,
		"allow_completion_with_rejects_or_waits":   types.BoolType,
		"on_last_iteration_require_vote":            types.BoolType,
		"on_each_iteration_require_vote":            types.BoolType,
		"on_push_reset_approved_votes":              types.BoolType,
		"on_push_reset_all_votes":                   types.BoolType,
		"last_pusher_cannot_approve":                types.BoolType,
		"scope":                                     types.ListType{ElemType: types.ObjectType{AttrTypes: scopeAttrTypes()}},
	}
}

// ── CRUD ─────────────────────────────────────────────────────────────────────

func (r *MinReviewersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model minReviewersModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyConfig, projectID, diags := expandMinReviewers(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.PolicyClient.CreatePolicyConfiguration(r.client.Ctx, policy.CreatePolicyConfigurationArgs{
		Configuration: policyConfig,
		Project:       projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating branch policy min reviewers", err.Error())
		return
	}
	model.ID = types.StringValue(strconv.Itoa(*created.Id))

	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *MinReviewersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model minReviewersModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *MinReviewersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan minReviewersModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state minReviewersModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	policyConfig, projectID, diags := expandMinReviewers(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.PolicyClient.UpdatePolicyConfiguration(r.client.Ctx, policy.UpdatePolicyConfigurationArgs{
		ConfigurationId: policyConfig.Id,
		Configuration:   policyConfig,
		Project:         projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating branch policy min reviewers", err.Error())
		return
	}

	resp.Diagnostics.Append(r.readIntoModel(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MinReviewersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model minReviewersModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policyID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid policy ID", err.Error())
		return
	}
	projectID := model.ProjectID.ValueString()

	err = r.client.PolicyClient.DeletePolicyConfiguration(r.client.Ctx, policy.DeletePolicyConfigurationArgs{
		ConfigurationId: &policyID,
		Project:         &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting branch policy min reviewers", err.Error())
	}
}

func (r *MinReviewersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPolicyState(ctx, req, resp)
}

// ── Internal helpers ─────────────────────────────────────────────────────────

func (r *MinReviewersResource) readIntoModel(ctx context.Context, model *minReviewersModel) diag.Diagnostics {
	var diags diag.Diagnostics

	policyID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid policy ID", err.Error())
		return diags
	}
	projectID := model.ProjectID.ValueString()

	policyConfig, err := r.client.PolicyClient.GetPolicyConfiguration(r.client.Ctx, policy.GetPolicyConfigurationArgs{
		Project:         &projectID,
		ConfigurationId: &policyID,
	})
	if utils.ResponseWasNotFound(err) || (policyConfig != nil && policyConfig.IsDeleted != nil && *policyConfig.IsDeleted) {
		model.ID = types.StringNull()
		return diags
	}
	if err != nil {
		diags.AddError("Error reading branch policy min reviewers", err.Error())
		return diags
	}

	diags.Append(flattenMinReviewers(ctx, model, policyConfig, &projectID)...)
	return diags
}

func expandMinReviewers(ctx context.Context, model *minReviewersModel) (*policy.PolicyConfiguration, *string, diag.Diagnostics) {
	var diags diag.Diagnostics

	projectID := model.ProjectID.ValueString()
	var settingsList []minReviewersSettingsModel
	diags.Append(model.Settings.ElementsAs(ctx, &settingsList, false)...)
	if diags.HasError() {
		return nil, nil, diags
	}
	s := settingsList[0]

	scopes, d := expandScopesFramework(ctx, s.Scope)
	diags.Append(d...)
	if diags.HasError() {
		return nil, nil, diags
	}

	typeID := MinReviewerCount
	policySettings := map[string]interface{}{
		"minimumApproverCount":        s.ReviewerCount.ValueInt64(),
		"creatorVoteCounts":           s.SubmitterCanVote.ValueBool(),
		"allowDownvotes":              s.AllowCompletionWithRejectsOrWaits.ValueBool(),
		"requireVoteOnLastIteration":  s.OnLastIterationRequireVote.ValueBool(),
		"resetOnSourcePush":           s.OnPushResetApprovedVotes.ValueBool(),
		"blockLastPusherVote":         s.LastPusherCannotApprove.ValueBool(),
		"requireVoteOnEachIteration":  s.OnEachIterationRequireVote.ValueBool(),
		"resetRejectionsOnSourcePush": s.OnPushResetAllVotes.ValueBool(),
		"scope":                       scopes,
	}
	if s.OnPushResetAllVotes.ValueBool() {
		policySettings["resetOnSourcePush"] = true
	}

	pc := &policy.PolicyConfiguration{
		IsEnabled:  boolPtr(model.Enabled.ValueBool()),
		IsBlocking: boolPtr(model.Blocking.ValueBool()),
		Type:       &policy.PolicyTypeRef{Id: &typeID},
		Settings:   policySettings,
	}

	if !model.ID.IsNull() && model.ID.ValueString() != "" {
		id, err := strconv.Atoi(model.ID.ValueString())
		if err == nil {
			pc.Id = &id
		}
	}

	return pc, &projectID, diags
}

func flattenMinReviewers(ctx context.Context, model *minReviewersModel, policyConfig *policy.PolicyConfiguration, projectID *string) diag.Diagnostics {
	var diags diag.Diagnostics

	model.ProjectID = types.StringValue(*projectID)
	model.Enabled = types.BoolValue(policyConfig.IsEnabled != nil && *policyConfig.IsEnabled)
	model.Blocking = types.BoolValue(policyConfig.IsBlocking != nil && *policyConfig.IsBlocking)
	model.ID = types.StringValue(strconv.Itoa(*policyConfig.Id))

	// Parse settings JSON
	settingsJSON, err := json.Marshal(policyConfig.Settings)
	if err != nil {
		diags.AddError("Failed to marshal policy settings", err.Error())
		return diags
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(settingsJSON, &raw); err != nil {
		diags.AddError("Failed to unmarshal policy settings", err.Error())
		return diags
	}

	// Parse scopes
	scopeList, d := flattenScopesFramework(ctx, raw)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	sm := minReviewersSettingsModel{
		ReviewerCount:                     types.Int64Value(int64Coerce(raw["minimumApproverCount"])),
		SubmitterCanVote:                  types.BoolValue(boolCoerce(raw["creatorVoteCounts"])),
		AllowCompletionWithRejectsOrWaits: types.BoolValue(boolCoerce(raw["allowDownvotes"])),
		OnLastIterationRequireVote:        types.BoolValue(boolCoerce(raw["requireVoteOnLastIteration"])),
		OnEachIterationRequireVote:        types.BoolValue(boolCoerce(raw["requireVoteOnEachIteration"])),
		OnPushResetApprovedVotes:          types.BoolValue(boolCoerce(raw["resetOnSourcePush"])),
		OnPushResetAllVotes:               types.BoolValue(boolCoerce(raw["resetRejectionsOnSourcePush"])),
		LastPusherCannotApprove:           types.BoolValue(boolCoerce(raw["blockLastPusherVote"])),
		Scope:                             scopeList,
	}

	settingsVal, d := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: minReviewersSettingsAttrTypes()},
		[]minReviewersSettingsModel{sm})
	diags.Append(d...)
	model.Settings = settingsVal
	return diags
}

// boolPtr returns a pointer to a bool value.
func boolPtr(b bool) *bool { return &b }

// ensure defaults import is used
var _ defaults.Bool = staticPolicyBool(false)
