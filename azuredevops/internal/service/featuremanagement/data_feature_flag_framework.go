package featuremanagement

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	featuremanagementapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/featuremanagement"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = (*FeatureFlagDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*FeatureFlagDataSource)(nil)
)

// FeatureFlagDataSource is the terraform-plugin-framework data source for
// betterado_feature_flag.
type FeatureFlagDataSource struct {
	client *client.AggregatedClient
}

// NewFeatureFlagDataSource returns a new datasource.DataSource for
// betterado_feature_flag.
func NewFeatureFlagDataSource() datasource.DataSource {
	return &FeatureFlagDataSource{}
}

// featureFlagDataModel is the Terraform state model for the
// betterado_feature_flag data source.
type featureFlagDataModel struct {
	FeatureID  types.String `tfsdk:"feature_id"`
	ScopeName  types.String `tfsdk:"scope_name"`
	ScopeValue types.String `tfsdk:"scope_value"`
	State      types.String `tfsdk:"state"`
	Overridden types.Bool   `tfsdk:"overridden"`
	Reason     types.String `tfsdk:"reason"`
}

// ── Metadata ──────────────────────────────────────────────────────────────────

func (d *FeatureFlagDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "betterado_feature_flag"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (d *FeatureFlagDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to read the current state of an Azure DevOps Feature Management feature flag at a given scope.",
		Attributes: map[string]schema.Attribute{
			"feature_id": schema.StringAttribute{
				Required:    true,
				Description: "The full contribution id of the feature (e.g. \"ms.vss-work.agile\").",
			},
			"scope_name": schema.StringAttribute{
				Required:    true,
				Description: "The setting scope name (e.g. \"project\", \"host\").",
			},
			"scope_value": schema.StringAttribute{
				Required:    true,
				Description: "The scope value (project ID for project-scope; empty string for host-scope).",
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "The current feature state: \"enabled\", \"disabled\", or \"undefined\".",
			},
			"overridden": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the feature state has been set by an override rule.",
			},
			"reason": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "Reason for the feature state.",
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (d *FeatureFlagDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	d.client = agg
}

// ── Read ──────────────────────────────────────────────────────────────────────

func (d *FeatureFlagDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Client not configured", "betterado_feature_flag data source Read: provider client not configured")
		return
	}

	var model featureFlagDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	featureID := model.FeatureID.ValueString()
	scopeName := model.ScopeName.ValueString()
	scopeValue := model.ScopeValue.ValueString()

	state, err := d.client.FeatureManagementClient.GetFeatureStateForScope(ctx, featuremanagementapi.GetFeatureStateForScopeArgs{
		FeatureId:  converter.String(featureID),
		UserScope:  converter.String("host"),
		ScopeName:  converter.String(scopeName),
		ScopeValue: converter.String(scopeValue),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading feature flag",
			fmt.Sprintf("GetFeatureStateForScope failed for feature %q scope %q/%q: %s", featureID, scopeName, scopeValue, err),
		)
		return
	}

	if state == nil || state.State == nil {
		resp.Diagnostics.AddError(
			"Feature flag not found",
			fmt.Sprintf("GetFeatureStateForScope returned nil for feature %q scope %q/%q", featureID, scopeName, scopeValue),
		)
		return
	}

	model.State = types.StringValue(string(*state.State))

	if state.Overridden != nil {
		model.Overridden = types.BoolValue(*state.Overridden)
	} else {
		model.Overridden = types.BoolValue(false)
	}

	if state.Reason != nil {
		model.Reason = types.StringValue(*state.Reason)
	} else {
		model.Reason = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
