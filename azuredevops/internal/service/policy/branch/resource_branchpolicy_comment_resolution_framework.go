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
	_ resource.Resource                = (*CommentResolutionResource)(nil)
	_ resource.ResourceWithConfigure   = (*CommentResolutionResource)(nil)
	_ resource.ResourceWithImportState = (*CommentResolutionResource)(nil)
)

// CommentResolutionResource is the framework implementation of betterado_branch_policy_comment_resolution.
type CommentResolutionResource struct {
	client *client.AggregatedClient
}

// NewCommentResolutionResource returns a new resource.Resource for betterado_branch_policy_comment_resolution.
func NewCommentResolutionResource() resource.Resource {
	return &CommentResolutionResource{}
}

func (r *CommentResolutionResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_branch_policy_comment_resolution"
}

func (r *CommentResolutionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CommentResolutionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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

type commentResolutionModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Blocking  types.Bool   `tfsdk:"blocking"`
	Settings  types.List   `tfsdk:"settings"`
}

type commentResolutionSettingsModel struct {
	Scope types.List `tfsdk:"scope"`
}

func commentResolutionSettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"scope": types.ListType{ElemType: types.ObjectType{AttrTypes: scopeAttrTypes()}},
	}
}

// ── CRUD ─────────────────────────────────────────────────────────────────────

func (r *CommentResolutionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model commentResolutionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pc, projectID, diags := expandCommentResolution(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.PolicyClient.CreatePolicyConfiguration(r.client.Ctx, policy.CreatePolicyConfigurationArgs{
		Configuration: pc, Project: projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating branch policy comment resolution", err.Error())
		return
	}
	model.ID = types.StringValue(strconv.Itoa(*created.Id))
	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *CommentResolutionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model commentResolutionModel
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

func (r *CommentResolutionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan commentResolutionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state commentResolutionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	pc, projectID, diags := expandCommentResolution(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.PolicyClient.UpdatePolicyConfiguration(r.client.Ctx, policy.UpdatePolicyConfigurationArgs{
		ConfigurationId: pc.Id, Configuration: pc, Project: projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating branch policy comment resolution", err.Error())
		return
	}
	resp.Diagnostics.Append(r.readIntoModel(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CommentResolutionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model commentResolutionModel
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
		ConfigurationId: &policyID, Project: &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting branch policy comment resolution", err.Error())
	}
}

func (r *CommentResolutionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPolicyState(ctx, req, resp)
}

func (r *CommentResolutionResource) readIntoModel(ctx context.Context, model *commentResolutionModel) diag.Diagnostics {
	var diags diag.Diagnostics
	policyID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid policy ID", err.Error())
		return diags
	}
	projectID := model.ProjectID.ValueString()
	pc, err := r.client.PolicyClient.GetPolicyConfiguration(r.client.Ctx, policy.GetPolicyConfigurationArgs{
		Project: &projectID, ConfigurationId: &policyID,
	})
	if utils.ResponseWasNotFound(err) || (pc != nil && pc.IsDeleted != nil && *pc.IsDeleted) {
		model.ID = types.StringNull()
		return diags
	}
	if err != nil {
		diags.AddError("Error reading branch policy comment resolution", err.Error())
		return diags
	}
	diags.Append(flattenCommentResolution(ctx, model, pc, &projectID)...)
	return diags
}

func expandCommentResolution(ctx context.Context, model *commentResolutionModel) (*policy.PolicyConfiguration, *string, diag.Diagnostics) {
	var diags diag.Diagnostics
	projectID := model.ProjectID.ValueString()

	var settingsList []commentResolutionSettingsModel
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

	typeID := CommentResolution
	policySettings := map[string]interface{}{
		"scope": scopes,
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

func flattenCommentResolution(ctx context.Context, model *commentResolutionModel, pc *policy.PolicyConfiguration, projectID *string) diag.Diagnostics {
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

	sm := commentResolutionSettingsModel{Scope: scopeList}
	settingsVal, d := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: commentResolutionSettingsAttrTypes()},
		[]commentResolutionSettingsModel{sm})
	diags.Append(d...)
	model.Settings = settingsVal
	return diags
}
