data "betterado_project" "example" {
  name = "ExampleProject"
}

data "betterado_git_repository" "example" {
  name       = "ExampleRepo"
  project_id = data.betterado_project.example.id
}

resource "betterado_repository_policy_author_email_pattern" "example" {
  project_id  = data.betterado_project.example.id
  enabled     = true
  blocking    = true

  repository_ids        = [data.betterado_git_repository.example.id]
  author_email_patterns = ["*@mycompany.com", "*@contractor.com"]
}
