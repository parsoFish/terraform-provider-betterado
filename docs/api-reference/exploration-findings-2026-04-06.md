# API Exploration Findings — 2026-04-06

## Method

Live API exploration against `davidgparsonson/DPLife` using browser-based fetch calls to `vsrm.dev.azure.com`. Created a multi-stage definition (ID 4) with manual approvals, gates, demands, environmentOptions, and workflow tasks, then read it back and compared input vs output.

## Key Behavioral Findings

### 1. Tags Are Not Persisted via Definitions API

Tags sent on both POST (create) and PUT (update) are silently ignored — the response always returns `[]`. Tags likely require a separate REST endpoint (possibly `_apis/release/tags`). **Our current schema includes `tags` but they will never round-trip.**

### 2. PUT Requires Full Object (No Partial Updates)

Sending a partial PUT with just `{ id, name, revision }` returns:
```
VS402875: Release pipeline needs to have at least one stage.
```
Updates must include the complete definition including all environments.

### 3. Revision Conflict Returns 400, Not 409

Sending a stale revision returns:
```json
{ "status": 400, "typeKey": "InvalidRequestException",
  "message": "You are using an old copy of the release pipeline. Refresh your copy and try again." }
```
**CLAUDE.md says 409 — this needs correction.** Our error handling should check for this 400 + message pattern.

### 4. Demands Format: Flat String Array, Not Objects

We sent demands as `[{ name: "Agent.OS", value: "Linux" }]` but the API stored and returned them as `["Agent.OS", "Linux"]` — a flat alternating string array of name/value pairs. The expand/flatten functions need to handle this format.

### 5. executionPolicy.queueDepthCount Only Accepts 0 or 1

Sending `queueDepthCount: 2` returns:
```
VS402973: Input not valid for executionPolicy:'queueDepthCount:2, valid values are: 0, 1.'
```
Schema validation should enforce this.

### 6. Secret Variables Return null on Read

Top-level variable `secret_var` sent with `isSecret: true, value: "s3cret"` came back as `{ value: null, isSecret: true }`. Environment-level variables with `isSecret: false` return the value normally but omit the `isSecret` and `allowOverride` fields when they're false. **Our flatten function already handles this correctly.**

### 7. environmentOptions: Server Respects Values on Create

The Dev environment where we explicitly set `enableAccessToken: true, badgeEnabled: true, autoLinkWorkItems: true, publishDeploymentStatus: true` had those values preserved in the read-back. Environments where we didn't set `environmentOptions` got server defaults (all false except `emailNotificationType` and `emailRecipients`).

### 8. Gates Structure Confirmed

`preDeploymentGates` accepts:
```json
{
  "gatesOptions": {
    "isEnabled": true, "timeout": 1440,
    "samplingInterval": 5, "stabilizationTime": 5,
    "minimumSuccessDuration": 0
  },
  "gates": []
}
```
Server assigns an `id` to the gates block. When not configured, `gatesOptions` is `null` and `gates` is `[]`.

### 9. Approval Options Confirmed

`approvalOptions` on the approvals block includes:
- `requiredApproverCount` (int or null — null means "all required")
- `releaseCreatorCanBeApprover` (bool)
- `autoTriggeredAndPreviousEnvironmentApprovedCanBeSkipped` (bool)
- `enforceIdentityRevalidation` (bool)
- `timeoutInMinutes` (int — 0 = no timeout, 1440 = 24h)
- `executionOrder` (string: "beforeGates" or "afterSuccessfulGates")

### 10. Workflow Task Extra Fields

API returns additional fields not in our schema:
- `timeoutInMinutes` (int, default 0)
- `retryCountOnTaskFailure` (int, default 0)
- `refName` (string, default "")
- `overrideInputs` (map, default {})
- `environment` (map, default {})

### 11. Deploy Phase deploymentInput Full Structure

```json
{
  "parallelExecution": { "parallelExecutionType": "none" },
  "agentSpecification": null,
  "skipArtifactsDownload": false,
  "artifactsDownloadInput": { "downloadInputs": [] },
  "queueId": 4,
  "demands": [],
  "enableAccessToken": false,
  "timeoutInMinutes": 0,
  "jobCancelTimeoutInMinutes": 1,
  "condition": "succeeded()",
  "overrideInputs": {}
}
```

---

## Full Environment Field Inventory (23 keys)

| Field | Current Schema | Status |
|-------|---------------|--------|
| id | Computed ✓ | OK |
| name | Required ✓ | OK |
| rank | Required ✓ | OK |
| owner | Optional ✓ | OK |
| variables | Optional ✓ | OK |
| variableGroups | Optional ✓ | OK |
| preDeployApprovals | Optional ✓ | Partial — missing approvalOptions |
| postDeployApprovals | Optional ✓ | Partial — missing approvalOptions |
| deployStep | Computed (server) | Skip — internal |
| deployPhases | Required ✓ | Partial — missing deploymentInput |
| conditions | Optional ✓ | OK |
| retentionPolicy | Optional ✓ | OK |
| **environmentOptions** | **MISSING** | **Add** |
| **executionPolicy** | **MISSING** | **Add** |
| **demands** | **MISSING** | **Add** |
| **preDeploymentGates** | **MISSING** | **Add** |
| **postDeploymentGates** | **MISSING** | **Add** |
| **schedules** | **MISSING** | **Add (low priority)** |
| **environmentTriggers** | **MISSING** | **Add (low priority)** |
| currentRelease | Computed (server) | Skip — runtime state |
| properties | Computed (server) | Skip — internal metadata |
| badgeUrl | Computed (server) | Skip — derived |
| processParameters | Computed (server) | Skip — internal |
