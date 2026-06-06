data "betterado_project" "example" {
  name = "MyProject"
}

data "betterado_release_definitions" "all" {
  project_id = data.betterado_project.example.id
}

data "betterado_release_definitions" "in_folder" {
  project_id = data.betterado_project.example.id
  path       = "\\MyFolder"
}

output "definition_names" {
  value = [for d in data.betterado_release_definitions.all.release_definitions : d.name]
}
