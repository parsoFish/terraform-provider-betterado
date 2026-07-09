package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccRepositoryPolicyAuthorEmailPatterns(t *testing.T) {
	testutils.RunTestsInSequence(t, map[string]map[string]func(t *testing.T){
		"RepositoryPolicies": {
			"basic":  testAccRepositoryPolicyAuthorEmailPatternsRepoPolicyBasic,
			"update": testAccRepositoryPolicyAuthorEmailPatternsRepoPolicyUpdate,
		},
		"ProjectPolicies": {
			"basic":  testAccRepositoryPolicyAuthorEmailPatternsProjectPolicyBasic,
			"update": testAccRepositoryPolicyAuthorEmailPatternsProjectPolicyUpdate,
		},
	})
}

func testAccRepositoryPolicyAuthorEmailPatternsRepoPolicyBasic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	authorEmailTfNode := "betterado_repository_policy_author_email_pattern.test"
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepositoryPolicyAuthorEmailPatternsResourceRepoPolicyBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(authorEmailTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(authorEmailTfNode, "author_email_patterns.0", "test1@test.com"),
				),
			}, {
				ResourceName:      authorEmailTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(authorEmailTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRepositoryPolicyAuthorEmailPatternsRepoPolicyUpdate(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	authorEmailTfNode := "betterado_repository_policy_author_email_pattern.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepositoryPolicyAuthorEmailPatternsResourceRepoPolicyBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(authorEmailTfNode, "enabled", "true"),
				),
			}, {
				Config: hclRepositoryPolicyAuthorEmailPatternsResourceRepoPolicyUpdate(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(authorEmailTfNode, "author_email_patterns.0", "test2@test.com"),
					resource.TestCheckResourceAttr(authorEmailTfNode, "enabled", "true"),
				),
			}, {
				ResourceName:      authorEmailTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(authorEmailTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRepositoryPolicyAuthorEmailPatternsProjectPolicyBasic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	authorEmailTfNode := "betterado_repository_policy_author_email_pattern.test"
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepositoryPolicyAuthorEmailPatternsResourceProjectPolicyBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(authorEmailTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(authorEmailTfNode, "author_email_patterns.#", "1"),
				),
			}, {
				ResourceName:      authorEmailTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(authorEmailTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRepositoryPolicyAuthorEmailPatternsProjectPolicyUpdate(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	authorEmailTfNode := "betterado_repository_policy_author_email_pattern.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepositoryPolicyAuthorEmailPatternsResourceProjectPolicyBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(authorEmailTfNode, "enabled", "true"),
				),
			}, {
				Config: hclRepositoryPolicyAuthorEmailPatternsResourceProjectPolicyUpdate(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(authorEmailTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(authorEmailTfNode, "author_email_patterns.#", "2"),
				),
			}, {
				ResourceName:      authorEmailTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(authorEmailTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func hclRepositoryPolicyAuthorEmailPatternsResourceTemplate(projectID string, repoName string) string {
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

func hclRepositoryPolicyAuthorEmailPatternsResourceRepoPolicyBasic(projectID string, repoName string) string {
	repoBlock := hclRepositoryPolicyAuthorEmailPatternsResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_author_email_pattern" "test" {
  project_id = %[2]q

  enabled  = true
  blocking = true

  author_email_patterns = ["test1@test.com"]
  repository_ids        = [betterado_git_repository.test.id]
}`, repoBlock, projectID)
}

func hclRepositoryPolicyAuthorEmailPatternsResourceRepoPolicyUpdate(projectID string, repoName string) string {
	repoBlock := hclRepositoryPolicyAuthorEmailPatternsResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_author_email_pattern" "test" {
  project_id = %[2]q

  enabled  = true
  blocking = true

  author_email_patterns = ["test2@test.com"]
  repository_ids        = [betterado_git_repository.test.id]
}`, repoBlock, projectID)
}

func hclRepositoryPolicyAuthorEmailPatternsResourceProjectPolicyBasic(projectID string, repoName string) string {
	repoBlock := hclRepositoryPolicyAuthorEmailPatternsResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_author_email_pattern" "test" {
  project_id = %[2]q

  enabled               = true
  blocking              = true
  author_email_patterns = ["test1@test.com"]
  depends_on            = [betterado_git_repository.test]
}`, repoBlock, projectID)
}

func hclRepositoryPolicyAuthorEmailPatternsResourceProjectPolicyUpdate(projectID string, repoName string) string {
	repoBlock := hclRepositoryPolicyAuthorEmailPatternsResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_author_email_pattern" "test" {
  project_id = %[2]q

  enabled  = true
  blocking = true

  author_email_patterns = ["test1@test.com", "test2@test.com"]
  depends_on            = [betterado_git_repository.test]
}`, repoBlock, projectID)
}
