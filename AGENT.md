# Agent Memory — WI-3

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

No forge brain queries made — WI spec was the single source of truth.

## What I've tried

### Iteration 1 (current)

**Gate failure from iter 0:** `AZDO_TEST_AAD_USER_EMAIL must be set` — the forge gate environment does NOT inject this env var.

**Root causes fixed:**
1. `TestAccNotificationSubscription_basic` called `testutils.PreCheck(t, &[]string{"AZDO_TEST_AAD_USER_EMAIL"})` which fatalf'd because the env var was missing.
2. The HCL template used `mail = email` for `betterado_identity_user` but the actual schema uses `name` + `search_filter = "MailAddress"`.

**What was changed (iter 1):**
- Rewrote `resource_notification_subscription_test.go`:
  - Removed `AZDO_TEST_AAD_USER_EMAIL` from `PreCheck`
  - Added `resolveNotificationSubscriberDirect(t)` — uses `GraphClient.ListUsers` (subjectTypes=aad,msa) to page through users, calls `resolveIdentityUUIDByEmailDirect` (IdentityClient.ReadIdentities with SearchFilter=MailAddress) to get UUID
  - `hclNotificationSubscriptionBasic` now takes `(email, subscriberID string)` and embeds the UUID directly as a literal string in HCL — no `betterado_identity_user` data source in HCL
  - Falls back to `AZDO_TEST_AAD_USER_EMAIL` if set
- `go build -tags all ./...` clean, `go test -list TestAccNotificationSubscription` finds the test

**Key ADO API facts (for next iteration):**
- `GraphClient.ListUsers` args: `ContinuationToken *string`; response `PagedGraphUsers.ContinuationToken *[]string` (extract `(*token)[0]`)
- `IdentityClient.ReadIdentities` with `SearchFilter:"MailAddress"` + `FilterValue:email` returns `[]identity.Identity` with `.Id` being the GUID

### Iteration 0

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

- `getDirectClient()` reuse from `resource_task_group_test.go` works for notification client.
- `GraphClient.ListUsers` + `IdentityClient.ReadIdentities(MailAddress)` is the right pattern for auto-discovering a real user without env var dependency.
- Embedding subscriber_id as a literal UUID in HCL avoids data source dependency and schema confusion.

## What didn't work / watch out for

- `betterado_identity_user` data source uses `name` + `search_filter` — NOT `mail`. The original HCL was wrong.
- AZDO_TEST_AAD_USER_EMAIL is NOT injected by the forge gate — tests that require it will fatalf, not skip.
- `PagedGraphUsers.ContinuationToken` is `*[]string`, not `*string` — use `(*page.ContinuationToken)[0]`.
- `ListUsersArgs.ContinuationToken` is `*string`.

### Iteration 2 (current)

**Gate failure from iter 1:** `provider still indicated an unknown value for betterado_notification_subscription.test.filter_criteria. All values must be known after apply`

**Root cause:** `flattenNotificationSubscription` never set `FilterCriteria` or `ChannelAddress` when the API didn't return them. Since these are `Optional+Computed`, their plan value was `Unknown` on the first Create (no prior state). After Create, the flatten left them as Unknown → Terraform errors with "unknown value after apply".

**ADO API facts confirmed:**
- `ISubscriptionChannel` Go struct has only `Type *string` — NO `Address` field. The `EmailHtmlSubscriptionChannel` struct has `Address` but the response is unmarshalled into `ISubscriptionChannel`.
- `ISubscriptionFilter` Go struct has only `EventType *string` + `Type *string` — NO `Criteria` field.

**Fix applied (iter 2):**
- `flattenNotificationSubscription`: after setting fields from API response, check if each `Optional+Computed` field is still `IsUnknown()` and if so, set it to `types.StringNull()`. This guarantees every field has a known value in state after every Create/Read.
- **Idempotency reasoning:**
  - On Create: model is populated from Plan before flatten, so `channel_address = "email"` is already known → not overwritten by null logic ✓
  - On Read: model is populated from State before flatten, so values from prior Create/state are already known → not overwritten ✓
  - Step 2 plan (PlanOnly): `filter_criteria = null` in state, not in config → plan = null → no diff ✓
  - Step 2 plan: `channel_address = "email"` in state, `email` in config → plan = "email" → no diff ✓

## What worked

- `getDirectClient()` reuse from `resource_task_group_test.go` works for notification client.
- `GraphClient.ListUsers` + `IdentityClient.ReadIdentities(MailAddress)` is the right pattern for auto-discovering a real user without env var dependency.
- Embedding subscriber_id as a literal UUID in HCL avoids data source dependency and schema confusion.
- **Pattern for Optional+Computed fields not returned by API**: in flatten, check `IsUnknown()` and set to `types.StringNull()` when true. Never call `types.StringValue("")` — use null to signal "API doesn't manage this value".

## What didn't work / watch out for

- `betterado_identity_user` data source uses `name` + `search_filter` — NOT `mail`. The original HCL was wrong.
- AZDO_TEST_AAD_USER_EMAIL is NOT injected by the forge gate — tests that require it will fatalf, not skip.
- `PagedGraphUsers.ContinuationToken` is `*[]string`, not `*string` — use `(*page.ContinuationToken)[0]`.
- `ListUsersArgs.ContinuationToken` is `*string`.
- `ISubscriptionChannel` Go struct has only `Type` — `channel_address` (the ADO `address` field) is not returned via the Go API wrapper. Preserve from state/plan, never read from API.
- `ISubscriptionFilter` Go struct has only `EventType` + `Type` — `filter_criteria` is not returned. Set to null after apply if not in config.
- Leaving ANY `Optional+Computed` field as Unknown after Create/Read causes TF error "provider still indicated an unknown value after apply".
