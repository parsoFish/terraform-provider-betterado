# List all accounts accessible to the authenticated user
data "betterado_accounts" "all" {}

output "account_names" {
  value = [for a in data.betterado_accounts.all.accounts : a.account_name]
}
