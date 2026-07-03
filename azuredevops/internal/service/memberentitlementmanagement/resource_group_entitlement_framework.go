package memberentitlementmanagement

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/licensing"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/memberentitlementmanagement"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// ── Static string default ─────────────────────────────────────────────────────

// geStaticStringDefault is a defaults.String that always returns a fixed value.
type geStaticStringDefault struct{ value string }

func geStringDefault(v string) defaults.String { return geStaticStringDefault{value: v} }

func (s geStaticStringDefault) Description(_ context.Context) string {
	return fmt.Sprintf("defaults to %q", s.value)
}

func (s geStaticStringDefault) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("defaults to `%s`", s.value)
}

func (s geStaticStringDefault) DefaultString(_ context.Context, _ defaults.StringRequest, resp *defaults.StringResponse) {
	resp.PlanValue = types.StringValue(s.value)
}

// ── Plan modifier helpers ─────────────────────────────────────────────────────

// gePlanRequiresReplace marks the resource for replacement when the string
// attribute value changes.
type gePlanRequiresReplace struct{}

func geRequiresReplace() planmodifier.String { return gePlanRequiresReplace{} }
func (m gePlanRequiresReplace) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m gePlanRequiresReplace) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m gePlanRequiresReplace) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// geUseStateForUnknown copies the prior state value when the plan value is
// unknown (prevents perpetual diffs for computed-only attributes).
type geUseStateForUnknown struct{}

func geStateForUnknown() planmodifier.String { return geUseStateForUnknown{} }
func (m geUseStateForUnknown) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m geUseStateForUnknown) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m geUseStateForUnknown) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// Compile-time interface checks.
var (
	_ resource.Resource                = (*GroupEntitlementResource)(nil)
	_ resource.ResourceWithConfigure   = (*GroupEntitlementResource)(nil)
	_ resource.ResourceWithImportState = (*GroupEntitlementResource)(nil)
)

// GroupEntitlementResource is the terraform-plugin-framework implementation of
// betterado_group_entitlement.
type GroupEntitlementResource struct {
	client *client.AggregatedClient
}

// NewGroupEntitlementResource returns a new resource.Resource for betterado_group_entitlement.
func NewGroupEntitlementResource() resource.Resource {
	return &GroupEntitlementResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *GroupEntitlementResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_group_entitlement"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *GroupEntitlementResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a group entitlement within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					geStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					geRequiresReplace(),
					geStateForUnknown(),
				},
				Description: "The display name of the group.",
			},
			"origin_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					geRequiresReplace(),
					geStateForUnknown(),
				},
				Description: "The origin ID of the group. Used when origin is set.",
			},
			"origin": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					geRequiresReplace(),
					geStateForUnknown(),
				},
				Description: "The origin of the group, e.g. 'aad' or 'vsts'.",
			},
			"account_license_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     geStringDefault(string(licensing.AccountLicenseTypeValues.Express)),
				Description: "The account license type of the group. Defaults to 'express'.",
			},
			"licensing_source": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     geStringDefault(string(licensing.LicensingSourceValues.Account)),
				Description: "The licensing source of the group. Defaults to 'account'.",
			},
			"principal_name": schema.StringAttribute{
				Computed:    true,
				Description: "The principal name of the group.",
				PlanModifiers: []planmodifier.String{
					geStateForUnknown(),
				},
			},
			"descriptor": schema.StringAttribute{
				Computed:    true,
				Description: "The descriptor of the group.",
				PlanModifiers: []planmodifier.String{
					geStateForUnknown(),
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *GroupEntitlementResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// groupEntitlementModel is the Terraform state model for betterado_group_entitlement.
type groupEntitlementModel struct {
	ID                 types.String `tfsdk:"id"`
	DisplayName        types.String `tfsdk:"display_name"`
	OriginID           types.String `tfsdk:"origin_id"`
	Origin             types.String `tfsdk:"origin"`
	AccountLicenseType types.String `tfsdk:"account_license_type"`
	LicensingSource    types.String `tfsdk:"licensing_source"`
	PrincipalName      types.String `tfsdk:"principal_name"`
	Descriptor         types.String `tfsdk:"descriptor"`
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *GroupEntitlementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model groupEntitlementModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been properly configured with Azure DevOps credentials.")
		return
	}

	accountLicenseType := model.AccountLicenseType.ValueString()
	if accountLicenseType == "" {
		accountLicenseType = string(licensing.AccountLicenseTypeValues.Express)
	}
	licensingSource := model.LicensingSource.ValueString()
	if licensingSource == "" {
		licensingSource = string(licensing.LicensingSourceValues.Account)
	}

	groupEntitlement, err := expandGroupEntitlementFramework(&model, accountLicenseType, licensingSource)
	if err != nil {
		resp.Diagnostics.AddError("Error expanding group entitlement", err.Error())
		return
	}

	addedGroupEntitlement, err := addGroupEntitlementFramework(r.client, groupEntitlement)
	if err != nil {
		resp.Diagnostics.AddError("Error creating group entitlement", err.Error())
		return
	}

	model.ID = types.StringValue(addedGroupEntitlement.Id.String())

	readErr := r.readIntoGroupModel(ctx, &model)
	if readErr != nil {
		resp.Diagnostics.AddError("Error reading group entitlement after create", readErr.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *GroupEntitlementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model groupEntitlementModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been properly configured with Azure DevOps credentials.")
		return
	}

	err := r.readIntoGroupModel(ctx, &model)
	if err != nil {
		if err.Error() == "resource_removed" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading group entitlement", err.Error())
		return
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *GroupEntitlementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan groupEntitlementModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state groupEntitlementModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been properly configured with Azure DevOps credentials.")
		return
	}

	id, err := uuid.Parse(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid group entitlement ID", err.Error())
		return
	}

	accountLicenseType := plan.AccountLicenseType.ValueString()
	if accountLicenseType == "" {
		accountLicenseType = string(licensing.AccountLicenseTypeValues.Express)
	}
	licensingSource := plan.LicensingSource.ValueString()
	if licensingSource == "" {
		licensingSource = string(licensing.LicensingSourceValues.Account)
	}

	alt, err := converter.AccountLicenseType(accountLicenseType)
	if err != nil {
		resp.Diagnostics.AddError("Invalid account_license_type", err.Error())
		return
	}

	patchResponse, err := r.client.MemberEntitleManagementClient.UpdateGroupEntitlement(r.client.Ctx,
		memberentitlementmanagement.UpdateGroupEntitlementArgs{
			GroupId: &id,
			Document: &[]webapi.JsonPatchOperation{
				{
					Op:   &webapi.OperationValues.Replace,
					From: nil,
					Path: converter.String("/accessLevel"),
					Value: struct {
						AccountLicenseType string `json:"accountLicenseType"`
						LicensingSource    string `json:"licensingSource"`
					}{
						string(*alt),
						licensingSource,
					},
				},
			},
		})
	if err != nil {
		resp.Diagnostics.AddError("Error updating group entitlement", err.Error())
		return
	}

	result := *patchResponse.Results
	if !*result[0].IsSuccess {
		opResults := []memberentitlementmanagement.GroupOperationResult{result[0]}
		resp.Diagnostics.AddError("Error updating group entitlement",
			getGroupEntitlementAPIErrorMessage(&opResults))
		return
	}

	plan.ID = state.ID
	readErr := r.readIntoGroupModel(ctx, &plan)
	if readErr != nil {
		resp.Diagnostics.AddError("Error reading group entitlement after update", readErr.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GroupEntitlementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model groupEntitlementModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been properly configured with Azure DevOps credentials.")
		return
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		return
	}

	id, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid group entitlement ID", err.Error())
		return
	}

	_, err = r.client.MemberEntitleManagementClient.DeleteGroupEntitlement(r.client.Ctx,
		memberentitlementmanagement.DeleteGroupEntitlementArgs{
			GroupId: &id,
		})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting group entitlement", err.Error())
		return
	}

	// Also delete the org-wide group if it's an Azure DevOps local group (origin == "vsts").
	// This matches the SDKv2 delete logic.
	origin := model.Origin.ValueString()
	if origin == "vsts" {
		descriptor := model.Descriptor.ValueString()
		err = r.client.GraphClient.DeleteGroup(r.client.Ctx, graph.DeleteGroupArgs{
			GroupDescriptor: &descriptor,
		})
		if err != nil {
			resp.Diagnostics.AddError("Error deleting Azure DevOps local group", err.Error())
		}
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

func (r *GroupEntitlementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been properly configured with Azure DevOps credentials.")
		return
	}

	importID := req.ID
	_, err := uuid.Parse(importID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Only UUID values can be used for import [%s]", importID),
		)
		return
	}

	var model groupEntitlementModel
	model.ID = types.StringValue(importID)

	readErr := r.readIntoGroupModel(ctx, &model)
	if readErr != nil {
		resp.Diagnostics.AddError("Error reading group entitlement on import", readErr.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Internal read helper ──────────────────────────────────────────────────────

// readIntoGroupModel fetches the group entitlement by ID and populates model.
// Clears model.ID (to types.StringNull) when the resource is gone.
func (r *GroupEntitlementResource) readIntoGroupModel(_ context.Context, model *groupEntitlementModel) error {
	id, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		return fmt.Errorf("parsing GroupEntitlementID: %s. %w", model.ID.ValueString(), err)
	}

	groupEntitlement, err := r.client.MemberEntitleManagementClient.GetGroupEntitlement(r.client.Ctx,
		memberentitlementmanagement.GetGroupEntitlementArgs{
			GroupId: &id,
		})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			model.ID = types.StringNull()
			return nil
		}
		return fmt.Errorf("reading group entitlement: %w", err)
	}

	if groupEntitlement == nil || groupEntitlement.Id == nil {
		log.Println("Group entitlement has been deleted")
		model.ID = types.StringNull()
		return nil
	}

	if groupEntitlement.Group != nil && groupEntitlement.Group.IsDeleted != nil && *groupEntitlement.Group.IsDeleted {
		log.Println("Group has been deleted")
		model.ID = types.StringNull()
		return nil
	}

	flattenGroupEntitlementFramework(model, groupEntitlement)
	return nil
}

// ── Expand / Flatten ──────────────────────────────────────────────────────────

func expandGroupEntitlementFramework(model *groupEntitlementModel, accountLicenseType, licensingSource string) (*memberentitlementmanagement.GroupEntitlement, error) {
	alt, err := converter.AccountLicenseType(accountLicenseType)
	if err != nil {
		return nil, err
	}
	ls, err := converter.AccountLicensingSource(licensingSource)
	if err != nil {
		return nil, err
	}

	origin := model.Origin.ValueString()
	originID := model.OriginID.ValueString()
	displayName := model.DisplayName.ValueString()

	return &memberentitlementmanagement.GroupEntitlement{
		LicenseRule: &licensing.AccessLevel{
			AccountLicenseType: alt,
			LicensingSource:    ls,
		},
		Group: &graph.GraphGroup{
			Origin:      &origin,
			OriginId:    &originID,
			DisplayName: &displayName,
			SubjectKind: converter.String("group"),
		},
	}, nil
}

func flattenGroupEntitlementFramework(model *groupEntitlementModel, ge *memberentitlementmanagement.GroupEntitlement) {
	if ge.Group != nil {
		if ge.Group.Descriptor != nil {
			model.Descriptor = types.StringValue(*ge.Group.Descriptor)
		}
		if ge.Group.Origin != nil {
			model.Origin = types.StringValue(*ge.Group.Origin)
		}
		if ge.Group.OriginId != nil {
			model.OriginID = types.StringValue(*ge.Group.OriginId)
		}
		if ge.Group.PrincipalName != nil {
			model.PrincipalName = types.StringValue(*ge.Group.PrincipalName)
		}
		if ge.Group.DisplayName != nil {
			model.DisplayName = types.StringValue(*ge.Group.DisplayName)
		}
	}
	if ge.LicenseRule != nil {
		if ge.LicenseRule.AccountLicenseType != nil {
			model.AccountLicenseType = types.StringValue(string(*ge.LicenseRule.AccountLicenseType))
		}
		if ge.LicenseRule.LicensingSource != nil {
			model.LicensingSource = types.StringValue(strings.ToLower(string(*ge.LicenseRule.LicensingSource)))
		}
	}
}

func addGroupEntitlementFramework(clients *client.AggregatedClient, groupEntitlement *memberentitlementmanagement.GroupEntitlement) (*memberentitlementmanagement.GroupEntitlement, error) {
	response, err := clients.MemberEntitleManagementClient.AddGroupEntitlement(clients.Ctx,
		memberentitlementmanagement.AddGroupEntitlementArgs{
			GroupEntitlement: groupEntitlement,
		})
	if err != nil {
		return nil, err
	}

	result := *response.Results
	if !*result[0].IsSuccess {
		opResults := []memberentitlementmanagement.GroupOperationResult{}
		if result[0].Errors != nil {
			opResults = append(opResults, result[0])
		}
		return nil, fmt.Errorf("adding group entitlement: %s", getGroupEntitlementAPIErrorMessage(&opResults))
	}

	return result[0].Result, nil
}
