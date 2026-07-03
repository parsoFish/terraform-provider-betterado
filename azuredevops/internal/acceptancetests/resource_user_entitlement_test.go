//go:build (all || resource_user_entitlement) && !exclude_resource_user_entitlement

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

func TestAccUserEntitlement_Create(t *testing.T) {
	if os.Getenv("AZDO_TEST_AAD_USER_EMAIL") == "" {
		t.Skip("Skip test due to `AZDO_TEST_AAD_USER_EMAIL` not set")
	}
	tfNode := "betterado_user_entitlement.user"
	principalName := os.Getenv("AZDO_TEST_AAD_USER_EMAIL")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, &[]string{"AZDO_TEST_AAD_USER_EMAIL"}) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkUserEntitlementDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclUserEntitlementResource(principalName),
				Check: resource.ComposeTestCheckFunc(
					checkUserEntitlementExists(principalName),
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					captureUserEntitlementEvidence(tfNode),
				),
			},
			// Idempotency: re-plan must show no diff.
			{
				Config:             hclUserEntitlementResource(principalName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// Given the principalName of an AzDO userEntitlement, this will return a function that will check whether
// or not the userEntitlement (1) exists in the state and (2) exist in AzDO and (3) has the correct name.
// Uses getDirectClient() because ProtoV6ProviderFactories does not wire the SDKv2 Meta() singleton.
func checkUserEntitlementExists(expectedPrincipalName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources["betterado_user_entitlement.user"]
		if !ok {
			return fmt.Errorf("Did not find a UserEntitlement in the TF state")
		}

		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("failed to build ADO client: %v", err)
		}

		id, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("Parsing UserEntitlement ID, got %s: %v", res.Primary.ID, err)
		}

		userEntitlement, err := clients.MemberEntitleManagementClient.GetUserEntitlement(clients.Ctx,
			memberentitlementmanagement.GetUserEntitlementArgs{
				UserId: &id,
			})
		if err != nil {
			return fmt.Errorf("UserEntitlement with ID=%s cannot be found! Error=%v", id, err)
		}

		if !strings.EqualFold(*userEntitlement.User.PrincipalName, expectedPrincipalName) {
			return fmt.Errorf("UserEntitlement with ID=%s has PrincipalName=%s, but expected Name=%s",
				res.Primary.ID, *userEntitlement.User.PrincipalName, expectedPrincipalName)
		}

		return nil
	}
}

// checkUserEntitlementDestroyed verifies that all user entitlements referenced in the state are
// destroyed (status == none or deleted) after terraform destroy.
// Uses getDirectClient() because ProtoV6ProviderFactories does not wire the SDKv2 Meta() singleton.
func checkUserEntitlementDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("checkUserEntitlementDestroyed: failed to build ADO client: %v", err)
	}

	// verify that every user referenced in the state does not exist in AzDO
	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_user_entitlement" {
			continue
		}

		id, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("Parsing UserEntitlement ID, got %s: %v", res.Primary.ID, err)
		}

		userEntitlement, err := clients.MemberEntitleManagementClient.GetUserEntitlement(clients.Ctx,
			memberentitlementmanagement.GetUserEntitlementArgs{
				UserId: &id,
			})
		if err != nil {
			if utils.ResponseWasNotFound(err) {
				return nil
			}
			return fmt.Errorf("Bad: Get UserEntitlement: %+v", err)
		}

		if userEntitlement != nil && userEntitlement.AccessLevel != nil &&
			string(*userEntitlement.AccessLevel.Status) != "none" {
			return fmt.Errorf("Status should be none: %s", string(*userEntitlement.AccessLevel.Status))
		}
	}

	return nil
}

// captureUserEntitlementEvidence performs a real live API GET of the created user entitlement
// and persists the response as forge demo live-evidence (before the resource is destroyed).
// Best-effort: a capture failure never fails the test.
func captureUserEntitlementEvidence(tfNode string) resource.TestCheckFunc {
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
		userEntitlement, err := clients.MemberEntitleManagementClient.GetUserEntitlement(clients.Ctx,
			memberentitlementmanagement.GetUserEntitlementArgs{
				UserId: &id,
			})
		if err != nil || userEntitlement == nil {
			return nil
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/_apis/memberentitlementmanagement/userentitlements/%s?api-version=7.1", orgURL, id)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-user-entitlement", url, userEntitlement)
		return nil
	}
}

func hclUserEntitlementResource(principalName string) string {
	return fmt.Sprintf(`
resource "betterado_user_entitlement" "user" {
  principal_name       = "%s"
  account_license_type = "express"
}`, principalName)
}
