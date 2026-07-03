data "betterado_project" "example" {
  name = "Example Project"
}

resource "betterado_environment" "example" {
  project_id  = data.betterado_project.example.id
  name        = "example-environment"
  description = "An example Azure DevOps environment."
}
