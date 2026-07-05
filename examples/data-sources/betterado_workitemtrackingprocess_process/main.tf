data "betterado_workitemtrackingprocess_process" "example" {
  id = "00000000-0000-0000-0000-000000000000" # process ID (UUID)
}

output "process_name" {
  value = data.betterado_workitemtrackingprocess_process.example.name
}
