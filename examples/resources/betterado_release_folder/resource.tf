data "betterado_project" "example" {
  name = "ExampleProject"
}

resource "betterado_release_folder" "example" {
  project_id  = data.betterado_project.example.id
  path        = "\\MyFolder"
  description = "Folder for MyFolder release pipelines"
}
