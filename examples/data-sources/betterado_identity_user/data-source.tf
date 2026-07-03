# Look up an identity user by name (using the default General search filter).
data "betterado_identity_user" "example" {
  name = "john.doe@example.com"
}

output "user_descriptor" {
  value = data.betterado_identity_user.example.descriptor
}
