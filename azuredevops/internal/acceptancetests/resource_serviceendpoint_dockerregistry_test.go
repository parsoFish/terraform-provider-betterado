//go:build (all || resource_serviceendpoint_dockerregistry) && !exclude_resource_serviceendpoint_dockerregistry

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

// TestAccServiceEndpointDockerRegistry_basic verifies the framework-migrated
// betterado_serviceendpoint_dockerregistry resource end-to-end:
//
//  1. apply — creates a DockerRegistry service endpoint in the pre-existing shared project
//  2. read-back — asserts name, project_id + captures live evidence
//  3. idempotency — re-plan produces no diff (ExpectNonEmptyPlan: false)
//  4. destroy — endpoint is removed cleanly
//
// Uses SharedFixtureProjectName (betterado-standing-demo) so no new ADO project
// is created — the org is at the 1000-project cap; project creates would fail.
func TestAccServiceEndpointDockerRegistry_basic(t *testing.T) {
	serviceEndpointName := testutils.GenerateResourceName()

	resourceType := "betterado_serviceendpoint_dockerregistry"
	tfSvcEpNode := resourceType + ".test"

	config := hclSvcEndpointDockerRegistryBasic(serviceEndpointName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkServiceEndpointDockerRegistryDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointName),
					resource.TestCheckResourceAttr(tfSvcEpNode, "docker_username", "testuser"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "docker_email", "test@email.com"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "registry_type", "DockerHub"),
					captureServiceEndpointDockerRegistryEvidence(tfSvcEpNode),
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

// TestAccServiceEndpointDockerRegistry_CreateAndUpdate validates apply + update.
func TestAccServiceEndpointDockerRegistry_CreateAndUpdate(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	serviceEndpointNameFirst := testutils.GenerateResourceName()
	serviceEndpointNameSecond := testutils.GenerateResourceName()

	resourceType := "betterado_serviceendpoint_dockerregistry"
	tfSvcEpNode := resourceType + ".serviceendpoint"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
		},
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkServiceEndpointDockerRegistryDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: testutils.HclServiceEndpointDockerRegistryResource(projectName, serviceEndpointNameFirst),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "docker_username", "testuser"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "docker_email", "test@email.com"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "docker_password", "secret"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameFirst),
				),
			}, {
				Config: testutils.HclServiceEndpointDockerRegistryResource(projectName, serviceEndpointNameSecond),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "docker_username"),
					resource.TestCheckResourceAttrSet(tfSvcEpNode, "docker_email"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "docker_password", "secret"),
					resource.TestCheckResourceAttr(tfSvcEpNode, "service_endpoint_name", serviceEndpointNameSecond),
				),
			},
		},
	})
}

// hclSvcEndpointDockerRegistryBasic returns HCL for a minimal DockerRegistry service endpoint
// using the pre-existing shared project (looked up by name via data source).
func hclSvcEndpointDockerRegistryBasic(serviceEndpointName string) string {
	return fmt.Sprintf(`
data "betterado_project" "test" {
  name = %[2]q
}

resource "betterado_serviceendpoint_dockerregistry" "test" {
  project_id            = data.betterado_project.test.id
  service_endpoint_name = %[1]q
  docker_username       = "testuser"
  docker_email          = "test@email.com"
  docker_password       = "testpassword"
  registry_type         = "DockerHub"
}
`, serviceEndpointName, SharedFixtureProjectName)
}

// checkServiceEndpointDockerRegistryDestroyed verifies all dockerregistry service endpoints in
// the state have been deleted. Uses getDirectClient (not testutils.GetProvider().Meta())
// because ProtoV6ProviderFactories does not wire the SDKv2 provider singleton's Meta.
func checkServiceEndpointDockerRegistryDestroyed(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		clients, err := getDirectClient()
		if err != nil {
			return fmt.Errorf("checkServiceEndpointDockerRegistryDestroyed: build client: %w", err)
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

// captureServiceEndpointDockerRegistryEvidence performs a real live API GET of the
// created service endpoint and persists the response as forge demo live-evidence
// (before destroy). Label "acceptance-resource-dockerregistry" matches the forge
// unifier checkpoint. Best-effort: a capture failure never fails the test.
func captureServiceEndpointDockerRegistryEvidence(tfNode string) resource.TestCheckFunc {
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
		_ = testutils.CaptureLiveEvidence("acceptance-resource-dockerregistry", endpointURL, ep)
		return nil
	}
}
