data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_git_repository" "example" {
  name       = "ExampleRepo"
  project_id = data.betterado_project.example.id
}

data "betterado_group" "example" {
  name       = "[ExampleProject]\\ExampleGroup"
  project_id = data.betterado_project.example.id
}

resource "betterado_branch_policy_auto_reviewers" "example" {
  project_id = data.betterado_project.example.id
  enabled    = true
  blocking   = true

  settings {
    auto_reviewer_ids = [data.betterado_group.example.id]
    submitter_can_vote = false
    message            = "Please review this change"

    scope {
      repository_id  = data.betterado_git_repository.example.id
      repository_ref = "refs/heads/main"
      match_type     = "Exact"
    }
  }
}
