resource "betterado_workitemtrackingprocess_group" "example" {
  process_id             = betterado_workitemtrackingprocess_process.example.id
  work_item_type_id      = betterado_workitemtrackingprocess_workitemtype.example.reference_name
  page_id                = betterado_workitemtrackingprocess_page.example.id
  section_id             = "Section1"
  label                  = "My Custom Group"
}
