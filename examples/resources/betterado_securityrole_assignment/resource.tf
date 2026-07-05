resource "betterado_project" "example" {
  name               = "Example Project"
  description        = "Managed by Terraform"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_group" "example" {
  scope        = betterado_project.example.id
  display_name = "ExampleGroup"
}

resource "betterado_environment" "example" {
  project_id  = betterado_project.example.id
  name        = "ExampleEnvironment"
  description = "Example pipeline deployment environment"
}

resource "betterado_securityrole_assignment" "example" {
  scope       = "distributedtask.environmentreferencerole"
  resource_id = "${betterado_project.example.id}_${betterado_environment.example.id}"
  identity_id = betterado_group.example.origin_id
  role_name   = "Administrator"
}
