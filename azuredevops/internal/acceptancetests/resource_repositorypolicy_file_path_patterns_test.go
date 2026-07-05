package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccRepositoryPolicyFilePathPatterns(t *testing.T) {
	testutils.RunTestsInSequence(t, map[string]map[string]func(t *testing.T){
		"RepositoryPolicies": {
			"basic":  testAccRepositoryPolicyFilePathPatternsRepoPolicyBasic,
			"update": testAccRepositoryPolicyFilePathPatternsRepoPolicyUpdate,
		},
		"ProjectPolicies": {
			"basic":  testAccRepositoryPolicyFilePathPatternsProjectPolicyBasic,
			"update": testAccRepositoryPolicyFilePathPatternsProjectPolicyUpdate,
		},
	})
}

func testAccRepositoryPolicyFilePathPatternsRepoPolicyBasic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	filePathTfNode := "betterado_repository_policy_file_path_pattern.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyFilePathPatternsResourceRepoPolicyBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(filePathTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(filePathTfNode, "filepath_patterns.#", "1"),
				),
			}, {
				ResourceName:      filePathTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(filePathTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRepositoryPolicyFilePathPatternsRepoPolicyUpdate(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	filePathTfNode := "betterado_repository_policy_file_path_pattern.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyFilePathPatternsResourceRepoPolicyBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(filePathTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(filePathTfNode, "filepath_patterns.#", "1"),
				),
			}, {
				Config: hclRepoPolicyFilePathPatternsResourceRepoPolicyUpdate(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(filePathTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(filePathTfNode, "filepath_patterns.#", "2"),
				),
			}, {
				ResourceName:      filePathTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(filePathTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRepositoryPolicyFilePathPatternsProjectPolicyBasic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	filePathTfNode := "betterado_repository_policy_file_path_pattern.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyFilePathPatternsResourceProjectPolicyBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(filePathTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(filePathTfNode, "filepath_patterns.#", "1"),
				),
			}, {
				ResourceName:      filePathTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(filePathTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRepositoryPolicyFilePathPatternsProjectPolicyUpdate(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	filePathTfNode := "betterado_repository_policy_file_path_pattern.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyFilePathPatternsResourceProjectPolicyBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(filePathTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(filePathTfNode, "filepath_patterns.#", "1"),
				),
			}, {
				Config: hclRepoPolicyFilePathPatternsResourceProjectPolicyUpdate(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(filePathTfNode, "filepath_patterns.#", "2"),
					resource.TestCheckResourceAttr(filePathTfNode, "enabled", "true"),
				),
			}, {
				ResourceName:      filePathTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(filePathTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func hclRepoPolicyFilePathPatternsResourceTemplate(projectID string, repoName string) string {
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

func hclRepoPolicyFilePathPatternsResourceRepoPolicyBasic(projectID string, repoName string) string {
	repoBlock := hclRepoPolicyFilePathPatternsResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_file_path_pattern" "test" {
  project_id        = %[2]q
  enabled           = true
  blocking          = true
  filepath_patterns = ["*.go"]
  repository_ids    = [betterado_git_repository.test.id]
}`, repoBlock, projectID)
}

func hclRepoPolicyFilePathPatternsResourceRepoPolicyUpdate(projectID string, repoName string) string {
	repoBlock := hclRepoPolicyFilePathPatternsResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_file_path_pattern" "test" {
  project_id        = %[2]q
  enabled           = true
  blocking          = true
  filepath_patterns = ["*.go", "/home/test/*.ts"]
  repository_ids    = [betterado_git_repository.test.id]
}`, repoBlock, projectID)
}

func hclRepoPolicyFilePathPatternsResourceProjectPolicyBasic(projectID string, repoName string) string {
	repoBlock := hclRepoPolicyFilePathPatternsResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_file_path_pattern" "test" {
  project_id        = %[2]q
  enabled           = true
  blocking          = true
  filepath_patterns = ["*.go"]
  depends_on        = [betterado_git_repository.test]
}`, repoBlock, projectID)
}

func hclRepoPolicyFilePathPatternsResourceProjectPolicyUpdate(projectID string, repoName string) string {
	repoBlock := hclRepoPolicyFilePathPatternsResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_file_path_pattern" "test" {
  project_id        = %[2]q
  enabled           = true
  blocking          = true
  filepath_patterns = ["*.go", "/home/test/*.ts"]
  depends_on        = [betterado_git_repository.test]
}`, repoBlock, projectID)
}
