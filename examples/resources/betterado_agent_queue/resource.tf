resource "betterado_project" "example" {
  name               = "Example Project"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_agent_pool" "example" {
  name           = "example-pool"
  auto_provision = false
  auto_update    = false
  pool_type      = "automation"
}

resource "betterado_agent_queue" "example" {
  project_id    = betterado_project.example.id
  agent_pool_id = betterado_agent_pool.example.id
}
