package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccCheckEndpoint(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	checkName := testutils.GenerateResourceName()
	serviceEndpointName := testutils.GenerateResourceName()
	branches := "refs/heads/main"

	resourceType := "betterado_check_branch_control"
	tfCheckNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclBranchControlCheckResourceBasicEndpoint(projectID, serviceEndpointName, checkName, branches),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckPipelineCheckExistsWithName(tfCheckNode, checkName),
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttr(tfCheckNode, "allowed_branches", branches),
					resource.TestCheckResourceAttr(tfCheckNode, "display_name", checkName),
				),
			},
		},
	})
}

func hclBranchControlCheckResourceBasicEndpoint(projectID string, serviceEndpointName string, checkName string, branches string) string {
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

func TestAccCheckEnvironment(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	checkName := testutils.GenerateResourceName()
	branches := "refs/heads/main"

	resourceType := "betterado_check_branch_control"
	tfCheckNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclBranchControlCheckResourceBasicEnvironment(projectID, checkName, branches),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckPipelineCheckExistsWithName(tfCheckNode, checkName),
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttr(tfCheckNode, "allowed_branches", branches),
					resource.TestCheckResourceAttr(tfCheckNode, "display_name", checkName),
				),
			},
		},
	})
}

func hclBranchControlCheckResourceBasicEnvironment(projectID string, checkName string, branches string) string {
	return fmt.Sprintf(`
resource "betterado_environment" "environment" {
  project_id = %q
  name       = "environment_test"
}

resource "betterado_check_branch_control" "test" {
  project_id           = %q
  display_name         = "%s"
  target_resource_id   = betterado_environment.environment.id
  target_resource_type = "environment"
  allowed_branches     = "%s"
}`, projectID, projectID, checkName, branches)
}

func TestAccCheckQueue(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	checkName := testutils.GenerateResourceName()
	branches := "refs/heads/main"

	resourceType := "betterado_check_branch_control"
	tfCheckNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclBranchControlCheckResourceBasicQueue(projectID, checkName, branches),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckPipelineCheckExistsWithName(tfCheckNode, checkName),
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttr(tfCheckNode, "allowed_branches", branches),
					resource.TestCheckResourceAttr(tfCheckNode, "display_name", checkName),
				),
			},
		},
	})
}

func hclBranchControlCheckResourceBasicQueue(projectID string, checkName string, branches string) string {
	poolName := testutils.GenerateResourceName()
	agentPoolHCL := testutils.HclAgentPoolResource(poolName)
	return fmt.Sprintf(`
%s

resource "betterado_agent_queue" "q" {
  project_id    = %q
  agent_pool_id = betterado_agent_pool.pool.id
}

resource "betterado_check_branch_control" "test" {
  project_id           = %q
  display_name         = "%s"
  target_resource_id   = betterado_agent_queue.q.id
  target_resource_type = "queue"
  allowed_branches     = "%s"
}`, agentPoolHCL, projectID, projectID, checkName, branches)
}

func TestAccCheckRepo(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	checkName := testutils.GenerateResourceName()
	branches := "refs/heads/main"

	resourceType := "betterado_check_branch_control"
	tfCheckNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclBranchControlCheckResourceBasicRepo(projectID, checkName, branches),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckPipelineCheckExistsWithName(tfCheckNode, checkName),
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttr(tfCheckNode, "allowed_branches", branches),
					resource.TestCheckResourceAttr(tfCheckNode, "display_name", checkName),
				),
			},
		},
	})
}

func hclBranchControlCheckResourceBasicRepo(projectID string, checkName string, branches string) string {
	repoName := testutils.GenerateResourceName()
	return fmt.Sprintf(`
resource "betterado_git_repository" "repository" {
  project_id = %q
  name       = "%s"
  initialization {
    init_type = "Clean"
  }
}

resource "betterado_check_branch_control" "test" {
  project_id           = %q
  display_name         = "%s"
  target_resource_id   = "%s.${betterado_git_repository.repository.id}"
  target_resource_type = "repository"
  allowed_branches     = "%s"
}`, projectID, repoName, projectID, checkName, projectID, branches)
}

func TestAccCheckVariableGroup(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	checkName := testutils.GenerateResourceName()
	branches := "refs/heads/main"

	resourceType := "betterado_check_branch_control"
	tfCheckNode := resourceType + ".test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		CheckDestroy:             testutils.CheckPipelineCheckDestroyed(resourceType),
		Steps: []resource.TestStep{
			{
				Config: hclBranchControlCheckResourceBasicVariableGroup(projectID, checkName, branches),
				Check: resource.ComposeTestCheckFunc(
					testutils.CheckPipelineCheckExistsWithName(tfCheckNode, checkName),
					resource.TestCheckResourceAttrSet(tfCheckNode, "project_id"),
					resource.TestCheckResourceAttr(tfCheckNode, "allowed_branches", branches),
					resource.TestCheckResourceAttr(tfCheckNode, "display_name", checkName),
				),
			},
		},
	})
}

func hclBranchControlCheckResourceBasicVariableGroup(projectID string, checkName string, branches string) string {
	variableGroupName := testutils.GenerateResourceName()
	return fmt.Sprintf(`
resource "betterado_variable_group" "vg" {
  project_id   = %q
  name         = "%s"
  description  = "A sample variable group."
  allow_access = true
  variable {
    name         = "key1"
    secret_value = "value1"
    is_secret    = true
  }
  variable {
    name  = "key2"
    value = "value2"
  }
  variable {
    name = "key3"
  }
}

resource "betterado_check_branch_control" "test" {
  project_id           = %q
  display_name         = "%s"
  target_resource_id   = betterado_variable_group.vg.id
  target_resource_type = "variablegroup"
  allowed_branches     = "%s"
}`, projectID, variableGroupName, projectID, checkName, branches)
}
