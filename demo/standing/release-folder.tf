# ── betterado_release_folder ──────────────────────────────────────────────────
# Organises release pipelines into a folder tree (both writable fields set).
resource "betterado_release_folder" "showcase" {
  project_id  = var.project_id
  path        = "\\betterado-standing-demo"
  description = "Folder holding the standing-demo release pipelines."
}
