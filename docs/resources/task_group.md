---
page_title: "betterado_task_group Resource - betterado"
subcategory: "Release Pipelines"
description: |-
  Manages a reusable task group in Azure DevOps.
---

# betterado_task_group (Resource)

Manages an Azure DevOps task group — a named, versioned collection of pipeline
tasks that can be referenced from release definitions (as `definition_type =
"metaTask"` workflow tasks) or build definitions.

Task groups support parameterized inputs that callers can override, making them
useful for standardizing deployment patterns across multiple pipelines.

## Example Usage

```terraform
# A reusable task group referenced from release/build definitions as a
# definition_type = "metaTask" workflow task.
resource "betterado_task_group" "example" {
  project_id           = var.project_id
  name                 = "deploy-webapp"
  friendly_name        = "Deploy Web App"
  description          = "Reusable deployment steps for the web application"
  category             = "Deploy"
  author               = "platform-team"
  instance_name_format = "Deploy $(environment)"
  icon_url             = "https://cdn.vsassets.io/v/someicon.png"

  version = [{
    major = 1
    minor = 0
    patch = 0
  }]

  # Parameterized input surfaced to consumers of the task group.
  input = [{
    name          = "environment"
    label         = "Target environment"
    type          = "string"
    default_value = "staging"
    required      = true
    help_markdown = "The environment slot to deploy to."
    visible_rule  = "targetType = filePath"
    properties    = { "EndpointId" = "" }
    aliases       = ["targetEnvAlias"]
  }]

  # Task steps executed when the group runs.
  task = [{
    display_name = "Run deploy script"
    task_id      = "d9bafed4-0b18-4f58-968d-86655b4d2ce9" # CmdLine@2
    task_version = "2.*"

    inputs = {
      script = "echo Deploying to $(environment)"
    }
  }]
}

variable "project_id" {
  type = string
}
```

## Schema

### Required

- `category` (String) The category of the task group (e.g. `"Build"`, `"Deploy"`).
- `friendly_name` (String) The user-facing display name for the task group.
- `name` (String) The internal name of the task group.
- `project_id` (String) The ID of the Azure DevOps project. Forces replacement when changed.
- `task` (List of Object, Min: 1) One or more task steps. See [nested schema for `task`](#nested-schema-for-task) below.

### Optional

- `author` (String) The author of the task group. Defaults to `""`.
- `description` (String) Human-readable description. Defaults to `""`.
- `icon_url` (String) URL of the icon shown in the ADO UI. Defaults to `""`.
- `input` (List of Object) Zero or more parameterized inputs exposed to consumers. See [nested schema for `input`](#nested-schema-for-input) below.
- `instance_name_format` (String) Display name template for instances of this task group. Defaults to `""`.
- `runs_on` (List of String) Execution environments this group can run on (e.g. `["Agent"]`).
- `version` (List of Object, Max: 1) Semantic version of the task group. See [nested schema for `version`](#nested-schema-for-version) below. Defaults to the version returned by ADO.

### Read-Only

- `definition_type` (String) The ADO definition type (always `"metaTask"` for task groups).
- `id` (String) The GUID of the task group resource.
- `revision` (Number) The ADO revision number; incremented on each update.

---

### Nested Schema for `task`

Each element of `task = [{ ... }]` has the following attributes:

#### Required

- `display_name` (String) Step label shown in the ADO UI.
- `task_id` (String) UUID of the ADO task definition (e.g. `"d9bafed4-0b18-4f58-968d-86655b4d2ce9"` for CmdLine@2).
- `task_version` (String) Version spec of the task (e.g. `"2.*"`).

#### Optional

- `always_run` (Boolean) Whether the step runs even if a previous step failed. Defaults to `false`.
- `condition` (String) Run condition expression. Defaults to `"succeeded()"`.
- `continue_on_error` (Boolean) Whether the pipeline continues if this step fails. Defaults to `false`.
- `enabled` (Boolean) Whether the step is enabled. Defaults to `true`.
- `environment` (Map of String) Environment variables injected into the step.
- `inputs` (Map of String) Input values passed to the task.
- `retry_count_on_task_failure` (Number) How many times to retry on failure. Defaults to `0`.
- `task_definition_type` (String) The definition type of the referenced task. Defaults to `"task"`.
- `timeout_in_minutes` (Number) Step timeout in minutes. `0` means no timeout. Defaults to `0`.

---

### Nested Schema for `input`

Each element of `input = [{ ... }]` has the following attributes:

#### Required

- `label` (String) Human-readable label shown to pipeline editors.
- `name` (String) Parameter name used in `$(name)` references.

#### Optional

- `aliases` (List of String) Alternative names for the parameter.
- `default_value` (String) Default value when the caller does not supply one. Defaults to `""`.
- `group_name` (String) UI group this parameter belongs to. Defaults to `""`.
- `help_markdown` (String) Markdown help text shown in the editor. Defaults to `""`.
- `options` (Map of String) Allowed values for pick-list inputs.
- `properties` (Map of String) Arbitrary metadata (e.g. `{ "EndpointId" = "" }`).
- `required` (Boolean) Whether callers must provide a value. Defaults to `false`.
- `type` (String) Input type (e.g. `"string"`, `"boolean"`, `"pickList"`). Defaults to `"string"`.
- `visible_rule` (String) Conditional visibility expression. Defaults to `""`.

---

### Nested Schema for `version`

Each element of `version = [{ ... }]` has the following attributes:

#### Required

- `major` (Number) Major version number.
- `minor` (Number) Minor version number.
- `patch` (Number) Patch version number.

#### Optional

- `is_test` (Boolean) Whether this is a test/preview version. Defaults to `false`.

---

## Import

Task groups are imported using the `projectID/taskGroupID` format:

```shell
terraform import betterado_task_group.example 00000000-0000-0000-0000-000000000000/00000000-0000-0000-0000-000000000001
```
