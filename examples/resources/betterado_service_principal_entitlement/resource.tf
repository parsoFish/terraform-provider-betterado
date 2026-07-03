resource "betterado_service_principal_entitlement" "example" {
  origin_id            = "00000000-0000-0000-0000-000000000000"
  origin               = "aad"
  account_license_type = "express"
  licensing_source     = "account"
}
