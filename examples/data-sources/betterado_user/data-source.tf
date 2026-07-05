data "betterado_project" "example" {
  name = "MyProject"
}

# Look up a group to get a member descriptor.
data "betterado_group" "contributors" {
  project_id = data.betterado_project.example.id
  name       = "Contributors"
}

# Look up a user by their descriptor.
data "betterado_user" "example" {
  descriptor = data.betterado_group.contributors.descriptor
}

output "user_principal_name" {
  value = data.betterado_user.example.principal_name
}
