variable "project_id" {
  description = "The Azure DevOps project ID"
  type        = string
}

variable "repository_id" {
  description = "The Git repository ID"
  type        = string
}

# Generate a security token for a specific Git repository
data "betterado_security_namespace_token" "git_repo" {
  namespace_name = "Git Repositories"
  identifiers = {
    project_id    = var.project_id
    repository_id = var.repository_id
  }
}

output "git_repo_token" {
  value = data.betterado_security_namespace_token.git_repo.token
}
