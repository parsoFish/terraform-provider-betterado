package approvalsandchecks

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/utils/sdk/pipelineschecksextras"
)

var (
	_ resource.Resource                = (*BusinessHoursResource)(nil)
	_ resource.ResourceWithConfigure   = (*BusinessHoursResource)(nil)
	_ resource.ResourceWithImportState = (*BusinessHoursResource)(nil)
)

type BusinessHoursResource struct {
	client *client.AggregatedClient
}

func NewBusinessHoursResource() resource.Resource {
	return &BusinessHoursResource{}
}

func (r *BusinessHoursResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_check_business_hours"
}

func (r *BusinessHoursResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BusinessHoursResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"display_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  staticCheckString("Managed by Terraform"),
			},
			"time_zone":  schema.StringAttribute{Required: true},
			"start_time": schema.StringAttribute{Required: true},
			"end_time":   schema.StringAttribute{Required: true},
			"monday":     schema.BoolAttribute{Optional: true, Computed: true, Default: staticCheckBool(false)},
			"tuesday":    schema.BoolAttribute{Optional: true, Computed: true, Default: staticCheckBool(false)},
			"wednesday":  schema.BoolAttribute{Optional: true, Computed: true, Default: staticCheckBool(false)},
			"thursday":   schema.BoolAttribute{Optional: true, Computed: true, Default: staticCheckBool(false)},
			"friday":     schema.BoolAttribute{Optional: true, Computed: true, Default: staticCheckBool(false)},
			"saturday":   schema.BoolAttribute{Optional: true, Computed: true, Default: staticCheckBool(false)},
			"sunday":     schema.BoolAttribute{Optional: true, Computed: true, Default: staticCheckBool(false)},
			"timeout": schema.Int64Attribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{checkUseStateForUnknownInt64Val()},
			},
			"version": schema.Int64Attribute{
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{checkVersionPlanModifierFn()},
			},
		},
	}
}

type businessHoursModel struct {
	ID                 types.String `tfsdk:"id"`
	ProjectID          types.String `tfsdk:"project_id"`
	TargetResourceID   types.String `tfsdk:"target_resource_id"`
	TargetResourceType types.String `tfsdk:"target_resource_type"`
	DisplayName        types.String `tfsdk:"display_name"`
	TimeZone           types.String `tfsdk:"time_zone"`
	StartTime          types.String `tfsdk:"start_time"`
	EndTime            types.String `tfsdk:"end_time"`
	Monday             types.Bool   `tfsdk:"monday"`
	Tuesday            types.Bool   `tfsdk:"tuesday"`
	Wednesday          types.Bool   `tfsdk:"wednesday"`
	Thursday           types.Bool   `tfsdk:"thursday"`
	Friday             types.Bool   `tfsdk:"friday"`
	Saturday           types.Bool   `tfsdk:"saturday"`
	Sunday             types.Bool   `tfsdk:"sunday"`
	Timeout            types.Int64  `tfsdk:"timeout"`
	Version            types.Int64  `tfsdk:"version"`
}

func (r *BusinessHoursResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model businessHoursModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	check := expandBusinessHoursFW(&model)
	created, err := r.client.PipelinesChecksClientExtras.AddCheckConfiguration(r.client.Ctx, pipelineschecksextras.AddCheckConfigurationArgs{
		Project: converter.String(model.ProjectID.ValueString()), Configuration: check,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating business hours check", err.Error())
		return
	}
	model.ID = types.StringValue(strconv.Itoa(*created.Id))
	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *BusinessHoursResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model businessHoursModel
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

func (r *BusinessHoursResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan businessHoursModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state businessHoursModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	plan.Version = state.Version
	check := expandBusinessHoursFW(&plan)
	_, err := r.client.PipelinesChecksClientExtras.UpdateCheckConfiguration(r.client.Ctx, pipelineschecksextras.UpdateCheckConfigurationArgs{
		Project: converter.String(plan.ProjectID.ValueString()), Configuration: check, Id: check.Id,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating business hours check", err.Error())
		return
	}
	resp.Diagnostics.Append(r.readIntoModel(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BusinessHoursResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model businessHoursModel
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
		resp.Diagnostics.AddError("Error deleting business hours check", err.Error())
	}
}

func (r *BusinessHoursResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importCheckState(ctx, req, resp)
}

func (r *BusinessHoursResource) readIntoModel(_ context.Context, model *businessHoursModel) diag.Diagnostics {
	var diags diag.Diagnostics
	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid check ID", err.Error())
		return diags
	}
	check, notFound, err := readCheckFromAPI(r.client, model.ProjectID.ValueString(), id)
	if err != nil {
		diags.AddError("Error reading business hours check", err.Error())
		return diags
	}
	if notFound {
		model.ID = types.StringNull()
		return diags
	}
	if err := flattenBusinessHoursFW(model, check); err != nil {
		diags.AddError("Error flattening business hours check", err.Error())
	}
	return diags
}

func expandBusinessHoursFW(model *businessHoursModel) *pipelineschecksextras.CheckConfiguration {
	checkID := 0
	if !model.ID.IsNull() && model.ID.ValueString() != "" {
		checkID, _ = strconv.Atoi(model.ID.ValueString()) //nolint:errcheck // ID was written as an integer by the provider; cannot fail
	}
	version := int(model.Version.ValueInt64())
	timeout := int(model.Timeout.ValueInt64())
	if timeout == 0 {
		timeout = 1440
	}
	dayMap := []struct {
		field types.Bool
		name  string
	}{
		{model.Monday, "Monday"},
		{model.Tuesday, "Tuesday"},
		{model.Wednesday, "Wednesday"},
		{model.Thursday, "Thursday"},
		{model.Friday, "Friday"},
		{model.Saturday, "Saturday"},
		{model.Sunday, "Sunday"},
	}
	var days []string
	for _, d := range dayMap {
		if d.field.ValueBool() {
			days = append(days, d.name)
		}
	}
	inputs := map[string]interface{}{
		"businessDays": strings.Join(days, ", "),
		"startTime":    model.StartTime.ValueString(),
		"endTime":      model.EndTime.ValueString(),
		"timeZone":     model.TimeZone.ValueString(),
	}
	settings := map[string]interface{}{
		"inputs":        inputs,
		"definitionRef": evaluateBusinessHoursDef,
		"displayName":   model.DisplayName.ValueString(),
	}
	check := &pipelineschecksextras.CheckConfiguration{
		Type:     approvalAndCheckType.BusinessHours,
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

func flattenBusinessHoursFW(model *businessHoursModel, check *pipelineschecksextras.CheckConfiguration) error {
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
	}
	if check.Version != nil {
		model.Version = types.Int64Value(int64(*check.Version))
	}
	settingsMap, err := settingsAsMap(check.Settings)
	if err != nil {
		return fmt.Errorf("parsing settings: %w", err)
	}
	if v, ok := settingsMap["displayName"]; ok {
		model.DisplayName = types.StringValue(fmt.Sprintf("%v", v))
	}
	if inputs, ok := settingsMap["inputs"].(map[string]interface{}); ok {
		if v, ok := inputs["timeZone"]; ok {
			model.TimeZone = types.StringValue(fmt.Sprintf("%v", v))
		}
		if v, ok := inputs["startTime"]; ok {
			model.StartTime = types.StringValue(fmt.Sprintf("%v", v))
		}
		if v, ok := inputs["endTime"]; ok {
			model.EndTime = types.StringValue(fmt.Sprintf("%v", v))
		}
		if businessDaysRaw, ok := inputs["businessDays"]; ok {
			businessDays := fmt.Sprintf("%v", businessDaysRaw)
			dayMap := []struct {
				field *types.Bool
				name  string
			}{
				{&model.Monday, "Monday"},
				{&model.Tuesday, "Tuesday"},
				{&model.Wednesday, "Wednesday"},
				{&model.Thursday, "Thursday"},
				{&model.Friday, "Friday"},
				{&model.Saturday, "Saturday"},
				{&model.Sunday, "Sunday"},
			}
			for _, d := range dayMap {
				*d.field = types.BoolValue(strings.Contains(businessDays, d.name))
			}
		}
	}
	return nil
}
