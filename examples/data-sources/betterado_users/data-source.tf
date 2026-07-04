# List all users originating from Azure AD.
data "betterado_users" "aad_users" {
  origin = "aad"
}

output "aad_user_count" {
  value = length(data.betterado_users.aad_users.users)
}
