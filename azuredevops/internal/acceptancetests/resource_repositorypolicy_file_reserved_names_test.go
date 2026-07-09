package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccRepositoryPolicyReservedNames(t *testing.T) {
	testutils.RunTestsInSequence(t, map[string]map[string]func(t *testing.T){
		"RepositoryPolicies": {
			"basic":  testAccRepoPolicyReservedNamesBasic,
			"update": testAccRepoPolicyReservedNamesUpdate,
		},
		"ProjectPolicies": {
			"basic":  testAccProjectPolicyReservedNamesBasic,
			"update": testAccProjectPolicyReservedNamesUpdate,
		},
	})
}

func testAccRepoPolicyReservedNamesBasic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	reservedNameTfNode := "betterado_repository_policy_reserved_names.test"
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyReservedNamesBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(reservedNameTfNode, "enabled", "true"),
				),
			}, {
				ResourceName:      reservedNameTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(reservedNameTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRepoPolicyReservedNamesUpdate(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	reservedNameTfNode := "betterado_repository_policy_reserved_names.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyReservedNamesBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(reservedNameTfNode, "enabled", "true"),
				),
			}, {
				Config: hclRepoPolicyReservedNamesUpdate(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(reservedNameTfNode, "enabled", "false"),
				),
			}, {
				ResourceName:      reservedNameTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(reservedNameTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccProjectPolicyReservedNamesBasic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	reservedNameTfNode := "betterado_repository_policy_reserved_names.test"
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclProjectPolicyReservedNamesBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(reservedNameTfNode, "enabled", "true"),
				),
			}, {
				ResourceName:      reservedNameTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(reservedNameTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccProjectPolicyReservedNamesUpdate(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	reservedNameTfNode := "betterado_repository_policy_reserved_names.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclProjectPolicyReservedNamesBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(reservedNameTfNode, "enabled", "true"),
				),
			}, {
				Config: hclProjectPolicyReservedNamesUpdate(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(reservedNameTfNode, "enabled", "false"),
				),
			}, {
				ResourceName:      reservedNameTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(reservedNameTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func hclPolicyReservedNamesResourceTemplate(projectID string, repoName string) string {
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

func hclRepoPolicyReservedNamesBasic(projectID string, repoName string) string {
	repoBlock := hclPolicyReservedNamesResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_reserved_names" "test" {
  project_id     = %[2]q
  enabled        = true
  blocking       = true
  repository_ids = [betterado_git_repository.test.id]
}`, repoBlock, projectID)
}

func hclRepoPolicyReservedNamesUpdate(projectID string, repoName string) string {
	repoBlock := hclPolicyReservedNamesResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_reserved_names" "test" {
  project_id     = %[2]q
  enabled        = false
  blocking       = true
  repository_ids = [betterado_git_repository.test.id]
}`, repoBlock, projectID)
}

func hclProjectPolicyReservedNamesBasic(projectID string, repoName string) string {
	repoBlock := hclPolicyReservedNamesResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_reserved_names" "test" {
  project_id = %[2]q
  enabled    = true
  blocking   = true
  depends_on = [betterado_git_repository.test]
}`, repoBlock, projectID)
}

func hclProjectPolicyReservedNamesUpdate(projectID string, repoName string) string {
	repoBlock := hclPolicyReservedNamesResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_reserved_names" "test" {
  project_id = %[2]q
  enabled    = false
  blocking   = true
  depends_on = [betterado_git_repository.test]
}`, repoBlock, projectID)
}
