resource "betterado_project" "example" {
  name               = "Example Project"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_serviceendpoint_azurerm" "example" {
  project_id                             = betterado_project.example.id
  service_endpoint_name                  = "example-service-connection"
  service_endpoint_authentication_scheme = "ServicePrincipal"
  credentials {
    serviceprincipalid  = "00000000-0000-0000-0000-000000000000"
    serviceprincipalkey = "supersecretkey"
  }
  azurerm_spn_tenantid      = "00000000-0000-0000-0000-000000000000"
  azurerm_subscription_id   = "00000000-0000-0000-0000-000000000000"
  azurerm_subscription_name = "Example Subscription"
}

resource "betterado_elastic_pool" "example" {
  name                   = "example-elastic-pool"
  service_endpoint_id    = betterado_serviceendpoint_azurerm.example.id
  service_endpoint_scope = betterado_project.example.id
  desired_idle           = 2
  max_capacity           = 5
  azure_resource_id      = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Compute/virtualMachineScaleSets/my-vmss"

  recycle_after_each_use = true
  agent_interactive_ui   = false
  time_to_live_minutes   = 30
  auto_provision         = false
  auto_update            = true
}
