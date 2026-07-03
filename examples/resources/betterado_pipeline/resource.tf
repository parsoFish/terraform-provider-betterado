data "betterado_project" "example" {
  name = "ExampleProject"
}

resource "betterado_pipeline" "example" {
  project_id = data.betterado_project.example.id
  name       = "My YAML Pipeline"
  folder     = "\\MyFolder"
}
