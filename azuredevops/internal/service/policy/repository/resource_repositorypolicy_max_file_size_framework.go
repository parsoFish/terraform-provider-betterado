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
	_ resource.Resource                = (*MaxFileSizeResource)(nil)
	_ resource.ResourceWithConfigure   = (*MaxFileSizeResource)(nil)
	_ resource.ResourceWithImportState = (*MaxFileSizeResource)(nil)
)

// MaxFileSizeResource is the framework implementation of
// betterado_repository_policy_max_file_size.
type MaxFileSizeResource struct {
	client *client.AggregatedClient
}

// NewMaxFileSizeResource returns a new resource.Resource.
func NewMaxFileSizeResource() resource.Resource {
	return &MaxFileSizeResource{}
}

func (r *MaxFileSizeResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_repository_policy_max_file_size"
}

func (r *MaxFileSizeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MaxFileSizeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"max_file_size": schema.Int64Attribute{
				Required: true,
			},
		},
	}
}

// ── State model ───────────────────────────────────────────────────────────────

type maxFileSizeModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Blocking      types.Bool   `tfsdk:"blocking"`
	RepositoryIDs types.List   `tfsdk:"repository_ids"`
	MaxFileSize   types.Int64  `tfsdk:"max_file_size"`
}

// ── CRUD ─────────────────────────────────────────────────────────────────────

func (r *MaxFileSizeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model maxFileSizeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pc, projectID, diags := expandMaxFileSize(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.PolicyClient.CreatePolicyConfiguration(r.client.Ctx, policy.CreatePolicyConfigurationArgs{
		Configuration: pc, Project: projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating repository policy max file size", err.Error())
		return
	}
	model.ID = types.StringValue(strconv.Itoa(*created.Id))
	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *MaxFileSizeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model maxFileSizeModel
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

func (r *MaxFileSizeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan maxFileSizeModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state maxFileSizeModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	pc, projectID, diags := expandMaxFileSize(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.PolicyClient.UpdatePolicyConfiguration(r.client.Ctx, policy.UpdatePolicyConfigurationArgs{
		ConfigurationId: pc.Id, Configuration: pc, Project: projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating repository policy max file size", err.Error())
		return
	}
	resp.Diagnostics.Append(r.readIntoModel(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MaxFileSizeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model maxFileSizeModel
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
		resp.Diagnostics.AddError("Error deleting repository policy max file size", err.Error())
	}
}

func (r *MaxFileSizeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importRepoPolicyState(ctx, req, resp)
}

func (r *MaxFileSizeResource) readIntoModel(ctx context.Context, model *maxFileSizeModel) diag.Diagnostics {
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
		diags.AddError("Error reading repository policy max file size", err.Error())
		return diags
	}
	diags.Append(flattenMaxFileSize(ctx, model, pc, &projectID)...)
	return diags
}

// maxFileSizeUnitBytes is 1 MB in bytes, matching the SDKv2 UNIT constant.
const maxFileSizeUnitBytes = 1024 * 1024

func expandMaxFileSize(ctx context.Context, model *maxFileSizeModel) (*policy.PolicyConfiguration, *string, diag.Diagnostics) {
	var diags diag.Diagnostics
	projectID := model.ProjectID.ValueString()

	scopes, d := expandRepositoryIDs(ctx, model.RepositoryIDs)
	diags.Append(d...)
	if diags.HasError() {
		return nil, nil, diags
	}

	typeID := FileSize
	policySettings := map[string]interface{}{
		"maximumGitBlobSizeInBytes": model.MaxFileSize.ValueInt64() * maxFileSizeUnitBytes,
		"scope":                     scopes,
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

func flattenMaxFileSize(ctx context.Context, model *maxFileSizeModel, pc *policy.PolicyConfiguration, projectID *string) diag.Diagnostics {
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

	var maxFileSize int64
	if v, ok := rawMap["maximumGitBlobSizeInBytes"]; ok && v != nil {
		if f, ok := v.(float64); ok {
			maxFileSize = int64(f) / maxFileSizeUnitBytes
		}
	}
	model.MaxFileSize = types.Int64Value(maxFileSize)
	return diags
}
