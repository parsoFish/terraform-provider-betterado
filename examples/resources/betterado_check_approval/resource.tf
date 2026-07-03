data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_environment" "example" {
  name       = "ExampleEnvironment"
  project_id = data.betterado_project.example.id
}

data "betterado_group" "example" {
  name       = "[ExampleProject]\\ExampleGroup"
  project_id = data.betterado_project.example.id
}

resource "betterado_check_approval" "example" {
  project_id            = data.betterado_project.example.id
  target_resource_id    = data.betterado_environment.example.id
  target_resource_type  = "environment"
  approvers             = [data.betterado_group.example.id]
  requester_can_approve = false
  timeout               = 43200
}
