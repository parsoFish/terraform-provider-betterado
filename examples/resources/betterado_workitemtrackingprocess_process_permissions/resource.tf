data "betterado_workitemtrackingprocess_process" "example" {
  name = "Agile"
}

data "betterado_group" "example" {
  name = "[TEAM FOUNDATION]\\Project Collection Administrators"
}

resource "betterado_workitemtrackingprocess_process_permissions" "example" {
  process_id = data.betterado_workitemtrackingprocess_process.example.id
  principal  = data.betterado_group.example.descriptor
  replace    = false

  permissions = {
    AdministerProcessPermissions = "allow"
    ReadProcessPermissions       = "allow"
    WriteProcessPermissions      = "deny"
    DeleteProcessPermissions     = "deny"
  }
}
