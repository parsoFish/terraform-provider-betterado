package memberentitlementmanagement

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/accounts"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/licensing"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/memberentitlementmanagement"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// ── Plan modifier helpers ─────────────────────────────────────────────────────

// uePlanRequiresReplace is a plan modifier that marks the resource for
// replacement when the string attribute value changes.
type uePlanRequiresReplace struct{}

func ueRequiresReplace() planmodifier.String { return uePlanRequiresReplace{} }
func (m uePlanRequiresReplace) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m uePlanRequiresReplace) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m uePlanRequiresReplace) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

// ueUseStateForUnknown copies the prior state value when the plan value is
// unknown (prevents perpetual diffs for computed-only attributes).
type ueUseStateForUnknown struct{}

func ueStateForUnknown() planmodifier.String { return ueUseStateForUnknown{} }
func (m ueUseStateForUnknown) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m ueUseStateForUnknown) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m ueUseStateForUnknown) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
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
	_ resource.Resource                = (*UserEntitlementResource)(nil)
	_ resource.ResourceWithConfigure   = (*UserEntitlementResource)(nil)
	_ resource.ResourceWithImportState = (*UserEntitlementResource)(nil)
)

// UserEntitlementResource is the terraform-plugin-framework implementation of
// betterado_user_entitlement.
type UserEntitlementResource struct {
	client *client.AggregatedClient
}

// NewUserEntitlementResource returns a new resource.Resource for betterado_user_entitlement.
func NewUserEntitlementResource() resource.Resource {
	return &UserEntitlementResource{}
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *UserEntitlementResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "betterado_user_entitlement"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *UserEntitlementResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a user entitlement within Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					ueStateForUnknown(),
				},
			},
			"principal_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					ueRequiresReplace(),
				},
				Description: "The principal name, i.e. the e-mail address, of the user.",
			},
			"origin_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					ueRequiresReplace(),
				},
				Description: "The origin ID of the user. Used when origin is set.",
			},
			"origin": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					ueRequiresReplace(),
				},
				Description: "The origin of the user, e.g. 'aad' or 'ghb'.",
			},
			"account_license_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The account license type of the user. Defaults to 'express'.",
			},
			"licensing_source": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The licensing source of the user. Defaults to 'account'.",
			},
			"descriptor": schema.StringAttribute{
				Computed:    true,
				Description: "The descriptor of the user.",
				PlanModifiers: []planmodifier.String{
					ueStateForUnknown(),
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *UserEntitlementResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// userEntitlementModel is the Terraform state model for betterado_user_entitlement.
type userEntitlementModel struct {
	ID                 types.String `tfsdk:"id"`
	PrincipalName      types.String `tfsdk:"principal_name"`
	OriginID           types.String `tfsdk:"origin_id"`
	Origin             types.String `tfsdk:"origin"`
	AccountLicenseType types.String `tfsdk:"account_license_type"`
	LicensingSource    types.String `tfsdk:"licensing_source"`
	Descriptor         types.String `tfsdk:"descriptor"`
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *UserEntitlementResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model userEntitlementModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been properly configured with Azure DevOps credentials.")
		return
	}

	// Apply defaults before expanding
	accountLicenseType := model.AccountLicenseType.ValueString()
	if accountLicenseType == "" {
		accountLicenseType = string(licensing.AccountLicenseTypeValues.Express)
	}
	licensingSource := model.LicensingSource.ValueString()
	if licensingSource == "" {
		licensingSource = string(licensing.LicensingSourceValues.Account)
	}

	userEntitlement, err := expandUserEntitlementFramework(&model, accountLicenseType, licensingSource)
	if err != nil {
		resp.Diagnostics.AddError("Error expanding user entitlement", err.Error())
		return
	}

	addedUserEntitlement, err := addUserEntitlementFramework(r.client, userEntitlement)
	if err != nil {
		resp.Diagnostics.AddError("Error creating user entitlement", err.Error())
		return
	}

	model.ID = types.StringValue(addedUserEntitlement.Id.String())

	readErr := r.readIntoModel(ctx, &model)
	if readErr != nil {
		resp.Diagnostics.AddError("Error reading user entitlement after create", readErr.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *UserEntitlementResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var model userEntitlementModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been properly configured with Azure DevOps credentials.")
		return
	}

	err := r.readIntoModel(ctx, &model)
	if err != nil {
		// Check if resource was removed
		if err.Error() == "resource_removed" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading user entitlement", err.Error())
		return
	}

	// ID cleared means resource was removed
	if model.ID.IsNull() || model.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *UserEntitlementResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan userEntitlementModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state userEntitlementModel
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
		resp.Diagnostics.AddError("Invalid user entitlement ID", err.Error())
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

	patchResponse, err := r.client.MemberEntitleManagementClient.UpdateUserEntitlement(r.client.Ctx,
		memberentitlementmanagement.UpdateUserEntitlementArgs{
			UserId: &id,
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
		resp.Diagnostics.AddError("Error updating user entitlement", err.Error())
		return
	}
	if !*patchResponse.IsSuccess {
		resp.Diagnostics.AddError("Error updating user entitlement",
			getUserEntitlementAPIErrorMessage(patchResponse.OperationResults))
		return
	}

	plan.ID = state.ID
	readErr := r.readIntoModel(ctx, &plan)
	if readErr != nil {
		resp.Diagnostics.AddError("Error reading user entitlement after update", readErr.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserEntitlementResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var model userEntitlementModel
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
		resp.Diagnostics.AddError("Invalid user entitlement ID", err.Error())
		return
	}

	err = r.client.MemberEntitleManagementClient.DeleteUserEntitlement(r.client.Ctx,
		memberentitlementmanagement.DeleteUserEntitlementArgs{
			UserId: &id,
		})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting user entitlement", err.Error())
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

var userEntitlementEmailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func (r *UserEntitlementResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider has not been properly configured with Azure DevOps credentials.")
		return
	}

	importID := req.ID
	_, err := uuid.Parse(importID)
	if err != nil {
		// Try UPN (email address)
		upn := importID
		if !userEntitlementEmailRegexp.MatchString(upn) {
			resp.Diagnostics.AddError(
				"Invalid import ID",
				fmt.Sprintf("Only UUID and UPN values can be used for import [%s]", upn),
			)
			return
		}

		result, err := r.client.IdentityClient.ReadIdentities(r.client.Ctx, identity.ReadIdentitiesArgs{
			SearchFilter: converter.String("General"),
			FilterValue:  &upn,
		})
		if err != nil {
			resp.Diagnostics.AddError("Error resolving UPN for import", err.Error())
			return
		}
		if result == nil || len(*result) == 0 {
			resp.Diagnostics.AddError("Not found", fmt.Sprintf("No entitlement found for [%s]", upn))
			return
		}
		if len(*result) > 1 {
			resp.Diagnostics.AddError("Ambiguous", fmt.Sprintf("More than one entitlement found for [%s]", upn))
			return
		}
		importID = (*result)[0].Id.String()
	}

	var model userEntitlementModel
	model.ID = types.StringValue(importID)

	readErr := r.readIntoModel(ctx, &model)
	if readErr != nil {
		resp.Diagnostics.AddError("Error reading user entitlement on import", readErr.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Internal read helper ──────────────────────────────────────────────────────

// readIntoModel fetches the user entitlement by ID and populates model.
// Returns a sentinel error "resource_removed" if the resource is gone.
func (r *UserEntitlementResource) readIntoModel(_ context.Context, model *userEntitlementModel) error {
	id, err := uuid.Parse(model.ID.ValueString())
	if err != nil {
		return fmt.Errorf("parsing UserEntitlementID: %s. %w", model.ID.ValueString(), err)
	}

	userEntitlement, err := r.client.MemberEntitleManagementClient.GetUserEntitlement(r.client.Ctx,
		memberentitlementmanagement.GetUserEntitlementArgs{
			UserId: &id,
		})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			model.ID = types.StringNull()
			return nil
		}
		return fmt.Errorf("reading user entitlement: %w", err)
	}

	if isUserDeletedFramework(userEntitlement) {
		model.ID = types.StringNull()
		return nil
	}

	flattenUserEntitlementFramework(model, userEntitlement)
	return nil
}

// ── Expand / Flatten ──────────────────────────────────────────────────────────

func expandUserEntitlementFramework(model *userEntitlementModel, accountLicenseType, licensingSource string) (*memberentitlementmanagement.UserEntitlement, error) {
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
	principalName := model.PrincipalName.ValueString()

	return &memberentitlementmanagement.UserEntitlement{
		AccessLevel: &licensing.AccessLevel{
			AccountLicenseType: alt,
			LicensingSource:    ls,
		},
		User: &graph.GraphUser{
			Origin:        &origin,
			OriginId:      &originID,
			PrincipalName: &principalName,
			SubjectKind:   converter.String("user"),
		},
	}, nil
}

func flattenUserEntitlementFramework(model *userEntitlementModel, ue *memberentitlementmanagement.UserEntitlement) {
	if ue.User != nil {
		if ue.User.Descriptor != nil {
			model.Descriptor = types.StringValue(*ue.User.Descriptor)
		}
		if ue.User.Origin != nil {
			model.Origin = types.StringValue(*ue.User.Origin)
		}
		if ue.User.OriginId != nil {
			model.OriginID = types.StringValue(*ue.User.OriginId)
		}
		if ue.User.PrincipalName != nil {
			model.PrincipalName = types.StringValue(*ue.User.PrincipalName)
		}
	}
	if ue.AccessLevel != nil {
		if ue.AccessLevel.AccountLicenseType != nil {
			model.AccountLicenseType = types.StringValue(string(*ue.AccessLevel.AccountLicenseType))
		}
		if ue.AccessLevel.LicensingSource != nil {
			model.LicensingSource = types.StringValue(strings.ToLower(string(*ue.AccessLevel.LicensingSource)))
		}
	}
}

func addUserEntitlementFramework(clients *client.AggregatedClient, userEntitlement *memberentitlementmanagement.UserEntitlement) (*memberentitlementmanagement.UserEntitlement, error) {
	response, err := clients.MemberEntitleManagementClient.AddUserEntitlement(clients.Ctx,
		memberentitlementmanagement.AddUserEntitlementArgs{
			UserEntitlement: userEntitlement,
		})
	if err != nil {
		return nil, err
	}
	if !*response.IsSuccess {
		opResults := []memberentitlementmanagement.UserEntitlementOperationResult{}
		if response.OperationResult != nil {
			opResults = append(opResults, *response.OperationResult)
		}
		return nil, fmt.Errorf("adding user entitlement: %s", getUserEntitlementAPIErrorMessage(&opResults))
	}
	return response.UserEntitlement, nil
}

func isUserDeletedFramework(userEntitlement *memberentitlementmanagement.UserEntitlement) bool {
	if userEntitlement == nil {
		return true
	}
	if userEntitlement.AccessLevel == nil || userEntitlement.AccessLevel.Status == nil {
		return false
	}
	return *userEntitlement.AccessLevel.Status == accounts.AccountUserStatusValues.Deleted ||
		*userEntitlement.AccessLevel.Status == accounts.AccountUserStatusValues.None
}
