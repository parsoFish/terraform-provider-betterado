# Example: Grant Readers group specific permissions on a project namespace

data "betterado_security_namespace" "project" {
  name = "Project"
}

data "betterado_security_namespace_token" "project" {
  namespace_name = "Project"
  identifiers = {
    project_id = var.project_id
  }
}

data "betterado_identity_group" "readers" {
  project_id = var.project_id
  name       = "[MyProject]\\Readers"
}

resource "betterado_security_permissions" "example" {
  namespace_id = data.betterado_security_namespace.project.id
  token        = data.betterado_security_namespace_token.project.token
  principal    = data.betterado_identity_group.readers.subject_descriptor

  permissions = {
    GENERIC_READ  = "allow"
    GENERIC_WRITE = "deny"
    DELETE        = "deny"
  }

  # replace = true means these permissions replace any existing ACE for this principal
  replace = true
}
