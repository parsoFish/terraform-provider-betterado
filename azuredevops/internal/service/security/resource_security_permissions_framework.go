package security

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/security"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// Ensure interface compliance.
var _ resource.Resource = &securityPermissionsFrameworkResource{}

// securityPermissionsFrameworkResource is the terraform-plugin-framework
// implementation of betterado_security_permissions.
type securityPermissionsFrameworkResource struct {
	client *client.AggregatedClient
}

// NewSecurityPermissionsResource returns a new framework resource for
// betterado_security_permissions.
func NewSecurityPermissionsResource() resource.Resource {
	return &securityPermissionsFrameworkResource{}
}

// ── Model ─────────────────────────────────────────────────────────────────────

type securityPermissionsModel struct {
	ID          types.String `tfsdk:"id"`
	NamespaceID types.String `tfsdk:"namespace_id"`
	Token       types.String `tfsdk:"token"`
	Principal   types.String `tfsdk:"principal"`
	Permissions types.Map    `tfsdk:"permissions"`
	Replace     types.Bool   `tfsdk:"replace"`
}

// ── Metadata / Schema ─────────────────────────────────────────────────────────

func (r *securityPermissionsFrameworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_permissions"
}

func (r *securityPermissionsFrameworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages generic security ACL permissions in Azure DevOps via the Security REST API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The resource ID, formed as `<namespace_id>/<token>/<principal>`.",
				PlanModifiers: []planmodifier.String{
					spUseStateForUnknownString(),
				},
			},
			"namespace_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the security namespace.",
				PlanModifiers: []planmodifier.String{
					spRequiresReplaceString(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`),
						"must be a valid UUID",
					),
				},
			},
			"token": schema.StringAttribute{
				Required:    true,
				Description: "The security token for the resource.",
				PlanModifiers: []planmodifier.String{
					spRequiresReplaceString(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"principal": schema.StringAttribute{
				Required:    true,
				Description: "The descriptor or identity ID of the principal (user or group).",
				PlanModifiers: []planmodifier.String{
					spRequiresReplaceString(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"permissions": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Map of permission names to values (allow, deny, or notset).",
			},
			"replace": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     spStaticBool(true),
				Description: "Replace existing permissions (true) or merge with existing (false).",
			},
		},
	}
}

// ── Provider data injection ───────────────────────────────────────────────────

func (r *securityPermissionsFrameworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *client.AggregatedClient, got %T", req.ProviderData),
		)
		return
	}
	r.client = c
}

// ── Create ────────────────────────────────────────────────────────────────────

func (r *securityPermissionsFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_security_permissions Create: provider client not configured")
		return
	}

	var model securityPermissionsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyPermissions(ctx, &model); err != nil {
		resp.Diagnostics.AddError("Create error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *securityPermissionsFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_security_permissions Read: provider client not configured")
		return
	}

	var model securityPermissionsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	namespaceID, err := uuid.Parse(model.NamespaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("invalid namespace_id: %v", err))
		return
	}

	token := model.Token.ValueString()
	principal := model.Principal.ValueString()

	// Get declared permissions from state to filter read-back.
	declaredPerms := map[string]string{}
	resp.Diagnostics.Append(model.Permissions.ElementsAs(ctx, &declaredPerms, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Resolve identity descriptor.
	identityDescriptor, err := resolveIdentityDescriptor(r.client, principal)
	if err != nil {
		log.Printf("[INFO] betterado_security_permissions Read: unable to resolve identity for principal %s, removing from state: %v", principal, err)
		resp.State.RemoveResource(ctx)
		return
	}

	// Get namespace actions for bit→name mapping.
	actionMap, err := r.getActionMap(namespaceID)
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("querying namespace actions: %v", err))
		return
	}

	// Query ACL.
	bTrue := true
	acls, err := r.client.SecurityClient.QueryAccessControlLists(r.client.Ctx, security.QueryAccessControlListsArgs{
		SecurityNamespaceId: &namespaceID,
		Token:               &token,
		Descriptors:         &identityDescriptor,
		IncludeExtendedInfo: &bTrue,
	})
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("querying ACL: %v", err))
		return
	}

	if acls == nil || len(*acls) == 0 {
		log.Printf("[INFO] betterado_security_permissions Read: no ACL for token %s, removing from state", token)
		resp.State.RemoveResource(ctx)
		return
	}

	acl := (*acls)[0]
	if acl.AcesDictionary == nil {
		log.Printf("[INFO] betterado_security_permissions Read: no ACEs for principal %s, removing from state", principal)
		resp.State.RemoveResource(ctx)
		return
	}

	ace, ok := (*acl.AcesDictionary)[identityDescriptor]
	if !ok {
		log.Printf("[INFO] betterado_security_permissions Read: ACE for principal %s not found, removing from state", principal)
		resp.State.RemoveResource(ctx)
		return
	}

	allowBits := 0
	denyBits := 0
	if ace.Allow != nil {
		allowBits = *ace.Allow
	}
	if ace.Deny != nil {
		denyBits = *ace.Deny
	}

	// Only return the keys that were declared in config (prevents perpetual diff).
	currentPermissions := make(map[string]string, len(declaredPerms))
	for permName := range declaredPerms {
		bit, exists := actionMap[permName]
		if !exists {
			resp.Diagnostics.AddError("Read error", fmt.Sprintf("permission '%s' not found in namespace %s", permName, namespaceID.String()))
			return
		}
		switch {
		case (allowBits & bit) != 0:
			currentPermissions[permName] = "allow"
		case (denyBits & bit) != 0:
			currentPermissions[permName] = "deny"
		default:
			currentPermissions[permName] = "notset"
		}
	}

	permMap, diags := types.MapValueFrom(ctx, types.StringType, currentPermissions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Permissions = permMap

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *securityPermissionsFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_security_permissions Update: provider client not configured")
		return
	}

	var model securityPermissionsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Merge in state for computed fields.
	var stateModel securityPermissionsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if model.ID.IsUnknown() || model.ID.IsNull() {
		model.ID = stateModel.ID
	}

	if err := r.applyPermissions(ctx, &model); err != nil {
		resp.Diagnostics.AddError("Update error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *securityPermissionsFrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_security_permissions Delete: provider client not configured")
		return
	}

	var model securityPermissionsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	namespaceID, err := uuid.Parse(model.NamespaceID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("invalid namespace_id: %v", err))
		return
	}

	token := model.Token.ValueString()
	principal := model.Principal.ValueString()

	// Resolve identity descriptor.
	identityDescriptor, err := resolveIdentityDescriptor(r.client, principal)
	if err != nil {
		log.Printf("[INFO] betterado_security_permissions Delete: unable to resolve identity for principal %s, assuming already removed: %v", principal, err)
		return
	}

	// Get namespace actions.
	actionMap, err := r.getActionMap(namespaceID)
	if err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("querying namespace actions: %v", err))
		return
	}

	// Get permissions map to determine managed bits.
	configPerms := map[string]string{}
	resp.Diagnostics.Append(model.Permissions.ElementsAs(ctx, &configPerms, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read current ACL.
	bTrue := true
	acls, err := r.client.SecurityClient.QueryAccessControlLists(r.client.Ctx, security.QueryAccessControlListsArgs{
		SecurityNamespaceId: &namespaceID,
		Token:               &token,
		Descriptors:         &identityDescriptor,
		IncludeExtendedInfo: &bTrue,
	})
	if err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("querying current ACL: %v", err))
		return
	}

	if acls == nil || len(*acls) == 0 {
		log.Printf("[INFO] betterado_security_permissions Delete: no ACL found for token %s", token)
		return
	}

	acl := (*acls)[0]
	if acl.AcesDictionary == nil {
		log.Printf("[INFO] betterado_security_permissions Delete: no ACEs for principal %s", principal)
		return
	}

	ace, ok := (*acl.AcesDictionary)[identityDescriptor]
	if !ok {
		log.Printf("[INFO] betterado_security_permissions Delete: ACE for principal %s not found", principal)
		return
	}

	currentAllowBits := 0
	currentDenyBits := 0
	if ace.Allow != nil {
		currentAllowBits = *ace.Allow
	}
	if ace.Deny != nil {
		currentDenyBits = *ace.Deny
	}

	managedBits := 0
	for permName := range configPerms {
		if bit, exists := actionMap[permName]; exists {
			managedBits |= bit
		}
	}

	newAllowBits := currentAllowBits &^ managedBits
	newDenyBits := currentDenyBits &^ managedBits

	updatedACE := security.AccessControlEntry{
		Descriptor: &identityDescriptor,
		Allow:      &newAllowBits,
		Deny:       &newDenyBits,
		ExtendedInfo: &security.AceExtendedInformation{
			EffectiveAllow: &newAllowBits,
			EffectiveDeny:  &newDenyBits,
			InheritedAllow: new(int),
			InheritedDeny:  new(int),
		},
	}

	bMerge := false
	container := struct {
		Token                *string                        `json:"token,omitempty"`
		Merge                *bool                          `json:"merge,omitempty"`
		AccessControlEntries *[]security.AccessControlEntry `json:"accessControlEntries,omitempty"`
	}{
		Token:                &token,
		Merge:                &bMerge,
		AccessControlEntries: &[]security.AccessControlEntry{updatedACE},
	}

	_, err = r.client.SecurityClient.SetAccessControlEntries(r.client.Ctx, security.SetAccessControlEntriesArgs{
		SecurityNamespaceId: &namespaceID,
		Container:           container,
	})
	if err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("removing managed permissions: %v", err))
		return
	}

	// Wait for managed bits to clear.
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		currentACL, pollErr := r.client.SecurityClient.QueryAccessControlLists(r.client.Ctx, security.QueryAccessControlListsArgs{
			SecurityNamespaceId: &namespaceID,
			Token:               &token,
			Descriptors:         &identityDescriptor,
			IncludeExtendedInfo: &[]bool{true}[0],
		})
		if pollErr != nil {
			break
		}
		if currentACL == nil || len(*currentACL) == 0 {
			break
		}
		pollACL := (*currentACL)[0]
		if pollACL.AcesDictionary == nil {
			break
		}
		aceEntry, aceOK := (*pollACL.AcesDictionary)[identityDescriptor]
		if !aceOK {
			break
		}
		curAllow := 0
		curDeny := 0
		if aceEntry.Allow != nil {
			curAllow = *aceEntry.Allow
		}
		if aceEntry.Deny != nil {
			curDeny = *aceEntry.Deny
		}
		if (curAllow&managedBits) == 0 && (curDeny&managedBits) == 0 {
			break
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}

	log.Printf("[INFO] betterado_security_permissions Delete: successfully removed managed permissions for principal %s", principal)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// getActionMap returns a map of action name → bit value for the given namespace.
func (r *securityPermissionsFrameworkResource) getActionMap(namespaceID uuid.UUID) (map[string]int, error) {
	namespaces, err := r.client.SecurityClient.QuerySecurityNamespaces(r.client.Ctx, security.QuerySecurityNamespacesArgs{
		SecurityNamespaceId: &namespaceID,
	})
	if err != nil {
		return nil, fmt.Errorf("querying security namespace: %v", err)
	}
	if namespaces == nil || len(*namespaces) == 0 {
		return nil, fmt.Errorf("namespace %s not found", namespaceID.String())
	}

	ns := (*namespaces)[0]
	actionMap := make(map[string]int)
	if ns.Actions != nil {
		for _, action := range *ns.Actions {
			if action.Name != nil && action.Bit != nil {
				actionMap[*action.Name] = *action.Bit
			}
		}
	}
	return actionMap, nil
}

// applyPermissions sets ACEs and waits for sync.
func (r *securityPermissionsFrameworkResource) applyPermissions(ctx context.Context, model *securityPermissionsModel) error {
	namespaceID, err := uuid.Parse(model.NamespaceID.ValueString())
	if err != nil {
		return fmt.Errorf("invalid namespace_id: %v", err)
	}

	token := model.Token.ValueString()
	principal := model.Principal.ValueString()
	replace := model.Replace.ValueBool()

	configPerms := map[string]string{}
	if diags := model.Permissions.ElementsAs(ctx, &configPerms, false); diags.HasError() {
		return fmt.Errorf("reading permissions from plan")
	}

	// Resolve identity descriptor.
	identityDescriptor, err := resolveIdentityDescriptor(r.client, principal)
	if err != nil {
		return fmt.Errorf("resolving identity for principal '%s': %v", principal, err)
	}

	// Get namespace actions.
	actionMap, err := r.getActionMap(namespaceID)
	if err != nil {
		return err
	}

	// Compute allow/deny bits.
	allowBits := 0
	denyBits := 0

	for permName, permValue := range configPerms {
		bit, exists := actionMap[permName]
		if !exists {
			return fmt.Errorf("permission '%s' not found in namespace %s", permName, namespaceID.String())
		}
		switch strings.ToLower(permValue) {
		case "allow":
			allowBits |= bit
		case "deny":
			denyBits |= bit
		case "notset":
			// notset — bits will not be included
		default:
			return fmt.Errorf("invalid permission value '%s' for permission '%s'. Must be allow, deny, or notset", permValue, permName)
		}
	}

	ace := security.AccessControlEntry{
		Descriptor: &identityDescriptor,
		Allow:      &allowBits,
		Deny:       &denyBits,
		ExtendedInfo: &security.AceExtendedInformation{
			EffectiveAllow: &allowBits,
			EffectiveDeny:  &denyBits,
			InheritedAllow: new(int),
			InheritedDeny:  new(int),
		},
	}

	bMerge := !replace
	container := struct {
		Token                *string                        `json:"token,omitempty"`
		Merge                *bool                          `json:"merge,omitempty"`
		AccessControlEntries *[]security.AccessControlEntry `json:"accessControlEntries,omitempty"`
	}{
		Token:                &token,
		Merge:                &bMerge,
		AccessControlEntries: &[]security.AccessControlEntry{ace},
	}

	_, err = r.client.SecurityClient.SetAccessControlEntries(r.client.Ctx, security.SetAccessControlEntriesArgs{
		SecurityNamespaceId: &namespaceID,
		Container:           container,
	})
	if err != nil {
		return fmt.Errorf("setting permissions: %v", err)
	}

	// Poll for sync.
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		currentACL, pollErr := r.client.SecurityClient.QueryAccessControlLists(r.client.Ctx, security.QueryAccessControlListsArgs{
			SecurityNamespaceId: &namespaceID,
			Token:               &token,
			Descriptors:         &identityDescriptor,
			IncludeExtendedInfo: &[]bool{true}[0],
		})
		if pollErr != nil {
			return fmt.Errorf("polling permissions: %v", pollErr)
		}

		if currentACL != nil && len(*currentACL) > 0 {
			acl := (*currentACL)[0]
			if acl.AcesDictionary != nil {
				if aceEntry, aceOK := (*acl.AcesDictionary)[identityDescriptor]; aceOK {
					curAllow := 0
					curDeny := 0
					if aceEntry.Allow != nil {
						curAllow = *aceEntry.Allow
					}
					if aceEntry.Deny != nil {
						curDeny = *aceEntry.Deny
					}
					if replace {
						if curAllow == allowBits && curDeny == denyBits {
							break
						}
					} else {
						if (curAllow&allowBits) == allowBits && (curDeny&denyBits) == denyBits {
							break
						}
					}
				}
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}

	model.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", namespaceID.String(), token, principal))
	return nil
}

// ── Inline plan modifiers ─────────────────────────────────────────────────────

func spRequiresReplaceString() planmodifier.String {
	return spRequiresReplaceStringModifier{}
}

type spRequiresReplaceStringModifier struct{}

func (m spRequiresReplaceStringModifier) Description(_ context.Context) string {
	return "Forces replacement when the value changes."
}

func (m spRequiresReplaceStringModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m spRequiresReplaceStringModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.StateValue.IsNull() && !req.PlanValue.Equal(req.StateValue) {
		resp.RequiresReplace = true
	}
}

func spUseStateForUnknownString() planmodifier.String {
	return spUseStateForUnknownStringModifier{}
}

type spUseStateForUnknownStringModifier struct{}

func (m spUseStateForUnknownStringModifier) Description(_ context.Context) string {
	return "Use state value when plan value is unknown."
}

func (m spUseStateForUnknownStringModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m spUseStateForUnknownStringModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

func spStaticBool(val bool) defaults.Bool {
	return spStaticBoolDefault{value: val}
}

type spStaticBoolDefault struct {
	value bool
}

func (d spStaticBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("Defaults to %v.", d.value)
}

func (d spStaticBoolDefault) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d spStaticBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.value)
}
