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

resource "betterado_area_permissions" "example" {
  project_id = betterado_project.example.id
  principal  = data.betterado_group.example.descriptor
  path       = "/"
  replace    = false

  permissions = {
    GENERIC_READ  = "allow"
    GENERIC_WRITE = "deny"
    CREATE_CHILDREN = "deny"
    DELETE        = "deny"
    WORK_ITEM_MOVE = "notset"
  }
}
