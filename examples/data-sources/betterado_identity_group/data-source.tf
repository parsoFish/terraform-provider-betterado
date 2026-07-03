data "betterado_project" "example" {
  name = "MyProject"
}

# Look up an identity group by name within a project.
data "betterado_identity_group" "contributors" {
  project_id = data.betterado_project.example.id
  name       = "[MyProject]\\Contributors"
}

output "contributors_descriptor" {
  value = data.betterado_identity_group.contributors.descriptor
}
