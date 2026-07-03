resource "betterado_serviceendpoint_generic_v2" "example" {
  project_id           = betterado_project.example.id
  name                 = "Example Generic V2 Service Connection"
  description          = "Managed by Terraform"
  type                 = "generic"
  server_url           = "https://example.com"
  authorization_scheme = "UsernamePassword"

  authorization_parameters = {
    username = "myuser"
    password = "mypassword"
  }
}
