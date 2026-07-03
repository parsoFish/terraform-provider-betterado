# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

No forge brain queries made — WI spec was the single source of truth.

## What I've tried

### Iteration 0 (current)

**Prior commits when iteration started:**
- `feat(notification): add betterado_notification_subscription framework resource` — full CRUD resource already existed at `azuredevops/internal/service/notification/resource_notification_subscription_framework.go`, unit tests at `resource_notification_subscription_framework_test.go`, and resource registered in `framework_provider.go` Resources().

**Gate failure:** `[no tests to run]` — `TestAccNotificationSubscription_basic` didn't exist.

**What was built:**
1. `data_notification_subscription_framework.go` — framework datasource `NewNotificationSubscriptionDataSource`, reads by ID, all attributes Computed.
2. `framework_provider.go` — registered `notification.NewNotificationSubscriptionDataSource` in `DataSources()`.
3. `azuredevops/internal/acceptancetests/resource_notification_subscription_test.go` — `TestAccNotificationSubscription_basic` with `ProtoV6ProviderFactories`, requires `AZDO_TEST_AAD_USER_EMAIL`, uses `betterado_identity_user` data source for subscriber_id, idempotency PlanOnly step, `captureNotificationSubscriptionEvidence`, `checkNotificationSubscriptionDestroyed`.
4. Examples: `examples/resources/betterado_notification_subscription/resource.tf`, `examples/data-sources/betterado_notification_subscription/data-source.tf`.
5. Templates: `templates/resources/notification_subscription.md.tmpl`, `templates/data-sources/notification_subscription.md.tmpl`.
6. Docs generated via `make docs`: `docs/resources/notification_subscription.md`, `docs/data-sources/notification_subscription.md`.
7. `CHANGELOG.md` — `## [Unreleased]` entry added.
8. `PROVIDER_VERSION.txt` — bumped `1.2.0` → `1.2.1`.

**Build status:** `go build -tags all ./...` clean, notification unit tests pass.
**Test discoverable:** `TestAccNotificationSubscription_basic` confirmed found by `go test -list`.

## What worked

- The resource was already committed from a prior WI-2 iteration — only the data source and acceptance test were missing.
- `getDirectClient()` reuse from `resource_task_group_test.go` works for the notification client too since `AggregatedClient.NotificationClient` is wired in `client.go`.
- `betterado_identity_user` data source is the right HCL approach for subscriber_id resolution from an email.

## What didn't work

_(nothing in this iteration)_

## Open questions

- **AZDO_TEST_AAD_USER_EMAIL in serve env**: the acceptance test requires this env var. If not set in forge's serve env, the test will skip (`t.Skip`) rather than fail, which may still count as `[no tests to run]`. If the next gate failure is still a skip, consider: (a) inlining a hardcoded known test user ID, or (b) using `getDirectClient()` in `TestMain` or a helper to discover the PAT owner's identity and store it in an env var.
- **channel_address round-trip**: the ADO `ISubscriptionChannel` struct may not return `address` on GET (it's set as email on the `EmailHtml` channel). The resource uses `notifUseStateForUnknown` plan modifier, so state preserves it — idempotency should work. But if ADO returns it in a nested struct we haven't mapped, the plan diff could re-appear. Watch the live gate output.

## Notes for reflection

- The `channel_address` field in `ISubscriptionChannel` may need to be fetched from a sub-field not currently mapped in `flattenNotificationSubscription`. If `ExpectNonEmptyPlan: false` fails live, check the actual JSON returned by ADO for the EmailHtml channel and update the flatten function.
