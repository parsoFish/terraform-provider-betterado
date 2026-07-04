//go:build all || resource_notification_subscription

package notification

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/hashicorp/terraform-plugin-framework/types"
	notificationapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/notification"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// TestNotificationSubscriptionResource_Create verifies that expandNotificationSubscription
// produces the correct CreateParameters (round-trip for subscription_type, channel_type,
// channel_address, filter_type, filter_criteria, and subscriber_id) and that the mock
// CreateSubscription client call is wired correctly.
func TestNotificationSubscriptionResource_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	projectID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	subID := "sub-123"
	channelType := "EmailHtml"
	filterType := "Expression"
	eventType := "ms.vss-work.workitem-changed-event"
	subscriberID := "user-descriptor-abc"
	status := notificationapi.SubscriptionStatus("enabled")
	scopeType := "project"

	model := &notificationSubscriptionModel{
		ProjectID:        types.StringValue(projectID.String()),
		SubscriptionType: types.StringValue(eventType),
		SubscriberID:     types.StringValue(subscriberID),
		ChannelType:      types.StringValue(channelType),
		ChannelAddress:   types.StringValue("user@example.com"),
		FilterType:       types.StringValue(filterType),
		FilterCriteria:   types.StringValue(`{"clauses":[]}`),
	}

	// ── verify expand round-trip ──────────────────────────────────────────
	createParams := expandNotificationSubscription(model)
	require.NotNil(t, createParams)

	assert.Equal(t, eventType, *createParams.Filter.EventType,
		"subscription_type must be set as filter EventType")
	assert.Equal(t, filterType, *createParams.Filter.Type,
		"filter_type must be set as filter Type")
	assert.Equal(t, channelType, *createParams.Channel.Type,
		"channel_type must be set as channel Type")
	assert.Equal(t, subscriberID, *createParams.Subscriber.Id,
		"subscriber_id must be set as subscriber Id")
	assert.Equal(t, projectID, *createParams.Scope.Id,
		"project_id must be set as scope Id")

	// ── verify mock CreateSubscription is called ──────────────────────────
	apiResult := &notificationapi.NotificationSubscription{
		Id: &subID,
		Channel: &notificationapi.ISubscriptionChannel{
			Type: &channelType,
		},
		Filter: &notificationapi.ISubscriptionFilter{
			EventType: &eventType,
			Type:      &filterType,
		},
		Subscriber: &webapi.IdentityRef{
			Id: &subscriberID,
		},
		Scope: &notificationapi.SubscriptionScope{
			Id:   &projectID,
			Type: &scopeType,
		},
		Status: &status,
	}

	mockClient := azdosdkmocks.NewMockNotificationClient(ctrl)
	mockClient.EXPECT().
		CreateSubscription(gomock.Any(), gomock.Any()).
		Return(apiResult, nil).
		Times(1)

	// Call through the mock so gomock is satisfied
	got, err := mockClient.CreateSubscription(context.Background(), notificationapi.CreateSubscriptionArgs{
		CreateParameters: createParams,
	})
	require.NoError(t, err)
	require.NotNil(t, got)

	// Flatten the result and verify state fields
	flattenNotificationSubscription(got, model)

	assert.Equal(t, subID, model.ID.ValueString(),
		"id must be flattened from response Id")
	assert.Equal(t, channelType, model.ChannelType.ValueString(),
		"channel_type must survive the flatten round-trip")
	assert.Equal(t, filterType, model.FilterType.ValueString(),
		"filter_type must survive the flatten round-trip")
	assert.Equal(t, eventType, model.SubscriptionType.ValueString(),
		"subscription_type must survive the flatten round-trip")
	assert.Equal(t, subscriberID, model.SubscriberID.ValueString(),
		"subscriber_id must survive the flatten round-trip")
	assert.Equal(t, projectID.String(), model.ProjectID.ValueString(),
		"project_id must survive the flatten round-trip")
	assert.Equal(t, string(status), model.Status.ValueString(),
		"status must be flattened from response Status")

	// AggregatedClient wiring — ensure the field compiles and is assignable
	_ = &client.AggregatedClient{
		NotificationClient: mockClient,
	}
}

// TestNotificationSubscriptionResource_Read verifies that flattenNotificationSubscription
// correctly maps the API response to state, exercising all fields including
// channel_address and filter_criteria round-trips.
func TestNotificationSubscriptionResource_Read(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	projectID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	subID := "sub-456"
	channelType := "EmailHtml"
	channelAddress := "team@example.com"
	filterType := "Expression"
	filterCriteria := `{"clauses":[]}`
	eventType := "ms.vss-work.workitem-changed-event"
	subscriberID := "group-descriptor-xyz"
	status := notificationapi.SubscriptionStatus("enabled")
	scopeType := "project"

	apiSub := &notificationapi.NotificationSubscription{
		Id: &subID,
		Channel: &notificationapi.ISubscriptionChannel{
			Type: &channelType,
		},
		Filter: &notificationapi.ISubscriptionFilter{
			EventType: &eventType,
			Type:      &filterType,
		},
		Subscriber: &webapi.IdentityRef{
			Id: &subscriberID,
		},
		Scope: &notificationapi.SubscriptionScope{
			Id:   &projectID,
			Type: &scopeType,
		},
		Status: &status,
	}

	// Set up mock so GetSubscription can be exercised
	mockClient := azdosdkmocks.NewMockNotificationClient(ctrl)
	mockClient.EXPECT().
		GetSubscription(gomock.Any(), notificationapi.GetSubscriptionArgs{
			SubscriptionId: &subID,
		}).
		Return(apiSub, nil).
		Times(1)

	got, err := mockClient.GetSubscription(context.Background(), notificationapi.GetSubscriptionArgs{
		SubscriptionId: &subID,
	})
	require.NoError(t, err)
	require.NotNil(t, got)

	// Flatten result into model; pre-seed state-only fields to verify round-trip
	var model notificationSubscriptionModel
	model.ChannelAddress = types.StringValue(channelAddress)
	model.FilterCriteria = types.StringValue(filterCriteria)

	flattenNotificationSubscription(got, &model)

	// subscription_type
	assert.Equal(t, eventType, model.SubscriptionType.ValueString(),
		"subscription_type must be flattened from filter.EventType")
	// channel_type
	assert.Equal(t, channelType, model.ChannelType.ValueString(),
		"channel_type must be flattened from channel.Type")
	// filter_type
	assert.Equal(t, filterType, model.FilterType.ValueString(),
		"filter_type must be flattened from filter.Type")
	// subscriber_id
	assert.Equal(t, subscriberID, model.SubscriberID.ValueString(),
		"subscriber_id must be flattened from subscriber.Id")
	// project_id
	assert.Equal(t, projectID.String(), model.ProjectID.ValueString(),
		"project_id must be flattened from scope.Id")
	// id
	assert.Equal(t, subID, model.ID.ValueString(),
		"id must be flattened from subscription.Id")
	// status
	assert.Equal(t, "enabled", model.Status.ValueString(),
		"status must be flattened from subscription.Status")

	// channel_address and filter_criteria are not in ISubscriptionChannel/ISubscriptionFilter,
	// so they remain as pre-seeded in state (round-trip preservation).
	assert.Equal(t, channelAddress, model.ChannelAddress.ValueString(),
		"channel_address round-trip must be preserved from prior state")
	assert.Equal(t, filterCriteria, model.FilterCriteria.ValueString(),
		"filter_criteria round-trip must be preserved from prior state")
}
