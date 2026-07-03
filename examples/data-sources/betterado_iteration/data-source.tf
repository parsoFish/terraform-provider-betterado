# Look up the root iteration node (with children) for a project.
data "betterado_iteration" "root" {
  project_id     = "00000000-0000-0000-0000-000000000000"
  fetch_children = true
}

# Look up a specific iteration path.
data "betterado_iteration" "sprint1" {
  project_id = "00000000-0000-0000-0000-000000000000"
  path       = "Sprint 1"
}
