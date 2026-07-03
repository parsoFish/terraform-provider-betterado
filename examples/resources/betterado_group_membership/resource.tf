data "betterado_project" "example" {
  name = "MyProject"
}

data "betterado_group" "contributors" {
  project_id = data.betterado_project.example.id
  name       = "Contributors"
}

data "betterado_user" "example_user" {
  descriptor = "vssgp.Uy0xLTktMTU1MTM3NDI0NS0..."
}

# Add a user as a member of a group.
resource "betterado_group_membership" "example" {
  group  = data.betterado_group.contributors.descriptor
  members = [
    data.betterado_user.example_user.descriptor,
  ]
}
