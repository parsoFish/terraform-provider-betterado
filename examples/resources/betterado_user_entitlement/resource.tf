resource "betterado_user_entitlement" "example" {
  principal_name       = "user@example.com"
  account_license_type = "express"
  licensing_source     = "account"
}
