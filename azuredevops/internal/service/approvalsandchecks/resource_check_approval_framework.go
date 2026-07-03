package approvalsandchecks

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/utils/sdk/pipelineschecksextras"
)

var (
	_ resource.Resource                = (*ApprovalResource)(nil)
	_ resource.ResourceWithConfigure   = (*ApprovalResource)(nil)
	_ resource.ResourceWithImportState = (*ApprovalResource)(nil)
)

type ApprovalResource struct {
	client *client.AggregatedClient
}

func NewApprovalResource() resource.Resource {
	return &ApprovalResource{}
}

func (r *ApprovalResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_check_approval"
}

func (r *ApprovalResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ApprovalResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"approvers": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.RegexMatches(
							regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
							"each approver must be a valid UUID",
						),
					),
				},
			},
			"instructions": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{checkUseStateForUnknown()},
			},
			"requester_can_approve": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  staticCheckBool(false),
			},
			"minimum_required_approvers": schema.Int64Attribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.Int64{checkUseStateForUnknownInt64Val()},
			},
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

type approvalModel struct {
	ID                       types.String `tfsdk:"id"`
	ProjectID                types.String `tfsdk:"project_id"`
	TargetResourceID         types.String `tfsdk:"target_resource_id"`
	TargetResourceType       types.String `tfsdk:"target_resource_type"`
	Approvers                types.List   `tfsdk:"approvers"`
	Instructions             types.String `tfsdk:"instructions"`
	RequesterCanApprove      types.Bool   `tfsdk:"requester_can_approve"`
	MinimumRequiredApprovers types.Int64  `tfsdk:"minimum_required_approvers"`
	Timeout                  types.Int64  `tfsdk:"timeout"`
	Version                  types.Int64  `tfsdk:"version"`
}

func (r *ApprovalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model approvalModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	check, diags := expandApprovalFW(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.client.PipelinesChecksClientExtras.AddCheckConfiguration(r.client.Ctx, pipelineschecksextras.AddCheckConfigurationArgs{
		Project: converter.String(model.ProjectID.ValueString()), Configuration: check,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating approval check", err.Error())
		return
	}
	model.ID = types.StringValue(strconv.Itoa(*created.Id))
	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ApprovalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model approvalModel
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

func (r *ApprovalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan approvalModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state approvalModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	plan.Version = state.Version
	check, diags := expandApprovalFW(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.PipelinesChecksClientExtras.UpdateCheckConfiguration(r.client.Ctx, pipelineschecksextras.UpdateCheckConfigurationArgs{
		Project: converter.String(plan.ProjectID.ValueString()), Configuration: check, Id: check.Id,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating approval check", err.Error())
		return
	}
	resp.Diagnostics.Append(r.readIntoModel(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ApprovalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model approvalModel
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
		if !utils.ResponseWasNotFound(err) {
			resp.Diagnostics.AddError("Error deleting approval check", err.Error())
		}
	}
}

func (r *ApprovalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importCheckState(ctx, req, resp)
}

func (r *ApprovalResource) readIntoModel(ctx context.Context, model *approvalModel) diag.Diagnostics {
	var diags diag.Diagnostics
	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid check ID", err.Error())
		return diags
	}
	check, notFound, err := readCheckFromAPI(r.client, model.ProjectID.ValueString(), id)
	if err != nil {
		diags.AddError("Error reading approval check", err.Error())
		return diags
	}
	if notFound {
		model.ID = types.StringNull()
		return diags
	}
	diags.Append(flattenApprovalFW(ctx, model, check)...)
	return diags
}

func expandApprovalFW(ctx context.Context, model *approvalModel) (*pipelineschecksextras.CheckConfiguration, diag.Diagnostics) {
	var diags diag.Diagnostics
	checkID := 0
	if !model.ID.IsNull() && model.ID.ValueString() != "" {
		var atoiErr error
		checkID, atoiErr = strconv.Atoi(model.ID.ValueString())
		if atoiErr != nil {
			diags.AddError("Invalid check ID", atoiErr.Error())
			return nil, diags
		}
	}
	version := int(model.Version.ValueInt64())
	timeout := int(model.Timeout.ValueInt64())
	if timeout == 0 {
		timeout = 43200
	}
	var approverIDs []string
	diags.Append(model.Approvers.ElementsAs(ctx, &approverIDs, false)...)
	if diags.HasError() {
		return nil, diags
	}
	approvers := make([]interface{}, 0, len(approverIDs))
	for _, id := range approverIDs {
		approvers = append(approvers, map[string]interface{}{"id": id})
	}
	settings := map[string]interface{}{
		"instructions":              model.Instructions.ValueString(),
		"minRequiredApprovers":      int(model.MinimumRequiredApprovers.ValueInt64()),
		"requesterCannotBeApprover": !model.RequesterCanApprove.ValueBool(),
		"approvers":                 approvers,
	}
	check := &pipelineschecksextras.CheckConfiguration{
		Type:     approvalAndCheckType.Approval,
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

func flattenApprovalFW(ctx context.Context, model *approvalModel, check *pipelineschecksextras.CheckConfiguration) diag.Diagnostics {
	var diags diag.Diagnostics
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
		diags.AddError("Error parsing approval check settings", err.Error())
		return diags
	}
	if v, ok := settingsMap["instructions"]; ok {
		model.Instructions = types.StringValue(fmt.Sprintf("%v", v))
	}
	if v, ok := settingsMap["minRequiredApprovers"]; ok {
		if val, ok := v.(float64); ok {
			model.MinimumRequiredApprovers = types.Int64Value(int64(val))
		}
	}
	if v, ok := settingsMap["requesterCannotBeApprover"]; ok {
		if b, ok := v.(bool); ok {
			model.RequesterCanApprove = types.BoolValue(!b)
		}
	}
	if approversRaw, ok := settingsMap["approvers"]; ok {
		if approverList, ok := approversRaw.([]interface{}); ok {
			ids := make([]string, 0, len(approverList))
			for _, a := range approverList {
				if aMap, ok := a.(map[string]interface{}); ok {
					if id, ok := aMap["id"].(string); ok {
						ids = append(ids, id)
					}
				}
			}
			approverVals, d := types.ListValueFrom(ctx, types.StringType, ids)
			diags.Append(d...)
			model.Approvers = approverVals
		}
	}
	return diags
}
