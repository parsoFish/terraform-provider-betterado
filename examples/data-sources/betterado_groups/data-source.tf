data "betterado_project" "example" {
  name = "MyProject"
}

# List all groups in a project.
data "betterado_groups" "project_groups" {
  project_id = data.betterado_project.example.id
}

output "group_descriptors" {
  value = [for g in data.betterado_groups.project_groups.groups : g.descriptor]
}
