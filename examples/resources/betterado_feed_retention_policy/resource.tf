resource "betterado_project" "example" {
  name = "Example Project"
}

resource "betterado_feed" "example" {
  name       = "example-feed"
  project_id = betterado_project.example.id
}

resource "betterado_feed_retention_policy" "example" {
  project_id                                = betterado_project.example.id
  feed_id                                   = betterado_feed.example.id
  count_limit                               = 100
  days_to_keep_recently_downloaded_packages = 365
}
