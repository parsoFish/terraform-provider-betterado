package acceptancetests

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	azuredevops "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/policy"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/acceptancetests/testutils"
	"github.com/parsoFish/terraform-provider-betterado/azuredevops/internal/client"
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
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: hclRepoPolicyFileSizeBasic(projectID, repoName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fileSizeTfNode, "enabled", "true"),
					resource.TestCheckResourceAttr(fileSizeTfNode, "max_file_size", "1"),
					captureRepoPolicyFileSizeEvidence(fileSizeTfNode),
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

// captureRepoPolicyFileSizeEvidence performs a live GET of the created repository policy
// configuration and writes it to .forge/live-evidence/acceptance-resource-repository_policy_max_file_size.json.
// Best-effort: failure does not fail the test (always returns nil).
func captureRepoPolicyFileSizeEvidence(tfNode string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if err := doCapturePolicyFileSizeEvidence(s, tfNode); err != nil {
			_ = err
		}
		return nil
	}
}

func doCapturePolicyFileSizeEvidence(s *terraform.State, tfNode string) error {
	res, ok := s.RootModule().Resources[tfNode]
	if !ok {
		return nil
	}
	policyID, err := strconv.Atoi(res.Primary.ID)
	if err != nil {
		return err
	}
	projectID := res.Primary.Attributes["project_id"]

	orgURL := os.Getenv("AZDO_ORG_SERVICE_URL")
	pat := os.Getenv("AZDO_PERSONAL_ACCESS_TOKEN")
	clients, err := client.GetAzdoClient(azuredevops.NewAuthProviderPAT(pat), orgURL)
	if err != nil {
		return err
	}

	cfg, err := clients.PolicyClient.GetPolicyConfiguration(clients.Ctx, policy.GetPolicyConfigurationArgs{
		Project:         &projectID,
		ConfigurationId: &policyID,
	})
	if err != nil {
		return err
	}
	if cfg == nil {
		return nil
	}

	url := fmt.Sprintf("%s/%s/_apis/policy/configurations/%d?api-version=7.1",
		orgURL, projectID, policyID)
	return testutils.CaptureLiveEvidence("acceptance-resource-repository_policy_max_file_size", url, cfg)
}

func testAccRepoPolicyFileSizeUpdate(t *testing.T) {
	projectID := SharedFixtureProjectID(t)
	fileSizeTfNode := "betterado_repository_policy_max_file_size.test"
	repoName := testutils.GenerateResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testutils.PreCheck(t, nil) },
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
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
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
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
		ProtoV6ProviderFactories: testutils.GetProviderFactories(),
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
