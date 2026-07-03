data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_git_repository" "example" {
  name       = "ExampleRepo"
  project_id = data.betterado_project.example.id
}

data "betterado_build_definition" "example" {
  name       = "ExamplePipeline"
  project_id = data.betterado_project.example.id
}

resource "betterado_branch_policy_build_validation" "example" {
  project_id = data.betterado_project.example.id
  enabled    = true
  blocking   = true

  settings {
    build_definition_id        = data.betterado_build_definition.example.id
    display_name               = "Build Validation"
    manual_queue_only          = false
    queue_on_source_update_only = true
    valid_duration             = 720

    scope {
      repository_id  = data.betterado_git_repository.example.id
      repository_ref = "refs/heads/main"
      match_type     = "Exact"
    }
  }
}
