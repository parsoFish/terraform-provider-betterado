resource "betterado_project" "example" {
  name               = "Example Project"
  description        = "Managed by Terraform"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_variable_group" "example" {
  project_id   = betterado_project.example.id
  name         = "example-variables"
  description  = "Example variable group"
  allow_access = true

  variable {
    name  = "EXAMPLE_VAR"
    value = "example-value"
  }
}

data "betterado_group" "example" {
  project_id = betterado_project.example.id
  name       = "[Example Project]\\Readers"
}

resource "betterado_variable_group_permissions" "example" {
  project_id        = betterado_project.example.id
  variable_group_id = betterado_variable_group.example.id
  principal         = data.betterado_group.example.descriptor
  replace           = false

  permissions = {
    View         = "allow"
    Use          = "allow"
    Administer   = "deny"
    Create       = "deny"
  }
}
