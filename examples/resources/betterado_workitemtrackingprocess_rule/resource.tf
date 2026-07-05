resource "betterado_workitemtrackingprocess_rule" "example" {
  process_id        = betterado_workitemtrackingprocess_process.example.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.example.reference_name
  name              = "Make Title Required When State Is Active"
  is_enabled        = true

  condition {
    condition_type = "when"
    field          = "System.State"
    value          = "Active"
  }

  action {
    action_type  = "makeRequired"
    target_field = "System.Title"
  }
}
