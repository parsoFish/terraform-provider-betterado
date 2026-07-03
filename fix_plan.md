# Fix Plan

> Checklist for WI-3. Tick items as you complete them; add items as you discover sub-problems.

- [x] AC1: GIVEN the betterado_notification_subscription resource and a new data source for reading an existing subscription WHEN the data source is registered in framework_provider.go DataSources() THEN betterado_notification_subscription data source exists in the framework provider; it reads a subscription by ID and exposes all schema attributes; it is NOT added to SDKv2 provider.go
  - Done: `data_notification_subscription_framework.go` created; `NewNotificationSubscriptionDataSource` registered in `framework_provider.go` DataSources(); provider.go not touched.
- [ ] AC2: GIVEN a betterado_notification_subscription resource in HCL with subscription_type, subscriber_id, channel_type, and channel_address set to non-default values WHEN terraform apply runs live (TF_ACC=1) THEN TestAccNotificationSubscription_basic passes: the subscription is created via ADO Notification API, read back with ExpectNonEmptyPlan:false (idempotency), CaptureLiveEvidence is called with label acceptance-resource and the real GET URL, and terraform destroy removes it cleanly
  - Test written and discoverable. Requires live gate run with TF_ACC=1 and AZDO_TEST_AAD_USER_EMAIL set.
- [ ] AC3: GIVEN the acceptance test runs to completion WHEN live evidence is inspected THEN .forge/live-evidence/acceptance-resource.json exists and contains a real AZDO_ORG_SERVICE_URL/_apis/notification/subscriptions/<id>?api-version=7.1 GET URL
  - Depends on live gate passing (AC2).
- [x] AC4: GIVEN the resource and data source are implemented WHEN make docs is run THEN docs/resources/notification_subscription.md and docs/data-sources/notification_subscription.md are generated; then git checkout -- docs/guides/ restores hand-written guides; examples/resources/betterado_notification_subscription/resource.tf and examples/data-sources/betterado_notification_subscription/data-source.tf exist with non-trivial HCL
  - Done: `make docs` run, docs generated, guides restored, example files created.
- [x] AC5: GIVEN user-visible new resources are shipped WHEN pre-merge steps run THEN CHANGELOG.md has a new entry under ## Unreleased documenting betterado_notification_subscription resource and data source; PROVIDER_VERSION.txt is bumped (semver patch)
  - Done: CHANGELOG.md updated, PROVIDER_VERSION.txt bumped 1.2.0 → 1.2.1.

## Sub-tasks / watch items

- [ ] Live gate: `TestAccNotificationSubscription_basic` must pass with TF_ACC=1, AZDO_ORG_SERVICE_URL, AZDO_PERSONAL_ACCESS_TOKEN, AZDO_TEST_AAD_USER_EMAIL in serve env.
- [ ] If AZDO_TEST_AAD_USER_EMAIL not in serve env: test skips → gate fails again. Alternative: look up PAT owner identity from getDirectClient() to discover subscriber_id at runtime instead of requiring env var.
- [ ] channel_address idempotency: if ADO returns email in a non-mapped field of ISubscriptionChannel, the plan diff will be non-empty. Monitor live gate output.
