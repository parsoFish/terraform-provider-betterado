# Task Groups API Reference

## Source

Live API exploration against `davidgparsonson/DPLife` on 2026-04-06.
Created, updated, and deleted a task group via `dev.azure.com/_apis/distributedtask/taskgroups`.

## API Endpoint

```
Base: https://dev.azure.com/{org}/{project}/_apis/distributedtask/taskgroups
Version: api-version=7.1
```

**Note:** Task groups use the core `dev.azure.com` host (NOT `vsrm.dev.azure.com` like release definitions).

## SDK Support

The Go SDK provides full CRUD via `taskagent.Client`:
- `AddTaskGroup(ctx, AddTaskGroupArgs)` — Create (POST)
- `GetTaskGroups(ctx, GetTaskGroupsArgs)` — Read (GET, supports single or list)
- `UpdateTaskGroup(ctx, UpdateTaskGroupArgs)` — Update (PUT)
- `DeleteTaskGroup(ctx, DeleteTaskGroupArgs)` — Delete (DELETE)

Create uses `TaskGroupCreateParameter`, update uses `TaskGroupUpdateParameter`.

**The `TaskAgentClient` is already initialized** in `azuredevops/internal/client/client.go` (line 252).

## CRUD Behavior

### Create (POST)
- **Endpoint:** `POST /_apis/distributedtask/taskgroups?api-version=7.1`
- Returns the full task group object with server-assigned `id` (UUID), `revision: 1`, and `definitionType: "metaTask"`
- `version` field is required; typical starting value is `{ major: 1, minor: 0, patch: 0, isTest: false }`

### Read (GET)
- **Single:** `GET /_apis/distributedtask/taskgroups/{taskGroupId}?api-version=7.1`
- **List:** `GET /_apis/distributedtask/taskgroups?api-version=7.1`
- GET returns a wrapper: `{ count: N, value: [...] }` even for a single task group by ID

### Update (PUT)
- **Endpoint:** `PUT /_apis/distributedtask/taskgroups/{taskGroupId}?api-version=7.1`
- Requires `revision` for optimistic concurrency
- Requires `version` field for the API to locate the correct version
- Omitted fields are **wiped** (set to empty/default), so always send the full object
- Increments `revision` on success

### Delete (DELETE)
- **Endpoint:** `DELETE /_apis/distributedtask/taskgroups/{taskGroupId}?api-version=7.1`
- Returns 204 with no body

---

## Behavioral Findings

### 1. Revision Conflict Returns 409

Unlike release definitions (which return 400), task groups return a proper **409 Conflict**:
```json
{ "status": 409, "typeKey": "TaskGroupAlreadyUpdatedException",
  "message": "Task group {id} already updated." }
```

### 2. Version Is Part of the Identity

The `version` field (major/minor/patch) is used to locate the task group during PUT. Changing the major version in an update causes a **404** — the API can't find a task group at the new version. Version bumping likely requires the separate publish/preview workflow (not standard CRUD).

**TF implication:** Treat `version.major` as ForceNew, or only allow minor/patch changes via update.

### 3. Partial Updates Erase Fields

Sending a partial body (e.g. just name + revision + version) succeeds but **wipes all omitted fields** (tasks becomes [], inputs becomes [], etc.). Always read-modify-write with the full object.

### 4. definitionType Is Server-Assigned

Always returns `"metaTask"` — this is not user-settable. Computed only.

### 5. ID Is a UUID (Not Int)

Unlike release definitions which use integer IDs, task groups use **UUIDs**. This affects the import function and ID parsing.

---

## Field Inventory

### User-Settable Fields (for TF schema)

| API Field | TF Attribute | Type | Required | Notes |
|-----------|-------------|------|----------|-------|
| `name` | `name` | string | Yes | Internal name |
| `friendlyName` | `friendly_name` | string | Yes | Display name |
| `description` | `description` | string | No | |
| `category` | `category` | string | Yes | e.g. "Deploy", "Build", "Utility" |
| `instanceNameFormat` | `instance_name_format` | string | No | Display format with `$(var)` refs |
| `version` | `version` block | object | Yes | `major`, `minor`, `patch`, `is_test` |
| `runsOn` | `runs_on` | list(string) | No | ["Agent"], ["Server"], ["DeploymentGroup"] |
| `inputs` | `input` block list | list(object) | No | Parameterized inputs |
| `tasks` | `task` block list | list(object) | Yes | The steps in the group |
| `author` | `author` | string | No | |
| `iconUrl` | `icon_url` | string | No | |

### Input Definition Schema

| API Field | TF Attribute | Type | Notes |
|-----------|-------------|------|-------|
| `name` | `name` | string | Required — variable name |
| `label` | `label` | string | Required — display label |
| `type` | `type` | string | "string", "boolean", "filePath", "multiLine", etc. |
| `defaultValue` | `default_value` | string | |
| `required` | `required` | bool | |
| `helpMarkDown` | `help_markdown` | string | |
| `groupName` | `group_name` | string | |
| `visibleRule` | `visible_rule` | string | Conditional visibility expression |
| `options` | `options` | map(string) | For picklist types |
| `properties` | `properties` | map(string) | Additional metadata |
| `aliases` | `aliases` | list(string) | Alternative names |

### Task Step Schema (TaskGroupStep)

| API Field | TF Attribute | Type | Notes |
|-----------|-------------|------|-------|
| `displayName` | `display_name` | string | Required |
| `enabled` | `enabled` | bool | Default true |
| `alwaysRun` | `always_run` | bool | Default false |
| `continueOnError` | `continue_on_error` | bool | Default false |
| `condition` | `condition` | string | Default "succeeded()" |
| `timeoutInMinutes` | `timeout_in_minutes` | int | Default 0 (infinite) |
| `retryCountOnTaskFailure` | `retry_count_on_task_failure` | int | Default 0 |
| `task.id` | `task_id` | string (UUID) | Required — built-in task ID |
| `task.versionSpec` | `task_version` | string | e.g. "3.*" |
| `task.definitionType` | `task_definition_type` | string | Usually "task" |
| `inputs` | `inputs` | map(string) | Task-specific inputs |
| `environment` | `environment` | map(string) | Env vars for the task |

### Computed Fields (read-only)

| API Field | Notes |
|-----------|-------|
| `id` | UUID, server-assigned |
| `revision` | Int, incremented on each update |
| `definitionType` | Always "metaTask" |
| `createdBy` | IdentityRef |
| `createdOn` | ISO 8601 datetime |
| `modifiedBy` | IdentityRef |
| `modifiedOn` | ISO 8601 datetime |
| `demands` | Server-computed from tasks |
| `groups` | Server-computed |
| `satisfies` | Server-computed |
| `sourceDefinitions` | Server-computed |
| `dataSourceBindings` | Server-computed |
| `execution` | Server-computed |
| `preJobExecution` | Server-computed |
| `postJobExecution` | Server-computed |
| `_buildConfigMapping` | Internal |

---

## Relationship to Release Definitions

Task groups can be referenced from release definition deploy phases as workflow tasks where:
- `taskId` = the task group's `id` (UUID)
- `definitionType` = `"metaTask"` (instead of `"task"` for built-in tasks)
- `version` = version spec like `"1.*"`

This means a `betterado_task_group` resource would naturally feed into `betterado_release_definition` workflow tasks.

---

## Example API Payloads

### Minimal Create
```json
{
  "name": "my-task-group",
  "friendlyName": "My Task Group",
  "category": "Deploy",
  "version": { "major": 1, "minor": 0, "patch": 0, "isTest": false },
  "tasks": [
    {
      "displayName": "Run Script",
      "enabled": true,
      "task": { "id": "6c731c3c-3c68-459a-a5c9-bde6e6595b5b", "versionSpec": "3.*", "definitionType": "task" },
      "inputs": { "targetType": "inline", "script": "echo hello" }
    }
  ]
}
```

### Full Response Structure (key fields)
```json
{
  "id": "74377f96-f756-40c6-815d-608f550d9e1a",
  "name": "my-task-group",
  "friendlyName": "My Task Group",
  "description": "...",
  "category": "Deploy",
  "definitionType": "metaTask",
  "instanceNameFormat": "...",
  "revision": 1,
  "version": { "major": 1, "minor": 0, "patch": 0, "isTest": false },
  "runsOn": ["Agent"],
  "inputs": [ { "name": "...", "label": "...", "type": "string", ... } ],
  "tasks": [ { "displayName": "...", "task": {...}, "inputs": {...}, ... } ],
  "demands": [],
  "groups": [],
  "satisfies": [],
  "createdBy": { "id": "...", "displayName": "..." },
  "createdOn": "2026-04-06T01:06:47.72Z",
  "modifiedBy": { "id": "...", "displayName": "..." },
  "modifiedOn": "2026-04-06T01:06:47.72Z"
}
```
