package acceptancetests

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccCheckApproval_basic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)

	resourceType := "betterado_check_approval"
	tfCheckNode := resourceType + ".test"
	principalName := os.Getenv("AZDO_TEST_AAD_USER_EMAIL")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, &[]string{"AZDO_TEST_AAD_USER_EMAIL"}) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclCheckApprovalResourceBasic(projectID, principalName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttr(tfCheckNode, "requester_can_approve", "false"),
					resource.TestCheckResourceAttr(tfCheckNode, "timeout", "43200"),
					resource.TestCheckResourceAttr(tfCheckNode, "approvers.#", "1"),
				),
			},
		},
	})
}

func TestAccCheckApproval_complete(t *testing.T) {
	projectID := SharedFixtureProjectID(t)

	resourceType := "betterado_check_approval"
	tfCheckNode := resourceType + ".test"
	principalName := os.Getenv("AZDO_TEST_AAD_USER_EMAIL")
	azdoGroupName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclCheckApprovalResourceComplete(projectID, principalName, azdoGroupName),
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

	resourceType := "betterado_check_approval"
	tfCheckNode := resourceType + ".test"
	principalName := os.Getenv("AZDO_TEST_AAD_USER_EMAIL")
	azdoGroupName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, &[]string{"AZDO_TEST_AAD_USER_EMAIL"}) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclCheckApprovalResourceBasic(projectID, principalName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttr(tfCheckNode, "approvers.#", "1"),
				),
			},
			{
				Config: hclCheckApprovalResourceComplete(projectID, principalName, azdoGroupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttr(tfCheckNode, "approvers.#", "2"),
					resource.TestCheckResourceAttr(tfCheckNode, "version", "2"),
				),
			},
		},
	})
}

func hclCheckApprovalResourceBasic(projectID string, principalName string) string {
	return fmt.Sprintf(`
resource "betterado_serviceendpoint_generic" "test" {
  project_id            = %q
  service_endpoint_name = "serviceendpoint"
  description           = "test"
  server_url            = "https://test/"
  username              = "test"
  password              = "test"
}

data "betterado_users" "test" {
  principal_name = "%s"
}

resource "betterado_check_approval" "test" {
  project_id           = %q
  target_resource_id   = betterado_serviceendpoint_generic.test.id
  target_resource_type = "endpoint"

  requester_can_approve = false
  approvers = [
    one(data.betterado_users.test.users).id,
  ]
}
`, projectID, principalName, projectID)
}

func hclCheckApprovalResourceComplete(projectID string, principalName string, azdoGroupName string) string {
	return fmt.Sprintf(`
resource "betterado_serviceendpoint_generic" "test" {
  project_id            = %q
  service_endpoint_name = "serviceendpoint"
  description           = "test"
  server_url            = "https://test/"
  username              = "test"
  password              = "test"
}

data "betterado_users" "test" {
  principal_name = "%s"
}

resource "betterado_group" "test" {
  display_name = "%s"
}

resource "betterado_check_approval" "test" {
  project_id           = %q
  target_resource_id   = betterado_serviceendpoint_generic.test.id
  target_resource_type = "endpoint"

  requester_can_approve = true
  approvers = [
    one(data.betterado_users.test.users).id,
    betterado_group.test.origin_id,
  ]

  timeout = 40000
}
`, projectID, principalName, azdoGroupName, projectID)
}
