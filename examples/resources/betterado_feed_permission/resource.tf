resource "betterado_project" "example" {
  name = "Example Project"
}

resource "betterado_feed" "example" {
  name       = "my-artifacts-feed"
  project_id = betterado_project.example.id
}

resource "betterado_feed_permission" "example" {
  feed_id             = betterado_feed.example.id
  project_id          = betterado_project.example.id
  role                = "contributor"
  identity_descriptor = "Microsoft.TeamFoundation.Identity;..."
  display_name        = "My Contributors"
}
