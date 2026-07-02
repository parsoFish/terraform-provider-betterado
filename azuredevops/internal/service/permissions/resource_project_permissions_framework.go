package permissions

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkschema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	securityhelper "github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/permissions/utils"
)

// Ensure interface compliance.
var _ resource.Resource = &projectPermissionsFrameworkResource{}

// projectPermissionsFrameworkResource is the terraform-plugin-framework
// implementation of betterado_project_permissions.
type projectPermissionsFrameworkResource struct {
	client *client.AggregatedClient
}

// NewProjectPermissionsResource returns a new framework resource for
// betterado_project_permissions.
func NewProjectPermissionsResource() resource.Resource {
	return &projectPermissionsFrameworkResource{}
}

// ── Model ─────────────────────────────────────────────────────────────────────

type projectPermissionsModel struct {
	ID          types.String `tfsdk:"id"`
	ProjectID   types.String `tfsdk:"project_id"`
	Principal   types.String `tfsdk:"principal"`
	Permissions types.Map    `tfsdk:"permissions"`
	Replace     types.Bool   `tfsdk:"replace"`
}

// ── Metadata / Schema ─────────────────────────────────────────────────────────

func (r *projectPermissionsFrameworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_permissions"
}

func (r *projectPermissionsFrameworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages project-level ACL permissions in Azure DevOps (Project security namespace).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The resource ID, formed as `<token>/<principal>`.",
				PlanModifiers: []planmodifier.String{
					ppUseStateForUnknownString(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Azure DevOps project.",
				PlanModifiers: []planmodifier.String{
					ppRequiresReplaceString(),
				},
			},
			"principal": schema.StringAttribute{
				Required:    true,
				Description: "The group or user descriptor to assign permissions to.",
				PlanModifiers: []planmodifier.String{
					ppRequiresReplaceString(),
				},
			},
			"permissions": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Map of permission bit names to `allow`, `deny`, or `notset` (lowercase). Values are stored and compared lowercase.",
			},
			"replace": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     ppStaticBool(true),
				Description: "If true (default), replaces existing ACEs for this principal. If false, merges.",
			},
		},
	}
}

// ── Provider data injection ───────────────────────────────────────────────────

func (r *projectPermissionsFrameworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectPermissionsFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_project_permissions Create: provider client not configured")
		return
	}

	var model projectPermissionsModel
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

func (r *projectPermissionsFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_project_permissions Read: provider client not configured")
		return
	}

	var model projectPermissionsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sn, err := r.newSecurityNamespace(&model)
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("creating security namespace: %s", err))
		return
	}

	principal := model.Principal.ValueString()
	principalPermissions, err := sn.GetPrincipalPermissions(&[]string{principal})
	if err != nil {
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("getting principal permissions: %s", err))
		return
	}
	if principalPermissions == nil {
		log.Printf("[INFO] Permissions for ACL token %q not found. Removing from state", sn.GetToken())
		resp.State.RemoveResource(ctx)
		return
	}

	// Filter returned permissions to only include keys the config declares.
	configPerms := map[string]string{}
	resp.Diagnostics.Append(model.Permissions.ElementsAs(ctx, &configPerms, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filtered := map[string]string{}
	for _, pp := range *principalPermissions {
		for action, perm := range pp.Permissions {
			if _, ok := configPerms[string(action)]; ok {
				filtered[string(action)] = string(perm)
			}
		}
	}

	permMap, diags := types.MapValueFrom(ctx, types.StringType, filtered)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.Permissions = permMap

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *projectPermissionsFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_project_permissions Update: provider client not configured")
		return
	}

	var model projectPermissionsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyPermissions(ctx, &model); err != nil {
		resp.Diagnostics.AddError("Update error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *projectPermissionsFrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_project_permissions Delete: provider client not configured")
		return
	}

	var model projectPermissionsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sn, err := r.newSecurityNamespace(&model)
	if err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("creating security namespace: %s", err))
		return
	}

	principal := model.Principal.ValueString()

	// Build a notset permission map covering all declared permissions.
	configPerms := map[string]string{}
	resp.Diagnostics.Append(model.Permissions.ElementsAs(ctx, &configPerms, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	notsetMap := make(map[securityhelper.ActionName]securityhelper.PermissionType, len(configPerms))
	for action := range configPerms {
		notsetMap[securityhelper.ActionName(action)] = securityhelper.PermissionTypeValues.NotSet
	}

	setPerm := []securityhelper.SetPrincipalPermission{
		{
			Replace: true,
			PrincipalPermission: securityhelper.PrincipalPermission{
				SubjectDescriptor: principal,
				Permissions:       notsetMap,
			},
		},
	}

	if err := sn.SetPrincipalPermissions(&setPerm); err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("removing principal permissions: %s", err))
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// newSecurityNamespace builds a SecurityNamespace for the Project namespace using
// the token derived from model. Token format: "$PROJECT:vstfs:///Classification/TeamProject/{projectId}"
func (r *projectPermissionsFrameworkResource) newSecurityNamespace(model *projectPermissionsModel) (*securityhelper.SecurityNamespace, error) {
	token := buildProjectToken(model)
	return securityhelper.NewSecurityNamespace(
		nil,
		r.client,
		securityhelper.SecurityNamespaceIDValues.Project,
		// TokenCreatorFunc — ignores *sdkschema.ResourceData (nil); returns captured token.
		func(_ *sdkschema.ResourceData, _ *client.AggregatedClient) (string, error) {
			return token, nil
		},
	)
}

// buildProjectToken produces the ACL token for a project:
//
//	"$PROJECT:vstfs:///Classification/TeamProject/{projectId}"
func buildProjectToken(model *projectPermissionsModel) string {
	return fmt.Sprintf("$PROJECT:vstfs:///Classification/TeamProject/%s", model.ProjectID.ValueString())
}

// applyPermissions writes the permissions and waits for them to be synced.
func (r *projectPermissionsFrameworkResource) applyPermissions(ctx context.Context, model *projectPermissionsModel) error {
	sn, err := r.newSecurityNamespace(model)
	if err != nil {
		return fmt.Errorf("creating security namespace: %w", err)
	}

	principal := model.Principal.ValueString()
	configPerms := map[string]string{}
	if diags := model.Permissions.ElementsAs(ctx, &configPerms, false); diags.HasError() {
		return fmt.Errorf("reading permissions from plan")
	}

	replace := model.Replace.ValueBool()

	permMap := make(map[securityhelper.ActionName]securityhelper.PermissionType, len(configPerms))
	for k, v := range configPerms {
		permMap[securityhelper.ActionName(k)] = securityhelper.PermissionType(strings.ToLower(v))
	}

	setPerm := []securityhelper.SetPrincipalPermission{
		{
			Replace: replace,
			PrincipalPermission: securityhelper.PrincipalPermission{
				SubjectDescriptor: principal,
				Permissions:       permMap,
			},
		},
	}

	if err := sn.SetPrincipalPermissions(&setPerm); err != nil {
		return fmt.Errorf("setting principal permissions: %w", err)
	}

	// Poll until the ACL is synced (mirrors SDKv2 wait logic).
	deadline := time.Now().Add(60 * time.Minute)
	for time.Now().Before(deadline) {
		currentPerms, err := sn.GetPrincipalPermissions(&[]string{principal})
		if err != nil {
			return fmt.Errorf("polling permissions: %w", err)
		}
		if currentPerms != nil && len(*currentPerms) == 1 {
			synced := true
			for key, want := range permMap {
				got, ok := (*currentPerms)[0].Permissions[key]
				if !ok || !strings.EqualFold(string(want), string(got)) {
					synced = false
					break
				}
			}
			if synced {
				break
			}
		}
		// Respect context cancellation during the poll.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}

	// Set the resource ID: "<token>/<principal>"
	token := buildProjectToken(model)
	model.ID = types.StringValue(fmt.Sprintf("%s/%s", token, principal))
	return nil
}

// ── Inline plan modifiers ─────────────────────────────────────────────────────

// ppRequiresReplaceString forces resource replacement when a string attribute changes.
func ppRequiresReplaceString() planmodifier.String {
	return ppRequiresReplaceStringModifier{}
}

type ppRequiresReplaceStringModifier struct{}

func (m ppRequiresReplaceStringModifier) Description(_ context.Context) string {
	return "Forces replacement when the value changes."
}

func (m ppRequiresReplaceStringModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m ppRequiresReplaceStringModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.StateValue.IsNull() && !req.PlanValue.Equal(req.StateValue) {
		resp.RequiresReplace = true
	}
}

// ppUseStateForUnknownString keeps the prior state value when the plan value is unknown (computed).
func ppUseStateForUnknownString() planmodifier.String {
	return ppUseStateForUnknownStringModifier{}
}

type ppUseStateForUnknownStringModifier struct{}

func (m ppUseStateForUnknownStringModifier) Description(_ context.Context) string {
	return "Use state value when plan value is unknown."
}

func (m ppUseStateForUnknownStringModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m ppUseStateForUnknownStringModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ppStaticBool provides a schema default value for a bool attribute.
func ppStaticBool(val bool) defaults.Bool {
	return ppStaticBoolDefault{value: val}
}

type ppStaticBoolDefault struct {
	value bool
}

func (d ppStaticBoolDefault) Description(_ context.Context) string {
	return fmt.Sprintf("Defaults to %v.", d.value)
}

func (d ppStaticBoolDefault) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d ppStaticBoolDefault) DefaultBool(_ context.Context, _ defaults.BoolRequest, resp *defaults.BoolResponse) {
	resp.PlanValue = types.BoolValue(d.value)
}

