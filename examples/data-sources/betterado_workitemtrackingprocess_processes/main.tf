data "betterado_workitemtrackingprocess_processes" "all" {}

output "process_names" {
  value = [for p in data.betterado_workitemtrackingprocess_processes.all.processes : p.name]
}
