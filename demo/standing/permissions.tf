# ── betterado_release_definition_permissions ──────────────────────────────────
# Every writable bit of the ADO ReleaseManagement security namespace
# (namespace c788c23e-…), keyed by the namespace's real action names. Values are
# case-insensitive: Allow | Deny | NotSet.
resource "betterado_release_definition_permissions" "showcase" {
  project_id            = var.project_id
  principal             = var.permission_principal_id
  release_definition_id = betterado_release_definition.showcase.id

  permissions = {
    ViewReleaseDefinition        = "Allow"
    EditReleaseDefinition        = "Allow"
    DeleteReleaseDefinition      = "Deny"
    ManageReleaseApprovers       = "Allow"
    ManageReleases               = "Allow"
    ViewReleases                 = "Allow"
    CreateReleases               = "Allow"
    EditReleaseEnvironment       = "Allow"
    DeleteReleaseEnvironment     = "Deny"
    AdministerReleasePermissions = "Deny"
    DeleteReleases               = "Deny"
    ManageDeployments            = "Allow"
    ManageReleaseSettings        = "Allow"
  }
}
