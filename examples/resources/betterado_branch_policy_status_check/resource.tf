data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_git_repository" "example" {
  name       = "ExampleRepo"
  project_id = data.betterado_project.example.id
}

resource "betterado_branch_policy_status_check" "example" {
  project_id = data.betterado_project.example.id
  enabled    = true
  blocking   = true

  settings {
    name                 = "my-external-check"
    genre                = "my-genre"
    invalidate_on_update = true
    applicability        = "conditional"

    scope {
      repository_id  = data.betterado_git_repository.example.id
      repository_ref = "refs/heads/main"
      match_type     = "Exact"
    }
  }
}
