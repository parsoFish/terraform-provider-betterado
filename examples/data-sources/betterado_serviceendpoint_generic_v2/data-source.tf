data "betterado_serviceendpoint_generic_v2" "example" {
  project_id = betterado_project.example.id
  name       = "Example Generic V2 Service Connection"
}

output "serviceendpoint_id" {
  value = data.betterado_serviceendpoint_generic_v2.example.id
}
