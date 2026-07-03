# List all available security namespaces
data "betterado_security_namespaces" "all" {}

output "namespace_count" {
  value = length(data.betterado_security_namespaces.all.namespaces)
}

output "namespace_names" {
  value = [for ns in data.betterado_security_namespaces.all.namespaces : ns.name]
}
