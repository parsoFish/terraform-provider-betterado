# ── betterado_release_definition ──────────────────────────────────────────────
# The centrepiece. Shapes are grounded in the live-proven acceptance fixtures
# (so this applies cleanly against real ADO), and exercise the v0.2.0 surface:
#   • stages (renamed from `environment`) — the headline change
#   • triggers: cd_artifact_trigger (branch + tag filters, create_release_on_build_tagging),
#     schedule_trigger, container_image_trigger                     ← coverage gaps closed
#   • workflow_task timeout_in_minutes + retry_count_on_task_failure ← coverage gaps
#   • pre/post_deployment_gates (Query Work Items server gate)       ← coverage gaps
#   • deployment_input parallel_execution (multi-config matrix)      ← coverage gap
#   • environment_trigger (rollback redeploy)                        ← coverage gap
#   • agent + agentless deploy phases, approvals + approval_options, conditions,
#     variables, environment_options, execution_policy, retention_policy, metaTask
resource "betterado_release_definition" "showcase" {
  project_id          = var.project_id
  name                = "betterado-standing-showcase"
  description         = "Standing demo — every betterado_release_definition feature shipped in v0.2.0, applied live."
  release_name_format = "Release-$(rev:r)"

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

  artifact {
    alias      = "_build"
    type       = "Build"
    is_primary = true
    definition_reference = {
      definition = tostring(var.build_definition_id)
      project    = var.project_id
    }
  }

  # ── Triggers — branch+tag CD, schedule, and container-image. ──────────────────
  # schedule_trigger omits branch_filter (ADO does not return it → perpetual diff)
  # and REQUIRES a `ReleaseStarted` event condition on the first stage (below).
  triggers {
    cd_artifact_trigger {
      artifact_alias = "_build"
      branch_filter {
        include = ["refs/heads/main"]
        exclude = [] # ADO does not persist exclude filters for artifact CD triggers
      }
      tag_filter {
        tags = ["prod-ready", "signed"]
      }
      use_build_definition_branch     = false
      create_release_on_build_tagging = true
    }

    schedule_trigger {
      schedule_only_with_changes = true
      start_hours                = 2
      start_minutes              = 0
      time_zone_id               = "UTC"
      days_to_release            = 62 # Mon–Fri
    }

    container_image_trigger {
      artifact_alias = "_build"
      label          = "latest"
    }
  }

  # ═══ Stage 1: Staging — agent + agentless phases, gates, parallel matrix ═════
  stages {
    name = "Staging"
    rank = 1

    # Required by ADO when a schedule_trigger is present.
    condition {
      name           = "ReleaseStarted"
      condition_type = "event"
      value          = ""
    }

    variable {
      name  = "TargetSlot"
      value = "staging"
    }

    # NOTE: environment_trigger (rollbackRedeploy) is a v0.2.0 feature proven by
    # TestAccReleaseDefinition_environmentConfig, but it is omitted here: in a
    # multi-stage definition ADO assigns its definition_environment_id to a
    # sibling stage on read-back, which a static config can't track (perpetual
    # diff). The standing demo stays idempotent; the dedicated acc test covers it.

    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
      approval_options {
        release_creator_can_be_approver = true
        timeout_in_minutes              = 720
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
      name       = "Agent phase"
      rank       = 1
      phase_type = "agentBasedDeployment"

      deployment_input {
        queue_id            = var.queue_id
        agent_specification = "ubuntu-22.04"
        timeout_in_minutes  = 120
        condition           = "succeeded()"
        demands             = ["Agent.Version -gtVersion 2.0"]

        # Matrix / parallel execution (gap closed in INIT-2).
        parallel_execution {
          type                 = "multiConfiguration"
          max_number_of_agents = 2
          multipliers          = ["Configuration"]
          continue_on_error    = false
        }
      }

      # Per-task timeout + retry (gap closed in INIT-2).
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

    # Agentless (server) phase — a Delay task.
    deploy_phase {
      name       = "Agentless phase"
      rank       = 2
      phase_type = "runOnServer"

      deployment_input {
        timeout_in_minutes            = 0
        job_cancel_timeout_in_minutes = 1
        condition                     = "succeeded()"
      }

      workflow_task {
        name    = "Delay"
        task_id = "28782b92-5e8e-4458-9751-a71cd1492bae"
        version = "1.*"
        enabled = true
        inputs = {
          delayForMinutes = "1"
        }
      }
    }

    environment_options {
      email_notification_type   = "Always"
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

    # Pre/post deployment gates — the "Query Work Items" server gate (gap, INIT-2).
    pre_deployment_gates {
      gates_options {
        is_enabled               = true
        timeout                  = 60
        sampling_interval        = 5
        stabilization_time       = 0
        minimum_success_duration = 0
      }
      gate {
        task {
          name    = "Query Work Items"
          task_id = "f1e4b0e6-017e-4819-8a48-ef19ae96e289"
          version = "0.*"
          enabled = true
          inputs = {
            queryId = betterado_workitemquery.gate_query.id
          }
        }
      }
    }
    post_deployment_gates {
      gates_options {
        is_enabled        = true
        timeout           = 60
        sampling_interval = 5
      }
      gate {
        task {
          name    = "Query Work Items"
          task_id = "f1e4b0e6-017e-4819-8a48-ef19ae96e289"
          version = "0.*"
          enabled = true
          inputs = {
            queryId = betterado_workitemquery.gate_query.id
          }
        }
      }
    }
  }

  # ═══ Stage 2: Production — human approval, condition chain ════════════════════
  stages {
    name = "Production"
    rank = 2

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

    # Manual approval before production (a real project group identity).
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
        agent_specification = "ubuntu-22.04"
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
    }

    environment_options {
      email_notification_type   = "Always"
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
