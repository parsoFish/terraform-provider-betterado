# ServiceHook Gap Matrix

> **Initiative:** INIT-2026-07-01-migrate-framework-servicehook
> **ADO API version compared:** ServiceHooks REST API v7.1
> **Status key:** `covered` · `gap-open` · `gap-deferred`

This document compares each resource's SDKv2 Terraform schema field-by-field against the ADO ServiceHooks REST API v7.1 subscription model (`consumerInputs`, `publisherInputs`). Each row records whether a field exists in the schema, and—for writable gaps—its resolution status.

---

## 1. `betterado_servicehook_storage_queue_pipelines`

Consumer: `azureStorageQueue` · Consumer action: `enqueue` · Publisher: `pipelines`

### 1.1 Consumer inputs (`consumerInputs`)

| API field (camelCase) | Terraform attribute | Type | Required? | SDKv2 schema | Status |
|---|---|---|---|---|---|
| `accountName` | `account_name` | string | yes | `schema.TypeString`, Required | **covered** |
| `accountKey` | `account_key` | string | yes | `schema.TypeString`, Required, Sensitive | **covered** |
| `queueName` | `queue_name` | string | yes | `schema.TypeString`, Required | **covered** |
| `visiTimeout` | `visi_timeout` | int (seconds) | no | `schema.TypeInt`, Optional, Default `0` | **covered** |
| `ttl` | `ttl` | int (seconds) | no | `schema.TypeInt`, Optional, Default `604800` | **covered** |
| `sasToken` | — | string | no | not in schema | **gap-deferred** |

> **`sasToken` note:** The v7.1 API accepts a SAS token as an alternative auth mechanism for the storage queue. The SDKv2 schema uses `account_key` (primary/secondary key) only. Exposing `sasToken` would require a mutually-exclusive `account_key` / `sas_token` credential model; deferred to a dedicated auth-enhancement work item.

### 1.2 Publisher inputs (`publisherInputs`) — `ms.vss-pipelines.run-state-changed-event`

API event type: `ms.vss-pipelines.run-state-changed-event`
Terraform block: `run_state_changed_event`

| API field (camelCase) | Terraform attribute | Type | SDKv2 schema | Status |
|---|---|---|---|---|
| `projectId` | `project_id` (top-level) | string | Required, top-level | **covered** |
| `pipelineId` | `pipeline_id` | string | Optional | **covered** |
| `runStateId` | `run_state_filter` | string | Optional, enum `InProgress\|Canceling\|Completed` | **covered** |
| `runResultId` | `run_result_filter` | string | Optional, enum `Canceled\|Failed\|Succeeded` | **covered** |

### 1.3 Publisher inputs (`publisherInputs`) — `ms.vss-pipelines.stage-state-changed-event`

API event type: `ms.vss-pipelines.stage-state-changed-event`
Terraform block: `stage_state_changed_event`

| API field (camelCase) | Terraform attribute | Type | SDKv2 schema | Status |
|---|---|---|---|---|
| `projectId` | `project_id` (top-level) | string | Required, top-level | **covered** |
| `pipelineId` | `pipeline_id` | string | Optional | **covered** |
| `stageNameId` | `stage_name` | string | Optional | **covered** |
| `stageStateId` | `stage_state_filter` | string | Optional, enum `NotStarted\|Waiting\|Running\|Completed` | **covered** |
| `stageResultId` | `stage_result_filter` | string | Optional, enum `Canceled\|Failed\|Rejected\|Skipped\|Succeeded` | **covered** |

### 1.4 Writable gaps summary — storage queue pipelines

| Gap field | API key | Resolution | Notes |
|---|---|---|---|
| SAS token auth | `sasToken` (consumerInput) | **deferred** | Requires auth-type discriminator; separate work item |

---

## 2. `betterado_servicehook_webhook_tfs`

Consumer: `webHooks` · Consumer action: `httpRequest` · Publisher: `tfs`

### 2.1 Consumer inputs (`consumerInputs`)

| API field (camelCase) | Terraform attribute | Type | Required? | SDKv2 schema | Status |
|---|---|---|---|---|---|
| `url` | `url` | string | yes | `schema.TypeString`, Required, validated HTTPS/HTTP | **covered** |
| `acceptUntrustedCerts` | `accept_untrusted_certs` | bool | no | `schema.TypeBool`, Optional, Default `false` | **covered** |
| `basicAuthUsername` | `basic_auth_username` | string | no | `schema.TypeString`, Optional | **covered** |
| `basicAuthPassword` | `basic_auth_password` | string | no | `schema.TypeString`, Optional, Sensitive | **covered** |
| `httpHeaders` | `http_headers` | map(string) | no | `schema.TypeMap`, Optional | **covered** |
| `resourceDetailsToSend` | `resource_details_to_send` | string | no | Optional, Default `all`, enum `all\|minimal\|none` | **covered** |
| `messagesToSend` | `messages_to_send` | string | no | Optional, Default `all`, enum `all\|text\|html\|markdown\|none` | **covered** |
| `detailedMessagesToSend` | `detailed_messages_to_send` | string | no | Optional, Default `all`, enum `all\|text\|html\|markdown\|none` | **covered** |

> **`resource_version`** is stored at the subscription level (`ResourceVersion` field, not `consumerInputs`). Exposed via `resource_version` attribute (Optional, Default `latest`). **covered**.

### 2.2 Publisher inputs (`publisherInputs`) — TFS event types

All 19 event types covered by the `tfs` publisher. Each event type block exposes filterable `publisherInputs` fields.

#### Event type 1: `build_completed` (`build.complete`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |
| `definitionName` | `definition_name` | Optional | **covered** |
| `buildStatus` | `build_status` | Optional, enum `Succeeded\|PartiallySucceeded\|Failed\|Stopped` | **covered** |

#### Event type 2: `git_push` (`git.push`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |
| `repository` | `repository_id` | Optional | **covered** |
| `branch` | `branch` | Optional | **covered** |
| `pushedBy` | `pushed_by` | Optional | **covered** |

#### Event type 3: `git_pull_request_created` (`git.pullrequest.created`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |
| `repository` | `repository_id` | Optional | **covered** |
| `branch` | `branch` | Optional | **covered** |
| `pullrequestCreatedBy` | `pull_request_created_by` | Optional | **covered** |
| `pullrequestReviewersContains` | `pull_request_reviewers_contains` | Optional | **covered** |

#### Event type 4: `git_pull_request_updated` (`git.pullrequest.updated`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |
| `repository` | `repository_id` | Optional | **covered** |
| `branch` | `branch` | Optional | **covered** |
| `notificationType` | `notification_type` | Optional, enum `PushNotification\|ReviewersUpdateNotification\|StatusUpdateNotification\|ReviewerVoteNotification` | **covered** |
| `pullrequestCreatedBy` | `pull_request_created_by` | Optional | **covered** |
| `pullrequestReviewersContains` | `pull_request_reviewers_contains` | Optional | **covered** |

#### Event type 5: `git_pull_request_merge_attempted` (`git.pullrequest.merged`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |
| `repository` | `repository_id` | Optional | **covered** |
| `branch` | `branch` | Optional | **covered** |
| `pullrequestCreatedBy` | `pull_request_created_by` | Optional | **covered** |
| `pullrequestReviewersContains` | `pull_request_reviewers_contains` | Optional | **covered** |
| `mergeResult` | `merge_result` | Optional, enum `Succeeded\|Unsuccessful\|Conflicts\|Failure\|RejectedByPolicy` | **covered** |

#### Event type 6: `git_pull_request_commented` (`ms.vss-code.git-pullrequest-comment-event`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |
| `repository` | `repository_id` | Optional | **covered** |
| `branch` | `branch` | Optional | **covered** |
| `commentPattern` | — | not in schema | **gap-open** |

> **`commentPattern` note:** The API supports filtering PR comment events by a text pattern in the comment body. This filter is absent from the SDKv2 `git_pull_request_commented` block. Adding it is straightforward — a single Optional string `publisherInput`; marked open for the framework migration work item.

#### Event type 7: `repository_created` (`git.repo.created`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level + nested alias) | Required at top-level | **covered** |

#### Event type 8: `repository_deleted` (`git.repo.deleted`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |
| `repository` | `repository_id` | Optional | **covered** |

#### Event type 9: `repository_forked` (`git.repo.forked`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |
| `repository` | `repository_id` | Optional | **covered** |

#### Event type 10: `repository_renamed` (`git.repo.renamed`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |
| `repository` | `repository_id` | Optional | **covered** |

#### Event type 11: `repository_status_changed` (`git.repo.statuschanged`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |
| `repository` | `repository_id` | Optional | **covered** |

#### Event type 12: `tfvc_checkin` (`tfvc.checkin`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |
| `path` | `path` | Required | **covered** |
| `checkedInBy` | — | not in schema | **gap-open** |

> **`checkedInBy` note:** The API supports filtering TFVC check-in events by the committer's identity. Absent from the SDKv2 schema; marked open for the migration work item.

#### Event type 13: `work_item_created` (`workitem.created`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |
| `workItemType` | `work_item_type` | Optional | **covered** |
| `areaPath` | `area_path` | Optional | **covered** |
| `tag` | `tag` | Optional | **covered** |
| `linksChanged` | `links_changed` | Optional, bool | **covered** |

#### Event type 14: `work_item_deleted` (`workitem.deleted`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |
| `workItemType` | `work_item_type` | Optional | **covered** |
| `areaPath` | `area_path` | Optional | **covered** |
| `tag` | `tag` | Optional | **covered** |

#### Event type 15: `work_item_restored` (`workitem.restored`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |
| `workItemType` | `work_item_type` | Optional | **covered** |
| `areaPath` | `area_path` | Optional | **covered** |
| `tag` | `tag` | Optional | **covered** |

#### Event type 16: `work_item_updated` (`workitem.updated`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |
| `workItemType` | `work_item_type` | Optional | **covered** |
| `areaPath` | `area_path` | Optional | **covered** |
| `tag` | `tag` | Optional | **covered** |
| `changedFields` | `changed_fields` | Optional | **covered** |
| `linksChanged` | `links_changed` | Optional, bool | **covered** |

#### Event type 17: `work_item_commented` (`workitem.commented`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |
| `workItemType` | `work_item_type` | Optional | **covered** |
| `areaPath` | `area_path` | Optional | **covered** |
| `tag` | `tag` | Optional | **covered** |
| `commentPattern` | `comment_pattern` | Optional | **covered** |

#### Event type 18: `service_connection_created` (`ms.vss-endpoint.endpoint-created`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |

> **Note:** The `service_connection_created` and `service_connection_updated` blocks in the schema expose a nested `project_id`, but `expandTfsEventConfig` maps it to `project` (API key), not `projectId`. The top-level `project_id` is always sent as `projectId`. The nested `project_id` in the block is a duplicate path; functionally redundant but harmless at read-back time. Flagged for cleanup in the framework migration.

#### Event type 19: `service_connection_updated` (`ms.vss-endpoint.endpoint-updated`)

| API field | Terraform attribute | SDKv2 schema | Status |
|---|---|---|---|
| `projectId` | `project_id` (top-level) | Required | **covered** |

### 2.3 Writable gaps summary — webhook TFS

| Gap field | API key | Event type | Resolution | Notes |
|---|---|---|---|---|
| PR comment pattern filter | `commentPattern` (publisherInput) | `git_pull_request_commented` | **open** | API supports comment body substring filter; absent from SDKv2 block |
| TFVC checked-in-by filter | `checkedInBy` (publisherInput) | `tfvc_checkin` | **open** | API supports committer identity filter; absent from SDKv2 block |
| SAS token auth | `sasToken` (consumerInput) | all | **deferred** | Same as storage-queue: requires auth-type discriminator |

---

## 3. Summary of all writable gaps

| Resource | Gap field | API key | Resolution |
|---|---|---|---|
| `betterado_servicehook_storage_queue_pipelines` | SAS token auth | `consumerInputs.sasToken` | **deferred** |
| `betterado_servicehook_webhook_tfs` | PR comment pattern filter | `publisherInputs.commentPattern` (on `git_pull_request_commented`) | **open** |
| `betterado_servicehook_webhook_tfs` | TFVC checked-in-by filter | `publisherInputs.checkedInBy` (on `tfvc_checkin`) | **open** |
| `betterado_servicehook_webhook_tfs` | SAS token auth | `consumerInputs.sasToken` | **deferred** |

**open** = in scope for the framework migration initiative (should be added when building the plugin-framework resource)
**deferred** = separate work item; out of scope for this initiative
