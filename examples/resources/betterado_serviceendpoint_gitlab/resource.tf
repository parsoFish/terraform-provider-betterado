# GitLab service connection using username and password (API token).
resource "betterado_serviceendpoint_gitlab" "example" {
  project_id            = var.project_id
  service_endpoint_name = "my-gitlab"
  username              = "myuser"
  password              = var.gitlab_token
  description           = "GitLab connection"
}

variable "project_id" {
  type = string
}
variable "gitlab_token" {
  type      = string
  sensitive = true
}
