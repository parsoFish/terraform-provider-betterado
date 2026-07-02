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

resource "betterado_git_repository_file" "example" {
  repository_id    = betterado_git_repository.example.id
  file             = "README.md"
  content          = "# Hello World"
  branch           = "refs/heads/main"
  commit_message   = "Initial commit"
  overwrite_on_create = true
}
