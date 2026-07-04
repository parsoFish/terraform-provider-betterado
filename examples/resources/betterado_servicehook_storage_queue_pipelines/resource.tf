data "betterado_project" "example" {
  name = "ExampleProject"
}

resource "betterado_servicehook_storage_queue_pipelines" "example" {
  project_id   = data.betterado_project.example.id
  account_name = "myexamplestorageaccount"
  account_key  = "base64encodedkey=="
  queue_name   = "my-pipeline-events"

  stage_state_changed_event {}
}
