data "betterado_project" "example" {
  name = "MyProject"
}

# Look up a group in a specific project.
data "betterado_group" "contributors" {
  project_id = data.betterado_project.example.id
  name       = "Contributors"
}

output "contributors_descriptor" {
  value = data.betterado_group.contributors.descriptor
}
