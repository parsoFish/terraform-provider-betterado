data "betterado_project" "example" {
  name = "MyProject"
}

data "betterado_release_folder" "example" {
  project_id = data.betterado_project.example.id
  path       = "\\MyFolder"
}

output "folder_description" {
  value = data.betterado_release_folder.example.description
}
