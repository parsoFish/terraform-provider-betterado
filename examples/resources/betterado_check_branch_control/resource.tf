data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_environment" "example" {
  name       = "ExampleEnvironment"
  project_id = data.betterado_project.example.id
}

resource "betterado_check_branch_control" "example" {
  project_id           = data.betterado_project.example.id
  target_resource_id   = data.betterado_environment.example.id
  target_resource_type = "environment"
  display_name         = "Branch Control Check"
  allowed_branches     = "refs/heads/main, refs/heads/release/*"
}
