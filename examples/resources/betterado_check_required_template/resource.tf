data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_environment" "example" {
  name       = "ExampleEnvironment"
  project_id = data.betterado_project.example.id
}

resource "betterado_check_required_template" "example" {
  project_id           = data.betterado_project.example.id
  target_resource_id   = data.betterado_environment.example.id
  target_resource_type = "environment"

  required_template {
    repository_type = "azureRepos"
    repository_name = "ExampleProject/ExampleRepo"
    repository_ref  = "refs/heads/main"
    template_path   = "pipelines/templates/deploy.yml"
  }
}
