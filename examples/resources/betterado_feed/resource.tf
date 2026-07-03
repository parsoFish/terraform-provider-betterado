resource "betterado_project" "example" {
  name = "Example Project"
}

resource "betterado_feed" "example" {
  name       = "my-artifacts-feed"
  project_id = betterado_project.example.id
}
