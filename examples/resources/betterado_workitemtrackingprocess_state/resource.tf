resource "betterado_workitemtrackingprocess_state" "example" {
  process_id        = betterado_workitemtrackingprocess_process.example.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.example.reference_name
  name              = "In Review"
  color             = "007ACC"
  state_category    = "InProgress"
}
