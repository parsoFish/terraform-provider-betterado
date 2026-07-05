data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_git_repository" "example" {
  name       = "ExampleRepo"
  project_id = data.betterado_project.example.id
}

resource "betterado_branch_policy_merge_types" "example" {
  project_id = data.betterado_project.example.id
  enabled    = true
  blocking   = true

  settings {
    allow_squash                  = true
    allow_rebase_and_fast_forward = true
    allow_basic_no_fast_forward   = false
    allow_rebase_with_merge       = false

    scope {
      repository_id  = data.betterado_git_repository.example.id
      repository_ref = "refs/heads/main"
      match_type     = "Exact"
    }
  }
}
