resource "betterado_workitemtrackingprocess_process" "example" {
  name                   = "My Custom Process"
  description            = "A custom Agile-based process"
  parent_process_type_id = "adcc42ab-9882-485e-a3ed-7678f01f66bc" # Agile process type ID
}
