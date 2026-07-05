resource "betterado_workitemtrackingprocess_system_control" "example" {
  process_id        = betterado_workitemtrackingprocess_process.example.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.example.reference_name
  control_id        = "System.AreaPath"
  label             = "Area Path"
  visible           = true
}
