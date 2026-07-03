package feed

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	feedapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/feed"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Ensure interface compliance at compile time.
var (
	_ resource.Resource                = &feedPermissionFrameworkResource{}
	_ resource.ResourceWithImportState = &feedPermissionFrameworkResource{}
)

// feedPermissionFrameworkResource is the terraform-plugin-framework implementation
// of betterado_feed_permission.
type feedPermissionFrameworkResource struct {
	client *client.AggregatedClient
}

// NewFeedPermissionResource returns a new framework resource for betterado_feed_permission.
func NewFeedPermissionResource() resource.Resource {
	return &feedPermissionFrameworkResource{}
}

// ── Model ─────────────────────────────────────────────────────────────────────

type feedPermissionModel struct {
	ID                 types.String `tfsdk:"id"`
	FeedID             types.String `tfsdk:"feed_id"`
	IdentityDescriptor types.String `tfsdk:"identity_descriptor"`
	IdentityID         types.String `tfsdk:"identity_id"`
	Role               types.String `tfsdk:"role"`
	ProjectID          types.String `tfsdk:"project_id"`
	DisplayName        types.String `tfsdk:"display_name"`
}

// ── Metadata / Schema ─────────────────────────────────────────────────────────

func (r *feedPermissionFrameworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_feed_permission"
}

func (r *feedPermissionFrameworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages feed permissions in Azure DevOps Artifacts.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of this feed permission resource.",
				PlanModifiers: []planmodifier.String{
					useStateForUnknown(),
				},
			},
			"feed_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the feed.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
				Validators: []validator.String{
					stringIsUUID(),
				},
			},
			"identity_descriptor": schema.StringAttribute{
				Required:    true,
				Description: "The identity descriptor of the group or user.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
				Validators: []validator.String{
					stringNotWhiteSpace(),
				},
			},
			"identity_id": schema.StringAttribute{
				Computed:    true,
				Description: "The identity ID resolved from the descriptor.",
				PlanModifiers: []planmodifier.String{
					useStateForUnknown(),
				},
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "The role to assign: reader, contributor, administrator, or collaborator.",
				Validators: []validator.String{
					stringOneOf("reader", "contributor", "administrator", "collaborator"),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "The project ID (UUID) for a project-scoped feed. Omit for org-scoped feeds.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
				Validators: []validator.String{
					stringIsUUID(),
				},
			},
			"display_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The display name for the permission entry.",
				PlanModifiers: []planmodifier.String{
					useStateForUnknown(),
				},
				Validators: []validator.String{
					stringNotWhiteSpace(),
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *feedPermissionFrameworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *feedPermissionFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feed_permission Create: provider client not configured")
		return
	}

	var model feedPermissionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	feedID := model.FeedID.ValueString()
	identityDescriptor := model.IdentityDescriptor.ValueString()
	role := feedapi.FeedRole(model.Role.ValueString())
	projectID := model.ProjectID.ValueString()
	displayName := model.DisplayName.ValueString()

	// Resolve the identity.
	identityResp, err := r.resolveIdentity(identityDescriptor)
	if err != nil {
		resp.Diagnostics.AddError("Identity resolution error",
			fmt.Sprintf("resolving identity %s: %s", identityDescriptor, err))
		return
	}

	// Check for conflict — permission must not already exist.
	existing, findErr := r.findPermission(feedID, projectID, *identityResp.Descriptor)
	if findErr != nil && !utils.ResponseWasNotFound(findErr) {
		resp.Diagnostics.AddError("Pre-create check error",
			fmt.Sprintf("checking existing feed permission for feed %s / identity %s: %s", feedID, identityDescriptor, findErr))
		return
	}
	if existing != nil {
		resp.Diagnostics.AddError("Conflict",
			fmt.Sprintf("Feed Permission for Feed : %s and Identity : %s already exists", feedID, identityDescriptor))
		return
	}

	_, err = r.client.FeedClient.SetFeedPermissions(r.client.Ctx, feedapi.SetFeedPermissionsArgs{
		FeedId:  &feedID,
		Project: nilIfEmpty(projectID),
		FeedPermission: &[]feedapi.FeedPermission{
			{
				DisplayName:        &displayName,
				IdentityDescriptor: identityResp.Descriptor,
				IdentityId:         identityResp.Id,
				Role:               &role,
			},
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Create error",
			fmt.Sprintf("creating feed permission for feed %s / identity %s: %s", feedID, identityDescriptor, err))
		return
	}

	// Poll until the permission appears (ADO propagation delay).
	if pollErr := r.pollPermissions(feedID, projectID, *identityResp.Descriptor); pollErr != nil {
		resp.Diagnostics.AddError("Sync error",
			fmt.Sprintf("waiting for feed permission to sync: %s", pollErr))
		return
	}

	id, idErr := uuid.NewUUID()
	if idErr != nil {
		resp.Diagnostics.AddError("ID generation error", fmt.Sprintf("generating resource ID: %s", idErr))
		return
	}
	model.ID = types.StringValue(fmt.Sprintf("fp-%s", id.String()))
	model.IdentityID = types.StringValue(identityResp.Id.String())

	// Re-read to populate computed display_name from ADO.
	// MUST set display_name to a known value; if it stays unknown Terraform
	// raises "provider still indicated an unknown value after apply".
	// We only overwrite if the config did not explicitly set display_name (i.e.
	// it was null or unknown in the plan), to avoid idempotency issues when the
	// user provides an explicit value that differs from what ADO echoes back.
	if model.DisplayName.IsUnknown() || model.DisplayName.IsNull() {
		perm, readErr := r.findPermission(feedID, projectID, *identityResp.Descriptor)
		if readErr == nil && perm != nil && perm.DisplayName != nil {
			model.DisplayName = types.StringValue(*perm.DisplayName)
		} else {
			// Fall back to empty string so the attribute is known after apply.
			model.DisplayName = types.StringValue("")
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *feedPermissionFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feed_permission Read: provider client not configured")
		return
	}

	var model feedPermissionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	feedID := model.FeedID.ValueString()
	identityDescriptor := model.IdentityDescriptor.ValueString()
	projectID := model.ProjectID.ValueString()

	identityResp, err := r.resolveIdentity(identityDescriptor)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Identity error",
			fmt.Sprintf("resolving identity %s: %s", identityDescriptor, err))
		return
	}

	perm, err := r.findPermission(feedID, projectID, *identityResp.Descriptor)
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read error",
			fmt.Sprintf("reading feed permission for feed %s / identity %s: %s", feedID, identityDescriptor, err))
		return
	}

	if perm == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if perm.DisplayName != nil {
		model.DisplayName = types.StringValue(*perm.DisplayName)
	}
	if perm.Role != nil {
		model.Role = types.StringValue(string(*perm.Role))
	}
	model.IdentityDescriptor = types.StringValue(identityDescriptor)
	model.IdentityID = types.StringValue(identityResp.Id.String())

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *feedPermissionFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feed_permission Update: provider client not configured")
		return
	}

	var plan feedPermissionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state feedPermissionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	feedID := state.FeedID.ValueString()
	identityDescriptor := state.IdentityDescriptor.ValueString()
	role := feedapi.FeedRole(plan.Role.ValueString())
	projectID := state.ProjectID.ValueString()
	displayName := plan.DisplayName.ValueString()

	identityResp, err := r.resolveIdentity(identityDescriptor)
	if err != nil {
		resp.Diagnostics.AddError("Identity error",
			fmt.Sprintf("resolving identity %s during update: %s", identityDescriptor, err))
		return
	}

	_, err = r.client.FeedClient.SetFeedPermissions(r.client.Ctx, feedapi.SetFeedPermissionsArgs{
		FeedId:  &feedID,
		Project: nilIfEmpty(projectID),
		FeedPermission: &[]feedapi.FeedPermission{
			{
				DisplayName:        &displayName,
				IdentityDescriptor: identityResp.Descriptor,
				IdentityId:         identityResp.Id,
				Role:               &role,
			},
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Update error",
			fmt.Sprintf("updating feed permission for feed %s / identity %s: %s", feedID, identityDescriptor, err))
		return
	}

	if pollErr := r.pollPermissions(feedID, projectID, *identityResp.Descriptor); pollErr != nil {
		resp.Diagnostics.AddError("Sync error",
			fmt.Sprintf("waiting for feed permission update to sync: %s", pollErr))
		return
	}

	plan.ID = state.ID
	plan.IdentityID = types.StringValue(identityResp.Id.String())

	// Ensure display_name is always a known value after update.
	if plan.DisplayName.IsUnknown() {
		if state.DisplayName.IsNull() || state.DisplayName.IsUnknown() {
			plan.DisplayName = types.StringValue("")
		} else {
			plan.DisplayName = state.DisplayName
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *feedPermissionFrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feed_permission Delete: provider client not configured")
		return
	}

	var model feedPermissionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	feedID := model.FeedID.ValueString()
	identityDescriptor := model.IdentityDescriptor.ValueString()
	projectID := model.ProjectID.ValueString()
	noneRole := feedapi.FeedRoleValues.None

	identityResp, err := r.resolveIdentity(identityDescriptor)
	if err != nil {
		resp.Diagnostics.AddError("Identity error",
			fmt.Sprintf("resolving identity %s during delete: %s", identityDescriptor, err))
		return
	}

	_, err = r.client.FeedClient.SetFeedPermissions(r.client.Ctx, feedapi.SetFeedPermissionsArgs{
		FeedId:  &feedID,
		Project: nilIfEmpty(projectID),
		FeedPermission: &[]feedapi.FeedPermission{
			{
				IdentityDescriptor: identityResp.Descriptor,
				Role:               &noneRole,
			},
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Delete error",
			fmt.Sprintf("deleting feed permission for feed %s / identity %s: %s", feedID, identityDescriptor, err))
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

func (r *feedPermissionFrameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "<feedId>/<identityDescriptor>" or "<projectId>/<feedId>/<identityDescriptor>"
	parts := splitImportID(req.ID, 3)
	model := feedPermissionModel{}

	switch len(parts) {
	case 2:
		model.FeedID = types.StringValue(parts[0])
		model.IdentityDescriptor = types.StringValue(parts[1])
	case 3:
		model.ProjectID = types.StringValue(parts[0])
		model.FeedID = types.StringValue(parts[1])
		model.IdentityDescriptor = types.StringValue(parts[2])
	default:
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Unexpected ID format (%q). Expected: <feedId>/<identityDescriptor> or <projectId>/<feedId>/<identityDescriptor>", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// resolveIdentity resolves a subject descriptor to a full Identity.
func (r *feedPermissionFrameworkResource) resolveIdentity(descriptor string) (*identity.Identity, error) {
	storageKey, err := r.client.GraphClient.GetStorageKey(r.client.Ctx, graph.GetStorageKeyArgs{
		SubjectDescriptor: &descriptor,
	})
	if err != nil {
		return nil, err
	}

	id, err := r.client.IdentityClient.ReadIdentity(r.client.Ctx, identity.ReadIdentityArgs{
		IdentityId: converter.String(storageKey.Value.String()),
	})
	if err != nil {
		return nil, err
	}
	return id, nil
}

// findPermission returns the FeedPermission matching the given identity descriptor,
// or a 404 WrappedError if not found.
func (r *feedPermissionFrameworkResource) findPermission(feedID, projectID, identityDescriptor string) (*feedapi.FeedPermission, error) {
	permissions, err := r.client.FeedClient.GetFeedPermissions(r.client.Ctx, feedapi.GetFeedPermissionsArgs{
		FeedId:  &feedID,
		Project: nilIfEmpty(projectID),
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			msg := fmt.Sprintf("error reading permission for Feed: %s and Identity: %s", feedID, identityDescriptor)
			return nil, azuredevops.WrappedError{
				StatusCode: converter.Int(http.StatusNotFound),
				Message:    &msg,
			}
		}
		return nil, err
	}

	for _, perm := range *permissions {
		if perm.IdentityDescriptor != nil && *perm.IdentityDescriptor == identityDescriptor {
			p := perm
			return &p, nil
		}
	}

	msg := fmt.Sprintf("error reading permission for Feed: %s and Identity: %s", feedID, identityDescriptor)
	return nil, azuredevops.WrappedError{
		StatusCode: converter.Int(http.StatusNotFound),
		Message:    &msg,
	}
}

// pollPermissions waits until the permission is visible in the ADO API.
func (r *feedPermissionFrameworkResource) pollPermissions(feedID, projectID, identityDescriptor string) error {
	stateConf := &retry.StateChangeConf{
		ContinuousTargetOccurence: 2,
		Delay:                     5 * time.Second,
		MinTimeout:                10 * time.Second,
		Timeout:                   10 * time.Minute,
		Pending:                   []string{syncing},
		Target:                    []string{succeed, failed},
		Refresh: func() (interface{}, string, error) {
			perm, err := r.findPermission(feedID, projectID, identityDescriptor)
			if err != nil {
				if utils.ResponseWasNotFound(err) {
					return nil, syncing, nil
				}
				return nil, failed, err
			}
			if perm == nil {
				return nil, syncing, nil
			}
			return perm, succeed, nil
		},
	}
	_, err := stateConf.WaitForStateContext(r.client.Ctx)
	if err != nil {
		return fmt.Errorf("waiting for feed permission: %w", err)
	}
	return nil
}

// splitImportID splits s by '/' into at most maxParts parts.
func splitImportID(s string, maxParts int) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '/' && len(parts) < maxParts-1 {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}
