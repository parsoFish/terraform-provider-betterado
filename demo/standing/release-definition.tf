# ── betterado_release_definition ──────────────────────────────────────────────
# The centrepiece. Exercises the full v0.2.0 surface:
#   • stages (renamed from `environment`) — the headline change
#   • triggers block: cd_artifact_trigger (tag filters + create_release_on_build_tagging),
#     schedule_trigger, source_repo_trigger, container_image_trigger  ← coverage gaps closed
#   • workflow_task timeout_in_minutes + retry_count_on_task_failure  ← coverage gaps
#   • pre/post_deployment_gates (gates_options + gate task)           ← coverage gaps
#   • deployment_input override_inputs + parallel_execution           ← coverage gaps
#   • environment_trigger                                             ← coverage gap
#   • approvals + approval_options, conditions, variables, variable groups,
#     environment_options, execution_policy, retention_policy
resource "betterado_release_definition" "showcase" {
  project_id          = var.project_id
  name                = "betterado-standing-showcase"
  description         = "Standing demo — exercises every betterado_release_definition feature shipped in v0.2.0."
  release_name_format = "Release-$(rev:r)"

  # Definition-level variables: plain, override-at-release, and secret.
  variable {
    name  = "AppVersion"
    value = "1.0.0"
  }
  variable {
    name           = "DeployTimeout"
    value          = "300"
    allow_override = true
  }
  variable {
    name      = "SigningKey"
    value     = "demo-signing-key"
    is_secret = true
  }

  # Artifact: the CI build that produces the deployable.
  artifact {
    alias      = "_ci"
    type       = "Build"
    is_primary = true
    definition_reference = {
      definition         = tostring(var.build_definition_id)
      project            = var.project_id
      defaultVersionType = "latestType"
    }
  }

  # ── Triggers — all four kinds, including the container_image_trigger gap. ──────
  triggers {
    # Continuous deployment off the build artifact, gated by tag + branch filters.
    cd_artifact_trigger {
      artifact_alias = "_ci"
      branch_filter {
        include = ["refs/heads/main", "refs/heads/release/*"]
        exclude = ["refs/heads/release/legacy"]
      }
      tag_filter {
        tags = ["prod-ready", "signed"]
      }
      use_build_definition_branch     = false
      create_release_on_build_tagging = true
    }

    # Scheduled release window (Mon–Fri 02:00 UTC).
    schedule_trigger {
      schedule_only_with_changes = true
      start_hours                = 2
      start_minutes              = 0
      time_zone_id               = "UTC"
      days_to_release            = 31 # Mon-Fri bitmask
    }

    # Source-repo trigger.
    source_repo_trigger {
      alias          = "_ci"
      branch_filters = ["refs/heads/main"]
    }

    # Container image trigger (gap closed in INIT-2): fire on a new image tag.
    container_image_trigger {
      artifact_alias = "_ci"
      label          = "latest"
    }
  }

  # ═══ Stage 1: Staging — multi-task phase, gates, parallel execution ══════════
  stages {
    name = "Staging"
    rank = 1

    variable {
      name  = "TargetSlot"
      value = "staging"
    }

    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
      approval_options {
        release_creator_can_be_approver = false
        timeout_in_minutes              = 1440
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
        agent_specification = "ubuntu-latest"
        timeout_in_minutes  = 120
        condition           = "succeeded()"
        demands             = ["docker"]

        # Phase-level task input overrides (gap closed in INIT-2).
        override_inputs = {
          "deploy_steps.target" = "staging"
        }

        # Matrix / parallel execution.
        parallel_execution {
          type                 = "multiConfiguration"
          max_number_of_agents = 2
          multipliers          = ["region"]
          continue_on_error    = false
        }
      }

      # A normal task, exercising the new per-task timeout + retry (gap, INIT-2).
      workflow_task {
        name                        = "Deploy artifacts"
        task_id                     = "d9bafed4-0b18-4f58-968d-86655b4d2ce9" # CmdLine v2
        version                     = "2.*"
        enabled                     = true
        condition                   = "succeeded()"
        definition_type             = "task"
        timeout_in_minutes          = 15
        retry_count_on_task_failure = 2
        inputs = {
          script = "echo Deploying $(AppVersion) to Staging"
        }
      }

      # The reusable task group as a metaTask.
      workflow_task {
        name            = "Standard Deploy Steps"
        task_id         = betterado_task_group.deploy_steps.id
        version         = "1.*"
        enabled         = true
        condition       = "succeeded()"
        definition_type = "metaTask"
        inputs = {
          target = "staging"
        }
      }
    }

    # Pre-deployment gate (gap closed in INIT-2): poll an REST gate before deploy.
    pre_deployment_gates {
      gates_options {
        is_enabled               = true
        timeout                  = 1440
        sampling_interval        = 15
        stabilization_time       = 5
        minimum_success_duration = 0
      }
      gate {
        task {
          name            = "Invoke REST API health gate"
          task_id         = "9c3e8943-130d-4eba-8ab6-69aa112af669" # Invoke REST API (gate)
          version         = "1.*"
          definition_type = "task"
          inputs = {
            connectionType = "connectedServiceName"
            method         = "GET"
          }
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

  # ═══ Stage 2: Production — human approval, condition chain, env trigger ══════
  stages {
    name = "Production"
    rank = 2

    # Only runs after Staging succeeds.
    condition {
      name           = "Staging"
      condition_type = "environmentState"
      value          = "4" # 4 = Succeeded
    }

    variable {
      name           = "MaintenanceWindow"
      value          = "02:00-04:00 UTC"
      allow_override = true
    }

    # Redeploy trigger (gap closed in INIT-2).
    environment_trigger {
      trigger_type = "rollbackRedeploy"
    }

    # Human approval before production.
    pre_deploy_approval {
      approver {
        id           = var.approver_id
        is_automated = false
        rank         = 1
      }
      approval_options {
        release_creator_can_be_approver = true
        enforce_identity_revalidation   = true
        timeout_in_minutes              = 1440
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
      name       = "Deploy to Production"
      phase_type = "agentBasedDeployment"
      rank       = 1

      deployment_input {
        queue_id            = var.queue_id
        agent_specification = "ubuntu-latest"
        timeout_in_minutes  = 180
        condition           = "succeeded()"
      }

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

    # Post-deployment gate.
    post_deployment_gates {
      gates_options {
        is_enabled        = true
        timeout           = 720
        sampling_interval = 30
      }
      gate {
        task {
          name            = "Post-deploy REST validation"
          task_id         = "9c3e8943-130d-4eba-8ab6-69aa112af669" # Invoke REST API (gate)
          version         = "1.*"
          definition_type = "task"
          inputs = {
            connectionType = "connectedServiceName"
            method         = "GET"
          }
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
