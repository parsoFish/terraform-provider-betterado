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
}

resource "betterado_resource_authorization" "example" {
  project_id  = data.betterado_project.example.id
  resource_id = data.betterado_git_repository.example.id
  definition_id = betterado_build_definition.example.id
  authorized  = true
  type        = "endpoint"
}
