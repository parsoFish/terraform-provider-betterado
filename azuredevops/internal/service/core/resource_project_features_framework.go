package core

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/featuremanagement"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource                = (*ProjectFeaturesResource)(nil)
	_ resource.ResourceWithConfigure   = (*ProjectFeaturesResource)(nil)
	_ resource.ResourceWithImportState = (*ProjectFeaturesResource)(nil)
)

// ProjectFeaturesResource is the terraform-plugin-framework implementation of
// betterado_project_features.
type ProjectFeaturesResource struct {
	client *client.AggregatedClient
}

// NewProjectFeaturesResource returns a new resource.Resource for
// betterado_project_features.
func NewProjectFeaturesResource() resource.Resource {
	return &ProjectFeaturesResource{}
}

// projectFeaturesModel is the Terraform state model for betterado_project_features.
type projectFeaturesModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Features  types.Map    `tfsdk:"features"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ProjectFeaturesResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_project_features"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ProjectFeaturesResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					projectStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
						"must be a valid UUID",
					),
				},
				PlanModifiers: []planmodifier.String{
					projectForceReplace(),
				},
			},
			"features": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Map{
					mapvalidator.KeysAre(stringvalidator.OneOf(
						"boards", "repositories", "pipelines", "testplans", "artifacts",
					)),
					mapvalidator.ValueStringsAre(stringvalidator.OneOf(
						"enabled", "disabled",
					)),
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *ProjectFeaturesResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectFeaturesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model projectFeaturesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()

	features, d := featuresFromMap(ctx, model.Features)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyFeatureStates(ctx, projectID, features); err != nil {
		resp.Diagnostics.AddError("setting project features", err.Error())
		return
	}

	model.ID = types.StringValue(projectID)

	d = r.readFeaturesIntoModel(ctx, &model)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectFeaturesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model projectFeaturesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	d := r.readFeaturesIntoModel(ctx, &model)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ProjectFeaturesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state projectFeaturesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Carry the ID from state to plan model.
	plan.ID = state.ID

	projectID := plan.ProjectID.ValueString()

	features, d := featuresFromMap(ctx, plan.Features)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyFeatureStates(ctx, projectID, features); err != nil {
		resp.Diagnostics.AddError("updating project features", err.Error())
		return
	}

	d = r.readFeaturesIntoModel(ctx, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete restores all managed features to "enabled" (matching SDKv2 behaviour).
func (r *ProjectFeaturesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model projectFeaturesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()

	features, d := featuresFromMap(ctx, model.Features)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Restore every managed feature to "enabled".
	for k := range features {
		features[k] = string(featuremanagement.ContributedFeatureEnabledValueValues.Enabled)
	}

	if err := r.applyFeatureStates(ctx, projectID, features); err != nil {
		resp.Diagnostics.AddError("restoring project features on delete", err.Error())
		return
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

// ImportState supports import by project UUID. All known feature types are
// imported and their current state is read from ADO.
func (r *ProjectFeaturesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectID := req.ID

	// Seed all known features so readFeaturesIntoModel can filter them.
	allFeatures := make(map[string]string, len(projectFeatureNameMapReverse))
	for k := range projectFeatureNameMapReverse {
		allFeatures[string(k)] = string(featuremanagement.ContributedFeatureEnabledValueValues.Enabled)
	}
	featuresVal, d := types.MapValueFrom(ctx, types.StringType, allFeatures)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	model := projectFeaturesModel{
		ID:        types.StringValue(projectID),
		ProjectID: types.StringValue(projectID),
		Features:  featuresVal,
	}

	diags := r.readFeaturesIntoModel(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// featuresFromMap converts a types.Map of string elements to map[string]string.
func featuresFromMap(ctx context.Context, m types.Map) (map[string]string, diag.Diagnostics) {
	elems := make(map[string]types.String, len(m.Elements()))
	d := m.ElementsAs(ctx, &elems, false)
	if d.HasError() {
		return nil, d
	}
	result := make(map[string]string, len(elems))
	for k, v := range elems {
		result[k] = v.ValueString()
	}
	return result, d
}

// readFeaturesIntoModel fetches the current feature states from ADO and updates
// model.Features with only the keys already present in model.Features (matching
// SDKv2 filter behaviour — we only track what the config declares).
func (r *ProjectFeaturesResource) readFeaturesIntoModel(ctx context.Context, model *projectFeaturesModel) diag.Diagnostics {
	var diags diag.Diagnostics

	projectID := model.ProjectID.ValueString()
	if projectID == "" {
		projectID = model.ID.ValueString()
	}

	// Keys tracked by this resource (from the current state/plan Features map).
	tracked, d := featuresFromMap(ctx, model.Features)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// Fetch all feature states from ADO.
	allStates, err := getProjectFeatureStates(ctx, r.client.FeatureManagementClient, projectID)
	if err != nil {
		diags.AddError("reading project feature states", err.Error())
		return diags
	}

	// Build the new features map containing only the keys the config manages.
	updated := make(map[string]string, len(tracked))
	for k := range tracked {
		if state, ok := (*allStates)[ProjectFeatureType(k)]; ok {
			updated[k] = string(state)
		}
	}

	featuresVal, d2 := types.MapValueFrom(ctx, types.StringType, updated)
	diags.Append(d2...)
	if diags.HasError() {
		return diags
	}
	model.Features = featuresVal
	return diags
}

// GetProjectFeatureStatesForEvidence is an exported helper for acceptance tests
// that need to capture live API evidence of project feature states.
// Returns the raw feature-state map from ADO (keyed by feature short-name,
// e.g. "testplans", "boards").
func GetProjectFeatureStatesForEvidence(clients *client.AggregatedClient, projectID string) (map[ProjectFeatureType]featuremanagement.ContributedFeatureEnabledValue, error) {
	ctx := clients.Ctx
	states, err := getProjectFeatureStates(ctx, clients.FeatureManagementClient, projectID)
	if err != nil {
		return nil, err
	}
	return *states, nil
}

// applyFeatureStates calls the ADO featuremanagement API to set the given
// feature → state mapping for a project. It verifies the returned state from
// SetFeatureStateForScope matches the intended state — if the API accepts the
// request but silently ignores it (e.g. license restrictions on TestPlans), a
// clear error is returned rather than letting Terraform emit a misleading
// "inconsistent result after apply" panic.
func (r *ProjectFeaturesResource) applyFeatureStates(ctx context.Context, projectID string, features map[string]string) error {
	for k, v := range features {
		enabledValue := featuremanagement.ContributedFeatureEnabledValue(v)
		featureAPIID, ok := projectFeatureNameMapReverse[ProjectFeatureType(k)]
		if !ok {
			return fmt.Errorf("unknown feature: %s, available: boards, repositories, pipelines, testplans, artifacts", k)
		}
		result, err := r.client.FeatureManagementClient.SetFeatureStateForScope(ctx, featuremanagement.SetFeatureStateForScopeArgs{
			Feature: &featuremanagement.ContributedFeatureState{
				FeatureId: converter.String(featureAPIID),
				State:     &enabledValue,
				Scope: &featuremanagement.ContributedFeatureSettingScope{
					SettingScope: converter.String("project"),
					UserScoped:   converter.Bool(false),
				},
			},
			FeatureId:  converter.String(featureAPIID),
			UserScope:  converter.String("host"),
			ScopeName:  converter.String("project"),
			ScopeValue: &projectID,
		})
		if err != nil {
			return fmt.Errorf("failed to update feature %s: %w", k, err)
		}
		// Verify the API actually applied the requested state. Some features
		// (e.g. ms.vss-test-web.test / testplans) require specific licensing;
		// the API returns HTTP 200 but leaves the state unchanged. Surface this
		// as an explicit error so callers see a meaningful message instead of a
		// Terraform "inconsistent result after apply" panic.
		if result != nil && result.State != nil && string(*result.State) != v {
			return fmt.Errorf(
				"feature %q: requested state %q but ADO returned %q — "+
					"the feature may not be available in this project or organization "+
					"(check licensing; TestPlans requires a Basic+TestPlans or Visual Studio subscription)",
				k, v, string(*result.State),
			)
		}
	}
	return nil
}
