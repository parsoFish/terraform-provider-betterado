package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccRepositoryPolicyCheckCredentials(t *testing.T) {
	testutils.RunTestsInSequence(t, map[string]map[string]func(t *testing.T){
		"RepositoryPolicies": {
			"basic":  testAccRepoPolicyCheckCredentialsBasic,
			"update": testAccRepoPolicyCheckCredentialsUpdate,
		},
		"ProjectPolicies": {
			"basic":  testAccProjectPolicyCheckCredentialsBasic,
			"update": testAccProjectPolicyCheckCredentialsUpdate,
		},
	})
}

func testAccRepoPolicyCheckCredentialsBasic(t *testing.T) {
	checkCredentialsTfNode := "betterado_repository_policy_check_credentials.test"
	projectName := testutils.GenerateResourceName()
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyCheckCredentialsBasic(projectName, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(checkCredentialsTfNode, "enabled", "true"),
				),
			}, {
				ResourceName:      checkCredentialsTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(checkCredentialsTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRepoPolicyCheckCredentialsUpdate(t *testing.T) {
	checkCredentialsTfNode := "betterado_repository_policy_check_credentials.test"
	projectName := testutils.GenerateResourceName()
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyCheckCredentialsBasic(projectName, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(checkCredentialsTfNode, "enabled", "true"),
				),
			}, {
				Config: hclRepoPolicyCheckCredentialsUpdate(projectName, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(checkCredentialsTfNode, "enabled", "false"),
				),
			}, {
				ResourceName:      checkCredentialsTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(checkCredentialsTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccProjectPolicyCheckCredentialsBasic(t *testing.T) {
	checkCredentialsTfNode := "betterado_repository_policy_check_credentials.test"
	projectName := testutils.GenerateResourceName()
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclProjectPolicyCheckCredentialsBasic(projectName, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(checkCredentialsTfNode, "enabled", "true"),
				),
			}, {
				ResourceName:      checkCredentialsTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(checkCredentialsTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccProjectPolicyCheckCredentialsUpdate(t *testing.T) {
	checkCredentialsTfNode := "betterado_repository_policy_check_credentials.test"
	projectName := testutils.GenerateResourceName()
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testutils.PreCheck(t, nil) },
		Providers: testutils.GetProviders(),
		Steps: []resource.TestStep{
			{
				Config: hclProjectPolicyCheckCredentialsBasic(projectName, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(checkCredentialsTfNode, "enabled", "true"),
				),
			}, {
				Config: hclProjectPolicyCheckCredentialsUpdate(projectName, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(checkCredentialsTfNode, "enabled", "false"),
				),
			}, {
				ResourceName:      checkCredentialsTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(checkCredentialsTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func hclPolicyCheckCredentialsResourceTemplate(projectName string, repoName string) string {
	return fmt.Sprintf(`
resource "betterado_project" "test" {
  name               = "%s"
  description        = "Test Project Description"
  visibility         = "private"
  version_control    = "Git"
  work_item_template = "Agile"
}

resource "betterado_git_repository" "test" {
  project_id = betterado_project.test.id
  name       = "%s"
  initialization {
    init_type = "Clean"
  }
}
`, projectName, repoName)
}

func hclRepoPolicyCheckCredentialsBasic(projectName string, repoName string) string {
	projectAndRepo := hclPolicyCheckCredentialsResourceTemplate(projectName, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_check_credentials" "test" {
  project_id = betterado_project.test.id

  enabled        = true
  blocking       = true
  repository_ids = [betterado_git_repository.test.id]
}`, projectAndRepo)
}

func hclRepoPolicyCheckCredentialsUpdate(projectName string, repoName string) string {
	projectAndRepo := hclPolicyCheckCredentialsResourceTemplate(projectName, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_check_credentials" "test" {
  project_id     = betterado_project.test.id
  enabled        = false
  blocking       = true
  repository_ids = [betterado_git_repository.test.id]
}`, projectAndRepo)
}

func hclProjectPolicyCheckCredentialsBasic(projectName string, repoName string) string {
	projectAndRepo := hclPolicyCheckCredentialsResourceTemplate(projectName, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_check_credentials" "test" {
  project_id = betterado_project.test.id
  enabled    = true
  blocking   = true
  depends_on = [betterado_git_repository.test]
}
`, projectAndRepo)
}

func hclProjectPolicyCheckCredentialsUpdate(projectName string, repoName string) string {
	projectAndRepo := hclPolicyCheckCredentialsResourceTemplate(projectName, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_check_credentials" "test" {
  project_id = betterado_project.test.id
  enabled    = false
  blocking   = true
  depends_on = [betterado_git_repository.test]
}`, projectAndRepo)
}
