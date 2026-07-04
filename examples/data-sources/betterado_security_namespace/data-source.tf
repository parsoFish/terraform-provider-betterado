# Look up a security namespace by name
data "betterado_security_namespace" "project" {
  name = "Project"
}

output "project_namespace_id" {
  value = data.betterado_security_namespace.project.id
}

output "project_namespace_actions" {
  value = data.betterado_security_namespace.project.actions
}
