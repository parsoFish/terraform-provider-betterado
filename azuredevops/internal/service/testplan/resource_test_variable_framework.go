package testplan

import (
	"context"
	"fmt"
	"strconv"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/testplan"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ resource.Resource              = (*TestVariableResource)(nil)
	_ resource.ResourceWithConfigure = (*TestVariableResource)(nil)
)

// TestVariableResource is the terraform-plugin-framework implementation of
// betterado_test_variable.
type TestVariableResource struct {
	client *client.AggregatedClient
}

// NewTestVariableResource returns a new resource.Resource for betterado_test_variable.
func NewTestVariableResource() resource.Resource {
	return &TestVariableResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *TestVariableResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_test_variable"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *TestVariableResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"allowed_values": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *TestVariableResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// testVariableModel is the Terraform state model for betterado_test_variable.
type testVariableModel struct {
	ID            types.String `tfsdk:"id"`
	ProjectID     types.String `tfsdk:"project_id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	AllowedValues types.List   `tfsdk:"allowed_values"`
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *TestVariableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model testVariableModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	createParams := expandTestVariable(ctx, &model)

	created, err := r.client.TestPlanClient.CreateTestVariable(r.client.Ctx, testplan.CreateTestVariableArgs{
		TestVariableCreateUpdateParameters: createParams,
		Project:                            &projectID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating test variable", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenTestVariable(ctx, &model, created)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TestVariableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model testVariableModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	varID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid test variable ID", err.Error())
		return
	}

	v, err := r.client.TestPlanClient.GetTestVariableById(r.client.Ctx, testplan.GetTestVariableByIdArgs{
		Project:        &projectID,
		TestVariableId: &varID,
	})
	if err != nil {
		if isNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading test variable", err.Error())
		return
	}
	if v == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(flattenTestVariable(ctx, &model, v)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *TestVariableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan testVariableModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state testVariableModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := plan.ProjectID.ValueString()
	varID, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid test variable ID", err.Error())
		return
	}

	updateParams := expandTestVariable(ctx, &plan)

	updated, err := r.client.TestPlanClient.UpdateTestVariable(r.client.Ctx, testplan.UpdateTestVariableArgs{
		TestVariableCreateUpdateParameters: updateParams,
		Project:                            &projectID,
		TestVariableId:                     &varID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating test variable", err.Error())
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(flattenTestVariable(ctx, &plan, updated)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *TestVariableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model testVariableModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := model.ProjectID.ValueString()
	varID, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid test variable ID", err.Error())
		return
	}

	err = r.client.TestPlanClient.DeleteTestVariable(r.client.Ctx, testplan.DeleteTestVariableArgs{
		Project:        &projectID,
		TestVariableId: &varID,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting test variable", err.Error())
	}
}

// ── Expand / Flatten ──────────────────────────────────────────────────────────

// expandTestVariable converts state model → API create/update params.
func expandTestVariable(ctx context.Context, model *testVariableModel) *testplan.TestVariableCreateUpdateParameters {
	params := &testplan.TestVariableCreateUpdateParameters{
		Name: converter.String(model.Name.ValueString()),
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		params.Description = converter.String(model.Description.ValueString())
	}

	if !model.AllowedValues.IsNull() && !model.AllowedValues.IsUnknown() {
		var values []string
		_ = model.AllowedValues.ElementsAs(ctx, &values, false)
		params.Values = &values
	}

	return params
}

// flattenTestVariable maps API response → state model.
func flattenTestVariable(ctx context.Context, model *testVariableModel, v *testplan.TestVariable) diag.Diagnostics {
	var diags diag.Diagnostics

	if v.Id != nil {
		model.ID = types.StringValue(strconv.Itoa(*v.Id))
	}
	model.Name = types.StringValue(converter.ToString(v.Name, ""))
	model.Description = types.StringValue(converter.ToString(v.Description, ""))

	// Convert *[]string → types.List
	var rawValues []string
	if v.Values != nil {
		rawValues = *v.Values
	}
	listVal, d := types.ListValueFrom(ctx, types.StringType, rawValues)
	diags.Append(d...)
	model.AllowedValues = listVal

	return diags
}
