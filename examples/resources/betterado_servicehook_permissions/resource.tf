resource "betterado_project" "example" {
  name               = "Example Project"
  description        = "Managed by Terraform"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

data "betterado_group" "example" {
  project_id = betterado_project.example.id
  name       = "[Example Project]\\Readers"
}

resource "betterado_servicehook_permissions" "example" {
  project_id = betterado_project.example.id
  principal  = data.betterado_group.example.descriptor
  replace    = false

  permissions = {
    ViewSubscriptions    = "allow"
    EditSubscriptions    = "deny"
    DeleteSubscriptions  = "deny"
    PublishEvents        = "deny"
  }
}
