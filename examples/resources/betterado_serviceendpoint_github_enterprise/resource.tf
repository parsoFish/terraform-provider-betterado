# GitHub Enterprise service connection using a Personal Access Token (PAT).
resource "betterado_serviceendpoint_github_enterprise" "example" {
  project_id            = var.project_id
  service_endpoint_name = "my-github-enterprise"
  github_enterprise_url = "https://github.example.com"
  personal_access_token = var.github_enterprise_pat
  description           = "GitHub Enterprise connection via PAT"
}

variable "project_id" {
  type = string
}
variable "github_enterprise_pat" {
  type      = string
  sensitive = true
}
