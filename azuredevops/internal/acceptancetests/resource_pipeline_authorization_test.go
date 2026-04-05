package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccPipelineAuthorization_allPipeline_queue(t *testing.T) {
	node := "betterado_pipeline_authorization.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
		},
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclAllPipelineAuthQueue(testutils.GenerateResourceName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "project_id"),
					resource.TestCheckResourceAttrSet(node, "resource_id"),
				),
			},
		},
	})
}

func TestAccPipelineAuthorization_pipeline_queue(t *testing.T) {
	node := "betterado_pipeline_authorization.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
		},
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclPipelineAuthQueue(testutils.GenerateResourceName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "project_id"),
					resource.TestCheckResourceAttrSet(node, "resource_id"),
				),
			},
		},
	})
}

func TestAccPipelineAuthorization_multiPipeline_queue(t *testing.T) {
	node := "betterado_pipeline_authorization.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
		},
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclMultiPipelineAuthQueue(testutils.GenerateResourceName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "project_id"),
					resource.TestCheckResourceAttrSet(node, "resource_id"),
				),
			},
		},
	})
}

func TestAccPipelineAuthorization_allPipelineWithPipeline_queue(t *testing.T) {
	node := "betterado_pipeline_authorization.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
		},
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclAllPipelineWithPipoelineAuthQueue(testutils.GenerateResourceName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "project_id"),
					resource.TestCheckResourceAttrSet(node, "resource_id"),
				),
			},
		},
	})
}

func TestAccPipelineAuthorization_allPipeline_environment(t *testing.T) {
	node := "betterado_pipeline_authorization.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
		},
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclAllPipelineAuthEnvironment(testutils.GenerateResourceName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "project_id"),
					resource.TestCheckResourceAttrSet(node, "resource_id"),
				),
			},
		},
	})
}

func TestAccPipelineAuthorization_pipeline_environment(t *testing.T) {
	node := "betterado_pipeline_authorization.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
		},
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclPipelineAuthEnvironment(testutils.GenerateResourceName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "project_id"),
					resource.TestCheckResourceAttrSet(node, "resource_id"),
				),
			},
		},
	})
}

func TestAccPipelineAuthorization_allPipeline_variableGroup(t *testing.T) {
	node := "betterado_pipeline_authorization.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
		},
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclAllPipelineAuthVariableGroup(testutils.GenerateResourceName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "project_id"),
					resource.TestCheckResourceAttrSet(node, "resource_id"),
				),
			},
		},
	})
}

func TestAccPipelineAuthorization_pipeline_variableGroup(t *testing.T) {
	node := "betterado_pipeline_authorization.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
		},
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclPipelineAuthVariableGroup(testutils.GenerateResourceName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "project_id"),
					resource.TestCheckResourceAttrSet(node, "resource_id"),
				),
			},
		},
	})
}

func TestAccPipelineAuthorization_allPipeline_endpoint(t *testing.T) {
	node := "betterado_pipeline_authorization.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
		},
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclAllPipelineAuthEndpoint(testutils.GenerateResourceName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "project_id"),
					resource.TestCheckResourceAttrSet(node, "resource_id"),
				),
			},
		},
	})
}

func TestAccPipelineAuthorization_pipeline_endpoint(t *testing.T) {
	node := "betterado_pipeline_authorization.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
		},
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclPipelineAuthEndpoint(testutils.GenerateResourceName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "project_id"),
					resource.TestCheckResourceAttrSet(node, "resource_id"),
				),
			},
		},
	})
}

func TestAccPipelineAuthorization_allPipeline_repository(t *testing.T) {
	node := "betterado_pipeline_authorization.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
		},
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclAllPipelineAuthRepository(testutils.GenerateResourceName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "project_id"),
					resource.TestCheckResourceAttrSet(node, "resource_id"),
				),
			},
		},
	})
}

func TestAccPipelineAuthorization_pipeline_repository(t *testing.T) {
	node := "betterado_pipeline_authorization.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
		},
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclPipelineAuthRepository(testutils.GenerateResourceName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "project_id"),
					resource.TestCheckResourceAttrSet(node, "resource_id"),
				),
			},
		},
	})
}

func TestAccPipelineAuthorization_pipeline_cross_project_repository(t *testing.T) {
	node := "betterado_pipeline_authorization.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testutils.PreCheck(t, nil)
		},
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclPipelineAuthCrossProjectRepository(testutils.GenerateResourceName()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(node, "project_id"),
					resource.TestCheckResourceAttrSet(node, "pipeline_project_id"),
					resource.TestCheckResourceAttrSet(node, "resource_id"),
				),
			},
			{
				ResourceName:      node,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[node]
					if !ok {
						return "", fmt.Errorf("resource not found in state: %s", node)
					}

					if rs.Primary == nil || rs.Primary.ID == "" {
						return "", fmt.Errorf("no primary instance or ID found for resource: %s", node)
					}

					return rs.Primary.ID, nil
				},
			},
		},
	})
}

func TestAccPipelineAuthorization_queue_import(t *testing.T) {
	projectName := testutils.GenerateResourceName()
	resourceName := "betterado_pipeline_authorization.import_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclPipelineAuthorizationQueueConfig(projectName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "type", "queue"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource not found in state: %s", resourceName)
					}
					projectID := rs.Primary.Attributes["project_id"]
					resourceID := rs.Primary.Attributes["resource_id"]
					typ := rs.Primary.Attributes["type"]

					if projectID == "" || typ == "" || resourceID == "" {
						return "", fmt.Errorf("missing required attributes for import id (project_id=%q type=%q resource_id=%q)", projectID, typ, resourceID)
					}

					return fmt.Sprintf("%s/%s/%s", projectID, typ, resourceID), nil
				},
			},
		},
	})
}

func hclAllPipelineAuthQueue(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = "Managed by Terraform"
}

resource "betterado_agent_pool" "test" {
  name           = "%[1]s"
  auto_provision = false
  auto_update    = false
}

resource "betterado_agent_queue" "test" {
  project_id    = betterado_project.test.id
  agent_pool_id = betterado_agent_pool.test.id
}

resource "betterado_pipeline_authorization" "test" {
  project_id  = betterado_project.test.id
  resource_id = betterado_agent_queue.test.id
  type        = "queue"
}
`, name)
}

func hclPipelineAuthQueue(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = "Managed by Terraform"
}

resource "betterado_agent_pool" "test" {
  name           = "%[1]s"
  auto_provision = false
  auto_update    = false
}

resource "betterado_agent_queue" "test" {
  project_id    = betterado_project.test.id
  agent_pool_id = betterado_agent_pool.test.id
}

data "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"
}

resource "betterado_build_definition" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"

  repository {
    repo_type = "TfsGit"
    repo_id   = data.betterado_git_repository.test.id
    yml_path  = "azure-pipelines.yml"
  }
}

resource "betterado_pipeline_authorization" "test" {
  project_id  = betterado_project.test.id
  resource_id = betterado_agent_queue.test.id
  pipeline_id = betterado_build_definition.test.id
  type        = "queue"
}
`, name)
}

func hclMultiPipelineAuthQueue(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = "Managed by Terraform"
}

resource "betterado_agent_pool" "test" {
  name           = "%[1]s"
  auto_provision = false
  auto_update    = false
}

resource "betterado_agent_queue" "test" {
  project_id    = betterado_project.test.id
  agent_pool_id = betterado_agent_pool.test.id
}

data "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"
}

resource "betterado_build_definition" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"

  repository {
    repo_type = "TfsGit"
    repo_id   = data.betterado_git_repository.test.id
    yml_path  = "azure-pipelines.yml"
  }
}

resource "betterado_build_definition" "test2" {
  project_id = betterado_project.test.id
  name       = "%[1]s2"

  repository {
    repo_type = "TfsGit"
    repo_id   = data.betterado_git_repository.test.id
    yml_path  = "azure-pipelines.yml"
  }
}

resource "betterado_pipeline_authorization" "test" {
  project_id  = betterado_project.test.id
  resource_id = betterado_agent_queue.test.id
  pipeline_id = betterado_build_definition.test.id
  type        = "queue"
}

resource "betterado_pipeline_authorization" "test2" {
  project_id  = betterado_project.test.id
  resource_id = betterado_agent_queue.test.id
  pipeline_id = betterado_build_definition.test2.id
  type        = "queue"
}
`, name)
}

func hclAllPipelineWithPipoelineAuthQueue(name string) string {
	{
		return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = "Managed by Terraform"
}

resource "betterado_agent_pool" "test" {
  name           = "%[1]s"
  auto_provision = false
  auto_update    = false
}

resource "betterado_agent_queue" "test" {
  project_id    = betterado_project.test.id
  agent_pool_id = betterado_agent_pool.test.id
}

data "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"
}

resource "betterado_build_definition" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"

  repository {
    repo_type = "TfsGit"
    repo_id   = data.betterado_git_repository.test.id
    yml_path  = "azure-pipelines.yml"
  }
}

resource "betterado_pipeline_authorization" "test" {
  project_id  = betterado_project.test.id
  resource_id = betterado_agent_queue.test.id
  pipeline_id = betterado_build_definition.test.id
  type        = "queue"
}

resource "betterado_pipeline_authorization" "test_all" {
  project_id  = betterado_project.test.id
  resource_id = betterado_agent_queue.test.id
  type        = "queue"
}

`, name)
	}
}

func hclAllPipelineAuthEnvironment(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = "Managed by Terraform"
}

resource "betterado_environment" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"
}

resource "betterado_pipeline_authorization" "test" {
  project_id  = betterado_project.test.id
  resource_id = betterado_environment.test.id
  type        = "environment"
}
`, name)
}

func hclPipelineAuthEnvironment(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = "Managed by Terraform"
}

resource "betterado_environment" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"
}

data "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"
}

resource "betterado_build_definition" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"

  repository {
    repo_type = "TfsGit"
    repo_id   = data.betterado_git_repository.test.id
    yml_path  = "azure-pipelines.yml"
  }
}

resource "betterado_pipeline_authorization" "test" {
  project_id  = betterado_project.test.id
  resource_id = betterado_environment.test.id
  type        = "environment"
  pipeline_id = betterado_build_definition.test.id
}
`, name)
}

func hclAllPipelineAuthVariableGroup(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = "Managed by Terraform"
}

resource "betterado_variable_group" "test" {
  project_id   = betterado_project.test.id
  name         = "%[1]s"
  allow_access = true

  variable {
    name  = "key1"
    value = "val1"
  }
}

resource "betterado_pipeline_authorization" "test" {
  project_id  = betterado_project.test.id
  resource_id = betterado_variable_group.test.id
  type        = "variablegroup"
}
`, name)
}

func hclPipelineAuthVariableGroup(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = "Managed by Terraform"
}

resource "betterado_variable_group" "test" {
  project_id   = betterado_project.test.id
  name         = "%[1]s"
  allow_access = true

  variable {
    name  = "key1"
    value = "val1"
  }
}

data "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"
}

resource "betterado_build_definition" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"

  repository {
    repo_type = "TfsGit"
    repo_id   = data.betterado_git_repository.test.id
    yml_path  = "azure-pipelines.yml"
  }
}

resource "betterado_pipeline_authorization" "test" {
  project_id  = betterado_project.test.id
  resource_id = betterado_variable_group.test.id
  type        = "variablegroup"
  pipeline_id = betterado_build_definition.test.id
}
`, name)
}

func hclAllPipelineAuthEndpoint(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = "Managed by Terraform"
}

resource "betterado_serviceendpoint_github" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = "%[1]s"

  auth_personal {
    personal_access_token = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  }
}

resource "betterado_pipeline_authorization" "test" {
  project_id  = betterado_project.test.id
  resource_id = betterado_serviceendpoint_github.test.id
  type        = "endpoint"
}
`, name)
}

func hclPipelineAuthEndpoint(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = "Managed by Terraform"
}

resource "betterado_serviceendpoint_github" "test" {
  project_id            = betterado_project.test.id
  service_endpoint_name = "%[1]s"

  auth_personal {
    personal_access_token = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  }
}

data "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"
}

resource "betterado_build_definition" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"

  repository {
    repo_type = "TfsGit"
    repo_id   = data.betterado_git_repository.test.id
    yml_path  = "azure-pipelines.yml"
  }
}

resource "betterado_pipeline_authorization" "test" {
  project_id  = betterado_project.test.id
  resource_id = betterado_serviceendpoint_github.test.id
  type        = "endpoint"
  pipeline_id = betterado_build_definition.test.id
}
`, name)
}

func hclAllPipelineAuthRepository(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = "Managed by Terraform"
}

data "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"
}

resource "betterado_pipeline_authorization" "test" {
  project_id  = betterado_project.test.id
  resource_id = data.betterado_git_repository.test.id
  type        = "repository"
}
`, name)
}

func hclPipelineAuthRepository(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = "Managed by Terraform"
}

data "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"
}

resource "betterado_build_definition" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"

  repository {
    repo_type = "TfsGit"
    repo_id   = data.betterado_git_repository.test.id
    yml_path  = "azure-pipelines.yml"
  }
}

resource "betterado_pipeline_authorization" "test" {
  project_id  = betterado_project.test.id
  resource_id = data.betterado_git_repository.test.id
  pipeline_id = betterado_build_definition.test.id
  type        = "repository"
}
`, name)
}

func hclPipelineAuthCrossProjectRepository(name string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%[1]s"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = "Managed by Terraform"
}

resource "betterado_project" "remote_repo" {
  name               = "%[1]s2"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
  description        = "Managed by Terraform"
}

data "betterado_git_repository" "remote_repo" {
  project_id = betterado_project.remote_repo.id
  name       = betterado_project.remote_repo.name
}

data "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"
}

resource "betterado_build_definition" "test" {
  project_id = betterado_project.test.id
  name       = "%[1]s"

  repository {
    repo_type = "TfsGit"
    repo_id   = data.betterado_git_repository.test.id
    yml_path  = "azure-pipelines.yml"
  }
}

resource "betterado_pipeline_authorization" "test" {
  project_id  = betterado_project.remote_repo.id
  resource_id = data.betterado_git_repository.remote_repo.id
  type        = "repository"

  pipeline_id         = betterado_build_definition.test.id
  pipeline_project_id = betterado_project.test.id
}
`, name)
}

func hclPipelineAuthorizationQueueConfig(projectName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name = %q
}

data "betterado_agent_queue" "default" {
  project_id = betterado_project.test.id
  name       = "Default"
}

resource "betterado_pipeline_authorization" "import_test" {
  project_id  = betterado_project.test.id
  type        = "queue"
  resource_id = data.betterado_agent_queue.default.id
}
`, projectName)
}
