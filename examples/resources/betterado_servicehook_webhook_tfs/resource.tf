data "betterado_project" "example" {
  name = "ExampleProject"
}

resource "betterado_servicehook_webhook_tfs" "example" {
  project_id = data.betterado_project.example.id
  url        = "https://example.com/webhook"

  git_push {}
}
