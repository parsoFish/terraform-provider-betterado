resource "betterado_workitemtrackingprocess_inherited_state" "example" {
  process_id        = betterado_workitemtrackingprocess_process.example.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.example.reference_name
  name              = "Active"
}
