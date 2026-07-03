# Bitbucket service connection using username and app password.
resource "betterado_serviceendpoint_bitbucket" "example" {
  project_id            = var.project_id
  service_endpoint_name = "my-bitbucket"
  username              = "myuser"
  password              = var.bitbucket_app_password
  description           = "Bitbucket connection"
}

variable "project_id" {
  type = string
}
variable "bitbucket_app_password" {
  type      = string
  sensitive = true
}
