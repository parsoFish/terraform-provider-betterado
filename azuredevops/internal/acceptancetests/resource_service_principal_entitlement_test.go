//go:build (all || resource_service_principal_entitlement) && !exclude_resource_service_principal_entitlement

package acceptancetests

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/memberentitlementmanagement"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils"
)

func TestAccServicePrincipalEntitlement_create(t *testing.T) {
	if os.Getenv("AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID") == "" {
		t.Skip("Skip test due to `AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID` not set")
	}
	tfNode := "betterado_service_principal_entitlement.service_principal"
	ServicePrincipalId := os.Getenv("AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, &[]string{"AZDO_TEST_AAD_SERVICE_PRINCIPAL_OBJECT_ID"}) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkServicePrincipalEntitlementDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclServicePrincipalEntitlementResource(ServicePrincipalId),
				Check: resource.ComposeTestCheckFunc(
					checkServicePrincipalEntitlementExists(ServicePrincipalId),
					resource.TestCheckResourceAttrSet(tfNode, "display_name"),
					resource.TestCheckResourceAttrSet(tfNode, "origin_id"),
					captureServicePrincipalEntitlementEvidence(tfNode),
				),
			},
			// Idempotency: re-plan must show no diff.
			{
				Config:             hclServicePrincipalEntitlementResource(ServicePrincipalId),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// checkServicePrincipalEntitlementExists verifies that the service principal entitlement
// (1) exists in the state, (2) exists in AzDO, and (3) has the correct OriginId.
// Uses getDirectClient() because ProtoV6ProviderFactories does not wire the SDKv2 Meta() singleton.
func checkServicePrincipalEntitlementExists(expectedServicePrincipalId string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources["betterado_service_principal_entitlement.service_principal"]
		if !ok {
			return fmt.Errorf("Did not find a ServicePrincipalEntitlement in the TF state")
		}

		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("failed to build ADO client: %v", err)
		}

		id, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("Parsing ServicePrincipalEntitlement ID, got %s: %v", res.Primary.ID, err)
		}

		servicePrincipalEntitlement, err := clients.MemberEntitleManagementClient.GetServicePrincipalEntitlement(clients.Ctx, memberentitlementmanagement.GetServicePrincipalEntitlementArgs{
			ServicePrincipalId: &id,
		})
		if err != nil {
			return fmt.Errorf("ServicePrincipalEntitlement with ID=%s cannot be found!. Error=%v", id, err)
		}

		if !strings.EqualFold(strings.ToLower(*servicePrincipalEntitlement.ServicePrincipal.OriginId), strings.ToLower(expectedServicePrincipalId)) {
			return fmt.Errorf("ServicePrincipalEntitlement with ID=%s has OriginId=%s, but expected OriginId=%s", res.Primary.ID, *servicePrincipalEntitlement.ServicePrincipal.OriginId, expectedServicePrincipalId)
		}

		return nil
	}
}

// checkServicePrincipalEntitlementDestroyed verifies that all service principal
// entitlements referenced in the state are destroyed after terraform destroy.
// Uses getDirectClient() because ProtoV6ProviderFactories does not wire the SDKv2 Meta() singleton.
func checkServicePrincipalEntitlementDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("checkServicePrincipalEntitlementDestroyed: failed to build ADO client: %v", err)
	}

	// verify that every service principal referenced in the state does not exist in AzDO
	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_service_principal_entitlement" {
			continue
		}

		id, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("Parsing ServicePrincipalEntitlement ID, got %s: %v", res.Primary.ID, err)
		}

		servicePrincipalEntitlement, err := clients.MemberEntitleManagementClient.GetServicePrincipalEntitlement(clients.Ctx, memberentitlementmanagement.GetServicePrincipalEntitlementArgs{
			ServicePrincipalId: &id,
		})
		if err != nil {
			if utils.ResponseWasNotFound(err) {
				return nil
			}
			return fmt.Errorf("Bad: Get ServicePrincipalEntitlement: %+v", err)
		}

		if servicePrincipalEntitlement != nil && servicePrincipalEntitlement.AccessLevel != nil && string(*servicePrincipalEntitlement.AccessLevel.Status) != "none" {
			return fmt.Errorf("Status should be none: %s", string(*servicePrincipalEntitlement.AccessLevel.Status))
		}
	}

	return nil
}

// captureServicePrincipalEntitlementEvidence performs a real live API GET of the
// created service principal entitlement and persists the response as forge demo
// live-evidence (before the resource is destroyed). Best-effort: a capture
// failure never fails the test.
func captureServicePrincipalEntitlementEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		id, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return nil
		}
		clients, err := getDirectClient()
		if err != nil {
			return nil // best-effort: client build failure does not fail the test
		}
		spe, err := clients.MemberEntitleManagementClient.GetServicePrincipalEntitlement(clients.Ctx,
			memberentitlementmanagement.GetServicePrincipalEntitlementArgs{
				ServicePrincipalId: &id,
			})
		if err != nil || spe == nil {
			return nil
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/_apis/memberentitlementmanagement/serviceprincipals/%s?api-version=7.1", orgURL, id)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-sp-entitlement", url, spe)
		return nil
	}
}

func hclServicePrincipalEntitlementResource(servicePrincipalId string) string {
	return fmt.Sprintf(`
resource "betterado_service_principal_entitlement" "service_principal" {
  origin_id            = "%s"
  origin               = "aad"
  account_license_type = "express"
}`, servicePrincipalId)
}
