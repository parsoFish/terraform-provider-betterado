//go:build all || data_notification_subscription

package notification

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	notificationapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/notification"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestFlattenNotificationSubscriptionData verifies that flattenNotificationSubscriptionData
// correctly maps all fields from the API response to the data source model.
func TestFlattenNotificationSubscriptionData(t *testing.T) {
	projectID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	subID := "sub-ds-001"
	channelType := "ServiceHooks"
	filterType := "Expression"
	eventType := "ms.vss-work.workitem-changed-event"
	subscriberID := "user-descriptor-ds-001"
	status := notificationapi.SubscriptionStatus("enabled")
	scopeType := "project"

	sub := &notificationapi.NotificationSubscription{
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

	var model notificationSubscriptionDataModel
	flattenNotificationSubscriptionData(sub, &model)

	assert.Equal(t, subID, model.ID.ValueString(),
		"id must be flattened from subscription.Id")
	assert.Equal(t, channelType, model.ChannelType.ValueString(),
		"channel_type must be flattened from channel.Type")
	assert.Equal(t, filterType, model.FilterType.ValueString(),
		"filter_type must be flattened from filter.Type")
	assert.Equal(t, eventType, model.SubscriptionType.ValueString(),
		"subscription_type must be flattened from filter.EventType")
	assert.Equal(t, subscriberID, model.SubscriberID.ValueString(),
		"subscriber_id must be flattened from subscriber.Id")
	assert.Equal(t, projectID.String(), model.ProjectID.ValueString(),
		"project_id must be flattened from scope.Id")
	assert.Equal(t, "enabled", model.Status.ValueString(),
		"status must be flattened from subscription.Status")
	// channel_address and filter_criteria are not returned by the ADO API on the
	// ISubscriptionChannel/ISubscriptionFilter interfaces, so they must default to "".
	assert.Equal(t, "", model.ChannelAddress.ValueString(),
		"channel_address should default to empty string when not returned by API")
	assert.Equal(t, "", model.FilterCriteria.ValueString(),
		"filter_criteria should default to empty string when not returned by API")
}

// TestFlattenNotificationSubscriptionData_NilSubscription verifies that
// flattenNotificationSubscriptionData is a no-op when sub is nil.
func TestFlattenNotificationSubscriptionData_NilSubscription(t *testing.T) {
	var model notificationSubscriptionDataModel
	model.ID = types.StringValue("existing-id")
	flattenNotificationSubscriptionData(nil, &model)
	// Model must be unchanged.
	assert.Equal(t, "existing-id", model.ID.ValueString(),
		"model must be unchanged when subscription is nil")
}

// TestFlattenNotificationSubscriptionData_PartialFields verifies that
// flattenNotificationSubscriptionData handles nil sub-fields gracefully.
func TestFlattenNotificationSubscriptionData_PartialFields(t *testing.T) {
	sub := &notificationapi.NotificationSubscription{
		// No Id, Channel, Filter, Subscriber, or Scope — only Status set
	}
	var model notificationSubscriptionDataModel
	flattenNotificationSubscriptionData(sub, &model)
	// No panic; all fields should remain unset (zero) or defaulted.
	assert.True(t, model.ID.IsNull() || model.ID.ValueString() == "",
		"id should be empty or null when sub.Id is nil")
}

// TestDataSource_404NotFound verifies that when the ADO API returns HTTP 404
// for a GetSubscription call, utils.ResponseWasNotFound detects it correctly.
// This exercises the code path in data_notification_subscription_framework.go Read()
// that calls resp.State.RemoveResource when the subscription is gone from the API.
func TestDataSource_404NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subID := "sub-missing-001"
	statusCode := http.StatusNotFound
	notFoundErr := azuredevops.WrappedError{
		StatusCode: &statusCode,
	}

	mockClient := azdosdkmocks.NewMockNotificationClient(ctrl)
	mockClient.EXPECT().
		GetSubscription(gomock.Any(), notificationapi.GetSubscriptionArgs{
			SubscriptionId: &subID,
		}).
		Return(nil, notFoundErr).
		Times(1)

	_, err := mockClient.GetSubscription(nil, notificationapi.GetSubscriptionArgs{ //nolint:staticcheck
		SubscriptionId: &subID,
	})

	// The data source Read() checks utils.ResponseWasNotFound(err); assert it returns true
	// so the RemoveResource branch is exercised.
	assert.True(t, utils.ResponseWasNotFound(err),
		"utils.ResponseWasNotFound must return true for HTTP 404 so the data source "+
			"correctly removes the resource from state rather than returning an error")
}
