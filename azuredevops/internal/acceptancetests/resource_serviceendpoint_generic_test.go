//go:build (all || resource_serviceendpoint_generic) && !exclude_resource_serviceendpoint_generic

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

// TestAccServiceEndpointGeneric_basic verifies the framework-migrated
// betterado_serviceendpoint_generic resource end-to-end:
//
//  1. apply — creates a Generic service endpoint in the pre-existing shared project
//  2. read-back — asserts name, url, project_id + captures live evidence
//  3. idempotency — re-plan produces no diff (ExpectNonEmptyPlan: false)
//  4. destroy — endpoint is removed cleanly
//
// Uses SharedFixtureProjectName (betterado-standing-demo) so no new ADO project
// is created — the org is at the 1000-project cap; project creates would fail.
func TestAccServiceEndpointGeneric_basic(t *testing.T) {
	serviceEndpointName := testutils.GenerateResourceName()

	resourceType := "betterado_serviceendpoint_generic"
	tfSvcEpNode := resourceType + ".test"

	config := hclSvcEndpointGenericBasic(serviceEndpointName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             checkServiceEndpointGenericDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "server_url", "https://some-server.example.com"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointName),
					captureServiceEndpointGenericEvidence(tfSvcEpNode),
				),
			},
			{
				// idempotency: re-plan must produce no diff
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// hclSvcEndpointGenericBasic returns HCL for a minimal Generic service endpoint
// using the pre-existing shared project (looked up by name via data source).
func hclSvcEndpointGenericBasic(serviceEndpointName string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_serviceendpoint_generic" "test" {
  project_id            = data.betterado_project.test.id
  service_endpoint_name = %[1]q
  server_url            = "https://some-server.example.com"
}
`, serviceEndpointName, SharedFixtureProjectName)
}

// checkServiceEndpointGenericDestroyed verifies all generic service endpoints in
// the state have been deleted. Uses getDirectClient (not testutils.GetProvider().Meta())
// because ProtoV6ProviderFactories does not wire the SDKv2 provider singleton's Meta.
func checkServiceEndpointGenericDestroyed(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("checkServiceEndpointGenericDestroyed: build client: %w", err)
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

// captureServiceEndpointGenericEvidence performs a real live API GET of the
// created service endpoint and persists the response as forge demo live-evidence
// (before destroy). Label "acceptance-resource-generic" matches the forge
// unifier checkpoint. Best-effort: a capture failure never fails the test.
func captureServiceEndpointGenericEvidence(tfNode string) resource.TestCheckFunc {
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
		_ = testutils.CaptureLiveEvidence("acceptance-resource-generic", endpointURL, ep)
		return nil
	}
}
