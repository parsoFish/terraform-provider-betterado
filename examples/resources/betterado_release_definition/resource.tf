# A classic release pipeline with a single Production stage fed by a build
# artifact — uses the new array (attribute-assignment) syntax introduced in
# the framework rewrite. Both block syntax and array syntax are accepted.
resource "betterado_release_definition" "example" {
  name       = "app-release"
  project_id = var.project_id

  # CI build that produces the deployable artifact (array syntax).
  artifact = [{
    alias      = "_build"
    type       = "Build"
    is_primary = true
    definition_reference = {
      definition = tostring(var.build_definition_id)
      project    = var.project_id
    }
  }]

  # Definition-level variables (a secret value is stored encrypted by ADO).
  variable {
    name  = "APP_ENV"
    value = "production"
  }
  variable {
    name      = "SIGNING_KEY"
    value     = "replace-me"
    is_secret = true
  }

  # Stages array syntax — each stage is an object in the list.
  stages = [{
    name = "Production"
    rank = 1

    # Stage-scoped secret (round-trips cleanly).
    variable = [{
      name      = "DEPLOY_TOKEN"
      value     = "replace-me"
      is_secret = true
    }]

    deploy_phase = [{
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }]

    retention_policy = [{
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }]

    # ADO requires both pre- and post-deploy approval blocks (VS402877).
    pre_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
    post_deploy_approval = [{
      approver = [{
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }]
    }]
  }]
}

variable "project_id" {
  type = string
}

variable "build_definition_id" {
  type = number
}
