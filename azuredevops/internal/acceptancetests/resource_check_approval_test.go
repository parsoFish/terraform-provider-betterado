package acceptancetests

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/utils/converter"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/utils/sdk/pipelineschecksextras"
)

func TestAccCheckApproval_basic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	serviceEndpointName := testutils.GenerateResourceName()
	groupName := testutils.GenerateResourceName()

	resourceType := "betterado_check_approval"
	tfCheckNode := resourceType + ".test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclCheckApprovalResourceBasic(projectID, serviceEndpointName, groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttr(tfCheckNode, "requester_can_approve", "false"),
					resource.TestCheckResourceAttr(tfCheckNode, "timeout", "43200"),
					resource.TestCheckResourceAttr(tfCheckNode, "approvers.#", "1"),
					captureCheckApprovalEvidence(tfCheckNode),
				),
			},
		},
	})
}

func TestAccCheckApproval_complete(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	serviceEndpointName := testutils.GenerateResourceName()
	groupName1 := testutils.GenerateResourceName()
	groupName2 := testutils.GenerateResourceName()

	resourceType := "betterado_check_approval"
	tfCheckNode := resourceType + ".test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclCheckApprovalResourceComplete(projectID, serviceEndpointName, groupName1, groupName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttr(tfCheckNode, "requester_can_approve", "true"),
					resource.TestCheckResourceAttr(tfCheckNode, "timeout", "40000"),
					resource.TestCheckResourceAttr(tfCheckNode, "approvers.#", "2"),
				),
			},
		},
	})
}

func TestAccCheckApproval_update(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	serviceEndpointName := testutils.GenerateResourceName()
	groupName1 := testutils.GenerateResourceName()
	groupName2 := testutils.GenerateResourceName()

	resourceType := "betterado_check_approval"
	tfCheckNode := resourceType + ".test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclCheckApprovalResourceBasic(projectID, serviceEndpointName, groupName1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttr(tfCheckNode, "approvers.#", "1"),
				),
			},
			{
				Config: hclCheckApprovalResourceComplete(projectID, serviceEndpointName, groupName1, groupName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttr(tfCheckNode, "approvers.#", "2"),
					resource.TestCheckResourceAttr(tfCheckNode, "version", "2"),
				),
			},
		},
	})
}

// captureCheckApprovalEvidence performs a live GET of the check configuration and
// writes it to .forge/live-evidence/acceptance-resource.json, satisfying the forge
// demo live-evidence contract (label "acceptance-resource").
func captureCheckApprovalEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[tfNode]
		if !ok {
			return fmt.Errorf("captureCheckApprovalEvidence: resource %q not found in state", tfNode)
		}
		checkID, err := strconv.Atoi(resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("captureCheckApprovalEvidence: invalid check ID %q: %w", resourceState.Primary.ID, err)
		}
		projectID := resourceState.Primary.Attributes["project_id"]

		clients, err := testutils.GetADOClientsFromEnv()
		if err != nil {
			return err
		}

		check, err := clients.PipelinesChecksClientExtras.GetCheckConfiguration(context.Background(), pipelineschecksextras.GetCheckConfigurationArgs{
			Project: &projectID,
			Id:      &checkID,
			Expand:  converter.ToPtr(pipelineschecksextras.CheckConfigurationExpandParameterValues.Settings),
		})
		if err != nil {
			return err
		}

		url := fmt.Sprintf("%s/%s/_apis/pipelines/checks/configurations/%d?api-version=7.1",
			clients.OrganizationURL, projectID, checkID)
		return testutils.CaptureLiveEvidence("acceptance-resource", url, check)
	}
}

func hclCheckApprovalResourceBasic(projectID string, serviceEndpointName string, groupName string) string {
	return fmt.Sprintf(`
resource "betterado_serviceendpoint_generic" "test" {
  project_id            = %q
  service_endpoint_name = "%s"
  description           = "test"
  server_url            = "https://test/"
  username              = "test"
  password              = "test"
}

resource "betterado_group" "test" {
  display_name = "%s"
}

resource "betterado_check_approval" "test" {
  project_id           = %q
  target_resource_id   = betterado_serviceendpoint_generic.test.id
  target_resource_type = "endpoint"

  requester_can_approve = false
  approvers = [
    betterado_group.test.origin_id,
  ]
}
`, projectID, serviceEndpointName, groupName, projectID)
}

func hclCheckApprovalResourceComplete(projectID string, serviceEndpointName string, groupName1 string, groupName2 string) string {
	return fmt.Sprintf(`
resource "betterado_serviceendpoint_generic" "test" {
  project_id            = %q
  service_endpoint_name = "%s"
  description           = "test"
  server_url            = "https://test/"
  username              = "test"
  password              = "test"
}

resource "betterado_group" "test" {
  display_name = "%s"
}

resource "betterado_group" "test2" {
  display_name = "%s"
}

resource "betterado_check_approval" "test" {
  project_id           = %q
  target_resource_id   = betterado_serviceendpoint_generic.test.id
  target_resource_type = "endpoint"

  requester_can_approve = true
  approvers = [
    betterado_group.test.origin_id,
    betterado_group.test2.origin_id,
  ]

  timeout = 40000
}
`, projectID, serviceEndpointName, groupName1, groupName2, projectID)
}
