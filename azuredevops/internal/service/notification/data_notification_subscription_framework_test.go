//go:build all || data_notification_subscription

package notification

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	notificationapi "github.com/microsoft/azure-devops-go-api/azuredevops/v7/notification"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/parsoFish/terraform-provider-betterado/azdosdkmocks"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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

// notificationDataSourceSchemaType returns the tftypes.Object type that matches
// the notificationSubscriptionDataSource schema, keyed by attribute name.
func notificationDataSourceSchemaType() tftypes.Object {
	return tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                tftypes.String,
			"project_id":        tftypes.String,
			"subscription_type": tftypes.String,
			"subscriber_id":     tftypes.String,
			"channel_type":      tftypes.String,
			"channel_address":   tftypes.String,
			"filter_type":       tftypes.String,
			"filter_criteria":   tftypes.String,
			"status":            tftypes.String,
		},
	}
}

// notificationDataSourceConfigRaw builds a tftypes.Value representing the config
// passed to the datasource Read() method, with only the subscription ID populated.
func notificationDataSourceConfigRaw(subID string) tftypes.Value {
	objType := notificationDataSourceSchemaType()
	return tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                tftypes.NewValue(tftypes.String, subID),
		"project_id":        tftypes.NewValue(tftypes.String, nil),
		"subscription_type": tftypes.NewValue(tftypes.String, nil),
		"subscriber_id":     tftypes.NewValue(tftypes.String, nil),
		"channel_type":      tftypes.NewValue(tftypes.String, nil),
		"channel_address":   tftypes.NewValue(tftypes.String, nil),
		"filter_type":       tftypes.NewValue(tftypes.String, nil),
		"filter_criteria":   tftypes.NewValue(tftypes.String, nil),
		"status":            tftypes.NewValue(tftypes.String, nil),
	})
}

// TestDataSource_404NotFound verifies that when the ADO API returns HTTP 404
// for a GetSubscription call, the datasource Read() actually invokes
// resp.State.RemoveResource(ctx), leaving the state Raw as a typed null.
//
// This drives the datasource Read() method directly (not just the mock client),
// exercising data_notification_subscription_framework.go lines 128-131 — the
// RemoveResource branch demanded by UWI-2 AC2 / UWI-4 AC2.
func TestDataSource_404NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	subID := "sub-missing-001"
	statusCode := http.StatusNotFound
	notFoundErr := azuredevops.WrappedError{
		StatusCode: &statusCode,
	}

	// Confirm that utils.ResponseWasNotFound detects the 404 correctly.
	assert.True(t, utils.ResponseWasNotFound(notFoundErr),
		"utils.ResponseWasNotFound must return true for HTTP 404")

	// Wire mock to return 404 for GetSubscription.
	mockClient := azdosdkmocks.NewMockNotificationClient(ctrl)
	mockClient.EXPECT().
		GetSubscription(gomock.Any(), notificationapi.GetSubscriptionArgs{
			SubscriptionId: &subID,
		}).
		Return(nil, notFoundErr).
		Times(1)

	// Build the datasource instance and inject the mock client via AggregatedClient.
	ds := &notificationSubscriptionDataSource{
		client: &client.AggregatedClient{
			NotificationClient: mockClient,
			Ctx:                ctx,
		},
	}

	// Obtain the framework schema from the datasource itself.
	var schemaResp datasource.SchemaResponse
	ds.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
	fwSchema := schemaResp.Schema

	// Build the config raw value (subscription ID required, rest null).
	rawVal := notificationDataSourceConfigRaw(subID)

	readReq := datasource.ReadRequest{
		Config: tfsdk.Config{
			Raw:    rawVal,
			Schema: fwSchema,
		},
	}
	// Pre-populate state with the same raw value so RemoveResource has something to nil.
	readResp := datasource.ReadResponse{
		State: tfsdk.State{
			Raw:    rawVal.Copy(),
			Schema: fwSchema,
		},
	}

	// Drive Read() directly — this exercises the 404 branch in Read():
	//   if utils.ResponseWasNotFound(err) { resp.State.RemoveResource(ctx); return }
	ds.Read(ctx, readReq, &readResp)

	// RemoveResource sets State.Raw to a typed null; no diagnostic error must be present.
	require.False(t, readResp.Diagnostics.HasError(),
		"Read must not return an error on 404; it should gracefully remove the resource")
	assert.True(t, readResp.State.Raw.IsNull(),
		"RemoveResource must set State.Raw to a typed null — "+
			"Terraform uses this signal to remove the data source from state")
}
