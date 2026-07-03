data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_git_repository" "example" {
  project_id = data.betterado_project.example.id
  name       = "ExampleRepo"
}

output "repo_remote_url" {
  value = data.betterado_git_repository.example.remote_url
}
