# GitHub service connection using a Personal Access Token (PAT).
resource "betterado_serviceendpoint_github" "pat_example" {
  project_id            = var.project_id
  service_endpoint_name = "my-github-pat"
  personal_access_token = var.github_pat
  description           = "GitHub connection via PAT"
}

# GitHub service connection using OAuth (requires an ADO OAuth app configuration).
resource "betterado_serviceendpoint_github" "oauth_example" {
  project_id             = var.project_id
  service_endpoint_name  = "my-github-oauth"
  oauth_configuration_id = var.oauth_configuration_id
  description            = "GitHub connection via OAuth"
}

variable "project_id" {
  type = string
}
variable "github_pat" {
  type      = string
  sensitive = true
}
variable "oauth_configuration_id" {
  type    = string
  default = ""
}
