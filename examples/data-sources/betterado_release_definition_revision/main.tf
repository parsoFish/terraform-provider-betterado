data "betterado_project" "example" {
  name = "MyProject"
}

# Retrieve the JSON content of revision 3 of release definition 42.
data "betterado_release_definition_revision" "example" {
  project_id            = data.betterado_project.example.id
  release_definition_id = 42
  revision              = 3
}

output "revision_json" {
  value = data.betterado_release_definition_revision.example.json_content
}
