//go:build (all || resource_serviceendpoint_azurerm) && !exclude_resource_serviceendpoint_azurerm

package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/serviceendpoint"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

// TestAccServiceEndpointAzureRm_CreateAndUpdate verifies the framework-migrated
// betterado_serviceendpoint_azurerm resource end-to-end:
//
//  1. apply step 1 — creates an AzureRM service endpoint in the pre-existing shared project
//  2. read-back — asserts attributes + captures live evidence
//  3. idempotency — re-plan produces no diff (ExpectNonEmptyPlan: false)
//  4. apply step 2 — updates the endpoint name
//  5. destroy — endpoint is removed cleanly
//
// Uses SharedFixtureProjectName (betterado-standing-demo) so no new ADO project
// is created — the org is at the 1000-project cap; project creates would fail.
func TestAccServiceEndpointAzureRm_CreateAndUpdate(t *testing.T) {
	serviceEndpointNameFirst := testutils.GenerateResourceName()
	serviceEndpointNameSecond := testutils.GenerateResourceName()
	serviceprincipalidFirst := uuid.New().String()
	serviceprincipalidSecond := uuid.New().String()
	serviceprincipalkeyFirst := uuid.New().String()
	serviceprincipalkeySecond := uuid.New().String()

	resourceType := "betterado_serviceendpoint_azurerm"
	tfSvcEpNode := resourceType + ".serviceendpointrm"

	configFirst := hclSvcEndpointAzureRMManual(serviceEndpointNameFirst, serviceprincipalidFirst, serviceprincipalkeyFirst)
	configSecond := hclSvcEndpointAzureRMManual(serviceEndpointNameSecond, serviceprincipalidSecond, serviceprincipalkeySecond)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkServiceEndpointAzureRMDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: configFirst,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_spn_tenantid"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameFirst),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_subscription_id"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_subscription_name"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "credentials.0.serviceprincipalid", serviceprincipalidFirst),
					captureServiceEndpointAzureRMEvidence(tfSvcEpNode),
				),
			},
			{
				// idempotency: re-plan must produce no diff
				Config:             configFirst,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
			{
				Config: configSecond,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_spn_tenantid"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameSecond),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_subscription_id"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "azurerm_subscription_name"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "credentials.0.serviceprincipalid", serviceprincipalidSecond),
				),
			},
		},
	})
}

// hclSvcEndpointAzureRMManual returns HCL for a manual ServicePrincipal AzureRM
// service endpoint using the pre-existing shared project.
func hclSvcEndpointAzureRMManual(serviceEndpointName, servicePrincipalID, servicePrincipalKey string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[4]q
}

resource "betterado_serviceendpoint_azurerm" "serviceendpointrm" {
  project_id            = data.betterado_project.test.id
  service_endpoint_name = %[1]q
  credentials {
    serviceprincipalid  = %[2]q
    serviceprincipalkey = %[3]q
  }
  azurerm_spn_tenantid                   = "9c59cbe5-2ca1-4516-b303-8968a070edd2"
  azurerm_subscription_id                = "3b0fee91-c36d-4d70-b1e9-fc4b9d608c3d"
  azurerm_subscription_name              = "Microsoft Azure DEMO"
  service_endpoint_authentication_scheme = "ServicePrincipal"
}
`, serviceEndpointName, servicePrincipalID, servicePrincipalKey, SharedFixtureProjectName)
}

// checkServiceEndpointAzureRMDestroyed verifies all azurerm service endpoints
// in the state have been deleted. Uses getDirectClient because
// ProtoV6ProviderFactories does not wire the SDKv2 provider singleton's Meta.
func checkServiceEndpointAzureRMDestroyed(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("checkServiceEndpointAzureRMDestroyed: build client: %w", err)
		}
		for _, res := range s.RootModule().Resources {
			if res.Type != resourceType {
				continue
			}
			id, err := uuid.Parse(res.Primary.ID)
			if err != nil {
				return fmt.Errorf("invalid service endpoint ID %q: %w", res.Primary.ID, err)
			}
			projectID := res.Primary.Attributes["project_id"]
			ep, err := clients.ServiceEndpointClient.GetServiceEndpointDetails(clients.Ctx,
				serviceendpoint.GetServiceEndpointDetailsArgs{
					EndpointId: &id,
					Project:    &projectID,
				})
			if err == nil && ep != nil && ep.Id != nil {
				return fmt.Errorf("service endpoint %s still exists after destroy", id)
			}
		}
		return nil
	}
}

// captureServiceEndpointAzureRMEvidence performs a real live API GET of the
// created service endpoint and persists the response as forge demo live-evidence
// (before destroy). Label "acceptance-resource-azurerm" matches the forge
// unifier checkpoint. Best-effort: a capture failure never fails the test.
func captureServiceEndpointAzureRMEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		id, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return nil
		}
		projectID := res.Primary.Attributes["project_id"]
		clients, err := getDirectClient()
		if err != nil {
			return nil
		}
		ep, err := clients.ServiceEndpointClient.GetServiceEndpointDetails(clients.Ctx,
			serviceendpoint.GetServiceEndpointDetailsArgs{
				EndpointId: &id,
				Project:    &projectID,
			})
		if err != nil || ep == nil {
			return nil
		}
		orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
		endpointURL := fmt.Sprintf("%s/_apis/serviceendpoint/endpoints/%s?api-version=7.1", orgURL, id)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-azurerm", endpointURL, ep)
		return nil
	}
}
