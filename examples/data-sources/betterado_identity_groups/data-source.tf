data "betterado_project" "example" {
  name = "MyProject"
}

# List all identity groups in a project.
data "betterado_identity_groups" "project_groups" {
  project_id = data.betterado_project.example.id
}

output "group_names" {
  value = [for g in data.betterado_identity_groups.project_groups.groups : g.name]
}
