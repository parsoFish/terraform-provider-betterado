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
