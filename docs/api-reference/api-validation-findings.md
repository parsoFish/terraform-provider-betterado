# API Validation Findings

## Source

Captured from a live ADO instance (`davidgparsonson/DPLife`) on 2026-04-04.
Definition created via UI "Empty job" template, read back via:
```
GET https://vsrm.dev.azure.com/{org}/{project}/_apis/release/definitions/1?api-version=7.1&$expand=environments,artifacts,triggers,variables,tags
```

---

## Top-Level Fields — Confirmed vs. Undocumented

### Confirmed (match our documented schema)
| Field | Type | Value from real API |
|-------|------|-------------------|
| `id` | int | 1 (server-assigned) |
| `name` | string | "tf-test-release-definition" |
| `path` | string | "\\" |
| `description` | string/null | null |
| `revision` | int | 1 |
| `source` | string | "userInterface" |
| `releaseNameFormat` | string | "Release-$(rev:r)" |
| `environments` | array | Populated (see below) |
| `artifacts` | array | [] (empty) |
| `triggers` | array | [] (empty) |
| `variables` | object | {} (empty) |
| `tags` | array | [] (empty) |
| `properties` | object | Contains typed $type/$value pairs |
| `url` | string | Full API URL |
| `_links` | object | self + web links |
| `createdBy` | IdentityRef | Full identity object |
| `createdOn` | datetime | ISO 8601 |
| `modifiedBy` | IdentityRef | Full identity object |
| `modifiedOn` | datetime | ISO 8601 |

### Newly Discovered (not in our initial docs)
| Field | Type | Value | Notes for TF Schema |
|-------|------|-------|-------------------|
| `isDeleted` | bool | false | Computed, read-only |
| `isDisabled` | bool | false | **Settable** — maps to TF `enabled` attribute |
| `variableGroups` | array[int] | [] | IDs of linked variable groups — **important** for TF |
| `projectReference` | object/null | null | Populated when expanded |

### Properties Object (typed values)
The `properties` field uses a `$type`/`$value` pattern instead of plain values:
```json
{
  "DefinitionCreationSource": {"$type": "System.String", "$value": "ReleaseNew"},
  "IntegrateBoardsWorkItems": {"$type": "System.String", "$value": "False"},
  "IntegrateJiraWorkItems": {"$type": "System.String", "$value": "false"}
}
```
**TF Implication:** We need to handle this typed format in expand/flatten. These are metadata properties, not user-facing config — probably ignore in schema or expose as computed.

---

## Environment Fields — Confirmed vs. Undocumented

### Core fields confirmed
- `id` (int, server-assigned: 1)
- `name` (string: "Stage 1")
- `rank` (int: 1)
- `owner` (IdentityRef)
- `variables` (object: {})
- `variableGroups` (array: [])
- `preDeployApprovals` / `postDeployApprovals` (see below)
- `deployPhases` (see below)
- `retentionPolicy` ({daysToKeep, releasesToKeep, retainBuild})
- `environmentTriggers` (array: [])
- `schedules` (array: [])
- `processParameters` (object: {})
- `properties` (typed $type/$value)

### Newly Discovered Environment Fields
| Field | Type | Value | Notes |
|-------|------|-------|-------|
| `deployStep` | object | {"id": 2} | Server-assigned deploy step reference |
| `environmentOptions` | object | See below | **Important** — notifications, badges, options |
| `demands` | array | [] | Environment-level agent demands |
| `conditions` | array | [{"name":"ReleaseStarted","conditionType":"event"}] | **Critical** — deployment conditions/triggers |
| `executionPolicy` | object | {"concurrencyCount":1,"queueDepthCount":0} | Concurrency control |
| `currentRelease` | object | {"id":0, ...} | Computed — current release state |
| `badgeUrl` | string | Full URL | Computed badge endpoint |
| `preDeploymentGates` | object | {"id":0,"gatesOptions":null,"gates":[]} | Gate configuration |
| `postDeploymentGates` | object | {"id":0,"gatesOptions":null,"gates":[]} | Gate configuration |

### environmentOptions (fully discovered)
```json
{
  "emailNotificationType": "OnlyOnFailure",
  "emailRecipients": "release.environment.owner;release.creator",
  "skipArtifactsDownload": false,
  "timeoutInMinutes": 0,
  "enableAccessToken": false,
  "publishDeploymentStatus": true,
  "badgeEnabled": false,
  "autoLinkWorkItems": false,
  "pullRequestDeploymentEnabled": false
}
```
**TF Implication:** Most of these should be optional attributes with server defaults.

### conditions (deployment trigger conditions)
```json
[{"name": "ReleaseStarted", "conditionType": "event", "value": "", "result": null}]
```
This is how ADO represents "After release" trigger. For multi-stage, dependent stages use:
```json
[{"name": "Stage 1", "conditionType": "environmentState", "value": "4"}]
```
Where value "4" = succeeded. **This must be mapped in TF for stage dependencies.**

---

## Approval Structure — Confirmed

### preDeployApprovals / postDeployApprovals
```json
{
  "approvals": [
    {"rank": 1, "isAutomated": true, "isNotificationOn": false, "id": 1}
  ],
  "approvalOptions": {
    "requiredApproverCount": null,
    "releaseCreatorCanBeApprover": false,
    "autoTriggeredAndPreviousEnvironmentApprovedCanBeSkipped": false,
    "enforceIdentityRevalidation": false,
    "timeoutInMinutes": 0,
    "executionOrder": "beforeGates"
  }
}
```

**Key findings:**
- `approvals[].id` is server-assigned (1, 3 — skipping 2 which is the deployStep)
- `executionOrder` differs between pre ("beforeGates") and post ("afterSuccessfulGates")
- `requiredApproverCount` can be null (means: all approvers required)

---

## DeployPhase Structure — Confirmed

### deploymentInput (full actual schema)
```json
{
  "parallelExecution": {"parallelExecutionType": "none"},
  "agentSpecification": null,
  "skipArtifactsDownload": false,
  "artifactsDownloadInput": {"downloadInputs": []},
  "queueId": 4,
  "demands": [],
  "enableAccessToken": false,
  "timeoutInMinutes": 0,
  "jobCancelTimeoutInMinutes": 1,
  "condition": "succeeded()",
  "overrideInputs": {}
}
```

**Key findings:**
- `queueId` was set to 4 (server-assigned default agent queue)
- `jobCancelTimeoutInMinutes` = 1 (default, not in original docs)
- `condition` = "succeeded()" (default YAML-style condition expression)
- `artifactsDownloadInput` has nested `downloadInputs` array
- `parallelExecution` is an object with `parallelExecutionType` enum, not just a string

---

## Save Dialog Fields

The UI save dialog reveals two fields that are part of the API:
1. **`path`** (Folder) — maps to `path` field, default "\\"
2. **`comment`** — sent with the create/update request, stored in revision history

---

## Summary: Schema Corrections for Terraform Resource

### Must Add to Schema
1. `isDisabled` → `enabled` (bool, Optional, default true)
2. `variableGroups` → `variable_group_ids` (list of ints, Optional)
3. `environment.conditions` → `condition` block (Required for multi-stage)
4. `environment.executionPolicy` → `execution_policy` block (Optional)
5. `environment.environmentOptions` → `environment_options` block (Optional, server defaults)
6. `environment.deploymentInput.jobCancelTimeoutInMinutes` → attribute
7. `environment.deploymentInput.condition` → deployment condition expression
8. `environment.demands` → agent demands list

### Computed-Only Fields
- `id`, `revision`, `url`, `_links`, `createdBy`, `createdOn`, `modifiedBy`, `modifiedOn`
- `isDeleted`
- `environment.id`, `environment.deployStep`, `environment.currentRelease`, `environment.badgeUrl`
- `approval.id`
- `properties` (typed metadata)

### Server Defaults to Document
- `releaseNameFormat`: "Release-$(rev:r)"
- `retentionPolicy`: {daysToKeep: 30, releasesToKeep: 3, retainBuild: true}
- `conditions`: [{"name":"ReleaseStarted","conditionType":"event"}] for first stage
- `executionPolicy`: {concurrencyCount: 1, queueDepthCount: 0}
- `environmentOptions.emailNotificationType`: "OnlyOnFailure"
- `jobCancelTimeoutInMinutes`: 1
