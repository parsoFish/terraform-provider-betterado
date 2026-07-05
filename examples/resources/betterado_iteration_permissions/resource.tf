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

resource "betterado_iteration_permissions" "example" {
  project_id = betterado_project.example.id
  path       = "/"
  principal  = data.betterado_group.example.descriptor
  replace    = false

  permissions = {
    GENERIC_READ       = "allow"
    CREATE_CHILDREN    = "deny"
    GENERIC_WRITE      = "deny"
    DELETE             = "deny"
    MANAGE_TEST_PLANS  = "deny"
    MANAGE_TEST_SUITES = "deny"
  }
}
