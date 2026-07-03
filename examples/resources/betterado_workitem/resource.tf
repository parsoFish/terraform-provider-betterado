data "betterado_project" "example" {
  name = "my-project"
}

resource "betterado_workitem" "example" {
  project_id = data.betterado_project.example.id
  type       = "Task"
  title      = "My example work item"
  description = "Created by Terraform."
  state      = "Active"
}
