resource "betterado_workitemtrackingprocess_list" "example" {
  name  = "My Allowed Values List"
  type  = "String"
  items = ["Option A", "Option B", "Option C"]
}
