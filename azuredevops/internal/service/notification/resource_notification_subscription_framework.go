package notification

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	notificationapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/notification"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

// ── Plan modifier helpers ─────────────────────────────────────────────────────

type notifRequiresReplaceModifier struct{}

func notifRequiresReplace() planmodifier.String { return notifRequiresReplaceModifier{} }

func (m notifRequiresReplaceModifier) Description(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m notifRequiresReplaceModifier) MarkdownDescription(_ context.Context) string {
	return "forces replacement when value changes"
}

func (m notifRequiresReplaceModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.StateValue.IsNull() {
		return
	}
	if req.PlanValue.Equal(req.StateValue) {
		return
	}
	resp.RequiresReplace = true
}

type notifUseStateForUnknownModifier struct{}

func notifUseStateForUnknown() planmodifier.String { return notifUseStateForUnknownModifier{} }

func (m notifUseStateForUnknownModifier) Description(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m notifUseStateForUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "uses prior state value when unknown"
}

func (m notifUseStateForUnknownModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !req.PlanValue.IsUnknown() {
		return
	}
	if req.StateValue.IsNull() {
		return
	}
	resp.PlanValue = req.StateValue
}

// ── Compile-time interface checks ─────────────────────────────────────────────

var (
	_ resource.Resource              = &notificationSubscriptionResource{}
	_ resource.ResourceWithConfigure = &notificationSubscriptionResource{}
)

// notificationSubscriptionResource is the terraform-plugin-framework implementation of
// betterado_notification_subscription.
type notificationSubscriptionResource struct {
	client *client.AggregatedClient
}

// NewNotificationSubscriptionResource returns a new framework resource for betterado_notification_subscription.
func NewNotificationSubscriptionResource() resource.Resource {
	return &notificationSubscriptionResource{}
}

// ── State model ───────────────────────────────────────────────────────────────

// notificationSubscriptionModel is the Terraform state model.
type notificationSubscriptionModel struct {
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

// ── Metadata ──────────────────────────────────────────────────────────────────

func (r *notificationSubscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_subscription"
}

// ── Schema ────────────────────────────────────────────────────────────────────

func (r *notificationSubscriptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Azure DevOps notification subscription.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the notification subscription.",
				PlanModifiers: []planmodifier.String{
					notifUseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the Azure DevOps project that scopes this subscription.",
				PlanModifiers: []planmodifier.String{
					notifRequiresReplace(),
				},
			},
			"subscription_type": schema.StringAttribute{
				Required:    true,
				Description: "The event type ID (e.g. `ms.vss-work.workitem-changed-event`).",
			},
			"subscriber_id": schema.StringAttribute{
				Required:    true,
				Description: "The subscriber identity descriptor or group descriptor.",
			},
			"channel_type": schema.StringAttribute{
				Required:    true,
				Description: "Delivery channel type (e.g. `EmailHtml`, `ServiceHooks`).",
			},
			"channel_address": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Email address for EmailHtml channel. Leave empty for other channel types.",
				PlanModifiers: []planmodifier.String{
					notifUseStateForUnknown(),
				},
			},
			"filter_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Filter type identifier. Defaults to `Expression`.",
				PlanModifiers: []planmodifier.String{
					notifUseStateForUnknown(),
				},
			},
			"filter_criteria": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Serialised expression criteria (JSON string).",
				PlanModifiers: []planmodifier.String{
					notifUseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Subscription status returned by the API.",
				PlanModifiers: []planmodifier.String{
					notifUseStateForUnknown(),
				},
			},
		},
	}
}

// ── Configure ─────────────────────────────────────────────────────────────────

func (r *notificationSubscriptionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (r *notificationSubscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddWarning("Provider not configured", "NotificationClient is nil; skipping create.")
		return
	}

	var model notificationSubscriptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createParams := expandNotificationSubscription(&model)
	created, err := r.client.NotificationClient.CreateSubscription(r.client.Ctx, notificationapi.CreateSubscriptionArgs{
		CreateParameters: createParams,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating notification subscription", err.Error())
		return
	}

	flattenNotificationSubscription(created, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *notificationSubscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if r.client == nil {
		resp.Diagnostics.AddWarning("Provider not configured", "NotificationClient is nil; skipping read.")
		return
	}

	var model notificationSubscriptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := model.ID.ValueString()
	sub, err := r.client.NotificationClient.GetSubscription(r.client.Ctx, notificationapi.GetSubscriptionArgs{
		SubscriptionId: &id,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading notification subscription", err.Error())
		return
	}

	// ADO soft-deletes notification subscriptions: after a DELETE call the
	// subscription transitions to "pendingDeletion" status before the record is
	// physically removed. Treat that as gone so Terraform removes it from state.
	if sub != nil && sub.Status != nil && *sub.Status == notificationapi.SubscriptionStatusValues.PendingDeletion {
		resp.State.RemoveResource(ctx)
		return
	}

	flattenNotificationSubscription(sub, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *notificationSubscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if r.client == nil {
		resp.Diagnostics.AddWarning("Provider not configured", "NotificationClient is nil; skipping update.")
		return
	}

	var model notificationSubscriptionModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateModel notificationSubscriptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &stateModel)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := stateModel.ID.ValueString()
	updateParams := expandNotificationSubscriptionUpdate(&model)
	updated, err := r.client.NotificationClient.UpdateSubscription(r.client.Ctx, notificationapi.UpdateSubscriptionArgs{
		UpdateParameters: updateParams,
		SubscriptionId:   &id,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating notification subscription", err.Error())
		return
	}

	flattenNotificationSubscription(updated, &model)
	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}

func (r *notificationSubscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if r.client == nil {
		resp.Diagnostics.AddWarning("Provider not configured", "NotificationClient is nil; skipping delete.")
		return
	}

	var model notificationSubscriptionModel
	resp.Diagnostics.Append(req.State.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := model.ID.ValueString()
	err := r.client.NotificationClient.DeleteSubscription(r.client.Ctx, notificationapi.DeleteSubscriptionArgs{
		SubscriptionId: &id,
	})
	if err != nil {
		if utils.ResponseWasNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting notification subscription", err.Error())
		return
	}
}

// ── expand / flatten ──────────────────────────────────────────────────────────

// expandNotificationSubscription converts Terraform state to API create parameters.
func expandNotificationSubscription(model *notificationSubscriptionModel) *notificationapi.NotificationSubscriptionCreateParameters {
	eventType := model.SubscriptionType.ValueString()
	filterType := model.FilterType.ValueString()
	if filterType == "" {
		filterType = "Expression"
	}

	filter := &notificationapi.ISubscriptionFilter{
		EventType: &eventType,
		Type:      &filterType,
	}

	channelType := model.ChannelType.ValueString()
	channel := &notificationapi.ISubscriptionChannel{
		Type: &channelType,
	}

	projectIDStr := model.ProjectID.ValueString()
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		// project_id schema is Required; if it's not a valid UUID we use the nil UUID.
		// The API will return a validation error on create, which surfaces the problem.
		projectID = uuid.Nil
	}
	scopeType := "project"
	scope := &notificationapi.SubscriptionScope{
		Id:   &projectID,
		Type: &scopeType,
	}

	subscriberIDStr := model.SubscriberID.ValueString()
	subscriber := &webapi.IdentityRef{
		Id: &subscriberIDStr,
	}

	return &notificationapi.NotificationSubscriptionCreateParameters{
		Filter:     filter,
		Channel:    channel,
		Scope:      scope,
		Subscriber: subscriber,
	}
}

// expandNotificationSubscriptionUpdate converts Terraform plan to API update parameters.
func expandNotificationSubscriptionUpdate(model *notificationSubscriptionModel) *notificationapi.NotificationSubscriptionUpdateParameters {
	eventType := model.SubscriptionType.ValueString()
	filterType := model.FilterType.ValueString()
	if filterType == "" {
		filterType = "Expression"
	}

	filter := &notificationapi.ISubscriptionFilter{
		EventType: &eventType,
		Type:      &filterType,
	}

	channelType := model.ChannelType.ValueString()
	channel := &notificationapi.ISubscriptionChannel{
		Type: &channelType,
	}

	return &notificationapi.NotificationSubscriptionUpdateParameters{
		Filter:  filter,
		Channel: channel,
	}
}

// flattenNotificationSubscription maps the API response struct to the Terraform state model.
//
// IMPORTANT: every Computed (or Optional+Computed) field MUST be set to a known
// value here, even if the API does not return it. Leaving a field as types.Unknown
// after Create/Read causes Terraform to error with "provider still indicated an
// unknown value … after apply".
//
// Fields that the ADO notification API does not return (ISubscriptionFilter has
// no Criteria field; ISubscriptionChannel has no Address field) are left at
// their current model value when that value is already known, or set to
// types.StringNull() when the model value is unknown/null. This preserves the
// config-supplied value on Create (model was populated from Plan before flatten)
// while guaranteeing a known value in state on every code path.
func flattenNotificationSubscription(sub *notificationapi.NotificationSubscription, model *notificationSubscriptionModel) {
	if sub == nil {
		return
	}

	if sub.Id != nil {
		model.ID = types.StringValue(*sub.Id)
	}

	// Channel type
	if sub.Channel != nil && sub.Channel.Type != nil {
		model.ChannelType = types.StringValue(*sub.Channel.Type)
	}

	// channel_address: the ADO ISubscriptionChannel interface only carries Type;
	// the address is not surfaced in the Go API wrapper. Keep the model value if
	// it is already known (set from Plan on Create / from State on Read), otherwise
	// ensure we write a known (null) value so Terraform does not see Unknown in state.
	if model.ChannelAddress.IsUnknown() {
		model.ChannelAddress = types.StringNull()
	}

	// Filter fields
	if sub.Filter != nil {
		if sub.Filter.Type != nil {
			model.FilterType = types.StringValue(*sub.Filter.Type)
		} else if model.FilterType.IsUnknown() {
			model.FilterType = types.StringNull()
		}
		if sub.Filter.EventType != nil {
			model.SubscriptionType = types.StringValue(*sub.Filter.EventType)
		}
	} else if model.FilterType.IsUnknown() {
		model.FilterType = types.StringNull()
	}

	// filter_criteria: ADO ISubscriptionFilter has no Criteria field in the Go
	// wrapper; keep the model value if known, otherwise write null so state is known.
	if model.FilterCriteria.IsUnknown() {
		model.FilterCriteria = types.StringNull()
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
	} else if model.Status.IsUnknown() {
		model.Status = types.StringNull()
	}
}
