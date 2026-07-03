package memberentitlementmanagement

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/accounts"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/licensing"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/memberentitlementmanagement"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// ── Plan modifier helpers ─────────────────────────────────────────────────────

// spePlanRequiresReplace marks the resource for replacement when the string
// attribute value changes.
type spePlanRequiresReplace struct{}

func speRequiresReplace() planmodifier.String { return spePlanRequiresReplace{} }
func (m spePlanRequiresReplace) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m spePlanRequiresReplace) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m spePlanRequiresReplace) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// speUseStateForUnknown copies the prior state value when the plan value is
// unknown (prevents perpetual diffs for computed-only attributes).
type speUseStateForUnknown struct{}

func speStateForUnknown() planmodifier.String { return speUseStateForUnknown{} }
func (m speUseStateForUnknown) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m speUseStateForUnknown) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m speUseStateForUnknown) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
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
	_ resource.Resource                = (*ServicePrincipalEntitlementResource)(nil)
	_ resource.ResourceWithConfigure   = (*ServicePrincipalEntitlementResource)(nil)
	_ resource.ResourceWithImportState = (*ServicePrincipalEntitlementResource)(nil)
)

// ServicePrincipalEntitlementResource is the terraform-plugin-framework implementation of
// betterado_service_principal_entitlement.
type ServicePrincipalEntitlementResource struct {
	client *client.AggregatedClient
}

// NewServicePrincipalEntitlementResource returns a new resource.Resource for betterado_service_principal_entitlement.
func NewServicePrincipalEntitlementResource() resource.Resource {
	return &ServicePrincipalEntitlementResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *ServicePrincipalEntitlementResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_service_principal_entitlement"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *ServicePrincipalEntitlementResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a service principal entitlement within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					speStateForUnknown(),
				},
			},
			"origin_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					speRequiresReplace(),
				},
				Description: "The AAD object ID of the service principal.",
			},
			"origin": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					speRequiresReplace(),
					speStateForUnknown(),
				},
				Description: "The origin of the service principal, e.g. 'aad'.",
			},
			"account_license_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     geStringDefault(string(licensing.AccountLicenseTypeValues.Express)),
				Description: "The account license type. Defaults to 'express'.",
			},
			"licensing_source": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     geStringDefault(string(licensing.LicensingSourceValues.Account)),
				Description: "The licensing source. Defaults to 'account'.",
			},
			"display_name": schema.StringAttribute{
				Computed:    true,
				Description: "The display name of the service principal.",
				PlanModifiers: []planmodifier.String{
					speStateForUnknown(),
				},
			},
			"descriptor": schema.StringAttribute{
				Computed:    true,
				Description: "The descriptor of the service principal.",
				PlanModifiers: []planmodifier.String{
					speStateForUnknown(),
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *ServicePrincipalEntitlementResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// servicePrincipalEntitlementModel is the Terraform state model for betterado_service_principal_entitlement.
type servicePrincipalEntitlementModel struct {
	ID                 types.String `tfsdk:"id"`
	OriginID           types.String `tfsdk:"origin_id"`
	Origin             types.String `tfsdk:"origin"`
	AccountLicenseType types.String `tfsdk:"account_license_type"`
	LicensingSource    types.String `tfsdk:"licensing_source"`
	DisplayName        types.String `tfsdk:"display_name"`
	Descriptor         types.String `tfsdk:"descriptor"`
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *ServicePrincipalEntitlementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model servicePrincipalEntitlementModel
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

	spe, err := expandServicePrincipalEntitlementFramework(&model, accountLicenseType, licensingSource)
	if err != nil {
		resp.Diagnostics.AddError("Error expanding service principal entitlement", err.Error())
		return
	}

	added, err := addServicePrincipalEntitlementFramework(r.client, spe)
	if err != nil {
		resp.Diagnostics.AddError("Error creating service principal entitlement", err.Error())
		return
	}

	model.ID = types.StringValue(added.Id.String())

	readErr := r.readIntoSPEModel(ctx, &model)
	if readErr != nil {
		resp.Diagnostics.AddError("Error reading service principal entitlement after create", readErr.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServicePrincipalEntitlementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model servicePrincipalEntitlementModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been properly configured with Azure DevOps credentials.")
		return
	}

	err := r.readIntoSPEModel(ctx, &model)
	if err != nil {
		if err.Error() == "resource_removed" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading service principal entitlement", err.Error())
		return
	}

	if model.ID.IsNull() || model.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *ServicePrincipalEntitlementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan servicePrincipalEntitlementModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state servicePrincipalEntitlementModel
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
		resp.Diagnostics.AddError("Invalid service principal entitlement ID", err.Error())
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

	patchResponse, err := r.client.MemberEntitleManagementClient.UpdateServicePrincipalEntitlement(r.client.Ctx,
		memberentitlementmanagement.UpdateServicePrincipalEntitlementArgs{
			ServicePrincipalId: &id,
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
		resp.Diagnostics.AddError("Error updating service principal entitlement", err.Error())
		return
	}

	if patchResponse != nil && patchResponse.IsSuccess != nil && !*patchResponse.IsSuccess {
		resp.Diagnostics.AddError("Error updating service principal entitlement",
			getServicePrincipalEntitlementAPIErrorMessage(patchResponse.OperationResults))
		return
	}

	plan.ID = state.ID
	readErr := r.readIntoSPEModel(ctx, &plan)
	if readErr != nil {
		resp.Diagnostics.AddError("Error reading service principal entitlement after update", readErr.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ServicePrincipalEntitlementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model servicePrincipalEntitlementModel
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
		resp.Diagnostics.AddError("Invalid service principal entitlement ID", err.Error())
		return
	}

	err = r.client.MemberEntitleManagementClient.DeleteServicePrincipalEntitlement(r.client.Ctx,
		memberentitlementmanagement.DeleteServicePrincipalEntitlementArgs{
			ServicePrincipalId: &id,
		})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting service principal entitlement", err.Error())
		return
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

func (r *ServicePrincipalEntitlementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

	var model servicePrincipalEntitlementModel
	model.ID = types.StringValue(importID)

	readErr := r.readIntoSPEModel(ctx, &model)
	if readErr != nil {
		resp.Diagnostics.AddError("Error reading service principal entitlement on import", readErr.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Internal read helper ──────────────────────────────────────────────────────

// readIntoSPEModel fetches the service principal entitlement by ID and populates model.
// Clears model.ID (to types.StringNull) when the resource is gone.
func (r *ServicePrincipalEntitlementResource) readIntoSPEModel(_ context.Context, model *servicePrincipalEntitlementModel) error {
	id, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		return fmt.Errorf("parsing ServicePrincipalEntitlementID: %s. %w", model.ID.ValueString(), err)
	}

	spe, err := r.client.MemberEntitleManagementClient.GetServicePrincipalEntitlement(r.client.Ctx,
		memberentitlementmanagement.GetServicePrincipalEntitlementArgs{
			ServicePrincipalId: &id,
		})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			model.ID = types.StringNull()
			return nil
		}
		return fmt.Errorf("reading service principal entitlement: %w", err)
	}

	if spe == nil || spe.Id == nil {
		log.Println("Service Principal Entitlement has been deleted")
		model.ID = types.StringNull()
		return nil
	}

	// Check if the service principal is in a deleted/none access state.
	if isServicePrincipalEntitlementDeleted(spe) {
		log.Println("Service Principal Entitlement access level is deleted/none")
		model.ID = types.StringNull()
		return nil
	}

	// Check IsDeletedInOrigin flag on the service principal.
	if spe.ServicePrincipal != nil && spe.ServicePrincipal.IsDeletedInOrigin != nil && *spe.ServicePrincipal.IsDeletedInOrigin {
		log.Println("Service Principal has been deleted in origin")
		model.ID = types.StringNull()
		return nil
	}

	flattenServicePrincipalEntitlementFramework(model, spe)
	return nil
}

// isServicePrincipalEntitlementDeleted returns true when the access level
// indicates the service principal has been removed (status == Deleted or None).
func isServicePrincipalEntitlementDeleted(spe *memberentitlementmanagement.ServicePrincipalEntitlement) bool {
	if spe == nil || spe.AccessLevel == nil || spe.AccessLevel.Status == nil {
		return false
	}
	return *spe.AccessLevel.Status == accounts.AccountUserStatusValues.Deleted ||
		*spe.AccessLevel.Status == accounts.AccountUserStatusValues.None
}

// ── Expand / Flatten ──────────────────────────────────────────────────────────

func expandServicePrincipalEntitlementFramework(model *servicePrincipalEntitlementModel, accountLicenseType, licensingSource string) (*memberentitlementmanagement.ServicePrincipalEntitlement, error) {
	alt, err := converter.AccountLicenseType(accountLicenseType)
	if err != nil {
		return nil, err
	}
	ls, err := converter.AccountLicensingSource(licensingSource)
	if err != nil {
		return nil, err
	}

	originID := model.OriginID.ValueString()
	origin := model.Origin.ValueString()

	return &memberentitlementmanagement.ServicePrincipalEntitlement{
		AccessLevel: &licensing.AccessLevel{
			AccountLicenseType: alt,
			LicensingSource:    ls,
		},
		ServicePrincipal: &graph.GraphServicePrincipal{
			Origin:      &origin,
			OriginId:    &originID,
			SubjectKind: converter.String("servicePrincipal"),
		},
	}, nil
}

func flattenServicePrincipalEntitlementFramework(model *servicePrincipalEntitlementModel, spe *memberentitlementmanagement.ServicePrincipalEntitlement) {
	if spe.ServicePrincipal != nil {
		if spe.ServicePrincipal.Origin != nil {
			model.Origin = types.StringValue(*spe.ServicePrincipal.Origin)
		}
		if spe.ServicePrincipal.OriginId != nil {
			model.OriginID = types.StringValue(*spe.ServicePrincipal.OriginId)
		}
		if spe.ServicePrincipal.DisplayName != nil {
			model.DisplayName = types.StringValue(*spe.ServicePrincipal.DisplayName)
		}
		if spe.ServicePrincipal.Descriptor != nil {
			model.Descriptor = types.StringValue(*spe.ServicePrincipal.Descriptor)
		}
	}
	if spe.AccessLevel != nil {
		if spe.AccessLevel.AccountLicenseType != nil {
			model.AccountLicenseType = types.StringValue(string(*spe.AccessLevel.AccountLicenseType))
		}
		if spe.AccessLevel.LicensingSource != nil {
			model.LicensingSource = types.StringValue(strings.ToLower(string(*spe.AccessLevel.LicensingSource)))
		}
	}
}

func addServicePrincipalEntitlementFramework(clients *client.AggregatedClient, spe *memberentitlementmanagement.ServicePrincipalEntitlement) (*memberentitlementmanagement.ServicePrincipalEntitlement, error) {
	response, err := clients.MemberEntitleManagementClient.AddServicePrincipalEntitlement(clients.Ctx,
		memberentitlementmanagement.AddServicePrincipalEntitlementArgs{
			ServicePrincipalEntitlement: spe,
		})
	if err != nil {
		return nil, err
	}

	if response.IsSuccess != nil && !*response.IsSuccess {
		opResults := []memberentitlementmanagement.ServicePrincipalEntitlementOperationResult{}
		if response.OperationResult != nil {
			opResults = append(opResults, *response.OperationResult)
		}
		return nil, fmt.Errorf("adding service principal entitlement: %s", getServicePrincipalEntitlementAPIErrorMessage(&opResults))
	}

	return response.ServicePrincipalEntitlement, nil
}
