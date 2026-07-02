package feed

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	feedapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/feed"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Ensure interface compliance at compile time.
var (
	_ resource.Resource                = &feedFrameworkResource{}
	_ resource.ResourceWithImportState = &feedFrameworkResource{}
)

// feedFrameworkResource is the terraform-plugin-framework implementation of
// betterado_feed.
type feedFrameworkResource struct {
	client *client.AggregatedClient
}

// NewFeedResource returns a new framework resource for betterado_feed.
func NewFeedResource() resource.Resource {
	return &feedFrameworkResource{}
}

// ── Models ────────────────────────────────────────────────────────────────────

type feedFeaturesModel struct {
	PermanentDelete types.Bool `tfsdk:"permanent_delete"`
	Restore         types.Bool `tfsdk:"restore"`
}

type feedModel struct {
	ID        types.String        `tfsdk:"id"`
	Name      types.String        `tfsdk:"name"`
	ProjectID types.String        `tfsdk:"project_id"`
	Features  []feedFeaturesModel `tfsdk:"features"`
}

// ── Metadata / Schema ─────────────────────────────────────────────────────────

func (r *feedFrameworkResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_feed"
}

func (r *feedFrameworkResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Azure DevOps Artifacts feed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the feed.",
				PlanModifiers: []planmodifier.String{
					useStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the feed.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the project the feed belongs to. Omit for an org-scoped feed.",
				PlanModifiers: []planmodifier.String{
					requiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"features": schema.ListNestedBlock{
				Description: "Optional feature flags for the feed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"permanent_delete": schema.BoolAttribute{
							Optional:    true,
							Description: "If true, permanently delete the feed on destroy.",
						},
						"restore": schema.BoolAttribute{
							Optional:    true,
							Description: "If true, restore a soft-deleted feed with the same name instead of creating.",
						},
					},
				},
			},
		},
	}
}

// ── Provider data injection ───────────────────────────────────────────────────

func (r *feedFrameworkResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *feedFrameworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feed Create: provider client not configured")
		return
	}

	var model feedModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := model.Name.ValueString()
	projectID := model.ProjectID.ValueString()

	// Restore path: if restore=true and a soft-deleted feed exists, restore it.
	if len(model.Features) > 0 && model.Features[0].Restore.ValueBool() {
		if r.isFeedRestorable(name, projectID) {
			if err := r.restoreFeed(name, projectID); err != nil {
				resp.Diagnostics.AddError("Restore error", fmt.Sprintf("restoring feed %s: %s", name, err))
				return
			}
			feedID, readErr := r.getFeedByName(name, projectID)
			if readErr != nil {
				resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading feed after restore %s: %s", name, readErr))
				return
			}
			model.ID = types.StringValue(feedID)
			resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
			return
		}
	}

	feedDetail, err := r.client.FeedClient.CreateFeed(r.client.Ctx, feedapi.CreateFeedArgs{
		Feed:    &feedapi.Feed{Name: &name},
		Project: nilIfEmpty(projectID),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create error", fmt.Sprintf("creating feed %s: %s", name, err))
		return
	}

	model.ID = types.StringValue(feedDetail.Id.String())
	model.Name = types.StringValue(*feedDetail.Name)
	if feedDetail.Project != nil {
		model.ProjectID = types.StringValue(feedDetail.Project.Id.String())
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (r *feedFrameworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feed Read: provider client not configured")
		return
	}

	var model feedModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	feedID := model.ID.ValueString()
	projectID := model.ProjectID.ValueString()

	feedDetail, err := r.client.FeedClient.GetFeed(r.client.Ctx, feedapi.GetFeedArgs{
		FeedId:  &feedID,
		Project: nilIfEmpty(projectID),
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read error", fmt.Sprintf("reading feed %s: %s", feedID, err))
		return
	}

	if feedDetail != nil {
		model.ID = types.StringValue(feedDetail.Id.String())
		model.Name = types.StringValue(*feedDetail.Name)
		if feedDetail.Project != nil {
			model.ProjectID = types.StringValue(feedDetail.Project.Id.String())
		} else {
			// Org-scoped feed: ADO returns no project. Preserve null so that a
			// config with no project_id attribute sees no diff (null == not set).
			// Do NOT write "" — empty-string != null in terraform-plugin-framework
			// and would cause a perpetual destroy+recreate via requiresReplace.
			model.ProjectID = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Update ────────────────────────────────────────────────────────────────────

func (r *feedFrameworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feed Update: provider client not configured")
		return
	}

	// name and project_id are ForceNew — only the features block is mutable.
	var planModel feedModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateModel feedModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := stateModel.Name.ValueString()
	projectID := stateModel.ProjectID.ValueString()

	_, err := r.client.FeedClient.UpdateFeed(r.client.Ctx, feedapi.UpdateFeedArgs{
		Feed:    &feedapi.FeedUpdate{},
		FeedId:  &name,
		Project: nilIfEmpty(projectID),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update error", fmt.Sprintf("updating feed %s: %s", name, err))
		return
	}

	// Carry forward the immutable fields from state; only features changed.
	planModel.ID = stateModel.ID
	planModel.Name = stateModel.Name
	planModel.ProjectID = stateModel.ProjectID

	resp.Diagnostics.Append(resp.State.Set(ctx, &planModel)...)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (r *feedFrameworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feed Delete: provider client not configured")
		return
	}

	var model feedModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := model.Name.ValueString()
	projectID := model.ProjectID.ValueString()

	if err := r.client.FeedClient.DeleteFeed(r.client.Ctx, feedapi.DeleteFeedArgs{
		FeedId:  &name,
		Project: nilIfEmpty(projectID),
	}); err != nil {
		resp.Diagnostics.AddError("Delete error", fmt.Sprintf("deleting feed %s: %s", name, err))
		return
	}

	if len(model.Features) > 0 && model.Features[0].PermanentDelete.ValueBool() {
		if err := r.client.FeedClient.PermanentDeleteFeed(r.client.Ctx, feedapi.PermanentDeleteFeedArgs{
			FeedId:  &name,
			Project: nilIfEmpty(projectID),
		}); err != nil {
			resp.Diagnostics.AddError("Permanent delete error", fmt.Sprintf("permanently deleting feed %s: %s", name, err))
		}
	}
}

// ── ImportState ───────────────────────────────────────────────────────────────

func (r *feedFrameworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "<feedId>" (org-scoped) or "<projectId>/<feedId>" (project-scoped).
	importID := req.ID
	model := feedModel{}

	slashIdx := -1
	for i, ch := range importID {
		if ch == '/' {
			slashIdx = i
			break
		}
	}

	if slashIdx < 0 {
		model.ID = types.StringValue(importID)
		model.Name = types.StringValue(importID)
		model.ProjectID = types.StringValue("")
	} else {
		projectID := importID[:slashIdx]
		feedID := importID[slashIdx+1:]
		model.ID = types.StringValue(feedID)
		model.Name = types.StringValue(feedID)
		model.ProjectID = types.StringValue(projectID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func (r *feedFrameworkResource) isFeedRestorable(name, projectID string) bool {
	change, err := r.client.FeedClient.GetFeedChange(r.client.Ctx, feedapi.GetFeedChangeArgs{
		FeedId:  &name,
		Project: nilIfEmpty(projectID),
	})
	return err == nil && change != nil && change.ChangeType != nil &&
		*(change.ChangeType) == feedapi.ChangeTypeValues.Delete
}

func (r *feedFrameworkResource) restoreFeed(name, projectID string) error {
	return r.client.FeedClient.RestoreDeletedFeed(r.client.Ctx, feedapi.RestoreDeletedFeedArgs{
		FeedId:  &name,
		Project: nilIfEmpty(projectID),
		PatchJson: &[]webapi.JsonPatchOperation{{
			From:  nil,
			Path:  converter.String("/isDeleted"),
			Op:    &webapi.OperationValues.Replace,
			Value: false,
		}},
	})
}

func (r *feedFrameworkResource) getFeedByName(name, projectID string) (string, error) {
	feedDetail, err := r.client.FeedClient.GetFeed(r.client.Ctx, feedapi.GetFeedArgs{
		FeedId:  &name,
		Project: nilIfEmpty(projectID),
	})
	if err != nil {
		return "", err
	}
	return feedDetail.Id.String(), nil
}

// nilIfEmpty returns nil if s is empty string, otherwise &s.
// Used to distinguish org-scoped API calls (no project parameter) from
// project-scoped ones (non-empty project ID).
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
