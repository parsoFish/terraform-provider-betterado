package featuremanagement

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	featuremanagementapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/featuremanagement"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource              = (*FeatureFlagResource)(nil)
	_ resource.ResourceWithConfigure = (*FeatureFlagResource)(nil)
)

// FeatureFlagResource is the terraform-plugin-framework implementation of
// betterado_feature_flag.
type FeatureFlagResource struct {
	client *client.AggregatedClient
}

// NewFeatureFlagResource returns a new resource.Resource for betterado_feature_flag.
func NewFeatureFlagResource() resource.Resource {
	return &FeatureFlagResource{}
}

// featureFlagModel is the Terraform state model for betterado_feature_flag.
type featureFlagModel struct {
	FeatureID  types.String `tfsdk:"feature_id"`
	ScopeName  types.String `tfsdk:"scope_name"`
	ScopeValue types.String `tfsdk:"scope_value"`
	State      types.String `tfsdk:"state"`
	Overridden types.Bool   `tfsdk:"overridden"`
	Reason     types.String `tfsdk:"reason"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *FeatureFlagResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_feature_flag"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *FeatureFlagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"feature_id": schema.StringAttribute{
				Required:    true,
				Description: "The full contribution id of the feature (e.g. \"ms.vss-work.agile\").",
				PlanModifiers: []planmodifier.String{
					ffRequiresReplace(),
				},
			},
			"scope_name": schema.StringAttribute{
				Required:    true,
				Description: "The setting scope name (e.g. \"project\", \"host\").",
				PlanModifiers: []planmodifier.String{
					ffRequiresReplace(),
				},
			},
			"scope_value": schema.StringAttribute{
				Required:    true,
				Description: "The scope value (project ID for project-scope; empty string for host-scope).",
				PlanModifiers: []planmodifier.String{
					ffRequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				Required:    true,
				Description: "The desired feature state. Must be \"enabled\" or \"disabled\".",
				Validators: []validator.String{
					stateOneOfValidator{},
				},
			},
			"overridden": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the feature state has been set by an override rule.",
				PlanModifiers: []planmodifier.Bool{
					ffUseStateForUnknownBool(),
				},
			},
			"reason": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Reason for the feature state.",
				PlanModifiers: []planmodifier.String{
					ffUseStateForUnknownString(),
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *FeatureFlagResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FeatureFlagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feature_flag Create: provider client not configured")
		return
	}

	var model featureFlagModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setFeatureFlag(ctx, r.client.FeatureManagementClient, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readDiags := readFeatureFlag(ctx, r.client.FeatureManagementClient, &model)
	if hasNotFoundDiag(readDiags) {
		resp.Diagnostics.AddError("Feature flag not found after create", "GetFeatureStateForScope returned not-found immediately after create")
		return
	}
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *FeatureFlagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feature_flag Read: provider client not configured")
		return
	}

	var model featureFlagModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readDiags := readFeatureFlag(ctx, r.client.FeatureManagementClient, &model)
	if hasNotFoundDiag(readDiags) {
		// 404 or state == "undefined": remove from Terraform state.
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *FeatureFlagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feature_flag Update: provider client not configured")
		return
	}

	var model featureFlagModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(setFeatureFlag(ctx, r.client.FeatureManagementClient, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readDiags := readFeatureFlag(ctx, r.client.FeatureManagementClient, &model)
	if hasNotFoundDiag(readDiags) {
		resp.Diagnostics.AddError("Feature flag not found after update", "GetFeatureStateForScope returned not-found immediately after update")
		return
	}
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *FeatureFlagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feature_flag Delete: provider client not configured")
		return
	}

	var model featureFlagModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(deleteFeatureFlag(ctx, r.client.FeatureManagementClient, &model)...)
}

// ── Internal helpers (called by CRUD + unit tests) ────────────────────────────

// setFeatureFlag calls SetFeatureStateForScope for Create and Update.
func setFeatureFlag(ctx context.Context, fc featuremanagementapi.Client, model *featureFlagModel) diag.Diagnostics {
	var d diag.Diagnostics

	featureID := model.FeatureID.ValueString()
	scopeName := model.ScopeName.ValueString()
	scopeValue := model.ScopeValue.ValueString()
	stateStr := model.State.ValueString()

	enabledValue := featuremanagementapi.ContributedFeatureEnabledValue(stateStr)

	_, err := fc.SetFeatureStateForScope(ctx, featuremanagementapi.SetFeatureStateForScopeArgs{
		Feature: &featuremanagementapi.ContributedFeatureState{
			FeatureId: converter.String(featureID),
			State:     &enabledValue,
			Scope: &featuremanagementapi.ContributedFeatureSettingScope{
				SettingScope: converter.String(scopeName),
				UserScoped:   converter.Bool(false),
			},
		},
		FeatureId:  converter.String(featureID),
		UserScope:  converter.String("host"),
		ScopeName:  converter.String(scopeName),
		ScopeValue: converter.String(scopeValue),
	})
	if err != nil {
		d.AddError(
			"Error setting feature flag",
			fmt.Sprintf("SetFeatureStateForScope failed for feature %q scope %q/%q: %s", featureID, scopeName, scopeValue, err),
		)
	}
	return d
}

// readFeatureFlag calls GetFeatureStateForScope and populates the model.
// Returns a diagnostic with summary "NotFound" when the API returns 404 or
// the state is "undefined" (resource no longer managed).
func readFeatureFlag(ctx context.Context, fc featuremanagementapi.Client, model *featureFlagModel) diag.Diagnostics {
	var d diag.Diagnostics

	featureID := model.FeatureID.ValueString()
	scopeName := model.ScopeName.ValueString()
	scopeValue := model.ScopeValue.ValueString()

	state, err := fc.GetFeatureStateForScope(ctx, featuremanagementapi.GetFeatureStateForScopeArgs{
		FeatureId:  converter.String(featureID),
		UserScope:  converter.String("host"),
		ScopeName:  converter.String(scopeName),
		ScopeValue: converter.String(scopeValue),
	})

	if err != nil {
		if utils.ResponseWasNotFound(err) {
			d.AddError("NotFound", fmt.Sprintf("feature flag %q at scope %q/%q not found (404)", featureID, scopeName, scopeValue))
			return d
		}
		d.AddError(
			"Error reading feature flag",
			fmt.Sprintf("GetFeatureStateForScope failed for feature %q scope %q/%q: %s", featureID, scopeName, scopeValue, err),
		)
		return d
	}

	if state == nil || state.State == nil || *state.State == featuremanagementapi.ContributedFeatureEnabledValueValues.Undefined {
		d.AddError("NotFound", fmt.Sprintf("feature flag %q at scope %q/%q is undefined (treated as deleted)", featureID, scopeName, scopeValue))
		return d
	}

	model.State = types.StringValue(string(*state.State))

	if state.Overridden != nil {
		model.Overridden = types.BoolValue(*state.Overridden)
	} else {
		model.Overridden = types.BoolValue(false)
	}

	if state.Reason != nil {
		model.Reason = types.StringValue(*state.Reason)
	} else {
		model.Reason = types.StringValue("")
	}

	return d
}

// deleteFeatureFlag sets the feature state to "undefined" (removes management).
func deleteFeatureFlag(ctx context.Context, fc featuremanagementapi.Client, model *featureFlagModel) diag.Diagnostics {
	var d diag.Diagnostics

	featureID := model.FeatureID.ValueString()
	scopeName := model.ScopeName.ValueString()
	scopeValue := model.ScopeValue.ValueString()

	undefinedState := featuremanagementapi.ContributedFeatureEnabledValueValues.Undefined

	_, err := fc.SetFeatureStateForScope(ctx, featuremanagementapi.SetFeatureStateForScopeArgs{
		Feature: &featuremanagementapi.ContributedFeatureState{
			FeatureId: converter.String(featureID),
			State:     &undefinedState,
			Scope: &featuremanagementapi.ContributedFeatureSettingScope{
				SettingScope: converter.String(scopeName),
				UserScoped:   converter.Bool(false),
			},
		},
		FeatureId:  converter.String(featureID),
		UserScope:  converter.String("host"),
		ScopeName:  converter.String(scopeName),
		ScopeValue: converter.String(scopeValue),
	})
	if err != nil {
		d.AddError(
			"Error deleting feature flag",
			fmt.Sprintf("SetFeatureStateForScope(undefined) failed for feature %q scope %q/%q: %s", featureID, scopeName, scopeValue, err),
		)
	}
	return d
}

// hasNotFoundDiag returns true if diags contains an error with summary "NotFound".
func hasNotFoundDiag(d diag.Diagnostics) bool {
	for _, diagItem := range d {
		if diagItem.Summary() == "NotFound" {
			return true
		}
	}
	return false
}

// ── Inline validators ─────────────────────────────────────────────────────────

// stateOneOfValidator enforces that the "state" attribute is one of
// "enabled" or "disabled". Avoids a dependency on
// terraform-plugin-framework-validators which is not yet vendored.
type stateOneOfValidator struct{}

func (v stateOneOfValidator) Description(_ context.Context) string {
	return `value must be one of: "enabled", "disabled"`
}

func (v stateOneOfValidator) MarkdownDescription(_ context.Context) string {
	return "value must be one of: `\"enabled\"`, `\"disabled\"`"
}

func (v stateOneOfValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	val := req.ConfigValue
	if val.IsNull() || val.IsUnknown() {
		return
	}
	s := val.ValueString()
	if s != "enabled" && s != "disabled" {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid feature state",
			fmt.Sprintf("Expected one of [enabled disabled], got: %q", s),
		)
	}
}

// ── Inline plan modifiers ─────────────────────────────────────────────────────

// ffRequiresReplaceModifier marks a string attribute for replacement when its
// value changes (mirrors stringplanmodifier.RequiresReplace).
type ffRequiresReplaceModifier struct{}

func ffRequiresReplace() planmodifier.String { return ffRequiresReplaceModifier{} }

func (m ffRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m ffRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m ffRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ffUseStateForUnknownStringModifier copies the prior state value when the plan
// value is unknown for a string attribute.
type ffUseStateForUnknownStringModifier struct{}

func ffUseStateForUnknownString() planmodifier.String { return ffUseStateForUnknownStringModifier{} }

func (m ffUseStateForUnknownStringModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m ffUseStateForUnknownStringModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m ffUseStateForUnknownStringModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ffUseStateForUnknownBoolModifier copies the prior state value when the plan
// value is unknown for a bool attribute.
type ffUseStateForUnknownBoolModifier struct{}

func ffUseStateForUnknownBool() planmodifier.Bool { return ffUseStateForUnknownBoolModifier{} }

func (m ffUseStateForUnknownBoolModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m ffUseStateForUnknownBoolModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m ffUseStateForUnknownBoolModifier) PlanModifyBool(_ context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}
