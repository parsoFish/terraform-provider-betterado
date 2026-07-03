resource "betterado_workitemtrackingprocess_control" "example" {
  process_id                   = betterado_workitemtrackingprocess_process.example.id
  work_item_type_reference_name = betterado_workitemtrackingprocess_workitemtype.example.reference_name
  group_id                     = betterado_workitemtrackingprocess_group.example.id
  control_id                   = "Custom.MyField"
  label                        = "My Field"
}
