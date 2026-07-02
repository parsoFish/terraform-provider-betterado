package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

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
	_ resource.Resource                = (*MaxPathLengthResource)(nil)
	_ resource.ResourceWithConfigure   = (*MaxPathLengthResource)(nil)
	_ resource.ResourceWithImportState = (*MaxPathLengthResource)(nil)
)

// MaxPathLengthResource is the framework implementation of
// betterado_repository_policy_max_path_length.
type MaxPathLengthResource struct {
	client *client.AggregatedClient
}

// NewMaxPathLengthResource returns a new resource.Resource.
func NewMaxPathLengthResource() resource.Resource {
	return &MaxPathLengthResource{}
}

func (r *MaxPathLengthResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_repository_policy_max_path_length"
}

func (r *MaxPathLengthResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MaxPathLengthResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					repoPolicyUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					repoPolicyRequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  staticRepoPolicyBool(true),
			},
			"blocking": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  staticRepoPolicyBool(true),
			},
			"repository_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     emptyRepoPolicyList(),
			},
			"max_path_length": schema.Int64Attribute{
				Required: true,
			},
		},
	}
}

// ── State model ───────────────────────────────────────────────────────────────

type maxPathLengthModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Blocking      types.Bool   `tfsdk:"blocking"`
	RepositoryIDs types.List   `tfsdk:"repository_ids"`
	MaxPathLength types.Int64  `tfsdk:"max_path_length"`
}

// ── CRUD ─────────────────────────────────────────────────────────────────────

func (r *MaxPathLengthResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model maxPathLengthModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pc, projectID, diags := expandMaxPathLength(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.PolicyClient.CreatePolicyConfiguration(r.client.Ctx, policy.CreatePolicyConfigurationArgs{
		Configuration: pc, Project: projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating repository policy max path length", err.Error())
		return
	}
	model.ID = types.StringValue(strconv.Itoa(*created.Id))
	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *MaxPathLengthResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model maxPathLengthModel
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

func (r *MaxPathLengthResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan maxPathLengthModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state maxPathLengthModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	pc, projectID, diags := expandMaxPathLength(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.PolicyClient.UpdatePolicyConfiguration(r.client.Ctx, policy.UpdatePolicyConfigurationArgs{
		ConfigurationId: pc.Id, Configuration: pc, Project: projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating repository policy max path length", err.Error())
		return
	}
	resp.Diagnostics.Append(r.readIntoModel(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MaxPathLengthResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model maxPathLengthModel
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
		resp.Diagnostics.AddError("Error deleting repository policy max path length", err.Error())
	}
}

func (r *MaxPathLengthResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importRepoPolicyState(ctx, req, resp)
}

func (r *MaxPathLengthResource) readIntoModel(ctx context.Context, model *maxPathLengthModel) diag.Diagnostics {
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
		diags.AddError("Error reading repository policy max path length", err.Error())
		return diags
	}
	diags.Append(flattenMaxPathLength(ctx, model, pc, &projectID)...)
	return diags
}

func expandMaxPathLength(ctx context.Context, model *maxPathLengthModel) (*policy.PolicyConfiguration, *string, diag.Diagnostics) {
	var diags diag.Diagnostics
	projectID := model.ProjectID.ValueString()

	scopes, d := expandRepositoryIDs(ctx, model.RepositoryIDs)
	diags.Append(d...)
	if diags.HasError() {
		return nil, nil, diags
	}

	typeID := PathLength
	policySettings := map[string]interface{}{
		"maxPathLength": model.MaxPathLength.ValueInt64(),
		"scope":         scopes,
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

func flattenMaxPathLength(ctx context.Context, model *maxPathLengthModel, pc *policy.PolicyConfiguration, projectID *string) diag.Diagnostics {
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

	repoIDs, d := flattenRepositoryIDs(ctx, rawMap)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}
	model.RepositoryIDs = repoIDs

	var maxPathLength int64
	if v, ok := rawMap["maxPathLength"]; ok && v != nil {
		if f, ok := v.(float64); ok {
			maxPathLength = int64(f)
		}
	}
	model.MaxPathLength = types.Int64Value(maxPathLength)
	return diags
}
