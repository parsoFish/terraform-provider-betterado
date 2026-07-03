data "betterado_project" "example" {
  name = "MyProject"
}

data "betterado_build_definition" "example" {
  project_id = data.betterado_project.example.id
  name       = "My Pipeline"
}

output "build_definition_id" {
  value = data.betterado_build_definition.example.id
}
