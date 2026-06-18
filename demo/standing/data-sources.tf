# ── Read-back data sources ────────────────────────────────────────────────────
# Prove the read path: each data source re-fetches a resource created above from
# the live API. After `apply`, these confirm the round-trip — the value we wrote
# is the value ADO returns.
data "betterado_release_definition" "showcase" {
  project_id            = var.project_id
  release_definition_id = betterado_release_definition.showcase.id
}

data "betterado_release_folder" "showcase" {
  project_id = var.project_id
  path       = betterado_release_folder.showcase.path
}

data "betterado_task_group" "deploy_steps" {
  project_id = var.project_id
  id         = betterado_task_group.deploy_steps.id
}
