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
	_ resource.Resource                = (*BranchControlResource)(nil)
	_ resource.ResourceWithConfigure   = (*BranchControlResource)(nil)
	_ resource.ResourceWithImportState = (*BranchControlResource)(nil)
)

type BranchControlResource struct {
	client *client.AggregatedClient
}

func NewBranchControlResource() resource.Resource {
	return &BranchControlResource{}
}

func (r *BranchControlResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_check_branch_control"
}

func (r *BranchControlResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BranchControlResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"allowed_branches": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  staticCheckString("*"),
			},
			"verify_branch_protection": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  staticCheckBool(false),
			},
			"ignore_unknown_protection_status": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  staticCheckBool(false),
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

type branchControlModel struct {
	ID                            types.String `tfsdk:"id"`
	ProjectID                     types.String `tfsdk:"project_id"`
	TargetResourceID              types.String `tfsdk:"target_resource_id"`
	TargetResourceType            types.String `tfsdk:"target_resource_type"`
	DisplayName                   types.String `tfsdk:"display_name"`
	AllowedBranches               types.String `tfsdk:"allowed_branches"`
	VerifyBranchProtection        types.Bool   `tfsdk:"verify_branch_protection"`
	IgnoreUnknownProtectionStatus types.Bool   `tfsdk:"ignore_unknown_protection_status"`
	Timeout                       types.Int64  `tfsdk:"timeout"`
	Version                       types.Int64  `tfsdk:"version"`
}

func (r *BranchControlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model branchControlModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	check := expandBranchControlFW(&model)
	created, err := r.client.PipelinesChecksClientExtras.AddCheckConfiguration(r.client.Ctx, pipelineschecksextras.AddCheckConfigurationArgs{
		Project: converter.String(model.ProjectID.ValueString()), Configuration: check,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating branch control check", err.Error())
		return
	}
	model.ID = types.StringValue(strconv.Itoa(*created.Id))
	resp.Diagnostics.Append(r.readIntoModel(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *BranchControlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model branchControlModel
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

func (r *BranchControlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan branchControlModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state branchControlModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = state.ID
	plan.Version = state.Version
	check := expandBranchControlFW(&plan)
	_, err := r.client.PipelinesChecksClientExtras.UpdateCheckConfiguration(r.client.Ctx, pipelineschecksextras.UpdateCheckConfigurationArgs{
		Project: converter.String(plan.ProjectID.ValueString()), Configuration: check, Id: check.Id,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating branch control check", err.Error())
		return
	}
	resp.Diagnostics.Append(r.readIntoModel(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BranchControlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model branchControlModel
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
		resp.Diagnostics.AddError("Error deleting branch control check", err.Error())
	}
}

func (r *BranchControlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importCheckState(ctx, req, resp)
}

func (r *BranchControlResource) readIntoModel(_ context.Context, model *branchControlModel) diag.Diagnostics {
	var diags diag.Diagnostics
	id, err := strconv.Atoi(model.ID.ValueString())
	if err != nil {
		diags.AddError("Invalid check ID", err.Error())
		return diags
	}
	check, notFound, err := readCheckFromAPI(r.client, model.ProjectID.ValueString(), id)
	if err != nil {
		diags.AddError("Error reading branch control check", err.Error())
		return diags
	}
	if notFound {
		model.ID = types.StringNull()
		return diags
	}
	if err := flattenBranchControlFW(model, check); err != nil {
		diags.AddError("Error flattening branch control check", err.Error())
	}
	return diags
}

func expandBranchControlFW(model *branchControlModel) *pipelineschecksextras.CheckConfiguration {
	checkID := 0
	if !model.ID.IsNull() && model.ID.ValueString() != "" {
		checkID, _ = strconv.Atoi(model.ID.ValueString()) //nolint:errcheck // ID was written as an integer by the provider; cannot fail
	}
	version := int(model.Version.ValueInt64())
	timeout := int(model.Timeout.ValueInt64())
	if timeout == 0 {
		timeout = 1440
	}
	verifyBranchProtection := model.VerifyBranchProtection.ValueBool()
	inputs := map[string]interface{}{
		"allowedBranches":          model.AllowedBranches.ValueString(),
		"ensureProtectionOfBranch": strconv.FormatBool(verifyBranchProtection),
	}
	if verifyBranchProtection {
		inputs["allowUnknownStatusBranch"] = strconv.FormatBool(model.IgnoreUnknownProtectionStatus.ValueBool())
	}
	settings := map[string]interface{}{
		"inputs":        inputs,
		"definitionRef": evaluateBranchProtectionDef,
		"displayName":   model.DisplayName.ValueString(),
	}
	check := &pipelineschecksextras.CheckConfiguration{
		Type:     approvalAndCheckType.BranchProtection,
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

func flattenBranchControlFW(model *branchControlModel, check *pipelineschecksextras.CheckConfiguration) error {
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
		if v, ok := inputs["allowedBranches"]; ok {
			model.AllowedBranches = types.StringValue(fmt.Sprintf("%v", v))
		}
		if v, ok := inputs["ensureProtectionOfBranch"]; ok {
			b, _ := strconv.ParseBool(fmt.Sprintf("%v", v)) //nolint:errcheck // API returns bool-like string; default false on parse failure is acceptable
			model.VerifyBranchProtection = types.BoolValue(b)
		}
		if v, ok := inputs["allowUnknownStatusBranch"]; ok {
			b, _ := strconv.ParseBool(fmt.Sprintf("%v", v)) //nolint:errcheck // API returns bool-like string; default false on parse failure is acceptable
			model.IgnoreUnknownProtectionStatus = types.BoolValue(b)
		} else {
			model.IgnoreUnknownProtectionStatus = types.BoolValue(false)
		}
	}
	if defRefRaw, ok := settingsMap["definitionRef"]; ok {
		if defRef, ok := defRefRaw.(map[string]interface{}); ok {
			if id, ok := defRef["id"].(string); ok {
				if !strings.EqualFold(id, evaluateBranchProtectionDefId) {
					return fmt.Errorf("invalid definitionRef id: %s", id)
				}
			}
		}
	}
	return nil
}
