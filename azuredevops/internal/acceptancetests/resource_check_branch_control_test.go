package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccCheckBranchControl_basic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	checkName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()
	branches := "refs/heads/main"

	resourceType := "betterado_check_branch_control"
	tfCheckNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclBranchControlCheckResourceBasic(projectID, serviceEndpointName, checkName, branches),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckPipelineCheckExistsWithName(tfCheckNode, checkName),
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttr(tfCheckNode, "allowed_branches", branches),
					resource.TestCheckResourceAttr(tfCheckNode, "display_name", checkName),
					resource.TestCheckResourceAttr(tfCheckNode, "timeout", "1440"),
				),
			},
		},
	})
}

func TestAccCheckBranchControl_complete(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	checkName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()
	branches := "refs/heads/main"

	resourceType := "betterado_check_branch_control"
	tfCheckNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclBranchControlCheckResourceComplete(projectID, serviceEndpointName, checkName, branches),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckPipelineCheckExistsWithName(tfCheckNode, checkName),
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttr(tfCheckNode, "allowed_branches", branches),
					resource.TestCheckResourceAttr(tfCheckNode, "display_name", checkName),
					resource.TestCheckResourceAttr(tfCheckNode, "timeout", "1440"),
					resource.TestCheckResourceAttr(tfCheckNode, "verify_branch_protection", "true"),
					resource.TestCheckResourceAttr(tfCheckNode, "ignore_unknown_protection_status", "false"),
				),
			},
		},
	})
}

func TestAccCheckBranchControl_update(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	serviceEndpointName := testutils.GenerateResourceName()
	checkNameFirst := testutils.GenerateResourceName()
	branchesFirst := "refs/heads/main"

	checkNameSecond := testutils.GenerateResourceName()
	branchesSecond := "refs/heads/master"

	resourceType := "betterado_check_branch_control"
	tfCheckNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclBranchControlCheckResourceBasic(projectID, serviceEndpointName, checkNameFirst, branchesFirst),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckPipelineCheckExistsWithName(tfCheckNode, checkNameFirst),
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttr(tfCheckNode, "allowed_branches", branchesFirst),
					resource.TestCheckResourceAttr(tfCheckNode, "display_name", checkNameFirst),
				),
			},
			{
				Config: hclBranchControlCheckResourceUpdate(projectID, serviceEndpointName, checkNameSecond, branchesSecond),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckPipelineCheckExistsWithName(tfCheckNode, checkNameSecond),
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttr(tfCheckNode, "allowed_branches", branchesSecond),
					resource.TestCheckResourceAttr(tfCheckNode, "display_name", checkNameSecond),
					resource.TestCheckResourceAttr(tfCheckNode, "version", "2"),
				),
			},
		},
	})
}

func hclBranchControlCheckResourceBasic(projectID string, serviceEndpointName string, checkName string, branches string) string {
	return fmt.Sprintf(`
resource "betterado_serviceendpoint_generic" "test" {
  project_id            = %q
  service_endpoint_name = "%s"
  description           = "test"
  server_url            = "https://test/"
  username              = "test"
  password              = "test"
}

resource "betterado_check_branch_control" "test" {
  project_id           = %q
  display_name         = "%s"
  target_resource_id   = betterado_serviceendpoint_generic.test.id
  allowed_branches     = "%s"
  target_resource_type = "endpoint"
}`, projectID, serviceEndpointName, projectID, checkName, branches)
}

func hclBranchControlCheckResourceComplete(projectID string, serviceEndpointName string, checkName string, branches string) string {
	return fmt.Sprintf(`
resource "betterado_serviceendpoint_generic" "test" {
  project_id            = %q
  service_endpoint_name = "%s"
  description           = "test"
  server_url            = "https://test/"
  username              = "test"
  password              = "test"
}

resource "betterado_check_branch_control" "test" {
  project_id                       = %q
  display_name                     = "%s"
  target_resource_id               = betterado_serviceendpoint_generic.test.id
  allowed_branches                 = "%s"
  verify_branch_protection         = true
  ignore_unknown_protection_status = false
  target_resource_type             = "endpoint"
}`, projectID, serviceEndpointName, projectID, checkName, branches)
}

func hclBranchControlCheckResourceUpdate(projectID string, serviceEndpointName string, checkName string, branches string) string {
	return fmt.Sprintf(`
resource "betterado_serviceendpoint_generic" "test" {
  project_id            = %q
  service_endpoint_name = "%s"
  description           = "test"
  server_url            = "https://test/"
  username              = "test"
  password              = "test"
}

resource "betterado_check_branch_control" "test" {
  project_id                       = %q
  display_name                     = "%s"
  target_resource_id               = betterado_serviceendpoint_generic.test.id
  target_resource_type             = "endpoint"
  allowed_branches                 = "%s"
  verify_branch_protection         = true
  ignore_unknown_protection_status = false
  timeout                          = 50000
}`, projectID, serviceEndpointName, projectID, checkName, branches)
}
