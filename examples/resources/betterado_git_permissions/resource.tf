resource "betterado_project" "example" {
  name               = "Example Project"
  description        = "Managed by Terraform"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_git_repository" "example" {
  project_id = betterado_project.example.id
  name       = "example-repo"
  initialization {
    init_type = "Clean"
  }
}

data "betterado_group" "example" {
  project_id = betterado_project.example.id
  name       = "[Example Project]\\Readers"
}

resource "betterado_git_permissions" "example" {
  project_id    = betterado_project.example.id
  repository_id = betterado_git_repository.example.id
  principal     = data.betterado_group.example.descriptor
  replace       = false

  permissions = {
    GenericRead   = "allow"
    GenericContribute = "deny"
    ForcePush     = "deny"
    CreateBranch  = "deny"
  }
}
