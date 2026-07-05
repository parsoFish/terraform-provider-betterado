data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_git_repository" "example" {
  name       = "ExampleRepo"
  project_id = data.betterado_project.example.id
}

resource "betterado_branch_policy_comment_resolution" "example" {
  project_id = data.betterado_project.example.id
  enabled    = true
  blocking   = true

  settings {
    scope {
      repository_id  = data.betterado_git_repository.example.id
      repository_ref = "refs/heads/main"
      match_type     = "Exact"
    }
  }
}
