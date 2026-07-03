resource "betterado_group_entitlement" "example" {
  display_name         = "MyTeamGroup"
  origin               = "vsts"
  origin_id            = "00000000-0000-0000-0000-000000000000"
  account_license_type = "express"
  licensing_source     = "account"
}
