# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the betterado_notification_subscription resource and a new data source for reading an existing subscription WHEN the data source is registered in framework_provider.go DataSources() THEN betterado_notification_subscription data source exists in the framework provider; it reads a subscription by ID and exposes all schema attributes; it is NOT added to SDKv2 provider.go
  - NewNotificationSubscriptionResource at line 210, NewNotificationSubscriptionDataSource at line 216 in framework_provider.go
  - NOT in azuredevops/provider.go (SDKv2)
  - Provider registration tests added: TestFrameworkProvider_HasNotificationSubscriptionResource and TestFrameworkProvider_HasNotificationSubscriptionDataSource

- [x] AC2: GIVEN a betterado_notification_subscription resource in HCL with subscription_type, subscriber_id, channel_type, and channel_address set to non-default values WHEN terraform apply runs live (TF_ACC=1) THEN TestAccNotificationSubscription_basic passes: the subscription is created via ADO Notification API, read back with ExpectNonEmptyPlan:false (idempotency), CaptureLiveEvidence is called with label acceptance-resource and the real GET URL, and terraform destroy removes it cleanly
  - Test passes (confirmed from prior iteration live run — live-evidence captured at acceptance-resource.json)
  - Auto-discovers subscriber via GraphClient.ListUsers if AZDO_TEST_AAD_USER_EMAIL not set
  - CheckDestroy treats pendingDeletion status as destroyed (ADO soft-delete)

- [x] AC3: GIVEN the acceptance test runs to completion WHEN live evidence is inspected THEN .forge/live-evidence/acceptance-resource.json exists and contains a real AZDO_ORG_SERVICE_URL/_apis/notification/subscriptions/<id>?api-version=7.1 GET URL
  - .forge/live-evidence/acceptance-resource.json exists
  - URL: https://dev.azure.com/davidgparsonson/_apis/notification/subscriptions/886547?api-version=7.1

- [x] AC4: GIVEN the resource and data source are implemented WHEN make docs is run THEN docs/resources/notification_subscription.md and docs/data-sources/notification_subscription.md are generated; then git checkout -- docs/guides/ restores hand-written guides; examples/resources/betterado_notification_subscription/resource.tf and examples/data-sources/betterado_notification_subscription/data-source.tf exist with non-trivial HCL
  - docs/resources/notification_subscription.md present
  - docs/data-sources/notification_subscription.md present
  - examples/resources/betterado_notification_subscription/resource.tf present (non-trivial HCL with project, identity data sources and optional filter_type)
  - examples/data-sources/betterado_notification_subscription/data-source.tf present (reads by ID, outputs 3 attributes)

- [x] AC5: GIVEN user-visible new resources are shipped WHEN pre-merge steps run THEN CHANGELOG.md has a new entry under ## Unreleased documenting betterado_notification_subscription resource and data source; PROVIDER_VERSION.txt is bumped (semver patch)
  - CHANGELOG.md has entry under ## [Unreleased] describing resource + data source
  - PROVIDER_VERSION.txt: 1.2.1 (bumped from 1.2.0)

## CI Gates (all green — iteration 4)
- make test: PASS (no failures, all relevant tests pass)
- golangci-lint run --new-from-rev=main ./azuredevops/...: 0 issues
- make terrafmt-check: PASS
