# Fix Plan

> Checklist for WI-2. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN AggregatedClient does not yet have a NotificationClient field WHEN the notification package is wired THEN azuredevops/internal/client/client.go gains a NotificationClient field of type notification.Client and GetAzdoClient initialises it; the build compiles
- [x] AC2: GIVEN a new notification service package with the framework resource WHEN go test is run against the package THEN TestNotificationSubscriptionResource_Create and TestNotificationSubscriptionResource_Read unit tests pass, verifying expand/flatten round-trips for subscription_type, channel_type, channel_address, filter_type, filter_criteria, and subscriber_id fields
- [x] AC3: GIVEN the betterado_notification_subscription framework resource implementation WHEN the resource is reviewed THEN it is registered only in framework_provider.go Resources() and NOT added to the SDKv2 provider.go ResourcesMap; Create/Read/Update/Delete methods use NotificationClient.CreateSubscription, GetSubscription, UpdateSubscription, DeleteSubscription; Read returns nil on 404 (external delete); schema uses TypeList+MaxItems:1 for filter block
