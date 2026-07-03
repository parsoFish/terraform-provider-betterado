data "betterado_workitemtrackingprocess_workitemtypes" "example" {
  process_id = betterado_workitemtrackingprocess_process.example.id
}

output "workitemtype_names" {
  value = [for wit in data.betterado_workitemtrackingprocess_workitemtypes.example.work_item_types : wit.name]
}
