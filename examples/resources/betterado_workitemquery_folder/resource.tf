data "betterado_project" "example" {
  name = "my-project"
}

resource "betterado_workitemquery_folder" "example" {
  project_id = data.betterado_project.example.id
  name       = "my-folder"
  area       = "My Queries"
}
