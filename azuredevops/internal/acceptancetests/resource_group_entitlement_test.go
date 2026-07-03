//go:build (all || resource_group_entitlement) && !exclude_resource_group_entitlement

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

func TestAccGroupEntitlement_Create(t *testing.T) {
	tfNode := "betterado_group_entitlement.test"
	displayName := "group-038c153d-c86e-443c-b6f6-3d97378025d0"
	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkGroupEntitlementDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGroupEntitlementResourceBasic(displayName),
				Check: resource.ComposeTestCheckFunc(
					checkGroupEntitlementExists(),
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					captureGroupEntitlementEvidence(tfNode),
				),
			},
			// Idempotency: re-plan must show no diff.
			{
				Config:             hclGroupEntitlementResourceBasic(displayName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccGroupEntitlement_AAD_Create(t *testing.T) {
	if os.Getenv("AZDO_TEST_AAD_GROUP_ID") == "" {
		t.Skip("Skip test due to `AZDO_TEST_AAD_GROUP_ID` not set")
	}
	tfNode := "betterado_group_entitlement.test"
	originId := os.Getenv("AZDO_TEST_AAD_GROUP_ID")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, &[]string{"AZDO_TEST_AAD_GROUP_ID"}) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkGroupEntitlementDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclGroupEntitlementResourceAAD(originId),
				Check: resource.ComposeTestCheckFunc(
					checkGroupEntitlementExists(),
					resource.TestCheckResourceAttrSet(tfNode, "descriptor"),
					resource.TestCheckResourceAttr(tfNode, "origin_id", originId),
				),
			},
		},
	})
}

// checkGroupEntitlementExists verifies that the group entitlement (1) exists in
// the state and (2) exists in AzDO.
// Uses getDirectClient() because ProtoV6ProviderFactories does not wire the SDKv2 Meta() singleton.
func checkGroupEntitlementExists() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources["betterado_group_entitlement.test"]
		if !ok {
			return fmt.Errorf("Did not find a GroupEntitlement in the TF state")
		}

		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("failed to build ADO client: %v", err)
		}

		id, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("Parsing GroupEntitlement ID, got %s: %v", res.Primary.ID, err)
		}

		groupEntitlement, err := clients.MemberEntitleManagementClient.GetGroupEntitlement(clients.Ctx,
			memberentitlementmanagement.GetGroupEntitlementArgs{
				GroupId: &id,
			})
		if err != nil {
			return fmt.Errorf("GroupEntitlement with ID=%s cannot be found! Error=%v", id, err)
		}

		if groupEntitlement == nil || groupEntitlement.Id == nil {
			return fmt.Errorf("GroupEntitlement with ID=%s cannot be found", id)
		}

		return nil
	}
}

// checkGroupEntitlementDestroyed verifies that all group entitlements referenced
// in the state are destroyed after terraform destroy.
// Uses getDirectClient() because ProtoV6ProviderFactories does not wire the SDKv2 Meta() singleton.
func checkGroupEntitlementDestroyed(s *terraform.State) error {
	clients, err := getDirectClient()
	if err != nil {
		return fmt.Errorf("checkGroupEntitlementDestroyed: failed to build ADO client: %v", err)
	}

	// verify that every group referenced in the state does not exist in AzDO
	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_group_entitlement" {
			continue
		}

		id, err := uuid.Parse(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("Parsing GroupEntitlement ID, got %s: %v", res.Primary.ID, err)
		}

		groupEntitlement, err := clients.MemberEntitleManagementClient.GetGroupEntitlement(clients.Ctx,
			memberentitlementmanagement.GetGroupEntitlementArgs{
				GroupId: &id,
			})
		if err != nil {
			if utils.ResponseWasNotFound(err) {
				return nil
			}
			return fmt.Errorf("Bad: Get GroupEntitlement: %+v", err)
		}

		if groupEntitlement != nil && groupEntitlement.LicenseRule != nil &&
			string(*groupEntitlement.LicenseRule.Status) != "none" {
			return fmt.Errorf("Status should be none: %s", string(*groupEntitlement.LicenseRule.Status))
		}
	}

	return nil
}

// captureGroupEntitlementEvidence performs a real live API GET of the created
// group entitlement and persists the response as forge demo live-evidence
// (before the resource is destroyed). Best-effort: a capture failure never
// fails the test.
func captureGroupEntitlementEvidence(tfNode string) resource.TestCheckFunc {
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
		groupEntitlement, err := clients.MemberEntitleManagementClient.GetGroupEntitlement(clients.Ctx,
			memberentitlementmanagement.GetGroupEntitlementArgs{
				GroupId: &id,
			})
		if err != nil || groupEntitlement == nil {
			return nil
		}
		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf("%s/_apis/memberentitlementmanagement/groupentitlements/%s?api-version=7.1", orgURL, id)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-group-entitlement", url, groupEntitlement)
		return nil
	}
}

func hclGroupEntitlementResourceBasic(displayName string) string {
	return fmt.Sprintf(`
resource "betterado_group_entitlement" "test" {
  display_name         = "%s"
  account_license_type = "express"
}`, displayName)
}

func hclGroupEntitlementResourceAAD(originId string) string {
	return fmt.Sprintf(`
resource "betterado_group_entitlement" "test" {
  origin_id            = "%s"
  origin               = "aad"
  account_license_type = "express"
}`, originId)
}
