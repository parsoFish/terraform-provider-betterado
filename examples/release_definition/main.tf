# =============================================================================
# Feature-Complete Release Definition Demo
# =============================================================================
#
# This self-contained example creates ALL prerequisite resources and showcases
# every schema feature of betterado_release_definition:
#
#   Prerequisites (created here):
#     - Git repository              (artifact source)
#     - Build definition            (artifact trigger)
#     - Variable group × 2          (shared + env-scoped)
#
#   Release definition features demonstrated:
#     - artifact                    (Build type, linked to CI pipeline)
#     - variable                    (plain, secret, allow_override)
#     - variable_groups             (top-level + env-scoped)
#     - environment × 3             (Dev, Staging, Production)
#     - condition                   (environmentState dependency chain)
#     - pre/post_deploy_approval    (automated + human approver)
#     - approval_options            (all 6 fields)
#     - deploy_phase                (agent-based with ranked phases)
#     - deployment_input            (queue, agent_spec, timeout, demands)
#     - workflow_task               (PowerShell, CmdLine, Bash; various flags)
#     - environment_options         (notifications, badges, email_recipients)
#     - execution_policy            (concurrency + queue depth)
#     - retention_policy            (per-environment retention rules)

terraform {
  required_providers {
    betterado = {
      source  = "parsoFish/betterado"
      version = "~> 0.1.0"
    }
  }
}

provider "betterado" {
  org_service_url       = var.ado_org_url
  personal_access_token = var.ado_pat
}

variable "ado_org_url" {
  description = "Azure DevOps organization URL"
  type        = string
}

variable "ado_pat" {
  description = "Personal Access Token"
  type        = string
  sensitive   = true
}

variable "project_id" {
  description = "Azure DevOps project ID (UUID)"
  type        = string
}

variable "project_name" {
  description = "Azure DevOps project name (used in output URLs)"
  type        = string
}

variable "queue_id" {
  description = "Agent pool queue ID (e.g. 9 for 'Azure Pipelines' hosted pool)"
  type        = number
}

variable "approver_id" {
  description = "UUID of the human approver for production deployments"
  type        = string
}

# =============================================================================
# 1. Prerequisites — resources the release definition depends on
# =============================================================================

# ── Git repo: source code repository for the build artifact ─────────────────
resource "betterado_git_repository" "app_repo" {
  project_id = var.project_id
  name       = "betterado-demo-app"

  initialization {
    init_type = "Clean"
  }
}

# ── Build definition: CI pipeline that produces deployable artifacts ─────────
resource "betterado_build_definition" "ci_pipeline" {
  project_id = var.project_id
  name       = "betterado-demo-ci"

  repository {
    repo_id   = betterado_git_repository.app_repo.id
    repo_type = "TfsGit"
  }
}

# ── Variable group: shared config available to all environments ──────────────
resource "betterado_variable_group" "shared_config" {
  project_id   = var.project_id
  name         = "betterado-shared-config"
  description  = "Shared configuration for all release environments"
  allow_access = true

  variable {
    name  = "REGION"
    value = "australiaeast"
  }

  variable {
    name  = "LOG_LEVEL"
    value = "info"
  }
}

# ── Variable group: production secrets (linked only to Prod environment) ─────
resource "betterado_variable_group" "prod_secrets" {
  project_id   = var.project_id
  name         = "betterado-prod-secrets"
  description  = "Sensitive values for the production environment"
  allow_access = false

  variable {
    name         = "DB_CONNECTION"
    secret_value = "Server=prod-sql.database.windows.net;Database=app;Integrated Security=true"
    is_secret    = true
  }

  variable {
    name         = "API_KEY"
    secret_value = "demo-api-key-not-real"
    is_secret    = true
  }
}

# =============================================================================
# 2. Task Group — reusable deploy steps referenced as a metaTask
# =============================================================================
resource "betterado_task_group" "deploy_steps" {
  project_id           = var.project_id
  name                 = "betterado-deploy-steps"
  friendly_name        = "Standard Deploy Steps"
  description          = "Reusable deployment task group managed by Terraform"
  category             = "Deploy"
  instance_name_format = "Deploy: $(target)"

  version {
    major   = 1
    minor   = 0
    patch   = 0
    is_test = false
  }

  input {
    name          = "target"
    label         = "Deployment Target"
    type          = "string"
    default_value = "default"
    required      = true
    help_markdown = "The target environment name to deploy to"
  }

  # Step 1: Pre-deploy health check
  task {
    display_name = "Health check"
    task_id      = "e213ff0f-5d5c-4791-802d-52ea3e7be1f1" # PowerShell v2
    task_version = "2.*"
    inputs = {
      targetType = "inline"
      script     = "Write-Host \"Health check for $(target)\""
    }
  }

  # Step 2: Deploy via shell
  task {
    display_name      = "Run deployment"
    task_id           = "6c731c3c-3c68-459a-a5c9-bde6e6595b5b" # Bash v3
    task_version      = "3.*"
    continue_on_error = true
    inputs = {
      targetType = "inline"
      script     = "echo 'Deploying to $(target)'"
    }
  }
}

# =============================================================================
# 3. Release Definition — exercises every schema feature
# =============================================================================
resource "betterado_release_definition" "example" {
  project_id          = var.project_id
  name                = "betterado-feature-complete"
  description         = "Managed by Terraform — showcases every betterado_release_definition feature"
  release_name_format = "Release-$(rev:r)"

  # ── Top-level variable groups (shared across all environments) ─────────────
  variable_groups = [betterado_variable_group.shared_config.id]

  # ── Top-level variables ────────────────────────────────────────────────────
  variable {
    name  = "AppVersion"
    value = "1.0.0"
  }

  variable {
    name           = "DeployTimeout"
    value          = "300"
    allow_override = true          # can be changed at release time
  }

  variable {
    name      = "SigningKey"
    value     = "demo-signing-key"
    is_secret = true               # masked in logs and API reads
  }

  # ── Artifact: links to the CI build pipeline ───────────────────────────────
  artifact {
    alias      = "_ci"
    type       = "Build"
    is_primary = true

    definition_reference = {
      definition         = betterado_build_definition.ci_pipeline.id
      project            = var.project_id
      defaultVersionType = "latestType"
    }
  }

  # ═══════════════════════════════════════════════════════════════════════════
  # Environment 1: Dev — fully automated, fast feedback loop
  # ═══════════════════════════════════════════════════════════════════════════
  environment {
    name = "Dev"
    rank = 1

    # Env-scoped variable (overrides top-level if same name)
    variable {
      name  = "TargetSlot"
      value = "dev"
    }

    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }

      approval_options {
        release_creator_can_be_approver                                = false
        enforce_identity_revalidation                                  = false
        timeout_in_minutes                                             = 0
        execution_order                                                = "beforeGates"
        auto_triggered_and_previous_environment_approved_can_be_skipped = false
      }
    }

    post_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }

    deploy_phase {
      name       = "Deploy to Dev"
      phase_type = "agentBasedDeployment"
      rank       = 1

      deployment_input {
        queue_id                      = var.queue_id
        agent_specification           = "ubuntu-latest"
        timeout_in_minutes            = 60
        job_cancel_timeout_in_minutes = 1
        condition                     = "succeeded()"
        skip_artifacts_download       = false
        enable_access_token           = false
      }

      workflow_task {
        name            = "Deploy application"
        task_id         = "e213ff0f-5d5c-4791-802d-52ea3e7be1f1" # PowerShell v2
        version         = "2.*"
        enabled         = true
        condition       = "succeeded()"
        definition_type = "task"
        inputs = {
          targetType = "inline"
          script     = "Write-Host \"Deploying v$(AppVersion) to Dev slot\""
        }
      }

      # metaTask: reference our reusable task group
      workflow_task {
        name            = "Standard Deploy Steps"
        task_id         = betterado_task_group.deploy_steps.id
        version         = "1.*"
        enabled         = true
        condition       = "succeeded()"
        definition_type = "metaTask"
        inputs = {
          target = "dev"
        }
      }
    }

    environment_options {
      email_notification_type         = "OnlyOnFailure"
      publish_deployment_status       = true
      badge_enabled                   = false
      auto_link_work_items            = true
      pull_request_deployment_enabled = false
    }

    execution_policy {
      concurrency_count = 1
      queue_depth_count = 0
    }

    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }
  }

  # ═══════════════════════════════════════════════════════════════════════════
  # Environment 2: Staging — depends on Dev, multi-step pipeline, demands
  # ═══════════════════════════════════════════════════════════════════════════
  environment {
    name = "Staging"
    rank = 2

    # Triggers only after Dev succeeds
    condition {
      name           = "Dev"
      condition_type = "environmentState"
      value          = "4" # 4 = Succeeded
    }

    # Env-scoped variables
    variable {
      name  = "TargetSlot"
      value = "staging"
    }

    # Link production secrets group to this environment too
    variable_groups = [betterado_variable_group.prod_secrets.id]

    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }

      approval_options {
        release_creator_can_be_approver = false
        timeout_in_minutes              = 1440 # 24 hours
        execution_order                 = "beforeGates"
      }
    }

    post_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }

    deploy_phase {
      name       = "Deploy to Staging"
      phase_type = "agentBasedDeployment"
      rank       = 1

      deployment_input {
        queue_id            = var.queue_id
        agent_specification = "windows-latest"
        timeout_in_minutes  = 120
        condition           = "succeeded()"
      }

      # Task 1: PowerShell pre-flight checks
      workflow_task {
        name            = "Pre-flight checks"
        task_id         = "e213ff0f-5d5c-4791-802d-52ea3e7be1f1" # PowerShell v2
        version         = "2.*"
        enabled         = true
        condition       = "succeeded()"
        definition_type = "task"
        inputs = {
          targetType = "inline"
          script     = "Write-Host \"Running pre-flight for $(AppVersion)\""
        }
      }

      # Task 2: CmdLine deploy (continue_on_error = true for non-blocking warnings)
      workflow_task {
        name              = "Deploy artifacts"
        task_id           = "d9bafed4-0b18-4f58-968d-86655b4d2ce9" # CmdLine v2
        version           = "2.*"
        enabled           = true
        continue_on_error = true
        condition         = "succeeded()"
        definition_type   = "task"
        inputs = {
          script = "echo Deploying $(AppVersion) to Staging"
        }
      }

      # Task 3: Bash smoke test (always_run = true — runs even if deploy fails)
      workflow_task {
        name            = "Smoke test"
        task_id         = "6c731c3c-3c68-459a-a5c9-bde6e6595b5b" # Bash v3
        version         = "3.*"
        enabled         = true
        always_run      = true
        condition       = "succeededOrFailed()"
        definition_type = "task"
        inputs = {
          targetType = "inline"
          script     = "echo 'Running smoke tests against Staging'"
        }
      }
    }

    environment_options {
      email_notification_type   = "Always"
      email_recipients          = "release.environment.owner;release.creator"
      publish_deployment_status = true
      badge_enabled             = true
      auto_link_work_items      = false
    }

    execution_policy {
      concurrency_count = 1
      queue_depth_count = 0
    }

    retention_policy {
      days_to_keep     = 60
      releases_to_keep = 5
      retain_build     = true
    }
  }

  # ═══════════════════════════════════════════════════════════════════════════
  # Environment 3: Production — human approval, strict retention, secrets
  # ═══════════════════════════════════════════════════════════════════════════
  environment {
    name = "Production"
    rank = 3

    # Triggers only after Staging succeeds
    condition {
      name           = "Staging"
      condition_type = "environmentState"
      value          = "4"
    }

    # Env-scoped variables: secret with allow_override
    variable {
      name  = "TargetSlot"
      value = "production"
    }

    variable {
      name           = "MaintenanceWindow"
      value          = "02:00-04:00 UTC"
      allow_override = true
    }

    # Prod-only variable group
    variable_groups = [betterado_variable_group.prod_secrets.id]

    # Human approval required before production deployment
    pre_deploy_approval {
      approver {
        id           = var.approver_id
        is_automated = false
        rank         = 1
      }

      approval_options {
        release_creator_can_be_approver                                = true
        enforce_identity_revalidation                                  = true
        timeout_in_minutes                                             = 1440
        execution_order                                                = "beforeGates"
        auto_triggered_and_previous_environment_approved_can_be_skipped = false
      }
    }

    post_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }

    deploy_phase {
      name       = "Deploy to Production"
      phase_type = "agentBasedDeployment"
      rank       = 1

      deployment_input {
        queue_id            = var.queue_id
        agent_specification = "ubuntu-latest"
        timeout_in_minutes  = 180
        condition           = "succeeded()"
      }

      # Task 1: Pre-deploy validation with custom condition
      workflow_task {
        name            = "Validate deployment window"
        task_id         = "e213ff0f-5d5c-4791-802d-52ea3e7be1f1" # PowerShell v2
        version         = "2.*"
        enabled         = true
        condition       = "and(succeeded(), eq(variables['Build.Reason'], 'Manual'))"
        definition_type = "task"
        inputs = {
          targetType = "inline"
          script     = "Write-Host \"Validated: deploying $(AppVersion) to Production\""
        }
      }

      # Task 2: Production deployment
      workflow_task {
        name            = "Deploy to production"
        task_id         = "d9bafed4-0b18-4f58-968d-86655b4d2ce9" # CmdLine v2
        version         = "2.*"
        enabled         = true
        condition       = "succeeded()"
        definition_type = "task"
        inputs = {
          script = "echo Deploying $(AppVersion) to Production"
        }
      }

      # Task 3: Cleanup task — always runs regardless of previous outcomes
      workflow_task {
        name            = "Post-deploy cleanup"
        task_id         = "6c731c3c-3c68-459a-a5c9-bde6e6595b5b" # Bash v3
        version         = "3.*"
        enabled         = true
        always_run      = true
        condition       = "always()"
        definition_type = "task"
        inputs = {
          targetType = "inline"
          script     = "echo 'Cleaning up temporary resources'"
        }
      }
    }

    environment_options {
      email_notification_type   = "Always"
      email_recipients          = "release.environment.owner;release.creator"
      publish_deployment_status = true
      badge_enabled             = true
      auto_link_work_items      = true
    }

    execution_policy {
      concurrency_count = 1
      queue_depth_count = 1
    }

    retention_policy {
      days_to_keep     = 7
      releases_to_keep = 1
      retain_build     = true
    }
  }
}

# =============================================================================
# 4. Outputs
# =============================================================================
output "release_definition_id" {
  value = betterado_release_definition.example.id
}

output "release_definition_url" {
  value = "${var.ado_org_url}/${var.project_name}/_release?definitionId=${betterado_release_definition.example.id}"
}

output "build_definition_id" {
  value = betterado_build_definition.ci_pipeline.id
}

output "git_repository_id" {
  value = betterado_git_repository.app_repo.id
}

output "shared_variable_group_id" {
  value = betterado_variable_group.shared_config.id
}

output "prod_secrets_variable_group_id" {
  value = betterado_variable_group.prod_secrets.id
}

output "task_group_id" {
  value = betterado_task_group.deploy_steps.id
}
