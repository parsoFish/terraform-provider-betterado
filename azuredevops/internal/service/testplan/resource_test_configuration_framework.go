package testplan

import (
	"context"
	"fmt"
	"strconv"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/test"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/testplan"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource              = (*TestConfigurationResource)(nil)
	_ resource.ResourceWithConfigure = (*TestConfigurationResource)(nil)
)

// TestConfigurationResource is the terraform-plugin-framework implementation of
// betterado_test_configuration.
type TestConfigurationResource struct {
	client *client.AggregatedClient
}

// NewTestConfigurationResource returns a new resource.Resource for betterado_test_configuration.
func NewTestConfigurationResource() resource.Resource {
	return &TestConfigurationResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *TestConfigurationResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_test_configuration"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *TestConfigurationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					tpUseStateForUnknown{},
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					tpRequiresReplace{},
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"is_default": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  tcStaticBool(false),
			},
			"values": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *TestConfigurationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ── State model ───────────────────────────────────────────────────────────────

// testConfigurationModel is the Terraform state model for betterado_test_configuration.
type testConfigurationModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
	Values      types.Map    `tfsdk:"values"`
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *TestConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model testConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	createParams := expandTestConfiguration(ctx, &model)

	created, err := r.client.TestPlanClient.CreateTestConfiguration(r.client.Ctx, testplan.CreateTestConfigurationArgs{
		TestConfigurationCreateUpdateParameters: createParams,
		Project:                                 &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating test configuration", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenTestConfiguration(ctx, &model, created)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TestConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model testConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	configID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid test configuration ID", err.Error())
		return
	}

	cfg, err := r.client.TestPlanClient.GetTestConfigurationById(r.client.Ctx, testplan.GetTestConfigurationByIdArgs{
		Project:             &projectID,
		TestConfigurationId: &configID,
	})
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading test configuration", err.Error())
		return
	}
	if cfg == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(flattenTestConfiguration(ctx, &model, cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TestConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan testConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state testConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ProjectID.ValueString()
	configID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid test configuration ID", err.Error())
		return
	}

	updateParams := expandTestConfiguration(ctx, &plan)

	updated, err := r.client.TestPlanClient.UpdateTestConfiguration(r.client.Ctx, testplan.UpdateTestConfigurationArgs{
		TestConfigurationCreateUpdateParameters: updateParams,
		Project:                                 &projectID,
		TestConfiguartionId:                     &configID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating test configuration", err.Error())
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(flattenTestConfiguration(ctx, &plan, updated)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TestConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model testConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	configID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid test configuration ID", err.Error())
		return
	}

	err = r.client.TestPlanClient.DeleteTestConfguration(r.client.Ctx, testplan.DeleteTestConfgurationArgs{
		Project:             &projectID,
		TestConfiguartionId: &configID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting test configuration", err.Error())
	}
}

// ── Expand / Flatten ──────────────────────────────────────────────────────────

// expandTestConfiguration converts state model → API create/update params.
func expandTestConfiguration(ctx context.Context, model *testConfigurationModel) *testplan.TestConfigurationCreateUpdateParameters {
	params := &testplan.TestConfigurationCreateUpdateParameters{
		Name: converter.String(model.Name.ValueString()),
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		params.Description = converter.String(model.Description.ValueString())
	}

	if !model.IsDefault.IsNull() && !model.IsDefault.IsUnknown() {
		v := model.IsDefault.ValueBool()
		params.IsDefault = &v
	}

	if !model.Values.IsNull() && !model.Values.IsUnknown() {
		rawMap := make(map[string]string)
		_ = model.Values.ElementsAs(ctx, &rawMap, false)
		pairs := make([]test.NameValuePair, 0, len(rawMap))
		for k, v := range rawMap {
			k, v := k, v
			pairs = append(pairs, test.NameValuePair{
				Name:  &k,
				Value: &v,
			})
		}
		params.Values = &pairs
	}

	return params
}

// ── Bool default helper ───────────────────────────────────────────────────────

// tcStaticBool provides a static bool schema default value.
func tcStaticBool(val bool) defaults.Bool {
	return tcStaticBoolDefault{value: val}
}

type tcStaticBoolDefault struct {
	value bool
}

func (d tcStaticBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("Defaults to %v.", d.value)
}

func (d tcStaticBoolDefault) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d tcStaticBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.value)
}

// flattenTestConfiguration maps API response → state model.
func flattenTestConfiguration(ctx context.Context, model *testConfigurationModel, cfg *testplan.TestConfiguration) diag.Diagnostics {
	var diags diag.Diagnostics

	if cfg.Id != nil {
		model.ID = types.StringValue(strconv.Itoa(*cfg.Id))
	}
	model.Name = types.StringValue(converter.ToString(cfg.Name, ""))
	model.Description = types.StringValue(converter.ToString(cfg.Description, ""))

	if cfg.IsDefault != nil {
		model.IsDefault = types.BoolValue(*cfg.IsDefault)
	} else {
		model.IsDefault = types.BoolValue(false)
	}

	// Convert []test.NameValuePair → map[string]string → types.Map
	rawMap := make(map[string]string)
	if cfg.Values != nil {
		for _, pair := range *cfg.Values {
			if pair.Name != nil && pair.Value != nil {
				rawMap[*pair.Name] = *pair.Value
			}
		}
	}
	valuesMap, d := types.MapValueFrom(ctx, types.StringType, rawMap)
	diags.Append(d...)
	model.Values = valuesMap

	return diags
}
