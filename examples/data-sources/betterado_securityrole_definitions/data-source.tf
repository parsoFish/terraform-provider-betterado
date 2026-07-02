# Look up security role definitions for a given scope
data "betterado_securityrole_definitions" "environment" {
  scope = "distributedtask.environmentreferencerole"
}

output "environment_roles" {
  value = data.betterado_securityrole_definitions.environment.definitions
}
