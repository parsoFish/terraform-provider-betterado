resource "betterado_workitemtrackingprocess_page" "example" {
  process_id        = betterado_workitemtrackingprocess_process.example.id
  work_item_type_id = betterado_workitemtrackingprocess_workitemtype.example.reference_name
  label             = "My Custom Page"
  visible           = true
}
