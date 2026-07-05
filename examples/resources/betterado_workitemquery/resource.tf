data "betterado_project" "example" {
  name = "my-project"
}

resource "betterado_workitemquery" "example" {
  project_id = data.betterado_project.example.id
  name       = "my-query"
  area       = "My Queries"
  wiql       = "SELECT [System.Id] FROM WorkItems WHERE [System.TeamProject] = @project"
}
