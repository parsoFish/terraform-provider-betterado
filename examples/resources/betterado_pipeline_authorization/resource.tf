data "betterado_project" "example" {
  name = "MyProject"
}

resource "betterado_agent_pool" "example" {
  name           = "my-pool"
  auto_provision = false
  auto_update    = false
}

resource "betterado_agent_queue" "example" {
  project_id    = data.betterado_project.example.id
  agent_pool_id = betterado_agent_pool.example.id
}

resource "betterado_pipeline_authorization" "example" {
  project_id  = data.betterado_project.example.id
  resource_id = betterado_agent_queue.example.id
  type        = "queue"
}
