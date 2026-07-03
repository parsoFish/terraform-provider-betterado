data "betterado_project" "example" {
  name = "ExampleProject"
}

# Example: project wiki
resource "betterado_wiki" "project_wiki" {
  project_id = data.betterado_project.example.id
  name       = "MyProjectWiki"
  type       = "projectWiki"
}

# Example: code wiki backed by a git repository
resource "betterado_git_repository" "example" {
  project_id = data.betterado_project.example.id
  name       = "MyWikiRepo"
  initialization {
    init_type = "Clean"
  }
}

resource "betterado_wiki" "code_wiki" {
  project_id    = data.betterado_project.example.id
  repository_id = betterado_git_repository.example.id
  name          = "MyCodeWiki"
  version       = "master"
  type          = "codeWiki"
  mapped_path   = "/"
}
