# betterado_notification_subscription resource and data source — ADO Notification API v7.1

> _Derived from `demo.json` (ADR 021). Essence:_ Before this initiative, the betterado Terraform provider had no support for ADO notification subscriptions. This initiative adds betterado_notification_subscription as a framework-native resource and companion data source, backed by the ADO Notification API v7.1. Users can now manage notification subscriptions (event type, delivery channel, expression filter, subscriber identity) entirely in Terraform. All CRUD operations are exercised by a live acceptance test (TestAccNotificationSubscription_basic) that applies, reads back, checks idempotency, and destroys a real subscription against the davidgparsonson ADO org.

## Intent & Outcome

> _Assessed intent:_ Before this initiative, the betterado Terraform provider had no support for ADO notification subscriptions. This initiative adds betterado_notification_subscription as a framework-native resource and companion data source, backed by the ADO Notification API v7.1. Users can now manage notification subscriptions (event type, delivery channel, expression filter, subscriber identity) entirely in Terraform. All CRUD operations are exercised by a live acceptance test (TestAccNotificationSubscription_basic) that applies, reads back, checks idempotency, and destroys a real subscription against the davidgparsonson ADO org.

| # | Acceptance criterion | Verdict | Evidence |
|---|---|---|---|
| 1 | GIVEN the ADO Notifications REST API v7.1 (subscriptions + subscription templates endpoints) WHEN the gap matrix is constructed by inspecting the vendored SDK models and client THEN docs/notification-gap-matrix.md exists and lists every API resource type, every field in NotificationSubscription, and its status (implement / read-only / out-of-scope) with explicit rationale for each deferral | ✓ met | docs/notification-gap-matrix.md committed (335 lines, 9bdde69d). Contains full field table for NotificationSubscription with implement/read-only/out-of-scope status and rationale columns. Present in git diff --name-only main...HEAD. |
| 2 | GIVEN the gap matrix describes NotificationSubscriptionTemplate WHEN the matrix is written THEN betterado_notification_subscription_template is explicitly triaged as a read-only data source (out-of-scope for this initiative) with rationale | ✓ met | docs/notification-gap-matrix.md §Deferred Resources section explicitly triages betterado_notification_subscription_template as out-of-scope with rationale: read-only data source, deferred to a follow-on initiative. |
| 3 | GIVEN AggregatedClient does not yet have a NotificationClient field WHEN the notification package is wired THEN azuredevops/internal/client/client.go gains a NotificationClient field of type notification.Client and GetAzdoClient initialises it; the build compiles | ✓ met | azuredevops/internal/client/client.go diff shows +5 lines adding NotificationClient notification.Client field and initialisation in GetAzdoClient. go build ./... succeeds (gate green). |
| 4 | GIVEN a new notification service package with the framework resource WHEN go test is run against the package THEN TestNotificationSubscriptionResource_Create and TestNotificationSubscriptionResource_Read unit tests pass, verifying expand/flatten round-trips for subscription_type, channel_type, channel_address, filter_type, filter_criteria, and subscriber_id fields | ✓ met | go test -tags all -count=1 -run TestNotificationSubscriptionResource_ ./azuredevops/internal/service/notification/ → PASS (both TestNotificationSubscriptionResource_Create and TestNotificationSubscriptionResource_Read green). The channel_type attribute is validated with stringvalidator.OneOf (resource_notification_subscription_framework.go:139) enforcing the exact ADO API channel type enum values (Block, EmailHtml, EmailPlaintext, Group, MessageQueue, ServiceBus, ServiceHooks, Soap, Unsupported, User, UserSystem); this framework-validators fix was added in UWI-2 AC1. |
| 5 | GIVEN the betterado_notification_subscription framework resource implementation WHEN the resource is reviewed THEN it is registered only in framework_provider.go Resources() and NOT added to the SDKv2 provider.go ResourcesMap; Create/Read/Update/Delete methods use NotificationClient.CreateSubscription, GetSubscription, UpdateSubscription, DeleteSubscription; Read returns nil on 404 (external delete); schema uses TypeList+MaxItems:1 for filter block | ✓ met | grep -n notification_subscription azuredevops/provider.go returns no output — zero SDKv2 registrations. framework_provider.go diff shows notification.NewNotificationSubscriptionResource added to Resources(). resource file implements all 4 CRUD methods using NotificationClient; Read calls resp.State.RemoveResource(ctx) on 404. |
| 6 | GIVEN the betterado_notification_subscription resource and a new data source for reading an existing subscription WHEN the data source is registered in framework_provider.go DataSources() THEN betterado_notification_subscription data source exists in the framework provider; it reads a subscription by ID and exposes all schema attributes; it is NOT added to SDKv2 provider.go | ✓ met | data_notification_subscription_framework.go committed (190 lines); framework_provider.go diff shows notification.NewNotificationSubscriptionDataSource added to DataSources(). grep -n notification_subscription azuredevops/provider.go → no output. Unit tests TestFlattenNotificationSubscriptionData (data_notification_subscription_framework_test.go) verify all flatten fields (id, channel_type, filter_type, subscription_type, subscriber_id, project_id, status, channel_address, filter_criteria). TestDataSource_404NotFound (data_notification_subscription_framework_test.go) drives datasource Read() directly against a 404 mock, asserts utils.ResponseWasNotFound returns true, and confirms resp.State.RemoveResource(ctx) is called — the RemoveResource branch at data_notification_subscription_framework.go:131 is exercised by test code. |
| 7 | GIVEN a betterado_notification_subscription resource in HCL with subscription_type, subscriber_id, channel_type, and channel_address set to non-default values WHEN terraform apply runs live (TF_ACC=1) THEN TestAccNotificationSubscription_basic passes: the subscription is created via ADO Notification API, read back with ExpectNonEmptyPlan:false (idempotency), CaptureLiveEvidence is called with label acceptance-resource and the real GET URL, and terraform destroy removes it cleanly | ✓ met | TestAccNotificationSubscription_basic → PASS. Live evidence: .forge/live-evidence/acceptance-resource.json capturedAt 2026-07-03T08:02:23Z, subscription ID 886548, GET URL https://dev.azure.com/davidgparsonson/_apis/notification/subscriptions/886548?api-version=7.1, status=enabled, channel.type=EmailHtml. |
| 8 | GIVEN the acceptance test runs to completion WHEN live evidence is inspected THEN .forge/live-evidence/acceptance-resource.json exists and contains a real AZDO_ORG_SERVICE_URL/_apis/notification/subscriptions/<id>?api-version=7.1 GET URL | ✓ met | .forge/live-evidence/acceptance-resource.json exists: url=https://dev.azure.com/davidgparsonson/_apis/notification/subscriptions/886548?api-version=7.1, response contains id=886548, status=enabled, subscriber=abby@kel.so, capturedAt=2026-07-03T08:02:23Z. |
| 9 | GIVEN the resource and data source are implemented WHEN make docs is run THEN docs/resources/notification_subscription.md and docs/data-sources/notification_subscription.md are generated; then git checkout -- docs/guides/ restores hand-written guides; examples/resources/betterado_notification_subscription/resource.tf and examples/data-sources/betterado_notification_subscription/data-source.tf exist with non-trivial HCL | ✓ met | docs/resources/notification_subscription.md (72 lines) and docs/data-sources/notification_subscription.md (51 lines) present in branch diff. examples/resources/betterado_notification_subscription/resource.tf (22 lines) and examples/data-sources/betterado_notification_subscription/data-source.tf (17 lines) committed with full HCL showing all attributes. |
| 10 | GIVEN user-visible new resources are shipped WHEN pre-merge steps run THEN CHANGELOG.md has a new entry under ## Unreleased documenting betterado_notification_subscription resource and data source; PROVIDER_VERSION.txt is bumped (semver patch) | ✓ met | CHANGELOG.md ## [Unreleased] ## FEATURES section added: betterado_notification_subscription resource and data source documented. PROVIDER_VERSION.txt bumped to 1.2.1 (was 1.2.0). |

## Visual Changes

### Quality gate (forge-mandated): servicehook package still green after notification wiring

- **Before:** servicehook tests passed on main before this initiative
- **After:** ok github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/servicehook 0.015s

### Notification package unit tests: resource expand/flatten, datasource flatten, and 404 RemoveResource branch

- **Before:** Package did not exist on main
- **After:** TestNotificationSubscriptionResource_Create, TestNotificationSubscriptionResource_Read, TestFlattenNotificationSubscriptionData, TestFlattenNotificationSubscriptionData_NilSubscription, TestFlattenNotificationSubscriptionData_PartialFields, and TestDataSource_404NotFound all pass

### docs/notification-gap-matrix.md catalogues every NotificationSubscription field with implement/read-only/out-of-scope status

- **Before:** File did not exist on main
- **After:** 335-line gap matrix documents every API field, rationale for deferrals, and subscription_template triage

### betterado_notification_subscription registered only in framework_provider.go — zero SDKv2 registrations

- **Before:** provider.go had no notification entries on main
- **After:** grep returns no matches — confirmed framework-only registration

### Live acceptance test creates subscription via ADO API, reads back, confirms idempotency, destroys

- **Before:** No acceptance test existed on main
- **After:** TestAccNotificationSubscription_basic passes: subscription 886548 created, GET confirmed, destroy clean
- **Live evidence (real API GET):** `https://dev.azure.com/davidgparsonson/_apis/notification/subscriptions/886548?api-version=7.1` _(captured 2026-07-03T08:02:23Z)_

```json
{
  "_links": {
    "edit": {
      "href": "https://dev.azure.com/davidgparsonson/_notifications?subscriptionId=886548&publisherId=ms.vss-work.work-event-publisher&action=view"
    }
  },
  "diagnostics": {},
  "filter": {
    "eventType": "ms.vss-work.workitem-changed-event",
    "type": "Expression"
  },
  "flags": "none",
  "channel": {
    "type": "EmailHtml"
  },
  "id": "886548",
  "lastModifiedBy": {
    "displayName": "david.g.parsonson",
    "id": "49e26c2f-ec33-6e72-b494-dedb0aee09e1",
    "uniqueName": "david.g.parsonson@gmail.com"
  },
  "modifiedDate": "2026-07-03T08:02:20.803Z",
  "permissions": "view, edit, delete",
  "scope": {
    "id": "6ddb680c-093d-4953-9561-2266eb7af800",
    "type": "none"
  },
  "status": "enabled",
  "subscriber": {
    "displayName": "Ambrose Kelso",
    "id": "0532ab9e-1c40-4efd-9906-b50d10fe13c2",
    "uniqueName": "abby@kel.so"
  },
  "url": "https://dev.azure.com/davidgparsonson/_apis/notification/Subscriptions/886548"
}
```

## Test Evidence

| test | result | delta |
|---|---|---|
| go test ./azuredevops/internal/service/release/... | pass | — |
| go test ./azuredevops/internal/service/taskagent/... | pass | — |
| go test ./azuredevops/internal/service/taskagent/validate/... | pass | — |
| TestNotificationSubscriptionResource_Create | pass | — |
| TestNotificationSubscriptionResource_Read | pass | — |
| TestFlattenNotificationSubscriptionData | pass | — |
| TestFlattenNotificationSubscriptionData_NilSubscription | pass | — |
| TestFlattenNotificationSubscriptionData_PartialFields | pass | — |
| TestDataSource_404NotFound | pass | — |
| TestAccNotificationSubscription_basic | pass | — |

> result: **pass**/**fail** · **skip** = not run in this gate (e.g. a live test with no credentials present) — not a failure · delta **new** = test added by this change.

## Files Changed

```
94 files changed, 5912 insertions(+), 256 deletions(-)
```
