data "betterado_project" "example" {
  name = "MyProject"
}

data "betterado_group" "admins" {
  project_id = data.betterado_project.example.id
  name       = "Build Administrators"
}

# List all members of a group by its descriptor.
data "betterado_group_membership" "admins_membership" {
  group_descriptor = data.betterado_group.admins.descriptor
}

output "admin_members" {
  value = data.betterado_group_membership.admins_membership.members
}
