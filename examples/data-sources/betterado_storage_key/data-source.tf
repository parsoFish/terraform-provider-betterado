data "betterado_project" "example" {
  name = "MyProject"
}

# Get the project descriptor first.
data "betterado_descriptor" "project" {
  storage_key = data.betterado_project.example.id
}

# Resolve the storage key (UUID) from a descriptor.
data "betterado_storage_key" "project" {
  descriptor = data.betterado_descriptor.project.descriptor
}

output "storage_key" {
  value = data.betterado_storage_key.project.storage_key
}
