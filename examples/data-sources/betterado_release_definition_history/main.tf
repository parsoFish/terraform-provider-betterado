data "betterado_project" "example" {
  name = "MyProject"
}

# List all revisions for release definition 42.
data "betterado_release_definition_history" "example" {
  project_id            = data.betterado_project.example.id
  release_definition_id = 42
}

output "latest_revision" {
  value = data.betterado_release_definition_history.example.revisions[0].revision
}

output "latest_changed_by" {
  value = data.betterado_release_definition_history.example.revisions[0].changed_by
}
