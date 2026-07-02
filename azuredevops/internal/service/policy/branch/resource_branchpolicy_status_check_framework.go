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
	_ resource.Resource                = (*StatusCheckResource)(nil)
	_ resource.ResourceWithConfigure   = (*StatusCheckResource)(nil)
	_ resource.ResourceWithImportState = (*StatusCheckResource)(nil)
)

// StatusCheckResource is the framework implementation of betterado_branch_policy_status_check.
type StatusCheckResource struct {
	client *client.AggregatedClient
}

// NewStatusCheckResource returns a new resource.Resource for betterado_branch_policy_status_check.
func NewStatusCheckResource() resource.Resource {
	return &StatusCheckResource{}
}

func (r *StatusCheckResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_branch_policy_status_check"
}

func (r *StatusCheckResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StatusCheckResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
						"name": schema.StringAttribute{
							Required: true,
						},
						"genre": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyString(""),
						},
						"author_id": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyString(""),
						},
						"invalidate_on_update": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyBool(false),
						},
						"applicability": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyString("default"),
						},
						"filename_patterns": schema.ListAttribute{
							ElementType: types.StringType,
							Optional:    true,
							Computed:    true,
						},
						"display_name": schema.StringAttribute{
							Optional: true,
							Computed: true,
							Default:  staticPolicyString(""),
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

type statusCheckModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
	Blocking  types.Bool   `tfsdk:"blocking"`
	Settings  types.List   `tfsdk:"settings"`
}

type statusCheckSettingsModel struct {
	Name               types.String `tfsdk:"name"`
	Genre              types.String `tfsdk:"genre"`
	AuthorID           types.String `tfsdk:"author_id"`
	InvalidateOnUpdate types.Bool   `tfsdk:"invalidate_on_update"`
	Applicability      types.String `tfsdk:"applicability"`
	FilenamePatterns   types.List   `tfsdk:"filename_patterns"`
	DisplayName        types.String `tfsdk:"display_name"`
	Scope              types.List   `tfsdk:"scope"`
}

func statusCheckSettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                 types.StringType,
		"genre":                types.StringType,
		"author_id":            types.StringType,
		"invalidate_on_update": types.BoolType,
		"applicability":        types.StringType,
		"filename_patterns":    types.ListType{ElemType: types.StringType},
		"display_name":         types.StringType,
		"scope":                types.ListType{ElemType: types.ObjectType{AttrTypes: scopeAttrTypes()}},
	}
}

// ── CRUD ─────────────────────────────────────────────────────────────────────

func (r *StatusCheckResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model statusCheckModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	pc, projectID, diags := expandStatusCheck(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.PolicyClient.CreatePolicyConfiguration(r.client.Ctx, policy.CreatePolicyConfigurationArgs{
		Configuration: pc, Project: projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating branch policy status check", err.Error())
		return
	}
	model.ID = types.StringValue(strconv.Itoa(*created.Id))
	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *StatusCheckResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model statusCheckModel
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

func (r *StatusCheckResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan statusCheckModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state statusCheckModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	pc, projectID, diags := expandStatusCheck(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.PolicyClient.UpdatePolicyConfiguration(r.client.Ctx, policy.UpdatePolicyConfigurationArgs{
		ConfigurationId: pc.Id, Configuration: pc, Project: projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating branch policy status check", err.Error())
		return
	}
	resp.Diagnostics.Append(r.readIntoModel(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *StatusCheckResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model statusCheckModel
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
		resp.Diagnostics.AddError("Error deleting branch policy status check", err.Error())
	}
}

func (r *StatusCheckResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importPolicyState(ctx, req, resp)
}

func (r *StatusCheckResource) readIntoModel(ctx context.Context, model *statusCheckModel) diag.Diagnostics {
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
		diags.AddError("Error reading branch policy status check", err.Error())
		return diags
	}
	diags.Append(flattenStatusCheck(ctx, model, pc, &projectID)...)
	return diags
}

func expandStatusCheck(ctx context.Context, model *statusCheckModel) (*policy.PolicyConfiguration, *string, diag.Diagnostics) {
	var diags diag.Diagnostics
	projectID := model.ProjectID.ValueString()

	var settingsList []statusCheckSettingsModel
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

	typeID := StatusCheck
	policySettings := map[string]interface{}{
		"statusName":               s.Name.ValueString(),
		"statusGenre":              s.Genre.ValueString(),
		"authorId":                 s.AuthorID.ValueString(),
		"invalidateOnSourceUpdate": s.InvalidateOnUpdate.ValueBool(),
		"defaultDisplayName":       s.DisplayName.ValueString(),
		"filenamePatterns":         patterns,
		"scope":                    scopes,
	}

	if s.Applicability.ValueString() == "conditional" {
		policySettings["policyApplicability"] = 1
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

func flattenStatusCheck(ctx context.Context, model *statusCheckModel, pc *policy.PolicyConfiguration, projectID *string) diag.Diagnostics {
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
		diags.AddError("Failed to unmarshal policy settings", err.Error())
		return diags
	}
	scopeList, d := flattenScopesFramework(ctx, rawMap)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// filename_patterns
	var patternsList types.List
	if patternsRaw, ok := rawMap["filenamePatterns"]; ok && patternsRaw != nil {
		patternsSlice, _ := patternsRaw.([]interface{})
		pats := make([]types.String, len(patternsSlice))
		for i, v := range patternsSlice {
			pats[i] = types.StringValue(stringCoerce(v))
		}
		patternsList, d = types.ListValueFrom(ctx, types.StringType, pats)
	} else {
		patternsList, d = types.ListValueFrom(ctx, types.StringType, []types.String{})
	}
	diags.Append(d...)

	// applicability
	applic := "default"
	if v, ok := rawMap["policyApplicability"]; ok && v != nil {
		if float64(1) == v.(float64) {
			applic = "conditional"
		}
	}

	sm := statusCheckSettingsModel{
		Name:               types.StringValue(stringCoerce(rawMap["statusName"])),
		Genre:              types.StringValue(stringCoerce(rawMap["statusGenre"])),
		AuthorID:           types.StringValue(stringCoerce(rawMap["authorId"])),
		InvalidateOnUpdate: types.BoolValue(boolCoerce(rawMap["invalidateOnSourceUpdate"])),
		Applicability:      types.StringValue(applic),
		FilenamePatterns:   patternsList,
		DisplayName:        types.StringValue(stringCoerce(rawMap["defaultDisplayName"])),
		Scope:              scopeList,
	}

	settingsVal, d := types.ListValueFrom(ctx,
		types.ObjectType{AttrTypes: statusCheckSettingsAttrTypes()},
		[]statusCheckSettingsModel{sm})
	diags.Append(d...)
	model.Settings = settingsVal
	return diags
}
