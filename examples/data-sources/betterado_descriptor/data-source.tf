data "betterado_project" "example" {
  name = "MyProject"
}

# Look up the descriptor for a project using its storage key (UUID).
data "betterado_descriptor" "project" {
  storage_key = data.betterado_project.example.id
}

output "project_descriptor" {
  value = data.betterado_descriptor.project.descriptor
}
