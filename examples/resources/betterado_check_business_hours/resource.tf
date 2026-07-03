data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_environment" "example" {
  name       = "ExampleEnvironment"
  project_id = data.betterado_project.example.id
}

resource "betterado_check_business_hours" "example" {
  project_id           = data.betterado_project.example.id
  target_resource_id   = data.betterado_environment.example.id
  target_resource_type = "environment"
  display_name         = "Business Hours Check"
  time_zone            = "UTC"
  start_time           = "09:00"
  end_time             = "17:00"
  monday               = true
  tuesday              = true
  wednesday            = true
  thursday             = true
  friday               = true
}
