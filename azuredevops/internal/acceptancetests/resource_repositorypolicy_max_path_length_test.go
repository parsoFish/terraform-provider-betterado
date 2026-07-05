package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccRepositoryPolicyPathLength(t *testing.T) {
	testutils.RunTestsInSequence(t, map[string]map[string]func(t *testing.T){
		"RepositoryPolicies": {
			"basic":  testAccRepoPolicyPathLengthBasic,
			"update": testAccRepoPolicyPathLengthUpdate,
		},
		"ProjectPolicies": {
			"basic":  testAccProjectPolicyPathLengthBasic,
			"update": testAccProjectPolicyPathLengthUpdate,
		},
	})
}

func testAccRepoPolicyPathLengthBasic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	pathLengthTfNode := "betterado_repository_policy_max_path_length.test"
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyPathLengthBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pathLengthTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(pathLengthTfNode, "max_path_length", "500"),
				),
			}, {
				ResourceName:      pathLengthTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(pathLengthTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRepoPolicyPathLengthUpdate(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	pathLengthTfNode := "betterado_repository_policy_max_path_length.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyPathLengthBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pathLengthTfNode, "enabled", "true"),
				),
			}, {
				Config: hclRepoPolicyPathLengthUpdate(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pathLengthTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(pathLengthTfNode, "max_path_length", "1000"),
				),
			}, {
				ResourceName:      pathLengthTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(pathLengthTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccProjectPolicyPathLengthBasic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	pathLengthTfNode := "betterado_repository_policy_max_path_length.test"
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclProjectPolicyPathLengthBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pathLengthTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(pathLengthTfNode, "max_path_length", "500"),
				),
			}, {
				ResourceName:      pathLengthTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(pathLengthTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccProjectPolicyPathLengthUpdate(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	pathLengthTfNode := "betterado_repository_policy_max_path_length.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclProjectPolicyPathLengthBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pathLengthTfNode, "enabled", "true"),
				),
			}, {
				Config: hclProjectPolicyPathLengthUpdate(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(pathLengthTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(pathLengthTfNode, "max_path_length", "1000"),
				),
			}, {
				ResourceName:      pathLengthTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(pathLengthTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func hclPolicyPathLengthResourceTemplate(projectID string, repoName string) string {
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

func hclRepoPolicyPathLengthBasic(projectID string, repoName string) string {
	repoBlock := hclPolicyPathLengthResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_max_path_length" "test" {
  project_id      = %[2]q
  enabled         = true
  blocking        = true
  max_path_length = 500
  repository_ids  = [betterado_git_repository.test.id]
}`, repoBlock, projectID)
}

func hclRepoPolicyPathLengthUpdate(projectID string, repoName string) string {
	repoBlock := hclPolicyPathLengthResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_max_path_length" "test" {
  project_id      = %[2]q
  enabled         = true
  blocking        = true
  max_path_length = 1000
  repository_ids  = [betterado_git_repository.test.id]
}`, repoBlock, projectID)
}

func hclProjectPolicyPathLengthBasic(projectID string, repoName string) string {
	repoBlock := hclPolicyPathLengthResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_max_path_length" "test" {
  project_id      = %[2]q
  enabled         = true
  blocking        = true
  max_path_length = 500
  depends_on      = [betterado_git_repository.test]
}`, repoBlock, projectID)
}

func hclProjectPolicyPathLengthUpdate(projectID string, repoName string) string {
	repoBlock := hclPolicyPathLengthResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_max_path_length" "test" {
  project_id      = %[2]q
  enabled         = true
  blocking        = true
  max_path_length = 1000
  depends_on      = [betterado_git_repository.test]
}`, repoBlock, projectID)
}
