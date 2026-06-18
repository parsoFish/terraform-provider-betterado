# ── Outputs ───────────────────────────────────────────────────────────────────
output "release_definition_id" {
  value = betterado_release_definition.showcase.id
}

output "release_definition_url" {
  description = "Open the pipeline in the ADO web portal."
  value       = "${var.ado_org_url}/${var.project_name}/_release?definitionId=${betterado_release_definition.showcase.id}"
}

output "release_folder_path" {
  value = betterado_release_folder.showcase.path
}

output "task_group_id" {
  value = betterado_task_group.deploy_steps.id
}

# Round-trip confirmation: the data source re-read returns the name we created.
output "roundtrip_release_name" {
  description = "release_name_format as the live API returns it via the data source."
  value       = data.betterado_release_definition.showcase.release_name_format
}

output "stage_names" {
  description = "Stage names from the managed resource (the headline `stages` rename)."
  value       = [for s in betterado_release_definition.showcase.stages : s.name]
}
