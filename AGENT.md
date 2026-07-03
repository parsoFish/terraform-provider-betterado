# Agent Memory — WI-2

> Institutional memory for this work item across Ralph iterations. Read at the start of every iteration; updated at the end.

## Brain context (loaded at iteration 1)

_(no brain context seeded — read theme files yourself if needed; the system prompt has the navigation index.)_

## What I've tried

_(updated by each iteration — most recent at the top)_

### Iteration 0 (WI-2 orientation pass)

All three ACs were already fully implemented in prior commits by WI-2 work:
- `ca623afd` — "docs(wi-2): update fix_plan and AGENT memory — all ACs complete"
- `a246f08c` — "feat(notification): add betterado_notification_subscription framework resource"

Current state confirmed by this iteration:
- AC1: `azuredevops/internal/client/client.go` has `NotificationClient notification.Client` at line 71 and initialisation at line 166/258. Build is clean (`go build ./...` → no errors).
- AC2: `resource_notification_subscription_framework_test.go` has build tag `//go:build all || resource_notification_subscription`. Running with `-tags resource_notification_subscription` both `TestNotificationSubscriptionResource_Create` and `TestNotificationSubscriptionResource_Read` PASS.
- AC3: `framework_provider.go` line 210 registers `notification.NewNotificationSubscriptionResource`; no SDKv2 provider.go (that file doesn't exist — the project uses only framework provider). CRUD methods call NotificationClient.* directly.

## What worked

- Running tests with explicit build tag: `go test -tags "resource_notification_subscription" -run "TestNotificationSubscriptionResource_Create|TestNotificationSubscriptionResource_Read" ./azuredevops/internal/service/notification/...`
- Without tags, `go test ./azuredevops/internal/service/notification/...` reports `[no test files]` — this is expected, not a failure.

## What didn't work

_(no dead-ends discovered this iteration — everything was already in place)_

## Open questions

_(none)_

## Notes for reflection

- The project only has `framework_provider.go` (no `provider.go` SDKv2 file), so AC3's "NOT added to SDKv2 provider.go ResourcesMap" is trivially satisfied.
- The `//go:build all || resource_notification_subscription` tag is the project convention for unit tests — the gate runner likely uses `-tags all` or the specific tag.
