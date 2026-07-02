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
	_ resource.Resource                = (*BuildValidationResource)(nil)
	_ resource.ResourceWithConfigure   = (*BuildValidationResource)(nil)
	_ resource.ResourceWithImportState = (*BuildValidationResource)(nil)
)

// BuildValidationResource is the framework implementation of betterado_branch_policy_build_validation.
type BuildValidationResource struct {
	client *client.AggregatedClient
}

// NewBuildValidationResource returns a new resource.Resource for betterado_branch_policy_build_validation.
func NewBuildValidationResource() resource.Resource {
	return &BuildValidationResource{}
}

func (r *BuildValidationResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_branch_policy_build_validation"
}

func (r *BuildValidationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BuildValidationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
						"build_definition_id": schema.Int64Attribute{
							Required: true,
						},
						"display_name": schema.StringAttribute{
							Required: true,
						},
						"manual_queue_only": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyBool(false),
						},
						"queue_on_source_update_only": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyBool(true),
						},
						"valid_duration": schema.Int64Attribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyInt64(720),
						},
						"filename_patterns": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
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

type buildValidationModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Blocking  types.Bool   `tfsdk:"blocking"`
	Settings  types.List   `tfsdk:"settings"`
}

type buildValidationSettingsModel struct {
	BuildDefinitionID       types.Int64  `tfsdk:"build_definition_id"`
	DisplayName             types.String `tfsdk:"display_name"`
	ManualQueueOnly         types.Bool   `tfsdk:"manual_queue_only"`
	QueueOnSourceUpdateOnly types.Bool   `tfsdk:"queue_on_source_update_only"`
	ValidDuration           types.Int64  `tfsdk:"valid_duration"`
	FilenamePatterns        types.List   `tfsdk:"filename_patterns"`
	Scope                   types.List   `tfsdk:"scope"`
}

func buildValidationSettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"build_definition_id":         types.Int64Type,
		"display_name":                types.StringType,
		"manual_queue_only":           types.BoolType,
		"queue_on_source_update_only": types.BoolType,
		"valid_duration":              types.Int64Type,
		"filename_patterns":           types.ListType{ElemType: types.StringType},
		"scope":                       types.ListType{ElemType: types.ObjectType{AttrTypes: scopeAttrTypes()}},
	}
}

// ── CRUD ─────────────────────────────────────────────────────────────────────

func (r *BuildValidationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model buildValidationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pc, projectID, diags := expandBuildValidation(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.PolicyClient.CreatePolicyConfiguration(r.client.Ctx, policy.CreatePolicyConfigurationArgs{
		Configuration: pc, Project: projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating branch policy build validation", err.Error())
		return
	}
	model.ID = types.StringValue(strconv.Itoa(*created.Id))
	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *BuildValidationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model buildValidationModel
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

func (r *BuildValidationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan buildValidationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state buildValidationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	pc, projectID, diags := expandBuildValidation(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.PolicyClient.UpdatePolicyConfiguration(r.client.Ctx, policy.UpdatePolicyConfigurationArgs{
		ConfigurationId: pc.Id, Configuration: pc, Project: projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating branch policy build validation", err.Error())
		return
	}
	resp.Diagnostics.Append(r.readIntoModel(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BuildValidationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model buildValidationModel
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
		resp.Diagnostics.AddError("Error deleting branch policy build validation", err.Error())
	}
}

func (r *BuildValidationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPolicyState(ctx, req, resp)
}

func (r *BuildValidationResource) readIntoModel(ctx context.Context, model *buildValidationModel) diag.Diagnostics {
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
		diags.AddError("Error reading branch policy build validation", err.Error())
		return diags
	}
	diags.Append(flattenBuildValidation(ctx, model, pc, &projectID)...)
	return diags
}

func expandBuildValidation(ctx context.Context, model *buildValidationModel) (*policy.PolicyConfiguration, *string, diag.Diagnostics) {
	var diags diag.Diagnostics
	projectID := model.ProjectID.ValueString()

	var settingsList []buildValidationSettingsModel
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

	var patterns []string
	if !s.FilenamePatterns.IsNull() && !s.FilenamePatterns.IsUnknown() {
		diags.Append(s.FilenamePatterns.ElementsAs(ctx, &patterns, false)...)
		if diags.HasError() {
			return nil, nil, diags
		}
	}

	typeID := BuildValidation
	policySettings := map[string]interface{}{
		"buildDefinitionId":       s.BuildDefinitionID.ValueInt64(),
		"displayName":             s.DisplayName.ValueString(),
		"manualQueueOnly":         s.ManualQueueOnly.ValueBool(),
		"queueOnSourceUpdateOnly": s.QueueOnSourceUpdateOnly.ValueBool(),
		"validDuration":           s.ValidDuration.ValueInt64(),
		"filenamePatterns":        patterns,
		"scope":                   scopes,
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

func flattenBuildValidation(ctx context.Context, model *buildValidationModel, pc *policy.PolicyConfiguration, projectID *string) diag.Diagnostics {
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
	var raw buildValidationPolicySettings
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

	patterns := make([]types.String, len(raw.FilenamePatterns))
	for i, v := range raw.FilenamePatterns {
		patterns[i] = types.StringValue(v)
	}
	patternsList, d := types.ListValueFrom(ctx, types.StringType, patterns)
	diags.Append(d...)

	sm := buildValidationSettingsModel{
		BuildDefinitionID:       types.Int64Value(int64(raw.BuildDefinitionID)),
		DisplayName:             types.StringValue(raw.PolicyDisplayName),
		ManualQueueOnly:         types.BoolValue(raw.ManualQueueOnly),
		QueueOnSourceUpdateOnly: types.BoolValue(raw.QueueOnSourceUpdateOnly),
		ValidDuration:           types.Int64Value(int64(raw.ValidDuration)),
		FilenamePatterns:        patternsList,
		Scope:                   scopeList,
	}

	settingsVal, d := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: buildValidationSettingsAttrTypes()},
		[]buildValidationSettingsModel{sm})
	diags.Append(d...)
	model.Settings = settingsVal
	return diags
}
