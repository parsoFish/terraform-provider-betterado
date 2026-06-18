# ── Prerequisites resolved/created in the demo project ────────────────────────

# A shared work-item query gives the "Query Work Items" deployment gate a real
# queryId (gates fail to apply without one).
resource "betterado_workitemquery" "gate_query" {
  project_id = var.project_id
  name       = "Standing Demo - Gate Check"
  area       = "Shared Queries"
  wiql       = "SELECT [System.Id] FROM WorkItems WHERE [System.TeamProject] = @project ORDER BY [System.Id]"
}
