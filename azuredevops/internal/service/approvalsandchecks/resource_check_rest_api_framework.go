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
	_ resource.Resource                = (*RestAPIResource)(nil)
	_ resource.ResourceWithConfigure   = (*RestAPIResource)(nil)
	_ resource.ResourceWithImportState = (*RestAPIResource)(nil)
)

type RestAPIResource struct {
	client *client.AggregatedClient
}

func NewRestAPIResource() resource.Resource {
	return &RestAPIResource{}
}

func (r *RestAPIResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_check_rest_api"
}

func (r *RestAPIResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RestAPIResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"display_name":                    schema.StringAttribute{Required: true},
			"connected_service_name_selector": schema.StringAttribute{Required: true},
			"connected_service_name":          schema.StringAttribute{Required: true},
			"method":                          schema.StringAttribute{Required: true},
			"body": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{checkUseStateForUnknown()},
			},
			"headers": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{checkUseStateForUnknown()},
			},
			"retry_interval": schema.Int64Attribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{checkUseStateForUnknownInt64Val()},
			},
			"success_criteria": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{checkUseStateForUnknown()},
			},
			"url_suffix": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{checkUseStateForUnknown()},
			},
			"variable_group_name": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{checkUseStateForUnknown()},
			},
			"completion_event": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  staticCheckString(string(CompleteEventValues.Callback)),
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

type restAPIModel struct {
	ID                           types.String `tfsdk:"id"`
	ProjectID                    types.String `tfsdk:"project_id"`
	TargetResourceID             types.String `tfsdk:"target_resource_id"`
	TargetResourceType           types.String `tfsdk:"target_resource_type"`
	DisplayName                  types.String `tfsdk:"display_name"`
	ConnectedServiceNameSelector types.String `tfsdk:"connected_service_name_selector"`
	ConnectedServiceName         types.String `tfsdk:"connected_service_name"`
	Method                       types.String `tfsdk:"method"`
	Body                         types.String `tfsdk:"body"`
	Headers                      types.String `tfsdk:"headers"`
	RetryInterval                types.Int64  `tfsdk:"retry_interval"`
	SuccessCriteria              types.String `tfsdk:"success_criteria"`
	URLSuffix                    types.String `tfsdk:"url_suffix"`
	VariableGroupName            types.String `tfsdk:"variable_group_name"`
	CompletionEvent              types.String `tfsdk:"completion_event"`
	Timeout                      types.Int64  `tfsdk:"timeout"`
	Version                      types.Int64  `tfsdk:"version"`
}

func (r *RestAPIResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model restAPIModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	check, diags := expandRestAPIFW(&model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.PipelinesChecksClientExtras.AddCheckConfiguration(r.client.Ctx, pipelineschecksextras.AddCheckConfigurationArgs{
		Project: converter.String(model.ProjectID.ValueString()), Configuration: check,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating REST API check", err.Error())
		return
	}
	model.ID = types.StringValue(strconv.Itoa(*created.Id))
	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *RestAPIResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model restAPIModel
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

func (r *RestAPIResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan restAPIModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state restAPIModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	plan.Version = state.Version
	check, diags := expandRestAPIFW(&plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.PipelinesChecksClientExtras.UpdateCheckConfiguration(r.client.Ctx, pipelineschecksextras.UpdateCheckConfigurationArgs{
		Project: converter.String(plan.ProjectID.ValueString()), Configuration: check, Id: check.Id,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating REST API check", err.Error())
		return
	}
	resp.Diagnostics.Append(r.readIntoModel(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RestAPIResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model restAPIModel
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
		resp.Diagnostics.AddError("Error deleting REST API check", err.Error())
	}
}

func (r *RestAPIResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importCheckState(ctx, req, resp)
}

func (r *RestAPIResource) readIntoModel(_ context.Context, model *restAPIModel) diag.Diagnostics {
	var diags diag.Diagnostics
	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid check ID", err.Error())
		return diags
	}
	check, notFound, err := readCheckFromAPI(r.client, model.ProjectID.ValueString(), id)
	if err != nil {
		diags.AddError("Error reading REST API check", err.Error())
		return diags
	}
	if notFound {
		model.ID = types.StringNull()
		return diags
	}
	if err := flattenRestAPIFW(model, check); err != nil {
		diags.AddError("Error flattening REST API check", err.Error())
	}
	return diags
}

func expandRestAPIFW(model *restAPIModel) (*pipelineschecksextras.CheckConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics
	checkID := 0
	if !model.ID.IsNull() && model.ID.ValueString() != "" {
		checkID, _ = strconv.Atoi(model.ID.ValueString())
	}
	version := int(model.Version.ValueInt64())
	timeout := int(model.Timeout.ValueInt64())
	if timeout == 0 {
		timeout = 1440
	}
	serviceNameSelector := model.ConnectedServiceNameSelector.ValueString()
	input := map[string]interface{}{
		"connectedServiceNameSelector": serviceNameSelector,
		"method":                       model.Method.ValueString(),
	}
	if serviceNameSelector == "connectedServiceName" {
		input["connectedServiceName"] = model.ConnectedServiceName.ValueString()
	} else if serviceNameSelector == "connectedServiceNameARM" {
		input["connectedServiceNameARM"] = model.ConnectedServiceName.ValueString()
	}
	if v := model.Headers.ValueString(); v != "" {
		input["headers"] = v
	}
	if v := model.Body.ValueString(); v != "" {
		input["body"] = v
	}
	if v := model.URLSuffix.ValueString(); v != "" {
		input["urlSuffix"] = v
	}
	completionEvent := model.CompletionEvent.ValueString()
	input["waitForCompletion"] = "true"
	if strings.EqualFold(completionEvent, string(CompleteEventValues.ApiResponse)) {
		input["waitForCompletion"] = "false"
		if v := model.SuccessCriteria.ValueString(); v != "" {
			input["successCriteria"] = v
		}
	}
	settings := map[string]interface{}{
		"definitionRef": map[string]string{
			"id":      "9c3e8943-130d-4c78-ac63-8af81df62dfb",
			"name":    "InvokeRESTAPI",
			"version": "1.220.0",
		},
		"displayName": model.DisplayName.ValueString(),
		"inputs":      input,
	}
	if retryInterval := model.RetryInterval.ValueInt64(); retryInterval > 0 {
		if strings.EqualFold(completionEvent, string(CompleteEventValues.Callback)) {
			diags.AddError("Invalid retry_interval", "Does not need to set retry_interval when completion_event=Callback")
			return nil, diags
		}
		settings["retryInterval"] = int(retryInterval)
	}
	if v := model.VariableGroupName.ValueString(); v != "" {
		settings["linkedVariableGroup"] = v
	}
	check := &pipelineschecksextras.CheckConfiguration{
		Type:     approvalAndCheckType.TaskCheck,
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
	return check, diags
}

func flattenRestAPIFW(model *restAPIModel, check *pipelineschecksextras.CheckConfiguration) error {
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
	if check.Settings == nil {
		return nil
	}
	settingsMap, err := settingsAsMap(check.Settings)
	if err != nil {
		return fmt.Errorf("parsing settings: %w", err)
	}
	if v, ok := settingsMap["displayName"]; ok {
		model.DisplayName = types.StringValue(fmt.Sprintf("%v", v))
	}
	if v, ok := settingsMap["retryInterval"]; ok {
		if val, ok := v.(float64); ok {
			model.RetryInterval = types.Int64Value(int64(val))
		}
	}
	if v, ok := settingsMap["linkedVariableGroup"]; ok {
		model.VariableGroupName = types.StringValue(fmt.Sprintf("%v", v))
	}
	if inputsRaw, ok := settingsMap["inputs"]; ok {
		if inputs, ok := inputsRaw.(map[string]interface{}); ok {
			if v, ok := inputs["connectedServiceNameSelector"]; ok {
				selector := fmt.Sprintf("%v", v)
				model.ConnectedServiceNameSelector = types.StringValue(selector)
				if selector == "connectedServiceName" {
					if v2, ok := inputs["connectedServiceName"]; ok {
						model.ConnectedServiceName = types.StringValue(fmt.Sprintf("%v", v2))
					}
				} else if selector == "connectedServiceNameARM" {
					if v2, ok := inputs["connectedServiceNameARM"]; ok {
						model.ConnectedServiceName = types.StringValue(fmt.Sprintf("%v", v2))
					}
				}
			}
			if v, ok := inputs["method"]; ok {
				model.Method = types.StringValue(fmt.Sprintf("%v", v))
			}
			if v, ok := inputs["headers"]; ok {
				model.Headers = types.StringValue(fmt.Sprintf("%v", v))
			}
			if v, ok := inputs["body"]; ok {
				model.Body = types.StringValue(fmt.Sprintf("%v", v))
			}
			if v, ok := inputs["waitForCompletion"]; ok {
				waitForCompletion, _ := strconv.ParseBool(fmt.Sprintf("%v", v))
				if waitForCompletion {
					model.CompletionEvent = types.StringValue(string(CompleteEventValues.Callback))
				} else {
					model.CompletionEvent = types.StringValue(string(CompleteEventValues.ApiResponse))
				}
			}
			if v, ok := inputs["successCriteria"]; ok {
				model.SuccessCriteria = types.StringValue(fmt.Sprintf("%v", v))
			}
			if v, ok := inputs["urlSuffix"]; ok {
				model.URLSuffix = types.StringValue(fmt.Sprintf("%v", v))
			}
		}
	}
	return nil
}
