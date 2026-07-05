data "betterado_workitemtrackingprocess_workitemtype" "example" {
  process_id     = betterado_workitemtrackingprocess_process.example.id
  reference_name = "Microsoft.VSTS.WorkItemTypes.Bug"
}

output "workitemtype_name" {
  value = data.betterado_workitemtrackingprocess_workitemtype.example.name
}
