data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_git_repository" "example" {
  project_id = data.betterado_project.example.id
  name       = "ExampleProject"
}

resource "betterado_pipeline" "example" {
  project_id         = data.betterado_project.example.id
  name               = "My YAML Pipeline"
  folder             = "\\MyFolder"
  configuration_type = "yaml"
  repo_id            = data.betterado_git_repository.example.id
  yaml_path          = "/azure-pipelines.yml"
}
