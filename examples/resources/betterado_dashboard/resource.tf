resource "betterado_project" "example" {
  name = "Example Project"
}

resource "betterado_team" "example" {
  project_id = betterado_project.example.id
  name       = "Example Team"
}

# Project-scoped dashboard
resource "betterado_dashboard" "project_dashboard" {
  project_id       = betterado_project.example.id
  name             = "My Project Dashboard"
  description      = "An example project-scoped dashboard managed by Terraform"
  refresh_interval = 5
}

# Team-scoped dashboard
resource "betterado_dashboard" "team_dashboard" {
  project_id       = betterado_project.example.id
  team_id          = betterado_team.example.id
  name             = "My Team Dashboard"
  description      = "An example team-scoped dashboard managed by Terraform"
  refresh_interval = 5
}
