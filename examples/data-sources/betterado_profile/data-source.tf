# Look up the authenticated user's profile
data "betterado_profile" "me" {
  id = "me"
}

output "display_name" {
  value = data.betterado_profile.me.display_name
}
