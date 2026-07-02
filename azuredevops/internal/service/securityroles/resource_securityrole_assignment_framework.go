package securityroles

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	sdkroles "github.com/parsoFish/terraform-provider-betterado/azuredevops/utils/sdk/securityroles"
)

// Ensure interface compliance.
var _ resource.Resource = &securityRoleAssignmentFrameworkResource{}

// securityRoleAssignmentFrameworkResource is the terraform-plugin-framework
// implementation of betterado_securityrole_assignment.
type securityRoleAssignmentFrameworkResource struct {
	client *client.AggregatedClient
}

// NewSecurityRoleAssignmentResource returns a new framework resource for
// betterado_securityrole_assignment.
func NewSecurityRoleAssignmentResource() resource.Resource {
	return &securityRoleAssignmentFrameworkResource{}
}

// ── Model ─────────────────────────────────────────────────────────────────────

type securityRoleAssignmentModel struct {
	ID         types.String `tfsdk:"id"`
	Scope      types.String `tfsdk:"scope"`
	ResourceID types.String `tfsdk:"resource_id"`
	IdentityID types.String `tfsdk:"identity_id"`
	RoleName   types.String `tfsdk:"role_name"`
}

// ── Metadata / Schema ─────────────────────────────────────────────────────────

func (r *securityRoleAssignmentFrameworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_securityrole_assignment"
}

func (r *securityRoleAssignmentFrameworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a security role assignment in Azure DevOps.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The resource ID, formed as `<scope>/<resource_id>/<identity_id>`.",
				PlanModifiers: []planmodifier.String{
					sraUseStateForUnknown(),
				},
			},
			"scope": schema.StringAttribute{
				Required:    true,
				Description: "The security role scope (e.g. `distributedtask.environmentreferencerole`).",
			},
			"resource_id": schema.StringAttribute{
				Required:    true,
				Description: "The resource ID within the scope. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					sraRequiresReplace(),
				},
			},
			"identity_id": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the identity (user or group) to assign the role to.",
			},
			"role_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the security role to assign (e.g. `Administrator`, `User`, `Reader`).",
			},
		},
	}
}

// ── Provider data injection ───────────────────────────────────────────────────

func (r *securityRoleAssignmentFrameworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	agg, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.AggregatedClient, got %T", req.ProviderData),
		)
		return
	}
	r.client = agg
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *securityRoleAssignmentFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan securityRoleAssignmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider was not configured with valid credentials.")
		return
	}

	identityID, err := uuid.Parse(plan.IdentityID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid identity_id", fmt.Sprintf("Could not parse identity_id as UUID: %s", err))
		return
	}

	scope := plan.Scope.ValueString()
	resourceID := plan.ResourceID.ValueString()
	roleName := plan.RoleName.ValueString()

	if err := r.client.SecurityRolesClient.SetSecurityRoleAssignment(ctx, &sdkroles.SetSecurityRoleAssignmentArgs{
		Scope:      &scope,
		ResourceId: &resourceID,
		IdentityId: &identityID,
		RoleName:   &roleName,
	}); err != nil {
		resp.Diagnostics.AddError("Error creating security role assignment", err.Error())
		return
	}

	// Poll until the assignment reflects the desired state.
	if err := r.waitForAssignment(ctx, scope, resourceID, identityID, roleName, 10*time.Minute); err != nil {
		resp.Diagnostics.AddError("Error waiting for security role assignment", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", scope, resourceID, identityID.String()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityRoleAssignmentFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state securityRoleAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider was not configured with valid credentials.")
		return
	}

	identityID, err := uuid.Parse(state.IdentityID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid identity_id", fmt.Sprintf("Could not parse identity_id as UUID: %s", err))
		return
	}

	scope := state.Scope.ValueString()
	resourceID := state.ResourceID.ValueString()

	assignment, err := r.client.SecurityRolesClient.GetSecurityRoleAssignment(ctx, &sdkroles.GetSecurityRoleAssignmentArgs{
		Scope:      &scope,
		ResourceId: &resourceID,
		IdentityId: &identityID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(strings.ToLower(err.Error()), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading security role assignment", err.Error())
		return
	}

	// Treat nil-identity + nil-role as gone.
	if assignment == nil || (assignment.Identity == nil && assignment.Role == nil) {
		resp.State.RemoveResource(ctx)
		return
	}

	if assignment.Role != nil && assignment.Role.Name != nil {
		state.RoleName = types.StringValue(*assignment.Role.Name)
	}
	// Note: do NOT overwrite state.Scope from assignment.Role.Scope.
	// Role.Scope is the role-definition's own scope field (may differ from the
	// query scope used to look up the assignment).  The query scope is the
	// resource's stable identity key and must not drift.
	if assignment.Identity != nil && assignment.Identity.ID != nil {
		state.IdentityID = types.StringValue(*assignment.Identity.ID)
	}
	state.ResourceID = types.StringValue(resourceID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *securityRoleAssignmentFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan securityRoleAssignmentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider was not configured with valid credentials.")
		return
	}

	identityID, err := uuid.Parse(plan.IdentityID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid identity_id", fmt.Sprintf("Could not parse identity_id as UUID: %s", err))
		return
	}

	scope := plan.Scope.ValueString()
	resourceID := plan.ResourceID.ValueString()
	roleName := plan.RoleName.ValueString()

	if err := r.client.SecurityRolesClient.SetSecurityRoleAssignment(ctx, &sdkroles.SetSecurityRoleAssignmentArgs{
		Scope:      &scope,
		ResourceId: &resourceID,
		IdentityId: &identityID,
		RoleName:   &roleName,
	}); err != nil {
		resp.Diagnostics.AddError("Error updating security role assignment", err.Error())
		return
	}

	// Poll until the assignment reflects the desired state.
	if err := r.waitForAssignment(ctx, scope, resourceID, identityID, roleName, 10*time.Minute); err != nil {
		resp.Diagnostics.AddError("Error waiting for security role assignment update", err.Error())
		return
	}

	// Preserve existing ID on update.
	var state securityRoleAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	plan.ID = state.ID

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *securityRoleAssignmentFrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state securityRoleAssignmentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if r.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider was not configured with valid credentials.")
		return
	}

	identityID, err := uuid.Parse(state.IdentityID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid identity_id", fmt.Sprintf("Could not parse identity_id as UUID: %s", err))
		return
	}

	scope := state.Scope.ValueString()
	resourceID := state.ResourceID.ValueString()

	if err := r.client.SecurityRolesClient.DeleteSecurityRoleAssignment(ctx, &sdkroles.DeleteSecurityRoleAssignmentArgs{
		Scope:      &scope,
		ResourceId: &resourceID,
		IdentityId: &identityID,
	}); err != nil {
		resp.Diagnostics.AddError("Error deleting security role assignment", err.Error())
		return
	}

	// Poll until the assignment is gone (ADO security-roles API is eventually
	// consistent — the PATCH delete can return 200 before the assignment
	// disappears from subsequent GET responses).
	if err := r.waitForDeletion(ctx, scope, resourceID, identityID, 10*time.Minute); err != nil {
		resp.Diagnostics.AddError("Error waiting for security role assignment deletion", err.Error())
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// waitForAssignment polls the API until the assignment reports the expected role
// (or until timeout). This mirrors the SDKv2 StateChangeConf approach.
func (r *securityRoleAssignmentFrameworkResource) waitForAssignment(
	ctx context.Context,
	scope, resourceID string,
	identityID uuid.UUID,
	roleName string,
	timeout time.Duration,
) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		assignment, err := r.client.SecurityRolesClient.GetSecurityRoleAssignment(ctx, &sdkroles.GetSecurityRoleAssignmentArgs{
			Scope:      &scope,
			ResourceId: &resourceID,
			IdentityId: &identityID,
		})
		if err != nil {
			// Treat as transient — keep polling.
			time.Sleep(5 * time.Second)
			continue
		}
		if assignment == nil || (assignment.Identity == nil && assignment.Role == nil) {
			time.Sleep(5 * time.Second)
			continue
		}
		if assignment.Identity != nil && assignment.Identity.ID != nil &&
			!strings.EqualFold(*assignment.Identity.ID, identityID.String()) {
			time.Sleep(5 * time.Second)
			continue
		}
		if assignment.Role != nil && assignment.Role.Name != nil &&
			strings.EqualFold(*assignment.Role.Name, roleName) &&
			(assignment.Role.Scope == nil || strings.EqualFold(*assignment.Role.Scope, scope)) {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("timed out waiting for security role assignment to reflect role %q", roleName)
}

// waitForDeletion polls the API until the explicit assignment for the given
// identity is gone (or until timeout). This handles ADO's eventual-consistency
// window after a PATCH revert-to-inherited request.
//
// ADO security roles distinguish "assigned" (explicit) from "inherited" access.
// The PATCH delete reverts an explicit assignment to inherited; if the identity
// has no inherited role the entry disappears entirely. Either outcome — no
// matching entry OR an entry with Access="inherited" — means the explicit
// Terraform-managed assignment has been successfully removed.
func (r *securityRoleAssignmentFrameworkResource) waitForDeletion(
	ctx context.Context,
	scope, resourceID string,
	identityID uuid.UUID,
	timeout time.Duration,
) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		assignment, err := r.client.SecurityRolesClient.GetSecurityRoleAssignment(ctx, &sdkroles.GetSecurityRoleAssignmentArgs{
			Scope:      &scope,
			ResourceId: &resourceID,
			IdentityId: &identityID,
		})
		if err != nil {
			// A 404 / "not found" response means the assignment was already removed.
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "404") {
				return nil
			}
			return fmt.Errorf("poll GetSecurityRoleAssignment: %w", err)
		}
		if assignment == nil || (assignment.Identity == nil && assignment.Role == nil) {
			// No matching assignment found — deletion confirmed.
			return nil
		}
		// If the remaining entry has inherited access, the explicit assignment
		// was successfully removed (inherited is expected ADO behaviour).
		if assignment.Access != nil && strings.EqualFold(*assignment.Access, "inherited") {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return fmt.Errorf("timed out waiting for security role assignment to be deleted (scope=%s resource_id=%s identity_id=%s)", scope, resourceID, identityID.String())
}

// ── Inline plan modifiers ─────────────────────────────────────────────────────

func sraRequiresReplace() planmodifier.String {
	return sraRequiresReplaceModifier{}
}

type sraRequiresReplaceModifier struct{}

func (m sraRequiresReplaceModifier) Description(_ context.Context) string {
	return "Forces replacement when the value changes."
}

func (m sraRequiresReplaceModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m sraRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.StateValue.IsNull() && !req.PlanValue.Equal(req.StateValue) {
		resp.RequiresReplace = true
	}
}

func sraUseStateForUnknown() planmodifier.String {
	return sraUseStateForUnknownModifier{}
}

type sraUseStateForUnknownModifier struct{}

func (m sraUseStateForUnknownModifier) Description(_ context.Context) string {
	return "Use state value when plan value is unknown."
}

func (m sraUseStateForUnknownModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m sraUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}
