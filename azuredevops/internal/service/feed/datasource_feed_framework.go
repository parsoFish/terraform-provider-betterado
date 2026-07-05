package feed

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	feedapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/feed"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// Ensure interface compliance at compile time.
var (
	_ datasource.DataSource                     = &feedDataSource{}
	_ datasource.DataSourceWithConfigValidators = &feedDataSource{}
)

// feedDataSource is the terraform-plugin-framework implementation of
// data.betterado_feed.
type feedDataSource struct {
	client *client.AggregatedClient
}

// NewFeedDataSource returns a new framework data source for data.betterado_feed.
func NewFeedDataSource() datasource.DataSource {
	return &feedDataSource{}
}

// ── Model ─────────────────────────────────────────────────────────────────────

type feedDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	FeedID    types.String `tfsdk:"feed_id"`
	ProjectID types.String `tfsdk:"project_id"`
}

// ── Metadata / Schema ─────────────────────────────────────────────────────────

func (d *feedDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_feed"
}

func (d *feedDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to access information about an existing Azure DevOps Artifacts feed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the feed.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the feed. Exactly one of name or feed_id must be set.",
				Validators: []validator.String{
					stringNotWhiteSpace(),
				},
			},
			"feed_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The UUID of the feed. Exactly one of name or feed_id must be set.",
				Validators: []validator.String{
					stringIsUUID(),
				},
			},
			"project_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The ID of the project the feed belongs to. Omit for an org-scoped feed.",
				Validators: []validator.String{
					stringIsUUID(),
				},
			},
		},
	}
}

// ── ConfigValidators ──────────────────────────────────────────────────────────

// ConfigValidators implements datasource.DataSourceWithConfigValidators.
// It enforces that 'name' and 'feed_id' are mutually exclusive at plan time
// using datasourcevalidator.Conflicting from terraform-plugin-framework-validators,
// replacing the SDKv2 ConflictsWith behaviour that was dropped during migration.
func (d *feedDataSource) ConfigValidators(_ context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("name"),
			path.MatchRoot("feed_id"),
		),
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *feedDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	agg, ok := req.ProviderData.(*client.AggregatedClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.AggregatedClient, got: %T", req.ProviderData),
		)
		return
	}
	d.client = agg
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (d *feedDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state feedDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.client == nil {
		resp.Diagnostics.AddError("Provider not configured", "The provider was not configured before reading data.betterado_feed.")
		return
	}

	// Validate: at least one of name/feed_id must be set.
	name := state.Name.ValueString()
	feedID := state.FeedID.ValueString()
	if name == "" && feedID == "" {
		resp.Diagnostics.AddError(
			"Missing required attribute",
			"At least one of 'name' or 'feed_id' must be set.",
		)
		return
	}

	// Prefer feed_id (UUID) over name when both are set, matching SDKv2 behaviour.
	identifier := feedID
	if identifier == "" {
		identifier = name
	}

	projectID := state.ProjectID.ValueString()

	getFeed, err := d.client.FeedClient.GetFeed(d.client.Ctx, feedapi.GetFeedArgs{
		FeedId:  &identifier,
		Project: nilIfEmpty(projectID),
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			// Data sources don't remove state on 404 — just return empty.
			return
		}
		resp.Diagnostics.AddError("Error reading feed", fmt.Sprintf("reading feed %q: %+v", identifier, err))
		return
	}

	if getFeed == nil {
		return
	}

	state.ID = types.StringValue(getFeed.Id.String())
	state.Name = types.StringValue(*getFeed.Name)
	state.FeedID = types.StringValue(getFeed.Id.String())
	if getFeed.Project != nil {
		state.ProjectID = types.StringValue(getFeed.Project.Id.String())
	} else if projectID == "" {
		// Org-scoped feed — no project; set computed field to null.
		state.ProjectID = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
