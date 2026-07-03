data "betterado_project" "example" {
  name = "MyProject"
}

data "betterado_git_repository" "example" {
  project_id = data.betterado_project.example.id
  name       = "MyRepo"
}

resource "betterado_build_definition" "example" {
  project_id = data.betterado_project.example.id
  name       = "My Pipeline"

  repository {
    repo_type   = "TfsGit"
    repo_id     = data.betterado_git_repository.example.id
    branch_name = data.betterado_git_repository.example.default_branch
    yml_path    = "azure-pipelines.yml"
  }

  variable {
    name  = "MY_VAR"
    value = "hello"
  }
}
