# betterado_release_definition

Manages a classic release definition in Azure DevOps.

This resource allows you to create, update, and delete classic release pipelines, including multi-stage environments, approval workflows, deployment phases, and workflow tasks.

## Example Usage

### Basic Release Definition

```hcl
resource "betterado_release_definition" "example" {
  project_id          = "00000000-0000-0000-0000-000000000000"
  name                = "My Release Pipeline"
  release_name_format = "Release-$(rev:r)"

  environment {
    name = "Dev"
    rank = 1

    pre_deploy_approvals {
      approval {
        is_automated = true
      }
    }

    post_deploy_approvals {
      approval {
        is_automated = true
      }
    }

    deploy_phase {
      name       = "Agent job"
      phase_type = "agentBasedDeployment"
      rank       = 1

      deployment_input {
        queue_id = 1
      }
    }
  }
}
```

### Multi-Stage with Approvals

```hcl
resource "betterado_release_definition" "multi_stage" {
  project_id          = "00000000-0000-0000-0000-000000000000"
  name                = "Multi-Stage Pipeline"
  release_name_format = "Release-$(rev:r)"

  variable {
    name  = "Environment"
    value = "default"
  }

  tags = ["managed-by-terraform"]

  environment {
    name = "Dev"
    rank = 1

    condition {
      name           = "ReleaseStarted"
      condition_type = "event"
    }

    pre_deploy_approvals {
      approval {
        is_automated = true
      }
    }

    post_deploy_approvals {
      approval {
        is_automated = true
      }
    }

    deploy_phase {
      name       = "Deploy"
      phase_type = "agentBasedDeployment"
      rank       = 1

      deployment_input {
        queue_id = 1
      }

      workflow_task {
        task_id = "46e4be58-730b-4389-8a2f-ea10b3e5e815"
        version = "2.*"
        name    = "Azure CLI"
        enabled = true

        inputs = {
          scriptType     = "bash"
          scriptLocation = "inlineScript"
          inlineScript   = "echo 'Deploying to Dev'"
        }
      }
    }

    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }
  }

  environment {
    name = "Production"
    rank = 2

    condition {
      name           = "Dev"
      condition_type = "environmentState"
      value          = "4"
    }

    pre_deploy_approvals {
      approval {
        is_automated = false
        approver_id  = "user-guid-here"
      }

      timeout_in_minutes = 1440
    }

    post_deploy_approvals {
      approval {
        is_automated = true
      }
    }

    deploy_phase {
      name       = "Deploy"
      phase_type = "agentBasedDeployment"
      rank       = 1

      deployment_input {
        queue_id = 1
      }
    }

    retention_policy {
      days_to_keep     = 90
      releases_to_keep = 10
      retain_build     = true
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Required, ForceNew) The ID of the Azure DevOps project (UUID).
* `name` - (Required) The name of the release definition.
* `path` - (Optional) The folder path. Default: `\`.
* `description` - (Optional) Description of the release definition.
* `release_name_format` - (Optional) Format string for release names. Default: `Release-$(rev:r)`.
* `enabled` - (Optional) Whether the definition is enabled. Default: `true`.
* `variable_group_ids` - (Optional) List of variable group IDs to link.
* `tags` - (Optional) Set of tags.
* `variable` - (Optional) Pipeline-level variables. See [Variable](#variable) below.
* `artifact` - (Optional) Build artifacts to link. See [Artifact](#artifact) below.
* `environment` - (Required, min 1) Release environments/stages. See [Environment](#environment) below.

### Variable

* `name` - (Required) Variable name.
* `value` - (Optional) Variable value. Empty string if not set.
* `is_secret` - (Optional) Whether the value is secret. Default: `false`.
* `allow_override` - (Optional) Whether the value can be overridden at release time. Default: `false`.

### Environment

* `name` - (Required) Stage name.
* `rank` - (Required) Stage order (1-based).
* `variable` - (Optional) Environment-level variables. Same structure as top-level [Variable](#variable).
* `variable_group_ids` - (Optional) Variable group IDs for this stage.
* `pre_deploy_approvals` - (Optional) Pre-deployment approval config. See [Approval Config](#approval-config).
* `post_deploy_approvals` - (Optional) Post-deployment approval config. See [Approval Config](#approval-config).
* `deploy_phase` - (Required, min 1) Deployment phases/jobs. See [Deploy Phase](#deploy-phase).
* `condition` - (Optional) Deployment trigger conditions. See [Condition](#condition).
* `environment_options` - (Optional) Environment settings. See [Environment Options](#environment-options).
* `execution_policy` - (Optional) Concurrency settings. See [Execution Policy](#execution-policy).
* `retention_policy` - (Optional) Retention settings. See [Retention Policy](#retention-policy).

### Approval Config

* `approval` - (Required, min 1) List of approval steps.
  * `is_automated` - (Required) Whether approval is automatic.
  * `is_notification_on` - (Optional) Send notifications. Default: `false`.
  * `approver_id` - (Optional) User ID of the approver (required if `is_automated = false`).
  * `rank` - (Optional) Approval order. Default: `1`.
* `timeout_in_minutes` - (Optional) Approval timeout. Default: `0` (no timeout).
* `execution_order` - (Optional, Computed) When approvals run relative to gates.
* `approval_options` - (Optional) Approval behavior settings. See [Approval Options](#approval-options).

### Approval Options

* `required_approver_count` - (Optional) Number of approvers required. Default: all.
* `release_creator_can_be_approver` - (Optional) Allow the release creator to approve. Default: `false`.
* `enforce_identity_revalidation` - (Optional) Require re-authentication. Default: `false`.
* `timeout_in_minutes` - (Optional) Approval-level timeout. Default: `0`.
* `execution_order` - (Optional) When approvals run relative to gates. Values: `beforeGates`, `afterSuccessfulGates`, `afterGatesAlways`.
* `auto_triggered_and_previous_environment_approved_can_be_skipped` - (Optional) Whether approvals can be skipped for auto-triggered releases when the previous environment was approved. Default: `false`.

### Environment Options

* `email_notification_type` - (Optional) Notification mode. Values: `OnlyOnFailure`, `Always`, `Never`. Default: `OnlyOnFailure`.
* `email_recipients` - (Optional) Email addresses for notifications.
* `badge_enabled` - (Optional) Show deployment badge. Default: `false`.
* `auto_link_work_items` - (Optional) Auto-link work items. Default: `false`.
* `pull_request_deployment_enabled` - (Optional) Enable PR deployment. Default: `false`.
* `publish_deployment_status` - (Optional) Publish deployment status to source. Default: `true`.
* `timeout_in_minutes` - (Optional) Environment-level timeout. Default: `0`.
* `enable_access_token` - (Optional) Make OAuth token available. Default: `false`.
* `skip_artifacts_download` - (Optional) Skip downloading artifacts. Default: `false`.

### Execution Policy

* `concurrency_count` - (Optional) Number of concurrent deployments. Default: `1`.
* `queue_depth_count` - (Optional) Number of deployments to queue. Values: `0` or `1`. Default: `0`.

### Deploy Phase

* `name` - (Required) Phase/job name.
* `phase_type` - (Optional) Deployment type. Default: `agentBasedDeployment`. Options: `agentBasedDeployment`, `runOnServer`, `machineGroupBasedDeployment`, `deploymentGroup`.
* `rank` - (Required) Phase order.
* `deployment_input` - (Optional) Deployment configuration. See [Deployment Input](#deployment-input).
* `workflow_task` - (Optional) List of tasks. See [Workflow Task](#workflow-task).

### Deployment Input

* `queue_id` - (Optional) Agent queue ID.
* `timeout_in_minutes` - (Optional) Job timeout. Default: `0`.
* `job_cancel_timeout_in_minutes` - (Optional) Grace period for cancellation. Default: `1`.
* `condition` - (Optional) Job condition expression. Default: `succeeded()`.
* `skip_artifacts_download` - (Optional) Skip downloading artifacts. Default: `false`.
* `enable_access_token` - (Optional) Make OAuth token available to tasks. Default: `false`.
* `agent_specification` - (Optional) Agent image specification (e.g., `ubuntu-latest`, `windows-latest`). Free-form string.

### Workflow Task

* `task_id` - (Required) Task GUID from the marketplace (or task group UUID).
* `version` - (Required) Task version (e.g., `2.*`).
* `name` - (Required) Display name.
* `definition_type` - (Optional) Type of task. `task` for built-in/marketplace tasks, `metaTask` for task group references. Default: `task`.
* `enabled` - (Optional) Whether the task runs. Default: `true`.
* `always_run` - (Optional) Run even if previous tasks failed. Default: `false`.
* `continue_on_error` - (Optional) Continue on failure. Default: `false`.
* `condition` - (Optional) Condition expression. Default: `succeeded()`.
* `inputs` - (Optional) Map of task input key-value pairs.

### Artifact

* `source_id` - (Required) Artifact source identifier (e.g., `projectId:buildDefinitionId`).
* `type` - (Required) Artifact type. Currently supported: `Build`.
* `alias` - (Required) Unique alias for this artifact (e.g., `_MyBuild`).
* `definition_reference` - (Required) Map of key-value pairs defining the artifact source. Common keys: `definition` (build def name), `defaultVersionType` (e.g., `latestType`), `project` (project name), `repository` (repo name).
* `is_primary` - (Optional) Whether this is the primary artifact. Default: `false`.
* `is_retained` - (Optional) Whether the artifact is retained. Default: `false`.

### Retention Policy

* `days_to_keep` - (Optional) Days to retain. Default: `30`.
* `releases_to_keep` - (Optional) Number of releases to retain. Default: `3`.
* `retain_build` - (Optional) Also retain the associated build. Default: `true`.

### Condition

* `name` - (Required) Condition name (e.g., `ReleaseStarted` or a stage name).
* `condition_type` - (Required) Either `event` (first stage) or `environmentState` (depends on another stage).
* `value` - (Optional) Condition value. For `environmentState`, use `4` for "succeeded".

## Attribute Reference

In addition to all arguments, the following attributes are exported:

* `id` - The release definition ID.
* `revision` - The current revision number.
* `url` - The API URL of the definition.
* `environment.*.id` - Server-assigned environment IDs.

## Import

Release definitions can be imported using the project ID and definition ID:

```
terraform import betterado_release_definition.example PROJECT_ID/DEFINITION_ID
```
