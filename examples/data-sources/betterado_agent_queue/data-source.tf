data "betterado_project" "example" {
  name = "Example Project"
}

data "betterado_agent_queue" "example" {
  project_id = data.betterado_project.example.id
  name       = "Azure Pipelines"
}
