data "betterado_project" "example" {
  name = "ExampleProject"
}

resource "betterado_git_repository" "example" {
  project_id = data.betterado_project.example.id
  name       = "ExampleRepo"

  initialization {
    init_type = "Clean"
  }
}

resource "betterado_git_repository_branch" "example" {
  repository_id = betterado_git_repository.example.id
  name          = "feature/my-branch"
  ref_branch    = "refs/heads/main"
}
