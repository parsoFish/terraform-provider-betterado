data "betterado_project" "example" {
  name = "MyProject"
}

resource "betterado_build_folder" "example" {
  project_id  = data.betterado_project.example.id
  path        = "\\MyFolder"
  description = "Folder for pipeline definitions"
}
