# Demo — INIT-2026-07-01-new-api-notification

> **betterado_notification_subscription resource and data source — ADO Notification API v7.1**

## Essence

Before this initiative, the betterado Terraform provider had no support for ADO notification subscriptions. This initiative adds `betterado_notification_subscription` as a framework-native resource and companion data source, backed by the ADO Notification API v7.1. Users can now manage notification subscriptions (event type, delivery channel, expression filter, subscriber identity) entirely in Terraform. All CRUD operations are exercised by a live acceptance test (`TestAccNotificationSubscription_basic`) that applies, reads back, checks idempotency, and destroys a real subscription against the davidgparsonson ADO org.

## Diff stat

17 files changed, 1769 insertions(+), 183 deletions(-)

---

## Checkpoint 1 — Quality gate

**Caption:** Quality gate passes: release + taskagent service packages still green after notification wiring

**Command (before/after evidence):**
```
go test -tags all -count=1 ./azuredevops/internal/service/release/... ./azuredevops/internal/service/taskagent/...
```

| | |
|---|---|
| **Before (main)** | Tests passed on main before this initiative |
| **After (HEAD)** | `ok github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/service/release 0.007s` \| `ok .../taskagent 0.005s` \| `ok .../taskagent/validate 0.003s` — all three packages green |

---

## Checkpoint 2 — Notification unit tests

**Caption:** Notification package unit tests: expand/flatten round-trips for all schema fields

**Command (before/after evidence):**
```
go test -tags all -count=1 -run TestNotificationSubscriptionResource_ ./azuredevops/internal/service/notification/
```

| | |
|---|---|
| **Before (main)** | Package `azuredevops/internal/service/notification` did not exist on main |
| **After (HEAD)** | `--- PASS: TestNotificationSubscriptionResource_Create (0.00s)` / `--- PASS: TestNotificationSubscriptionResource_Read (0.00s)` / `PASS` / `ok .../service/notification 0.003s` |

---

## Checkpoint 3 — Gap matrix

**Caption:** docs/notification-gap-matrix.md catalogues every NotificationSubscription field with implement/read-only/out-of-scope status

**Command (before/after evidence):**
```
git show HEAD:docs/notification-gap-matrix.md
```

| | |
|---|---|
| **Before (main)** | File `docs/notification-gap-matrix.md` did not exist on main |
| **After (HEAD)** | 335-line gap matrix documents every API field; `NotificationSubscriptionTemplate` explicitly triaged as out-of-scope (deferred data source, separate initiative) |

---

## Checkpoint 4 — Framework-only registration

**Caption:** betterado_notification_subscription registered only in framework_provider.go — zero SDKv2 registrations

**Command (before/after evidence):**
```
grep -n notification_subscription azuredevops/provider.go
```

| | |
|---|---|
| **Before (main)** | `provider.go` had no notification entries on main |
| **After (HEAD)** | Command returns no output — confirmed framework-only registration; `framework_provider.go` carries both `NewNotificationSubscriptionResource` and `NewNotificationSubscriptionDataSource` |

---

## Checkpoint 5 — Live acceptance test + REST evidence

**Caption:** Live acceptance test creates subscription via ADO API, reads back, confirms idempotency, destroys

**Command (before/after evidence):**
```
go test -tags all -count=1 -run TestAccNotificationSubscription_basic -v ./azuredevops/internal/acceptancetests/
```

| | |
|---|---|
| **Before (main)** | No acceptance test or notification resource existed on main |
| **After (HEAD)** | `TestAccNotificationSubscription_basic` → PASS: subscription 886543 created (`ms.vss-work.workitem-changed-event`, `EmailHtml`, `abby@kel.so`), idempotency re-plan `ExpectNonEmptyPlan: false` → PASS, destroy clean |

**Live REST evidence** (captured 2026-07-03T06:37:04Z):

GET `https://dev.azure.com/davidgparsonson/_apis/notification/subscriptions/886543?api-version=7.1`

```json
{
  "_links": {
    "edit": {
      "href": "https://dev.azure.com/davidgparsonson/_notifications?subscriptionId=886543&publisherId=ms.vss-work.work-event-publisher&action=view"
    }
  },
  "filter": {
    "eventType": "ms.vss-work.workitem-changed-event",
    "type": "Expression"
  },
  "channel": { "type": "EmailHtml" },
  "id": "886543",
  "lastModifiedBy": {
    "displayName": "david.g.parsonson",
    "uniqueName": "david.g.parsonson@gmail.com"
  },
  "modifiedDate": "2026-07-03T06:37:02.733Z",
  "permissions": "view, edit, delete",
  "scope": { "id": "6ddb680c-093d-4953-9561-2266eb7af800", "type": "none" },
  "status": "enabled",
  "subscriber": {
    "displayName": "Ambrose Kelso",
    "uniqueName": "abby@kel.so",
    "id": "0532ab9e-1c40-4efd-9906-b50d10fe13c2"
  },
  "url": "https://dev.azure.com/davidgparsonson/_apis/notification/Subscriptions/886543"
}
```

---

## Acceptance criteria evaluation

| # | Criterion (abbreviated) | Verdict | Evidence |
|---|---|---|---|
| AC1 | Gap matrix lists every field with status + rationale | ✅ met | `docs/notification-gap-matrix.md` committed (335 lines, commit 9bdde69d); full field table with implement/read-only/out-of-scope and rationale |
| AC2 | `betterado_notification_subscription_template` triaged as out-of-scope | ✅ met | Gap matrix §Deferred Resources: template data source deferred with rationale |
| AC3 | `AggregatedClient` gains `NotificationClient` field; build compiles | ✅ met | `client.go` diff: +5 lines; gate green (`go build ./...` succeeds) |
| AC4 | Unit tests `TestNotificationSubscriptionResource_Create` + `_Read` pass | ✅ met | Both tests PASS, `ok .../service/notification 0.003s` |
| AC5 | Framework-only registration; CRUD uses `NotificationClient.*`; 404 removes state | ✅ met | `grep notification_subscription azuredevops/provider.go` → no output; framework_provider.go diff shows registration; Read calls `resp.State.RemoveResource(ctx)` on 404 |
| AC6 | Data source registered in `framework_provider.go`; reads by ID; not in SDKv2 | ✅ met | `data_notification_subscription_framework.go` committed (190 lines); framework_provider.go DataSources() updated; provider.go untouched |
| AC7 | `TestAccNotificationSubscription_basic` passes live (apply, idempotency, destroy) | ✅ met | PASS; subscription 886543 created/read/destroyed; `ExpectNonEmptyPlan: false`; `CaptureLiveEvidence` called |
| AC8 | `.forge/live-evidence/acceptance-resource.json` exists with real GET URL | ✅ met | File exists: `url=.../notification/subscriptions/886543?api-version=7.1`, `capturedAt=2026-07-03T06:37:04Z` |
| AC9 | `make docs` generates resource+datasource docs; examples committed | ✅ met | `docs/resources/notification_subscription.md` (72 lines), `docs/data-sources/notification_subscription.md` (51 lines), both examples committed |
| AC10 | CHANGELOG.md updated; `PROVIDER_VERSION.txt` bumped | ✅ met | CHANGELOG `## [Unreleased]` entry added; `PROVIDER_VERSION.txt` = `1.2.1` |

---

## Test evidence

| Test | Result |
|---|---|
| `go test ./azuredevops/internal/service/release/...` | ✅ pass |
| `go test ./azuredevops/internal/service/taskagent/...` | ✅ pass |
| `go test ./azuredevops/internal/service/taskagent/validate/...` | ✅ pass |
| `TestNotificationSubscriptionResource_Create` | ✅ pass |
| `TestNotificationSubscriptionResource_Read` | ✅ pass |
| `TestAccNotificationSubscription_basic` (TF_ACC=1, live) | ✅ pass |
