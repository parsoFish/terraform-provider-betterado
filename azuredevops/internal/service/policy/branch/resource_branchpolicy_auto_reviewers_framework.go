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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/policy"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = (*AutoReviewersResource)(nil)
	_ resource.ResourceWithConfigure   = (*AutoReviewersResource)(nil)
	_ resource.ResourceWithImportState = (*AutoReviewersResource)(nil)
)

// AutoReviewersResource is the framework implementation of betterado_branch_policy_auto_reviewers.
type AutoReviewersResource struct {
	client *client.AggregatedClient
}

// NewAutoReviewersResource returns a new resource.Resource for betterado_branch_policy_auto_reviewers.
func NewAutoReviewersResource() resource.Resource {
	return &AutoReviewersResource{}
}

func (r *AutoReviewersResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_branch_policy_auto_reviewers"
}

func (r *AutoReviewersResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AutoReviewersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
						"auto_reviewer_ids": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
						"path_filters": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
						},
						"message": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyString(""),
						},
						"submitter_can_vote": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyBool(false),
						},
						"minimum_number_of_reviewers": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyInt64(1),
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

// ── State models ─────────────────────────────────────────────────────────────

type autoReviewersModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Blocking  types.Bool   `tfsdk:"blocking"`
	Settings  types.List   `tfsdk:"settings"`
}

type autoReviewersSettingsModel struct {
	AutoReviewerIDs          types.List   `tfsdk:"auto_reviewer_ids"`
	PathFilters              types.List   `tfsdk:"path_filters"`
	Message                  types.String `tfsdk:"message"`
	SubmitterCanVote         types.Bool   `tfsdk:"submitter_can_vote"`
	MinimumNumberOfReviewers types.Int64  `tfsdk:"minimum_number_of_reviewers"`
	Scope                    types.List   `tfsdk:"scope"`
}

func autoReviewersSettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"auto_reviewer_ids":           types.ListType{ElemType: types.StringType},
		"path_filters":                types.ListType{ElemType: types.StringType},
		"message":                     types.StringType,
		"submitter_can_vote":          types.BoolType,
		"minimum_number_of_reviewers": types.Int64Type,
		"scope":                       types.ListType{ElemType: types.ObjectType{AttrTypes: scopeAttrTypes()}},
	}
}

// ── CRUD ─────────────────────────────────────────────────────────────────────

func (r *AutoReviewersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model autoReviewersModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pc, projectID, diags := expandAutoReviewers(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.client.PolicyClient.CreatePolicyConfiguration(r.client.Ctx, policy.CreatePolicyConfigurationArgs{
		Configuration: pc,
		Project:       projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating branch policy auto reviewers", err.Error())
		return
	}
	model.ID = types.StringValue(strconv.Itoa(*created.Id))

	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *AutoReviewersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model autoReviewersModel
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

func (r *AutoReviewersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan autoReviewersModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state autoReviewersModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID

	pc, projectID, diags := expandAutoReviewers(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.PolicyClient.UpdatePolicyConfiguration(r.client.Ctx, policy.UpdatePolicyConfigurationArgs{
		ConfigurationId: pc.Id,
		Configuration:   pc,
		Project:         projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating branch policy auto reviewers", err.Error())
		return
	}

	resp.Diagnostics.Append(r.readIntoModel(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AutoReviewersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model autoReviewersModel
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
		resp.Diagnostics.AddError("Error deleting branch policy auto reviewers", err.Error())
	}
}

func (r *AutoReviewersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPolicyState(ctx, req, resp)
}

// ── Internal helpers ─────────────────────────────────────────────────────────

func (r *AutoReviewersResource) readIntoModel(ctx context.Context, model *autoReviewersModel) diag.Diagnostics {
	var diags diag.Diagnostics
	policyID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid policy ID", err.Error())
		return diags
	}
	projectID := model.ProjectID.ValueString()

	pc, err := r.client.PolicyClient.GetPolicyConfiguration(r.client.Ctx, policy.GetPolicyConfigurationArgs{
		Project:         &projectID,
		ConfigurationId: &policyID,
	})
	if utils.ResponseWasNotFound(err) || (pc != nil && pc.IsDeleted != nil && *pc.IsDeleted) {
		model.ID = types.StringNull()
		return diags
	}
	if err != nil {
		diags.AddError("Error reading branch policy auto reviewers", err.Error())
		return diags
	}
	diags.Append(flattenAutoReviewers(ctx, model, pc, &projectID)...)
	return diags
}

func expandAutoReviewers(ctx context.Context, model *autoReviewersModel) (*policy.PolicyConfiguration, *string, diag.Diagnostics) {
	var diags diag.Diagnostics
	projectID := model.ProjectID.ValueString()

	var settingsList []autoReviewersSettingsModel
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

	var reviewerIDs []string
	diags.Append(s.AutoReviewerIDs.ElementsAs(ctx, &reviewerIDs, false)...)
	if diags.HasError() {
		return nil, nil, diags
	}

	var pathFilters []string
	if !s.PathFilters.IsNull() && !s.PathFilters.IsUnknown() {
		diags.Append(s.PathFilters.ElementsAs(ctx, &pathFilters, false)...)
		if diags.HasError() {
			return nil, nil, diags
		}
	}

	typeID := AutoReviewers
	policySettings := map[string]interface{}{
		"creatorVoteCounts":    s.SubmitterCanVote.ValueBool(),
		"message":              s.Message.ValueString(),
		"minimumApproverCount": s.MinimumNumberOfReviewers.ValueInt64(),
		"requiredReviewerIds":  reviewerIDs,
		"filenamePatterns":     pathFilters,
		"scope":                scopes,
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

func flattenAutoReviewers(ctx context.Context, model *autoReviewersModel, pc *policy.PolicyConfiguration, projectID *string) diag.Diagnostics {
	var diags diag.Diagnostics
	model.ProjectID = types.StringValue(*projectID)
	model.Enabled = types.BoolValue(pc.IsEnabled != nil && *pc.IsEnabled)
	model.Blocking = types.BoolValue(pc.IsBlocking != nil && *pc.IsBlocking)
	model.ID = types.StringValue(strconv.Itoa(*pc.Id))

	settingsJSON, err := json.Marshal(pc.Settings)
	if err != nil {
		diags.AddError("Failed to marshal policy settings", err.Error())
		return diags
	}
	var raw autoReviewerPolicySettings
	if err := json.Unmarshal(settingsJSON, &raw); err != nil {
		diags.AddError("Failed to unmarshal policy settings", err.Error())
		return diags
	}

	var rawMap map[string]interface{}
	if err := json.Unmarshal(settingsJSON, &rawMap); err != nil {
		diags.AddError("Failed to unmarshal policy settings map", err.Error())
		return diags
	}

	scopeList, d := flattenScopesFramework(ctx, rawMap)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	reviewerIDs := make([]types.String, len(raw.AutoReviewerIds))
	for i, v := range raw.AutoReviewerIds {
		reviewerIDs[i] = types.StringValue(v)
	}
	reviewerIDsList, d := types.ListValueFrom(ctx, types.StringType, reviewerIDs)
	diags.Append(d...)

	pathFilters := make([]types.String, len(raw.PathFilters))
	for i, v := range raw.PathFilters {
		pathFilters[i] = types.StringValue(v)
	}
	pathFiltersList, d := types.ListValueFrom(ctx, types.StringType, pathFilters)
	diags.Append(d...)

	sm := autoReviewersSettingsModel{
		AutoReviewerIDs:          reviewerIDsList,
		PathFilters:              pathFiltersList,
		Message:                  types.StringValue(raw.DisplayMessage),
		SubmitterCanVote:         types.BoolValue(raw.SubmitterCanVote),
		MinimumNumberOfReviewers: types.Int64Value(int64(raw.MinimumApproverCount)),
		Scope:                    scopeList,
	}

	settingsVal, d := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: autoReviewersSettingsAttrTypes()},
		[]autoReviewersSettingsModel{sm})
	diags.Append(d...)
	model.Settings = settingsVal
	return diags
}
