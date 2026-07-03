resource "betterado_serviceendpoint_generic_git" "example" {
  project_id            = betterado_project.example.id
  service_endpoint_name = "Example Generic Git Service Connection"
  description           = "Managed by Terraform"
  repository_url        = "https://github.com/example/repo.git"
  username              = "myuser"
  password              = "mypassword"
  enable_pipelines_access = true
}
