data "betterado_project" "example" {
  name = "MyProject"
}

# Create a custom group scoped to a project.
resource "betterado_group" "example" {
  scope        = data.betterado_project.example.id
  display_name = "my-custom-group"
  description  = "A custom group managed by Terraform."
}

output "group_descriptor" {
  value = betterado_group.example.descriptor
}
