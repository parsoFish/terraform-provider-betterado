# ── betterado_release_definition_permissions ──────────────────────────────────
# Sets ACL bits on the showcase pipeline for a principal.
#
# The keys below are the four exercised by the live acceptance test
# (TestAccReleaseDefinitionPermissions_*), so they are verified to round-trip
# against the real ADO ReleaseManagement security namespace. The namespace
# exposes further actions; only add a key here once it is covered by a live test
# (the permission map is validated by ADO at apply time, not by the schema).
# Values are case-insensitive: Allow | Deny | NotSet.
resource "betterado_release_definition_permissions" "showcase" {
  project_id            = var.project_id
  principal             = var.permission_principal_id
  release_definition_id = betterado_release_definition.showcase.id

  permissions = {
    ViewReleases           = "Allow"
    CreateReleases         = "Allow"
    EditReleaseEnvironment = "Allow"
    DeleteReleases         = "Deny"
  }
}
