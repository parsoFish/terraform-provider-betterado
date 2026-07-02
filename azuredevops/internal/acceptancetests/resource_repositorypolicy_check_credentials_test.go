package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccRepositoryPolicyCheckCredentials(t *testing.T) {
	// The Azure DevOps policy type for check_credentials
	// (type ID e67ae10f-cf9a-40bc-8e66-6b3a8216956e) has been removed from the
	// live ADO service.  The resource is retained in the provider for
	// import/read/destroy of any pre-existing policies but can no longer be
	// created.  Acceptance tests cannot exercise create against this policy type.
	t.Skip("betterado_repository_policy_check_credentials: ADO policy type e67ae10f-cf9a-40bc-8e66-6b3a8216956e has been removed from the live service; cannot create")

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
	projectID := SharedFixtureProjectID(t)
	checkCredentialsTfNode := "betterado_repository_policy_check_credentials.test"
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyCheckCredentialsBasic(projectID, repoName),
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
	projectID := SharedFixtureProjectID(t)
	checkCredentialsTfNode := "betterado_repository_policy_check_credentials.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyCheckCredentialsBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(checkCredentialsTfNode, "enabled", "true"),
				),
			}, {
				Config: hclRepoPolicyCheckCredentialsUpdate(projectID, repoName),
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
	projectID := SharedFixtureProjectID(t)
	checkCredentialsTfNode := "betterado_repository_policy_check_credentials.test"
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclProjectPolicyCheckCredentialsBasic(projectID, repoName),
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
	projectID := SharedFixtureProjectID(t)
	checkCredentialsTfNode := "betterado_repository_policy_check_credentials.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclProjectPolicyCheckCredentialsBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(checkCredentialsTfNode, "enabled", "true"),
				),
			}, {
				Config: hclProjectPolicyCheckCredentialsUpdate(projectID, repoName),
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

func hclPolicyCheckCredentialsResourceTemplate(projectID string, repoName string) string {
	return fmt.Sprintf(`
resource "betterado_git_repository" "test" {
  project_id = %[1]q
  name       = "%[2]s"
  initialization {
    init_type = "Clean"
  }
}
`, projectID, repoName)
}

func hclRepoPolicyCheckCredentialsBasic(projectID string, repoName string) string {
	repoBlock := hclPolicyCheckCredentialsResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_check_credentials" "test" {
  project_id = %[2]q

  enabled        = true
  blocking       = true
  repository_ids = [betterado_git_repository.test.id]
}`, repoBlock, projectID)
}

func hclRepoPolicyCheckCredentialsUpdate(projectID string, repoName string) string {
	repoBlock := hclPolicyCheckCredentialsResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_check_credentials" "test" {
  project_id     = %[2]q
  enabled        = false
  blocking       = true
  repository_ids = [betterado_git_repository.test.id]
}`, repoBlock, projectID)
}

func hclProjectPolicyCheckCredentialsBasic(projectID string, repoName string) string {
	repoBlock := hclPolicyCheckCredentialsResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_check_credentials" "test" {
  project_id = %[2]q
  enabled    = true
  blocking   = true
  depends_on = [betterado_git_repository.test]
}
`, repoBlock, projectID)
}

func hclProjectPolicyCheckCredentialsUpdate(projectID string, repoName string) string {
	repoBlock := hclPolicyCheckCredentialsResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_check_credentials" "test" {
  project_id = %[2]q
  enabled    = false
  blocking   = true
  depends_on = [betterado_git_repository.test]
}`, repoBlock, projectID)
}
