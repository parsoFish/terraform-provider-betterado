data "betterado_project" "example" {
  name = "MyProject"
}

# Look up by ID
data "betterado_release_definition" "by_id" {
  project_id            = data.betterado_project.example.id
  release_definition_id = 42
}

# Look up by name
data "betterado_release_definition" "by_name" {
  project_id = data.betterado_project.example.id
  name       = "My Release Pipeline"
}

output "definition_path" {
  value = data.betterado_release_definition.by_id.path
}
