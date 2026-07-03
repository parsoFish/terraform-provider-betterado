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
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
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

// ── CRUD stubs (WI-3 fills in the implementation) ─────────────────────────────

func (r *FeatureFlagResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.Append(notImplemented("Create")...)
}

func (r *FeatureFlagResource) Read(_ context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	resp.Diagnostics.Append(notImplemented("Read")...)
}

func (r *FeatureFlagResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.Append(notImplemented("Update")...)
}

func (r *FeatureFlagResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.Append(notImplemented("Delete")...)
}

// notImplemented returns an error diagnostic for stub CRUD methods.
func notImplemented(op string) diag.Diagnostics {
	var d diag.Diagnostics
	d.AddError(op+" not implemented", "CRUD logic will be added in WI-3")
	return d
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
