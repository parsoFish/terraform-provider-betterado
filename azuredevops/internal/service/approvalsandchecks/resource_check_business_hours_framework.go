package approvalsandchecks

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

// validTimezoneIds mirrors the SDKv2 validation.StringInSlice list for SDKv2 parity.
var validTimezoneIds = []string{
	"AUS Central Standard Time", "AUS Eastern Standard Time", "Afghanistan Standard Time",
	"Alaskan Standard Time", "Aleutian Standard Time", "Altai Standard Time",
	"Arab Standard Time", "Arabian Standard Time", "Arabic Standard Time",
	"Argentina Standard Time", "Astrakhan Standard Time", "Atlantic Standard Time",
	"Aus Central W. Standard Time", "Azerbaijan Standard Time", "Azores Standard Time",
	"Bahia Standard Time", "Bangladesh Standard Time", "Belarus Standard Time",
	"Bougainville Standard Time", "Canada Central Standard Time", "Cape Verde Standard Time",
	"Caucasus Standard Time", "Cen. Australia Standard Time", "Central America Standard Time",
	"Central Asia Standard Time", "Central Brazilian Standard Time", "Central Europe Standard Time",
	"Central European Standard Time", "Central Pacific Standard Time",
	"Central Standard Time (Mexico)", "Central Standard Time", "Chatham Islands Standard Time",
	"China Standard Time", "Cuba Standard Time", "Dateline Standard Time",
	"E. Africa Standard Time", "E. Australia Standard Time", "E. Europe Standard Time",
	"E. South America Standard Time", "Easter Island Standard Time",
	"Eastern Standard Time (Mexico)", "Eastern Standard Time", "Egypt Standard Time",
	"Ekaterinburg Standard Time", "FLE Standard Time", "Fiji Standard Time",
	"GMT Standard Time", "GTB Standard Time", "Georgian Standard Time",
	"Greenland Standard Time", "Greenwich Standard Time", "Haiti Standard Time",
	"Hawaiian Standard Time", "India Standard Time", "Iran Standard Time",
	"Israel Standard Time", "Jordan Standard Time", "Kaliningrad Standard Time",
	"Kamchatka Standard Time", "Korea Standard Time", "Libya Standard Time",
	"Line Islands Standard Time", "Lord Howe Standard Time", "Magadan Standard Time",
	"Magallanes Standard Time", "Marquesas Standard Time", "Mauritius Standard Time",
	"Mid-Atlantic Standard Time", "Middle East Standard Time", "Montevideo Standard Time",
	"Morocco Standard Time", "Mountain Standard Time (Mexico)", "Mountain Standard Time",
	"Myanmar Standard Time", "N. Central Asia Standard Time", "Namibia Standard Time",
	"Nepal Standard Time", "New Zealand Standard Time", "Newfoundland Standard Time",
	"Norfolk Standard Time", "North Asia East Standard Time", "North Asia Standard Time",
	"North Korea Standard Time", "Omsk Standard Time", "Pacific SA Standard Time",
	"Pacific Standard Time (Mexico)", "Pacific Standard Time", "Pakistan Standard Time",
	"Paraguay Standard Time", "Qyzylorda Standard Time", "Romance Standard Time",
	"Russia Time Zone 10", "Russia Time Zone 11", "Russia Time Zone 3",
	"Russian Standard Time", "SA Eastern Standard Time", "SA Pacific Standard Time",
	"SA Western Standard Time", "SE Asia Standard Time", "Saint Pierre Standard Time",
	"Sakhalin Standard Time", "Samoa Standard Time", "Sao Tome Standard Time",
	"Saratov Standard Time", "Singapore Standard Time", "South Africa Standard Time",
	"South Sudan Standard Time", "Sri Lanka Standard Time", "Sudan Standard Time",
	"Syria Standard Time", "Taipei Standard Time", "Tasmania Standard Time",
	"Tocantins Standard Time", "Tokyo Standard Time", "Tomsk Standard Time",
	"Tonga Standard Time", "Transbaikal Standard Time", "Turkey Standard Time",
	"Turks And Caicos Standard Time", "US Eastern Standard Time", "US Mountain Standard Time",
	"UTC", "UTC+12", "UTC+13", "UTC-02", "UTC-08", "UTC-09", "UTC-11",
	"Ulaanbaatar Standard Time", "Venezuela Standard Time", "Vladivostok Standard Time",
	"Volgograd Standard Time", "W. Australia Standard Time", "W. Central Africa Standard Time",
	"W. Europe Standard Time", "W. Mongolia Standard Time", "West Asia Standard Time",
	"West Bank Standard Time", "West Pacific Standard Time", "Yakutsk Standard Time",
	"Yukon Standard Time",
}

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

// businessHoursTimeRegexp validates 24-hour HH:MM with range checks (SDKv2 parity: ^(0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$).
var businessHoursTimeRegexp = regexp.MustCompile(`^(0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$`)

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
			"time_zone": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(validTimezoneIds...),
				},
			},
			"start_time": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						businessHoursTimeRegexp,
						"must be a 24-hour time with leading zeros and valid range, e.g. 09:00 (hours 00-23, minutes 00-59)",
					),
				},
			},
			"end_time": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						businessHoursTimeRegexp,
						"must be a 24-hour time with leading zeros and valid range, e.g. 18:00 (hours 00-23, minutes 00-59)",
					),
				},
			},
			"monday":    schema.BoolAttribute{Optional: true, Computed: true, Default: staticCheckBool(false)},
			"tuesday":   schema.BoolAttribute{Optional: true, Computed: true, Default: staticCheckBool(false)},
			"wednesday": schema.BoolAttribute{Optional: true, Computed: true, Default: staticCheckBool(false)},
			"thursday":  schema.BoolAttribute{Optional: true, Computed: true, Default: staticCheckBool(false)},
			"friday":    schema.BoolAttribute{Optional: true, Computed: true, Default: staticCheckBool(false)},
			"saturday":  schema.BoolAttribute{Optional: true, Computed: true, Default: staticCheckBool(false)},
			"sunday":    schema.BoolAttribute{Optional: true, Computed: true, Default: staticCheckBool(false)},
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
