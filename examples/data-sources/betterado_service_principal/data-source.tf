# Look up a service principal by display name.
data "betterado_service_principal" "build_service" {
  display_name = "Project Collection Build Service (myorg)"
}

output "build_service_descriptor" {
  value = data.betterado_service_principal.build_service.descriptor
}
