data "betterado_project" "example" {
  name = "Example Project"
}

data "betterado_environment" "example" {
  project_id = data.betterado_project.example.id
  name       = "example-environment"
}
