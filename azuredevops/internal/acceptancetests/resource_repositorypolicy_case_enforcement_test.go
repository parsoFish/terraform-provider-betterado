package acceptancetests

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
)

func TestAccRepositoryPolicyCaseEnforcement(t *testing.T) {
	testutils.RunTestsInSequence(t, map[string]map[string]func(t *testing.T){
		"RepositoryPolicies": {
			"basic":  testAccRepoPolicyEnforceConsistentCaseBasic,
			"update": testAccRepoPolicyEnforceConsistentCaseUpdate,
		},
		"ProjectPolicies": {
			"basic":  testAccProjectPolicyEnforceConsistentCaseBasic,
			"update": testAccProjectPolicyEnforceConsistentCaseUpdate,
		},
	})
}

func testAccRepoPolicyEnforceConsistentCaseBasic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	caseEnforceTfNode := "betterado_repository_policy_case_enforcement.test"
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyEnforceConsistentCaseBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(caseEnforceTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(caseEnforceTfNode, "enforce_consistent_case", "true"),
				),
			}, {
				ResourceName:      caseEnforceTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(caseEnforceTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRepoPolicyEnforceConsistentCaseUpdate(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	caseEnforceTfNode := "betterado_repository_policy_case_enforcement.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyEnforceConsistentCaseBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(caseEnforceTfNode, "enabled", "true"),
				),
			}, {
				Config: hclRepoPolicyEnforceConsistentCaseUpdate(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(caseEnforceTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(caseEnforceTfNode, "enforce_consistent_case", "false"),
				),
			}, {
				ResourceName:      "betterado_repository_policy_case_enforcement.test",
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID("betterado_repository_policy_case_enforcement.test"),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccProjectPolicyEnforceConsistentCaseBasic(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	caseEnforceTfNode := "betterado_repository_policy_case_enforcement.test"
	repoName := testutils.GenerateResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclProjectPolicyEnforceConsistentCaseBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(caseEnforceTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(caseEnforceTfNode, "enforce_consistent_case", "true"),
				),
			}, {
				ResourceName:      caseEnforceTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(caseEnforceTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccProjectPolicyEnforceConsistentCaseUpdate(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	caseEnforceTfNode := "betterado_repository_policy_case_enforcement.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclProjectPolicyEnforceConsistentCaseBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(caseEnforceTfNode, "enabled", "true"),
				),
			}, {
				Config: hclProjectPolicyEnforceConsistentCaseUpdate(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(caseEnforceTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(caseEnforceTfNode, "enforce_consistent_case", "false"),
				),
			}, {
				ResourceName:      caseEnforceTfNode,
				ImportStateIdFunc: testutils.ComputeProjectQualifiedResourceImportID(caseEnforceTfNode),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func hclPolicyEnforceConsistentCaseResourceTemplate(projectID string, repoName string) string {
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

func hclRepoPolicyEnforceConsistentCaseBasic(projectID string, repoName string) string {
	repoBlock := hclPolicyEnforceConsistentCaseResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_case_enforcement" "test" {
  project_id = %[2]q

  enabled                 = true
  blocking                = true
  enforce_consistent_case = true
  repository_ids          = [betterado_git_repository.test.id]
}`, repoBlock, projectID)
}

func hclRepoPolicyEnforceConsistentCaseUpdate(projectID string, repoName string) string {
	repoBlock := hclPolicyEnforceConsistentCaseResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_case_enforcement" "test" {
  project_id = %[2]q

  enabled                 = true
  blocking                = true
  enforce_consistent_case = false
  repository_ids          = [betterado_git_repository.test.id]
}`, repoBlock, projectID)
}

func hclProjectPolicyEnforceConsistentCaseBasic(projectID string, repoName string) string {
	repoBlock := hclPolicyEnforceConsistentCaseResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_case_enforcement" "test" {
  project_id              = %[2]q
  enabled                 = true
  blocking                = true
  enforce_consistent_case = true
  depends_on              = [betterado_git_repository.test]
}`, repoBlock, projectID)
}

func hclProjectPolicyEnforceConsistentCaseUpdate(projectID string, repoName string) string {
	repoBlock := hclPolicyEnforceConsistentCaseResourceTemplate(projectID, repoName)
	return fmt.Sprintf(`
%s

resource "betterado_repository_policy_case_enforcement" "test" {
  project_id = %[2]q

  enabled                 = true
  blocking                = true
  enforce_consistent_case = false
  depends_on              = [betterado_git_repository.test]
}`, repoBlock, projectID)
}
