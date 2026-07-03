data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_git_repositories" "example" {
  project_id = data.betterado_project.example.id
}

output "repository_names" {
  value = [for r in data.betterado_git_repositories.example.repositories : r.name]
}
