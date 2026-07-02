package approvalsandchecks

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
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/utils/sdk/pipelineschecksextras"
)

var (
	_ resource.Resource                = (*ExclusiveLockResource)(nil)
	_ resource.ResourceWithConfigure   = (*ExclusiveLockResource)(nil)
	_ resource.ResourceWithImportState = (*ExclusiveLockResource)(nil)
)

type ExclusiveLockResource struct {
	client *client.AggregatedClient
}

func NewExclusiveLockResource() resource.Resource {
	return &ExclusiveLockResource{}
}

func (r *ExclusiveLockResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_check_exclusive_lock"
}

func (r *ExclusiveLockResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	agg, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Provider Data", fmt.Sprintf("Expected *client.AggregatedClient, got: %T", req.ProviderData))
		return
	}
	r.client = agg
}

func (r *ExclusiveLockResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{checkUseStateForUnknown()},
			},
			"project_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{checkRequiresReplace()},
			},
			"target_resource_id": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{checkRequiresReplace()},
			},
			"target_resource_type": schema.StringAttribute{
				Required:      true,
				PlanModifiers: []planmodifier.String{checkRequiresReplace()},
			},
			"timeout": schema.Int64Attribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{checkUseStateForUnknownInt64Val()},
			},
			"version": schema.Int64Attribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{checkUseStateForUnknownInt64Val()},
			},
		},
	}
}

type exclusiveLockModel struct {
	ID                 types.String `tfsdk:"id"`
	ProjectID          types.String `tfsdk:"project_id"`
	TargetResourceID   types.String `tfsdk:"target_resource_id"`
	TargetResourceType types.String `tfsdk:"target_resource_type"`
	Timeout            types.Int64  `tfsdk:"timeout"`
	Version            types.Int64  `tfsdk:"version"`
}

func (r *ExclusiveLockResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model exclusiveLockModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	check := expandExclusiveLockFW(&model)
	created, err := r.client.PipelinesChecksClientExtras.AddCheckConfiguration(r.client.Ctx, pipelineschecksextras.AddCheckConfigurationArgs{
		Project: converter.String(model.ProjectID.ValueString()), Configuration: check,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating exclusive lock check", err.Error())
		return
	}
	model.ID = types.StringValue(strconv.Itoa(*created.Id))
	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ExclusiveLockResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model exclusiveLockModel
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

func (r *ExclusiveLockResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan exclusiveLockModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state exclusiveLockModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	plan.Version = state.Version
	check := expandExclusiveLockFW(&plan)
	_, err := r.client.PipelinesChecksClientExtras.UpdateCheckConfiguration(r.client.Ctx, pipelineschecksextras.UpdateCheckConfigurationArgs{
		Project: converter.String(plan.ProjectID.ValueString()), Configuration: check, Id: check.Id,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating exclusive lock check", err.Error())
		return
	}
	resp.Diagnostics.Append(r.readIntoModel(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ExclusiveLockResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model exclusiveLockModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid check ID", err.Error())
		return
	}
	projectID := model.ProjectID.ValueString()
	if err := r.client.PipelinesChecksClientExtras.DeleteCheckConfiguration(r.client.Ctx, pipelineschecksextras.DeleteCheckConfigurationArgs{
		Project: &projectID, Id: &id,
	}); err != nil {
		resp.Diagnostics.AddError("Error deleting exclusive lock check", err.Error())
	}
}

func (r *ExclusiveLockResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importCheckState(ctx, req, resp)
}

func (r *ExclusiveLockResource) readIntoModel(_ context.Context, model *exclusiveLockModel) diag.Diagnostics {
	var diags diag.Diagnostics
	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid check ID", err.Error())
		return diags
	}
	projectID := model.ProjectID.ValueString()
	expand := pipelineschecksextras.CheckConfigurationExpandParameterValues.Settings
	check, err := r.client.PipelinesChecksClientExtras.GetCheckConfiguration(r.client.Ctx, pipelineschecksextras.GetCheckConfigurationArgs{
		Project: &projectID, Id: &id, Expand: &expand,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			model.ID = types.StringNull()
			return diags
		}
		diags.AddError("Error reading exclusive lock check", err.Error())
		return diags
	}
	flattenExclusiveLockFW(model, check)
	return diags
}

func expandExclusiveLockFW(model *exclusiveLockModel) *pipelineschecksextras.CheckConfiguration {
	checkID := 0
	if !model.ID.IsNull() && model.ID.ValueString() != "" {
		checkID, _ = strconv.Atoi(model.ID.ValueString())
	}
	version := int(model.Version.ValueInt64())
	timeout := int(model.Timeout.ValueInt64())
	if timeout == 0 {
		timeout = 43200
	}
	settings := map[string]interface{}{}
	check := &pipelineschecksextras.CheckConfiguration{
		Type:     approvalAndCheckType.ExclusiveLock,
		Settings: settings,
		Resource: &pipelineschecksextras.Resource{
			Id: converter.String(model.TargetResourceID.ValueString()), Type: converter.String(model.TargetResourceType.ValueString()),
		},
		Version: &version,
		Timeout: &timeout,
	}
	if checkID != 0 {
		check.Id = &checkID
	}
	return check
}

func flattenExclusiveLockFW(model *exclusiveLockModel, check *pipelineschecksextras.CheckConfiguration) {
	model.ID = types.StringValue(strconv.Itoa(*check.Id))
	if check.Resource != nil {
		if check.Resource.Id != nil {
			model.TargetResourceID = types.StringValue(*check.Resource.Id)
		}
		if check.Resource.Type != nil {
			model.TargetResourceType = types.StringValue(*check.Resource.Type)
		}
	}
	if check.Timeout != nil {
		model.Timeout = types.Int64Value(int64(*check.Timeout))
	} else {
		model.Timeout = types.Int64Value(43200)
	}
	if check.Version != nil {
		model.Version = types.Int64Value(int64(*check.Version))
	}
}

func readCheckFromAPI(c *client.AggregatedClient, projectID string, checkID int) (*pipelineschecksextras.CheckConfiguration, bool, error) {
	expand := pipelineschecksextras.CheckConfigurationExpandParameterValues.Settings
	check, err := c.PipelinesChecksClientExtras.GetCheckConfiguration(c.Ctx, pipelineschecksextras.GetCheckConfigurationArgs{
		Project: &projectID, Id: &checkID, Expand: &expand,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			return nil, true, nil
		}
		return nil, false, err
	}
	return check, false, nil
}

func settingsAsMap(settings interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}
