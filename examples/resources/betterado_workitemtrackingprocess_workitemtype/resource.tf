resource "betterado_workitemtrackingprocess_workitemtype" "example" {
  process_id  = betterado_workitemtrackingprocess_process.example.id
  name        = "My Work Item Type"
  color       = "CC293D"
  description = "A custom work item type"
  icon        = "icon_airplane"
}
