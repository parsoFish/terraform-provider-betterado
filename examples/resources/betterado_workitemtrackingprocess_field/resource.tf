resource "betterado_workitemtrackingprocess_field" "example" {
  process_id        = betterado_workitemtrackingprocess_process.example.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.example.reference_name
  field_id          = "Custom.MyField"
  required          = false
}
