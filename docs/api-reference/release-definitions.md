# ADO Release Definitions REST API Reference

## Base URL

```
https://vsrm.dev.azure.com/{organization}/{project}/_apis/release/definitions
```

**API Version:** 7.1

## Endpoints

### List Definitions

```
GET /_apis/release/definitions?api-version=7.1
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `searchText` | string | Filter by name |
| `expand` | string | Properties to expand: `environments`, `artifacts`, `triggers`, `variables`, `tags`, `lastRelease` |
| `artifactType` | string | Filter by artifact type |
| `isExactNameMatch` | bool | Exact name matching |
| `$top` | int | Results per page (default: 50) |
| `$skip` | int | Pagination offset |

### Get Definition

```
GET /_apis/release/definitions/{definitionId}?api-version=7.1
```

**Query Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| `revision` | int | Specific revision number |

### Create Definition

```
POST /_apis/release/definitions?api-version=7.1
```

**Request Body:** Full `ReleaseDefinition` object (see schema below)

### Update Definition

```
PUT /_apis/release/definitions/{definitionId}?api-version=7.1
```

**Request Body:** Full `ReleaseDefinition` object with current `revision` number

### Delete Definition

```
DELETE /_apis/release/definitions/{definitionId}?api-version=7.1
```

---

## Schema: ReleaseDefinition

```json
{
  "id": 0,
  "name": "string",
  "path": "string",
  "description": "string",
  "comment": "string",
  "releaseNameFormat": "Release-$(rev:r)",
  "revision": 0,
  "source": "restApi|ui|serviceImport",
  "environments": [ReleaseDefinitionEnvironment],
  "artifacts": [Artifact],
  "triggers": [ReleaseTrigger],
  "variables": { "key": ConfigurationVariableValue },
  "tags": ["string"],
  "retentionPolicy": RetentionPolicy,
  "properties": { "key": "string" }
}
```

## Schema: ReleaseDefinitionEnvironment

```json
{
  "id": 0,
  "name": "string",
  "rank": 1,
  "description": "string",
  "owner": IdentityRef,
  "variables": { "key": ConfigurationVariableValue },
  "variableGroups": [0],
  "preDeployApprovals": {
    "approvals": [{
      "rank": 1,
      "isAutomated": false,
      "isNotificationOn": false,
      "approver": IdentityRef
    }],
    "approvalOptions": {
      "requiredApproverCount": 0,
      "releaseCreatorCanBeApprover": true,
      "autoTriggeredAndPreviousEnvironmentApprovedCanBeSkipped": false,
      "enforceIdentityRevalidation": false,
      "timeoutInMinutes": 0,
      "executionOrder": "beforeGates|afterSuccessfulGates|afterGatesAlways"
    }
  },
  "postDeployApprovals": {
    "approvals": [{
      "rank": 1,
      "isAutomated": true,
      "isNotificationOn": false,
      "approver": IdentityRef
    }],
    "approvalOptions": { ... }
  },
  "preDeploymentGates": {
    "gates": [{
      "tasks": [WorkflowTask]
    }],
    "gatesOptions": {
      "isEnabled": false,
      "timeout": 0,
      "samplingInterval": 0,
      "stabilizationTime": 0,
      "minimumSuccessDuration": 0
    }
  },
  "postDeploymentGates": { ... },
  "deployPhases": [DeployPhase],
  "environmentOptions": {
    "emailNotificationType": "string",
    "emailRecipients": "string",
    "skipArtifactsDownload": false,
    "timeoutInMinutes": 0,
    "enableAccessToken": false,
    "publishDeploymentStatus": true,
    "badgeEnabled": false,
    "autoLinkWorkItems": false,
    "pullRequestDeploymentEnabled": false
  },
  "environmentTriggers": [EnvironmentTrigger],
  "schedules": [ReleaseSchedule],
  "retentionPolicy": {
    "daysToKeep": 30,
    "releasesToKeep": 3,
    "retainBuild": true
  },
  "processParameters": {},
  "properties": {}
}
```

## Schema: DeployPhase

```json
{
  "deploymentInput": {
    "parallelExecution": { "parallelExecutionType": "none|multiConfiguration|multiMachine" },
    "agentSpecification": { "identifier": "string" },
    "skipArtifactsDownload": false,
    "artifactsDownloadInput": {},
    "queueId": 0,
    "demands": [],
    "enableAccessToken": false,
    "timeoutInMinutes": 0,
    "jobCancelTimeoutInMinutes": 1,
    "condition": "string",
    "overrideInputs": {}
  },
  "rank": 1,
  "phaseType": "agentBasedDeployment|runOnServer|machineGroupBasedDeployment|deploymentGroup",
  "name": "string",
  "refName": "string",
  "workflowTasks": [WorkflowTask]
}
```

## Schema: WorkflowTask

```json
{
  "taskId": "guid-string",
  "version": "1.*",
  "name": "string",
  "refName": "string",
  "enabled": true,
  "alwaysRun": false,
  "continueOnError": false,
  "timeoutInMinutes": 0,
  "retryCountOnTaskFailure": 0,
  "definitionType": "task|metaTask",
  "overrideInputs": {},
  "condition": "succeeded()",
  "inputs": { "key": "string" },
  "environment": {}
}
```

## Schema: Artifact

```json
{
  "sourceId": "string",
  "type": "Build|Git|GitHub|TFVC|ExternalTfsBuild|Jenkins|NuGet|AzureContainerRegistry",
  "alias": "string",
  "isPrimary": true,
  "isRetained": false,
  "definitionReference": {
    "definition": { "id": "string", "name": "string" },
    "project": { "id": "string", "name": "string" },
    "defaultVersionType": { "id": "latestType", "name": "Latest" },
    "defaultVersionBranch": { "id": "refs/heads/main", "name": "" },
    "defaultVersionSpecific": { "id": "", "name": "" },
    "defaultVersionTags": { "id": "", "name": "" },
    "artifactSourceDefinitionUrl": { "id": "url", "name": "" }
  }
}
```

## Schema: Trigger Types

### ContinuousDeploymentTrigger
```json
{
  "triggerType": "artifactSource",
  "artifactAlias": "_BuildPipeline",
  "triggerConditions": [{
    "sourceBranch": "refs/heads/main",
    "tags": [],
    "useBuildDefinitionBranch": false,
    "createReleaseOnBuildTagging": false
  }]
}
```

### ScheduleTrigger
```json
{
  "triggerType": "schedule",
  "schedule": {
    "daysToRelease": 31,
    "jobId": "guid",
    "startHours": 3,
    "startMinutes": 0,
    "timeZoneId": "UTC"
  }
}
```

## Schema: Common Types

### ConfigurationVariableValue
```json
{
  "value": "string",
  "isSecret": false,
  "allowOverride": true
}
```

### IdentityRef
```json
{
  "id": "guid",
  "displayName": "string",
  "uniqueName": "string",
  "descriptor": "string"
}
```

### RetentionPolicy
```json
{
  "daysToKeep": 30,
  "releasesToKeep": 3,
  "retainBuild": true
}
```

---

## Notes

- The release API host is `vsrm.dev.azure.com`, NOT `dev.azure.com`
- Environment IDs are assigned server-side; don't set them on create
- The `revision` field must match the current server revision for updates (optimistic concurrency)
- `expand` parameter on list/get is important for performance — don't expand what you don't need
- Variable values with `isSecret: true` are write-only; reads return null for the value
- Task IDs are GUIDs that reference tasks from the marketplace or built-in task catalog

## Terraform Resource Mapping

| API Field | Terraform Attribute | Type | Notes |
|-----------|-------------------|------|-------|
| `id` | `id` | Computed | Set as resource ID |
| `name` | `name` | Required | |
| `path` | `path` | Optional, default `\\` | |
| `description` | `description` | Optional | |
| `releaseNameFormat` | `release_name_format` | Optional | |
| `revision` | `revision` | Computed | Used internally for updates |
| `environments` | `environment` | Required, list of blocks | Complex nested |
| `artifacts` | `artifact` | Optional, list of blocks | |
| `triggers` | `trigger` | Optional, list of blocks | |
| `variables` | `variable` | Optional, map or blocks | |
| `tags` | `tags` | Optional, set of strings | |
| `retentionPolicy` | `retention_policy` | Optional, block | |
