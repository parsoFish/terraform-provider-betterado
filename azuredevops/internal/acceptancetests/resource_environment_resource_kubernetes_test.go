package acceptancetests

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/taskagent"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
)

// TestAccEnvironmentResourceKubernetes_createUpdate verifies the full lifecycle
// (create, read-back with idempotency check, update name, destroy) of a
// betterado_environment_resource_kubernetes resource via the framework provider.
func TestAccEnvironmentResourceKubernetes_createUpdate(t *testing.T) {
	environmentName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()
	resourceNameFirst := testutils.GenerateResourceName()
	resourceNameSecond := testutils.GenerateResourceName()
	tfNode := "betterado_environment_resource_kubernetes.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             checkEnvironmentKubernetesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: hclEnvironmentKubernetes(environmentName, serviceEndpointName, resourceNameFirst),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", resourceNameFirst),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "environment_id"),
					resource.TestCheckResourceAttrSet(tfNode, "service_endpoint_id"),
					resource.TestCheckResourceAttr(tfNode, "namespace", "test"),
					resource.TestCheckResourceAttr(tfNode, "cluster_name", "k8scluster"),
					checkEnvironmentKubernetesExists(tfNode, resourceNameFirst),
					captureEnvironmentKubernetesEvidence(tfNode),
				),
				ExpectNonEmptyPlan: false,
			},
			{
				Config: hclEnvironmentKubernetes(environmentName, serviceEndpointName, resourceNameSecond),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(tfNode, "name", resourceNameSecond),
					resource.TestCheckResourceAttrSet(tfNode, "project_id"),
					resource.TestCheckResourceAttrSet(tfNode, "environment_id"),
					resource.TestCheckResourceAttrSet(tfNode, "service_endpoint_id"),
					resource.TestCheckResourceAttr(tfNode, "namespace", "test"),
					resource.TestCheckResourceAttr(tfNode, "cluster_name", "k8scluster"),
					checkEnvironmentKubernetesExists(tfNode, resourceNameSecond),
				),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// checkEnvironmentKubernetesExists verifies that the resource:
// (1) exists in the Terraform state and
// (2) exists in AzDO with the correct name.
func checkEnvironmentKubernetesExists(tfNode string, expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return fmt.Errorf("Did not find an resource in the TF state")
		}

		clients, err := testutils.GetDirectClient()
		if err != nil {
			return fmt.Errorf("building direct client: %v", err)
		}
		id, err := strconv.Atoi(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("Parse resource id, ID:  %v !. Error= %v", res.Primary.ID, err)
		}
		projectId := res.Primary.Attributes["project_id"]
		environmentIdStr := res.Primary.Attributes["environment_id"]
		environmentId, err := strconv.Atoi(environmentIdStr)
		if err != nil {
			return fmt.Errorf("Parse environment_id error, ID:  %v !. Error= %v", environmentIdStr, err)
		}

		readResource, err := readEnvironmentKubernetes(clients, projectId, environmentId, id)
		if err != nil {
			return fmt.Errorf("Resource with ID=%d cannot be found!. Error=%v", id, err)
		}

		if *readResource.Name != expectedName {
			return fmt.Errorf("Resource with ID=%d has Name=%s, but expected Name=%s", id, *readResource.Name, expectedName)
		}
		return nil
	}
}

// checkEnvironmentKubernetesDestroyed verifies that every kubernetes resource
// referenced in the state is gone after destroy.
func checkEnvironmentKubernetesDestroyed(s *terraform.State) error {
	clients, err := testutils.GetDirectClient()
	if err != nil {
		return fmt.Errorf("building direct client: %v", err)
	}

	for _, res := range s.RootModule().Resources {
		if res.Type != "betterado_environment_resource_kubernetes" {
			continue
		}

		id, err := strconv.Atoi(res.Primary.ID)
		if err != nil {
			return fmt.Errorf("Parse resource id, ID:  %v !. Error= %v", res.Primary.ID, err)
		}
		projectId := res.Primary.Attributes["project_id"]
		environmentIdStr := res.Primary.Attributes["environment_id"]
		environmentId, err := strconv.Atoi(environmentIdStr)
		if err != nil {
			return fmt.Errorf("Parse environment_id error, ID:  %v !. Error= %v", environmentIdStr, err)
		}

		// indicates the environment still exists - this should fail the test
		if _, err := readEnvironmentKubernetes(clients, projectId, environmentId, id); err == nil {
			return fmt.Errorf("Resource ID %d should not exist", id)
		}
	}
	return nil
}

// readEnvironmentKubernetes looks up a KubernetesResource by project, environment, and resource ID.
func readEnvironmentKubernetes(clients *client.AggregatedClient, projectId string, environmentId int, resourceId int) (*taskagent.KubernetesResource, error) {
	return clients.TaskAgentClient.GetKubernetesResource(
		clients.Ctx,
		taskagent.GetKubernetesResourceArgs{
			Project:       &projectId,
			EnvironmentId: &environmentId,
			ResourceId:    &resourceId,
		},
	)
}

// captureEnvironmentKubernetesEvidence performs a live API GET of the Kubernetes resource
// and writes forge demo live-evidence to
// .forge/live-evidence/acceptance-resource-environment-kubernetes.json (AC3).
// Best-effort: a capture failure never fails the test.
func captureEnvironmentKubernetesEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		res, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return nil
		}
		resourceIDStr := res.Primary.ID
		projectID := res.Primary.Attributes["project_id"]
		environmentIDStr := res.Primary.Attributes["environment_id"]

		resourceID, parseErr := strconv.Atoi(resourceIDStr)
		if parseErr != nil {
			return nil //nolint:nilerr // best-effort: parse failure must not fail the test
		}
		environmentID, envParseErr := strconv.Atoi(environmentIDStr)
		if envParseErr != nil {
			return nil //nolint:nilerr // best-effort: parse failure must not fail the test
		}

		clients, clientErr := testutils.GetDirectClient()
		if clientErr != nil {
			return nil //nolint:nilerr // best-effort: client failure must not fail the test
		}

		k8sResource, readErr := readEnvironmentKubernetes(clients, projectID, environmentID, resourceID)
		if readErr != nil || k8sResource == nil {
			return nil //nolint:nilerr // best-effort: API failure must not fail the test
		}

		orgURL := strings.TrimRight(os.Getenv("AZDO_ORG_SERVICE_URL"), "/")
		url := fmt.Sprintf(
			"%s/%s/_apis/distributedtask/environments/%s/providers/kubernetes/%s?api-version=7.1",
			orgURL, projectID, environmentIDStr, resourceIDStr,
		)
		_ = testutils.CaptureLiveEvidence("acceptance-resource-environment-kubernetes", url, k8sResource)
		return nil
	}
}

// hclEnvironmentKubernetes creates a Kubernetes environment resource in the standing
// fixture project to avoid the 1000-project ADO org limit.
func hclEnvironmentKubernetes(environmentName, serviceEndpointName, k8sName string) string {
	return fmt.Sprintf(`
data "betterado_project" "fixture" {
  name = %[4]q
}

resource "betterado_environment" "test" {
  project_id = data.betterado_project.fixture.id
  name       = %[1]q
}

resource "betterado_serviceendpoint_kubernetes" "test" {
  project_id            = data.betterado_project.fixture.id
  service_endpoint_name = %[2]q
  apiserver_url         = "https://test-dns-r9lconkh.hcp.eastus.azmk8s.io:443"
  authorization_type    = "ServiceAccount"
  service_account {
    token   = "test"
    ca_cert = "test"
  }
}

resource "betterado_environment_resource_kubernetes" "test" {
  project_id          = data.betterado_project.fixture.id
  environment_id      = betterado_environment.test.id
  service_endpoint_id = betterado_serviceendpoint_kubernetes.test.id
  name                = %[3]q
  namespace           = "test"
  cluster_name        = "k8scluster"
  tags                = ["tag1", "tag2"]
}
`, environmentName, serviceEndpointName, k8sName, SharedFixtureProjectName)
}
