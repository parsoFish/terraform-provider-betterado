resource "betterado_project" "example" {
  name               = "ExampleProject"
  description        = "Example project for release definition permissions"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_release_definition" "example" {
  project_id = betterado_project.example.id
  name       = "ExamplePipeline"

  stages {
    name = "Production"
    rank = 1

    deploy_phase {
      name       = "Agent job"
      rank       = 1
      phase_type = "agentBasedDeployment"
    }

    retention_policy {
      days_to_keep     = 30
      releases_to_keep = 3
      retain_build     = true
    }

    pre_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }

    post_deploy_approval {
      approver {
        id           = "00000000-0000-0000-0000-000000000000"
        is_automated = true
        rank         = 1
      }
    }
  }
}

data "betterado_group" "readers" {
  project_id = betterado_project.example.id
  name       = "Readers"
}

resource "betterado_release_definition_permissions" "example" {
  project_id            = betterado_project.example.id
  principal             = data.betterado_group.readers.id
  release_definition_id = betterado_release_definition.example.id

  # Keyed by the ReleaseManagement namespace action names (validated by ADO at apply).
  permissions = {
    ViewReleaseDefinition        = "Allow"
    EditReleaseDefinition        = "Allow"
    DeleteReleaseDefinition      = "Deny"
    ManageReleaseApprovers       = "NotSet"
    ManageReleases               = "Allow"
    ViewReleases                 = "Allow"
    CreateReleases               = "Allow"
    EditReleaseEnvironment       = "Deny"
    DeleteReleaseEnvironment     = "Deny"
    AdministerReleasePermissions = "Deny"
    DeleteReleases               = "Deny"
    ManageDeployments            = "Allow"
    ManageReleaseSettings        = "Allow"
  }
}
