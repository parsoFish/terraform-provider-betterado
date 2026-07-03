data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_environment" "example" {
  name       = "ExampleEnvironment"
  project_id = data.betterado_project.example.id
}

resource "betterado_check_rest_api" "example" {
  project_id                      = data.betterado_project.example.id
  target_resource_id              = data.betterado_environment.example.id
  target_resource_type            = "environment"
  display_name                    = "REST API Gate"
  connected_service_name_selector = "connectedServiceName"
  connected_service_name          = "my-service-connection-id"
  method                          = "POST"
  url_suffix                      = "/api/gate"
  success_criteria                = "eq(root['status'], 'approved')"
  completion_event                = "Callback"
  timeout                         = 43200
}
