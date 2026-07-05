data "betterado_feed" "example" {
  name       = "my-artifacts-feed"
  project_id = betterado_project.example.id
}
