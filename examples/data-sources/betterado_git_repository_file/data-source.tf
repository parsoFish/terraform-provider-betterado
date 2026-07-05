data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_git_repository" "example" {
  project_id = data.betterado_project.example.id
  name       = "ExampleRepo"
}

data "betterado_git_repository_file" "example" {
  repository_id = data.betterado_git_repository.example.id
  file          = "README.md"
  branch        = "refs/heads/main"
}

output "file_content" {
  value = data.betterado_git_repository_file.example.content
}
