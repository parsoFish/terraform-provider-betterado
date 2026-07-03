package feed

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	feedapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/feed"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Ensure interface compliance at compile time.
var (
	_ resource.Resource                = &feedRetentionPolicyFrameworkResource{}
	_ resource.ResourceWithImportState = &feedRetentionPolicyFrameworkResource{}
)

// feedRetentionPolicyFrameworkResource is the terraform-plugin-framework
// implementation of betterado_feed_retention_policy.
type feedRetentionPolicyFrameworkResource struct {
	client *client.AggregatedClient
}

// NewFeedRetentionPolicyResource returns a new framework resource for
// betterado_feed_retention_policy.
func NewFeedRetentionPolicyResource() resource.Resource {
	return &feedRetentionPolicyFrameworkResource{}
}

// ── Model ──────────────────────────────────────────────────────────────────────

type feedRetentionPolicyModel struct {
	ID                                   types.String `tfsdk:"id"`
	FeedID                               types.String `tfsdk:"feed_id"`
	ProjectID                            types.String `tfsdk:"project_id"`
	CountLimit                           types.Int64  `tfsdk:"count_limit"`
	DaysToKeepRecentlyDownloadedPackages types.Int64  `tfsdk:"days_to_keep_recently_downloaded_packages"`
}

// ── Metadata / Schema ──────────────────────────────────────────────────────────

func (r *feedRetentionPolicyFrameworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_feed_retention_policy"
}

func (r *feedRetentionPolicyFrameworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the retention policy for an Azure DevOps Artifacts feed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The feed ID (same as feed_id), used as the resource identifier.",
				PlanModifiers: []planmodifier.String{
					useStateForUnknown(),
				},
			},
			"feed_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the feed to set the retention policy on.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "The project ID (UUID) for a project-scoped feed. Omit for org-scoped feeds.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"count_limit": schema.Int64Attribute{
				Required:    true,
				Description: "The maximum number of versions per package to keep (1–5000).",
			},
			"days_to_keep_recently_downloaded_packages": schema.Int64Attribute{
				Required:    true,
				Description: "The number of days to keep recently downloaded packages (1–4000).",
			},
		},
	}
}

// ── Configure ──────────────────────────────────────────────────────────────────

func (r *feedRetentionPolicyFrameworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ── Create ─────────────────────────────────────────────────────────────────────

func (r *feedRetentionPolicyFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feed_retention_policy Create: provider client not configured")
		return
	}

	var model feedRetentionPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	feedID := model.FeedID.ValueString()
	projectID := model.ProjectID.ValueString()
	countLimit := int(model.CountLimit.ValueInt64())
	days := int(model.DaysToKeepRecentlyDownloadedPackages.ValueInt64())

	_, err := r.client.FeedClient.SetFeedRetentionPolicies(r.client.Ctx, feedapi.SetFeedRetentionPoliciesArgs{
		Policy: &feedapi.FeedRetentionPolicy{
			CountLimit:                           converter.Int(countLimit),
			DaysToKeepRecentlyDownloadedPackages: converter.Int(days),
		},
		FeedId:  &feedID,
		Project: nilIfEmpty(projectID),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create error",
			fmt.Sprintf("creating feed retention policy for feed %s: %s", feedID, err))
		return
	}

	model.ID = types.StringValue(feedID)

	// Re-read to confirm applied values from ADO.
	d := r.readPolicy(feedID, projectID, &model)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ───────────────────────────────────────────────────────────────────────

func (r *feedRetentionPolicyFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feed_retention_policy Read: provider client not configured")
		return
	}

	var model feedRetentionPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	feedID := model.ID.ValueString()
	if feedID == "" {
		feedID = model.FeedID.ValueString()
	}
	projectID := model.ProjectID.ValueString()

	d := r.readPolicy(feedID, projectID, &model)
	if d.HasError() {
		// Check if it's a 404 — if so, remove from state.
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// readPolicy calls GetFeedRetentionPolicies and populates model. Returns
// diagnostics with an error if the API call fails (including 404).
func (r *feedRetentionPolicyFrameworkResource) readPolicy(feedID, projectID string, model *feedRetentionPolicyModel) diag.Diagnostics {
	var diags diag.Diagnostics

	policy, err := r.client.FeedClient.GetFeedRetentionPolicies(r.client.Ctx, feedapi.GetFeedRetentionPoliciesArgs{
		FeedId:  &feedID,
		Project: nilIfEmpty(projectID),
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			diags.AddError("Not found", fmt.Sprintf("feed retention policy for feed %s was not found", feedID))
			return diags
		}
		diags.AddError("Read error",
			fmt.Sprintf("reading feed retention policy for feed %s: %s", feedID, err))
		return diags
	}

	model.FeedID = types.StringValue(feedID)
	if model.ID.IsNull() || model.ID.IsUnknown() || model.ID.ValueString() == "" {
		model.ID = types.StringValue(feedID)
	}
	if policy != nil {
		if policy.CountLimit != nil {
			model.CountLimit = types.Int64Value(int64(*policy.CountLimit))
		}
		if policy.DaysToKeepRecentlyDownloadedPackages != nil {
			model.DaysToKeepRecentlyDownloadedPackages = types.Int64Value(int64(*policy.DaysToKeepRecentlyDownloadedPackages))
		}
	}
	return diags
}

// ── Update ─────────────────────────────────────────────────────────────────────

func (r *feedRetentionPolicyFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feed_retention_policy Update: provider client not configured")
		return
	}

	var plan feedRetentionPolicyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state feedRetentionPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	feedID := state.ID.ValueString()
	projectID := state.ProjectID.ValueString()
	countLimit := int(plan.CountLimit.ValueInt64())
	days := int(plan.DaysToKeepRecentlyDownloadedPackages.ValueInt64())

	_, err := r.client.FeedClient.SetFeedRetentionPolicies(r.client.Ctx, feedapi.SetFeedRetentionPoliciesArgs{
		Policy: &feedapi.FeedRetentionPolicy{
			CountLimit:                           converter.Int(countLimit),
			DaysToKeepRecentlyDownloadedPackages: converter.Int(days),
		},
		FeedId:  &feedID,
		Project: nilIfEmpty(projectID),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update error",
			fmt.Sprintf("updating feed retention policy for feed %s: %s", feedID, err))
		return
	}

	plan.ID = state.ID
	plan.FeedID = state.FeedID
	plan.ProjectID = state.ProjectID

	// Re-read to confirm applied values.
	d := r.readPolicy(feedID, projectID, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// ── Delete ─────────────────────────────────────────────────────────────────────

func (r *feedRetentionPolicyFrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feed_retention_policy Delete: provider client not configured")
		return
	}

	var model feedRetentionPolicyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	feedID := model.ID.ValueString()
	projectID := model.ProjectID.ValueString()

	err := r.client.FeedClient.DeleteFeedRetentionPolicies(r.client.Ctx, feedapi.DeleteFeedRetentionPoliciesArgs{
		FeedId:  &feedID,
		Project: nilIfEmpty(projectID),
	})
	if err != nil {
		resp.Diagnostics.AddError("Delete error",
			fmt.Sprintf("deleting feed retention policy for feed %s: %s", feedID, err))
	}
}

// ── ImportState ────────────────────────────────────────────────────────────────

func (r *feedRetentionPolicyFrameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "<feedId>" or "<projectId>/<feedId>"
	parts := strings.SplitN(req.ID, "/", 2)
	model := feedRetentionPolicyModel{}

	switch len(parts) {
	case 1:
		model.ID = types.StringValue(parts[0])
		model.FeedID = types.StringValue(parts[0])
	case 2:
		model.ProjectID = types.StringValue(parts[0])
		model.ID = types.StringValue(parts[1])
		model.FeedID = types.StringValue(parts[1])
	default:
		resp.Diagnostics.AddError("Invalid import ID",
			fmt.Sprintf("Unexpected ID format (%q). Expected: <feedId> or <projectId>/<feedId>", req.ID))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
