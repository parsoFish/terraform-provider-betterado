package notification

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	notificationapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/notification"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// Compile-time interface checks.
var (
	_ datasource.DataSource              = &notificationSubscriptionDataSource{}
	_ datasource.DataSourceWithConfigure = &notificationSubscriptionDataSource{}
)

// notificationSubscriptionDataSource implements the betterado_notification_subscription data source.
type notificationSubscriptionDataSource struct {
	client *client.AggregatedClient
}

// NewNotificationSubscriptionDataSource returns a new framework data source for
// betterado_notification_subscription.
func NewNotificationSubscriptionDataSource() datasource.DataSource {
	return &notificationSubscriptionDataSource{}
}

// notificationSubscriptionDataModel is the Terraform state model for the data source.
type notificationSubscriptionDataModel struct {
	ID               types.String `tfsdk:"id"`
	ProjectID        types.String `tfsdk:"project_id"`
	SubscriptionType types.String `tfsdk:"subscription_type"`
	SubscriberID     types.String `tfsdk:"subscriber_id"`
	ChannelType      types.String `tfsdk:"channel_type"`
	ChannelAddress   types.String `tfsdk:"channel_address"`
	FilterType       types.String `tfsdk:"filter_type"`
	FilterCriteria   types.String `tfsdk:"filter_criteria"`
	Status           types.String `tfsdk:"status"`
}

// Metadata sets the data source type name.
func (d *notificationSubscriptionDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_subscription"
}

// Schema defines the data source schema. All attributes except id are Computed.
func (d *notificationSubscriptionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to read an existing Azure DevOps notification subscription by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the notification subscription to look up.",
			},
			"project_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Azure DevOps project that scopes this subscription.",
			},
			"subscription_type": schema.StringAttribute{
				Computed:    true,
				Description: "The event type ID (e.g. `ms.vss-work.workitem-changed-event`).",
			},
			"subscriber_id": schema.StringAttribute{
				Computed:    true,
				Description: "The subscriber identity descriptor or group descriptor.",
			},
			"channel_type": schema.StringAttribute{
				Computed:    true,
				Description: "Delivery channel type (e.g. `EmailHtml`, `ServiceHooks`).",
			},
			"channel_address": schema.StringAttribute{
				Computed:    true,
				Description: "Email address for EmailHtml channel.",
			},
			"filter_type": schema.StringAttribute{
				Computed:    true,
				Description: "Filter type identifier.",
			},
			"filter_criteria": schema.StringAttribute{
				Computed:    true,
				Description: "Serialised expression criteria (JSON string).",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Subscription status returned by the API.",
			},
		},
	}
}

// Configure injects the provider client.
func (d *notificationSubscriptionDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches the subscription by ID from the ADO Notification API.
func (d *notificationSubscriptionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddWarning("Provider not configured", "NotificationClient is nil; skipping read.")
		return
	}

	var model notificationSubscriptionDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := model.ID.ValueString()
	sub, err := d.client.NotificationClient.GetSubscription(d.client.Ctx, notificationapi.GetSubscriptionArgs{
		SubscriptionId: &id,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			// Subscription is gone — remove from state gracefully.
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading notification subscription", err.Error())
		return
	}

	flattenNotificationSubscriptionData(sub, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

// flattenNotificationSubscriptionData maps the API response to the data source model.
func flattenNotificationSubscriptionData(sub *notificationapi.NotificationSubscription, model *notificationSubscriptionDataModel) {
	if sub == nil {
		return
	}

	if sub.Id != nil {
		model.ID = types.StringValue(*sub.Id)
	}

	// Channel fields
	if sub.Channel != nil && sub.Channel.Type != nil {
		model.ChannelType = types.StringValue(*sub.Channel.Type)
	}

	// Filter fields
	if sub.Filter != nil {
		if sub.Filter.Type != nil {
			model.FilterType = types.StringValue(*sub.Filter.Type)
		}
		if sub.Filter.EventType != nil {
			model.SubscriptionType = types.StringValue(*sub.Filter.EventType)
		}
	}

	// Subscriber
	if sub.Subscriber != nil && sub.Subscriber.Id != nil {
		model.SubscriberID = types.StringValue(*sub.Subscriber.Id)
	}

	// Scope → project_id
	if sub.Scope != nil && sub.Scope.Id != nil {
		model.ProjectID = types.StringValue(sub.Scope.Id.String())
	}

	// Status
	if sub.Status != nil {
		model.Status = types.StringValue(string(*sub.Status))
	}

	// channel_address and filter_criteria: the ADO API does not return these on
	// ISubscriptionChannel/ISubscriptionFilter directly; they default to empty.
	if model.ChannelAddress.IsNull() || model.ChannelAddress.IsUnknown() {
		model.ChannelAddress = types.StringValue("")
	}
	if model.FilterCriteria.IsNull() || model.FilterCriteria.IsUnknown() {
		model.FilterCriteria = types.StringValue("")
	}
}
