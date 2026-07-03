# Look up the root area node (with children) for a project.
data "betterado_area" "root" {
  project_id     = "00000000-0000-0000-0000-000000000000"
  fetch_children = true
}

# Look up a specific area path.
data "betterado_area" "sprint" {
  project_id = "00000000-0000-0000-0000-000000000000"
  path       = "MyArea/SubArea"
}
