package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccRepositoryPolicyFileSize(t *testing.T) {
	testutils.RunTestsInSequence(t, map[string]map[string]func(t *testing.T){
		"RepositoryPolicies": {
			"basic":  testAccRepoPolicyFileSizeBasic,
			"update": testAccRepoPolicyFileSizeUpdate,
		},
		"ProjectPolicies": {
			"basic":  testAccProjectPolicyFileSizeBasic,
			"update": testAccProjectPolicyFileSizeUpdate,
		},
	})
}

func testAccRepoPolicyFileSizeBasic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	fileSizeTfNode := "betterado_repository_policy_max_file_size.test"
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyFileSizeBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fileSizeTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(fileSizeTfNode, "max_file_size", "1"),
				),
			}, {
				ResourceName:      fileSizeTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(fileSizeTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRepoPolicyFileSizeUpdate(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	fileSizeTfNode := "betterado_repository_policy_max_file_size.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyFileSizeBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fileSizeTfNode, "enabled", "true"),
				),
			}, {
				Config: hclRepoPolicyFileSizeUpdate(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fileSizeTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(fileSizeTfNode, "max_file_size", "2"),
				),
			}, {
				ResourceName:      fileSizeTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(fileSizeTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccProjectPolicyFileSizeBasic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	fileSizeTfNode := "betterado_repository_policy_max_file_size.test"
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclProjectPolicyFileSizeBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fileSizeTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(fileSizeTfNode, "max_file_size", "1"),
				),
			}, {
				ResourceName:      fileSizeTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(fileSizeTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccProjectPolicyFileSizeUpdate(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	fileSizeTfNode := "betterado_repository_policy_max_file_size.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetMuxedProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclProjectPolicyFileSizeBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fileSizeTfNode, "enabled", "true"),
				),
			}, {
				Config: hclProjectPolicyFileSizeUpdate(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fileSizeTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(fileSizeTfNode, "max_file_size", "2"),
				),
			}, {
				ResourceName:      fileSizeTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(fileSizeTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func hclPolicyFileSizeResourceTemplate(projectID string, repoName string) string {
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

func hclRepoPolicyFileSizeBasic(projectID string, repoName string) string {
	repoBlock := hclPolicyFileSizeResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_max_file_size" "test" {
  project_id     = %[2]q
  enabled        = true
  blocking       = true
  max_file_size  = 1
  repository_ids = [betterado_git_repository.test.id]
}`, repoBlock, projectID)
}

func hclRepoPolicyFileSizeUpdate(projectID string, repoName string) string {
	repoBlock := hclPolicyFileSizeResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_max_file_size" "test" {
  project_id     = %[2]q
  enabled        = true
  blocking       = true
  max_file_size  = 2
  repository_ids = [betterado_git_repository.test.id]
}`, repoBlock, projectID)
}

func hclProjectPolicyFileSizeBasic(projectID string, repoName string) string {
	repoBlock := hclPolicyFileSizeResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_max_file_size" "test" {
  project_id    = %[2]q
  enabled       = true
  blocking      = true
  max_file_size = 1
  depends_on    = [betterado_git_repository.test]
}`, repoBlock, projectID)
}

func hclProjectPolicyFileSizeUpdate(projectID string, repoName string) string {
	repoBlock := hclPolicyFileSizeResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_max_file_size" "test" {
  project_id    = %[2]q
  enabled       = true
  blocking      = true
  max_file_size = 2
  depends_on    = [betterado_git_repository.test]
}`, repoBlock, projectID)
}
